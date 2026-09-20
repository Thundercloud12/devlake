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
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/apache/devlake/helpers/pluginhelper/api"
)

// RefreshOpenIncidents makes real API calls and can rewrite raw data, so
// DevLake's own "Re-transform Data" filter (server/services/blueprint.go's
// removeCollectorTasks, which strips any subtask whose name contains
// "collect") must recognize it as a collector, the same way it already does
// for collectIncidents. If this ever regresses, a "Re-transform Data" run
// silently deletes every already-resolved incident from the tool/domain
// tables — confirmed live once, see grafana_irm_plan.md §12/§13.
func TestRefreshOpenIncidentsMetaNameIsRecognizedAsACollector(t *testing.T) {
	assert.Contains(t, strings.ToLower(RefreshOpenIncidentsMeta.Name), "collect")
}

// The expected strings below are the exact query shapes verified live against
// a real stack (see grafana_irm_plan.md §10.1): `isdrill:false`, and the
// `or(...)` date-range group added for an incremental sync.
func TestBuildIncidentsQueryString(t *testing.T) {
	since := time.Date(2026, 9, 19, 11, 2, 0, 0, time.UTC)
	until := time.Date(2026, 9, 19, 23, 59, 59, 0, time.UTC)

	cases := []struct {
		name     string
		since    *time.Time
		until    *time.Time
		expected string
	}{
		{
			name:     "full sync",
			expected: "isdrill:false",
		},
		{
			name:     "incremental",
			since:    &since,
			until:    &until,
			expected: "isdrill:false or(declared:2026-09-19T11:02:00Z,2026-09-19T23:59:59Z resolved:2026-09-19T11:02:00Z,2026-09-19T23:59:59Z)",
		},
		{
			name:     "a since with no until falls back to full sync",
			since:    &since,
			expected: "isdrill:false",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, buildIncidentsQueryString(tc.since, tc.until))
		})
	}
}

// Regression test for a real panic hit on a live pipeline run:
// NewDalCursorIterator (see RefreshOpenIncidents) hands back *simplifiedIncident,
// not simplifiedIncident — reflect.New always yields a pointer — so asserting
// the value type here panicked with "interface conversion: interface {} is
// *tasks.simplifiedIncident, not tasks.simplifiedIncident" the first time this
// path actually ran against a connection with an unresolved incident already
// synced. Neither the unit tests nor the e2e fixtures exercised this iterator
// before that.
func TestRefreshOpenIncidentsRequestBody(t *testing.T) {
	reqData := &api.RequestData{Input: &simplifiedIncident{Id: "42"}}
	body := refreshOpenIncidentsRequestBody(reqData)
	assert.Equal(t, map[string]interface{}{"incidentID": "42"}, body)
}
