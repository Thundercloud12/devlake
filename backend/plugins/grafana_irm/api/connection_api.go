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

package api

import (
	"github.com/apache/devlake/core/errors"
	"github.com/apache/devlake/core/plugin"
)

// TestConnection and TestExistingConnection are intentionally not
// implemented yet: verifying a Grafana IRM connection means calling a real
// IncidentsService RPC method (e.g. QueryIncidents) over POST, and which
// call/request-body to use is a logic decision deferred alongside the
// collector (see grafana_irm_plan.md). The route is wired up so config-ui's
// "test connection" button has somewhere to call once that's decided.

// TestConnection test grafana_irm connection
// @Summary test grafana_irm connection
// @Description Test Grafana IRM Connection
// @Tags plugins/grafana_irm
// @Param body body models.GrafanaIrmConn true "json body"
// @Success 200  {object} shared.ApiBody "Success"
// @Failure 400  {string} errcode.Error "Bad Request"
// @Failure 500  {string} errcode.Error "Internal Error"
// @Router /plugins/grafana_irm/test [POST]
func TestConnection(input *plugin.ApiResourceInput) (*plugin.ApiResourceOutput, errors.Error) {
	return nil, errors.Default.New("grafana_irm TestConnection is not implemented yet")
}

// TestExistingConnection test grafana_irm connection
// @Summary test grafana_irm connection
// @Description Test Grafana IRM Connection
// @Tags plugins/grafana_irm
// @Param connectionId path int true "connection ID"
// @Success 200  {object} shared.ApiBody "Success"
// @Failure 400  {string} errcode.Error "Bad Request"
// @Failure 500  {string} errcode.Error "Internal Error"
// @Router /plugins/grafana_irm/connections/{connectionId}/test [POST]
func TestExistingConnection(input *plugin.ApiResourceInput) (*plugin.ApiResourceOutput, errors.Error) {
	return nil, errors.Default.New("grafana_irm TestExistingConnection is not implemented yet")
}

// @Summary create grafana_irm connection
// @Description Create Grafana IRM connection
// @Tags plugins/grafana_irm
// @Param body body models.GrafanaIrmConnection true "json body"
// @Success 200  {object} models.GrafanaIrmConnection
// @Failure 400  {string} errcode.Error "Bad Request"
// @Failure 500  {string} errcode.Error "Internal Error"
// @Router /plugins/grafana_irm/connections [POST]
func PostConnections(input *plugin.ApiResourceInput) (*plugin.ApiResourceOutput, errors.Error) {
	return dsHelper.ConnApi.Post(input)
}

// @Summary patch grafana_irm connection
// @Description Patch Grafana IRM connection
// @Tags plugins/grafana_irm
// @Param body body models.GrafanaIrmConnection true "json body"
// @Success 200  {object} models.GrafanaIrmConnection
// @Failure 400  {string} errcode.Error "Bad Request"
// @Failure 500  {string} errcode.Error "Internal Error"
// @Router /plugins/grafana_irm/connections/{connectionId} [PATCH]
func PatchConnection(input *plugin.ApiResourceInput) (*plugin.ApiResourceOutput, errors.Error) {
	return dsHelper.ConnApi.Patch(input)
}

// @Summary delete grafana_irm connection
// @Description Delete Grafana IRM connection
// @Tags plugins/grafana_irm
// @Success 200  {object} models.GrafanaIrmConnection
// @Failure 400  {string} errcode.Error "Bad Request"
// @Failure 409  {object} services.BlueprintProjectPairs "References exist to this connection"
// @Failure 500  {string} errcode.Error "Internal Error"
// @Router /plugins/grafana_irm/connections/{connectionId} [DELETE]
func DeleteConnection(input *plugin.ApiResourceInput) (*plugin.ApiResourceOutput, errors.Error) {
	return dsHelper.ConnApi.Delete(input)
}

// @Summary list grafana_irm connections
// @Description List Grafana IRM connections
// @Tags plugins/grafana_irm
// @Success 200  {object} models.GrafanaIrmConnection
// @Failure 400  {string} errcode.Error "Bad Request"
// @Failure 500  {string} errcode.Error "Internal Error"
// @Router /plugins/grafana_irm/connections [GET]
func ListConnections(input *plugin.ApiResourceInput) (*plugin.ApiResourceOutput, errors.Error) {
	return dsHelper.ConnApi.GetAll(input)
}

// @Summary get grafana_irm connection
// @Description Get Grafana IRM connection
// @Tags plugins/grafana_irm
// @Success 200  {object} models.GrafanaIrmConnection
// @Failure 400  {string} errcode.Error "Bad Request"
// @Failure 500  {string} errcode.Error "Internal Error"
// @Router /plugins/grafana_irm/connections/{connectionId} [GET]
func GetConnection(input *plugin.ApiResourceInput) (*plugin.ApiResourceOutput, errors.Error) {
	return dsHelper.ConnApi.GetDetail(input)
}
