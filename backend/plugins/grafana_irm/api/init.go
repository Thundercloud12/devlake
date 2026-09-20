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
	"github.com/apache/devlake/core/context"
	"github.com/apache/devlake/core/plugin"
	"github.com/apache/devlake/helpers/pluginhelper/api"
	"github.com/apache/devlake/plugins/grafana_irm/models"
	"github.com/go-playground/validator/v10"
)

var vld *validator.Validate
var basicRes context.BasicRes

var dsHelper *api.DsHelper[models.GrafanaIrmConnection, models.GrafanaIrmScope, models.GrafanaIrmScopeConfig]

// No remote-scope helpers (raProxy/raScopeList/raScopeSearch) are wired up
// here: the Grafana Incident API has no remote-listable service/team
// resource to browse, so scopes are created by hand in config-ui rather
// than picked from a remote list (see grafana_irm_plan.md §4).

func Init(br context.BasicRes, p plugin.PluginMeta) {
	vld = validator.New()
	basicRes = br
	dsHelper = api.NewDataSourceHelper[
		models.GrafanaIrmConnection, models.GrafanaIrmScope, models.GrafanaIrmScopeConfig,
	](
		br,
		p.Name(),
		[]string{"name"},
		func(c models.GrafanaIrmConnection) models.GrafanaIrmConnection {
			return c.Sanitize()
		},
		nil,
		nil,
	)
}
