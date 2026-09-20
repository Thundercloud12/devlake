/*
Licensed to the Apache Software Foundation (ASF) under one or more
contributor license agreements.  See the NOTICE file distributed with
this work for additional information regarding copyright ownership.
The ASF licenses this file to You under the Apache License, Version 2.0
(the "License"); you may not use this file except in compliance with
the License.  You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package tasks

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/apache/devlake/core/dal"
	"github.com/apache/devlake/core/errors"
	"github.com/apache/devlake/core/plugin"
	"github.com/apache/devlake/helpers/pluginhelper/api"
	"github.com/apache/devlake/plugins/grafana_irm/models"
)

const RAW_INCIDENTS_TABLE = "grafana_irm_incidents"

const incidentsPageSize = 100

var _ plugin.SubTaskEntryPoint = CollectIncidents
var _ plugin.SubTaskEntryPoint = RefreshOpenIncidents

// queryIncidentsResponse is the envelope IncidentsService.QueryIncidents
// returns; verified live, see grafana_irm_plan.md §3.1/§10.1.
type queryIncidentsResponse struct {
	Incidents []json.RawMessage `json:"incidents"`
	Cursor    struct {
		NextValue string `json:"nextValue"`
		HasMore   bool   `json:"hasMore"`
	} `json:"cursor"`
}

// getIncidentResponse is IncidentsService.GetIncident's envelope.
type getIncidentResponse struct {
	Incident json.RawMessage `json:"incident"`
}

// simplifiedIncident is the minimal shape read back from our own tool table
// to drive RefreshOpenIncidents' input iterator below.
type simplifiedIncident struct {
	Id string
}

var CollectIncidentsMeta = plugin.SubTaskMeta{
	Name:             "collectIncidents",
	EntryPoint:       CollectIncidents,
	EnabledByDefault: true,
	Description:      "Collect Grafana IRM incidents (new and recently resolved)",
	DomainTypes:      []string{plugin.DOMAIN_TYPE_TICKET},
	ProductTables:    []string{RAW_INCIDENTS_TABLE},
}

// CollectIncidents runs the main incremental pass designed in
// grafana_irm_plan.md §10.2: `IncidentsQuery.DateFrom`/`DateTo` were verified
// live to have no filtering effect at all, so incremental filtering goes
// through `queryString`'s `declared:`/`resolved:` date-range syntax instead
// — verified live to combine correctly with a bare `isdrill:false` term via
// `or(...)` (§10.1). Real incidents only; drills are never synced (§11).
func CollectIncidents(taskCtx plugin.SubTaskContext) errors.Error {
	data := taskCtx.GetData().(*GrafanaIrmTaskData)

	collector, err := api.NewStatefulApiCollector(api.RawDataSubTaskArgs{
		Ctx:     taskCtx,
		Options: data.Options,
		Table:   RAW_INCIDENTS_TABLE,
	})
	if err != nil {
		return err
	}

	queryString := buildIncidentsQueryString(collector.GetSince(), collector.GetUntil())

	var lastCursor string
	var lastHasMore bool

	err = collector.InitCollector(api.ApiCollectorArgs{
		ApiClient:   data.Client,
		Method:      http.MethodPost,
		UrlTemplate: "api/plugins/grafana-irm-app/resources/api/v1/IncidentsService.QueryIncidents",
		PageSize:    incidentsPageSize,
		RequestBody: func(reqData *api.RequestData) map[string]interface{} {
			query := map[string]interface{}{
				"limit":          reqData.Pager.Size,
				"orderDirection": "ASC",
				"queryString":    queryString,
			}
			body := map[string]interface{}{"query": query}
			if cursor, ok := reqData.CustomData.(string); ok && cursor != "" {
				body["cursor"] = map[string]interface{}{"nextValue": cursor}
			}
			return body
		},
		GetNextPageCustomData: func(prevReqData *api.RequestData, prevPageResponse *http.Response) (interface{}, errors.Error) {
			// lastCursor/lastHasMore are set in ResponseParser below and read
			// from that closure rather than prevPageResponse.Body here: the
			// body is a single-read stream and is already drained by the time
			// this hook fires (same constraint incidentio's collector notes).
			if !lastHasMore || lastCursor == "" {
				return nil, api.ErrFinishCollect
			}
			return lastCursor, nil
		},
		ResponseParser: func(res *http.Response) ([]json.RawMessage, errors.Error) {
			envelope := &queryIncidentsResponse{}
			if err := api.UnmarshalResponse(res, envelope); err != nil {
				return nil, err
			}
			lastCursor = envelope.Cursor.NextValue
			lastHasMore = envelope.Cursor.HasMore
			return envelope.Incidents, nil
		},
	})
	if err != nil {
		return err
	}
	return collector.Execute()
}

// buildIncidentsQueryString implements the incremental-sync recipe verified
// live against the real API (grafana_irm_plan.md §10.1/§10.2). since/until
// come from the framework's own collector state tracking; a nil since means
// a full sync, so no date restriction is applied. A connection has exactly
// one scope covering its whole incident stream (§4.2), so there is no
// per-scope filter term to add here.
func buildIncidentsQueryString(since, until *time.Time) string {
	terms := []string{"isdrill:false"}
	if since != nil && until != nil {
		from := since.UTC().Format(time.RFC3339)
		to := until.UTC().Format(time.RFC3339)
		terms = append(terms, fmt.Sprintf("or(declared:%s,%s resolved:%s,%s)", from, to, from, to))
	}
	return strings.Join(terms, " ")
}

var RefreshOpenIncidentsMeta = plugin.SubTaskMeta{
	Name:             "refreshOpenIncidents",
	EntryPoint:       RefreshOpenIncidents,
	EnabledByDefault: true,
	Description:      "Re-fetch incidents this connection last saw as unresolved, to catch changes no date-range query can see",
	DomainTypes:      []string{plugin.DOMAIN_TYPE_TICKET},
	ProductTables:    []string{RAW_INCIDENTS_TABLE},
}

// RefreshOpenIncidents is the second half of the design in
// grafana_irm_plan.md §10.2: `queryString` only supports `declared:`/
// `started:`/`resolved:`/`ended:` date ranges (verified live; `modified:`/
// `updated:` are not valid properties there), so a status change, label
// edit, or role assignment on an incident that's already synced and still
// open would never be picked up by CollectIncidents alone. This mirrors
// pagerduty's CollectUnfinishedDetails pattern: re-fetch (via GetIncident)
// every incident our own tool table still has as non-resolved from a prior
// sync.
func RefreshOpenIncidents(taskCtx plugin.SubTaskContext) errors.Error {
	data := taskCtx.GetData().(*GrafanaIrmTaskData)
	db := taskCtx.GetDal()

	// Deliberately connection-wide, NOT filtered to this scope's label: the
	// tool tables are shared across a connection's scopes, and this pass is
	// what keeps every open incident's labels current. That matters because
	// scope membership is recomputed from those labels at convert time — if
	// an incident is relabelled out of this scope, neither scope's list pass
	// would re-fetch it (relabelling changes no declared/resolved date), so
	// this refresh is the only thing that notices.
	cursor, err := db.Cursor(
		dal.Select("id"),
		dal.From(&models.Incident{}),
		dal.Where("connection_id = ? AND status != ?", data.Options.ConnectionId, "resolved"),
	)
	if err != nil {
		return err
	}
	defer cursor.Close()

	iterator, err := api.NewDalCursorIterator(db, cursor, reflect.TypeOf(simplifiedIncident{}))
	if err != nil {
		return err
	}

	collector, err := api.NewApiCollector(api.ApiCollectorArgs{
		RawDataSubTaskArgs: api.RawDataSubTaskArgs{
			Ctx:     taskCtx,
			Options: data.Options,
			Table:   RAW_INCIDENTS_TABLE,
		},
		// This pass only ever adds to what CollectIncidents just wrote in the
		// same run; it must never wipe the raw table, regardless of whether
		// the overall pipeline run is a full or incremental sync.
		Incremental: true,
		ApiClient:   data.Client,
		Input:       iterator,
		Method:      http.MethodPost,
		UrlTemplate: "api/plugins/grafana-irm-app/resources/api/v1/IncidentsService.GetIncident",
		RequestBody: func(reqData *api.RequestData) map[string]interface{} {
			input := reqData.Input.(simplifiedIncident)
			return map[string]interface{}{"incidentID": input.Id}
		},
		ResponseParser: func(res *http.Response) ([]json.RawMessage, errors.Error) {
			envelope := &getIncidentResponse{}
			if err := api.UnmarshalResponse(res, envelope); err != nil {
				return nil, err
			}
			return []json.RawMessage{envelope.Incident}, nil
		},
	})
	if err != nil {
		return err
	}
	return collector.Execute()
}
