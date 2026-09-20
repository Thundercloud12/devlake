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

package models

import (
	"github.com/apache/devlake/core/models/common"
)

// GrafanaIrmScopeConfig defines what subset of a connection's incidents a
// scope covers. Grafana IRM has no listable service/team resource to scope
// by (see grafana_irm_plan.md §4), but incident labels ARE filterable
// server-side via the query DSL's `field:<key>:<value>` term (verified live,
// §4.1) — so a scope is defined by a single label key/value pair.
//
// Both fields empty means "no filter": the scope covers every real incident
// on the connection. Scopes are deliberately self-contained — a no-filter
// scope means literally "everything", NOT "everything the other scopes
// didn't claim", so two scopes whose definitions overlap will both include
// the same incident. That overlap is explicit and predictable; making it
// exclusive would require every scope to know its siblings' filters.
type GrafanaIrmScopeConfig struct {
	common.ScopeConfig `mapstructure:",squash" json:",inline" gorm:"embedded"`
	LabelKey           string `mapstructure:"labelKey,omitempty" json:"labelKey"`
	LabelValue         string `mapstructure:"labelValue,omitempty" json:"labelValue"`
}

// LabelFilterTerm renders this scope's filter as a query-DSL term, or "" when
// the scope has no filter. The `field:<key>:<value>` form is the one verified
// live to work — a bare `<key>:<value>` term fails to parse (§4.1).
func (sc GrafanaIrmScopeConfig) LabelFilterTerm() string {
	if sc.LabelKey == "" || sc.LabelValue == "" {
		return ""
	}
	return "field:" + sc.LabelKey + ":" + sc.LabelValue
}

// Matches reports whether an incident carrying the given labels belongs to
// this scope. Membership is recomputed from the incident's CURRENT labels on
// every converter run rather than stored on the incident row: an incident can
// legitimately match several scopes at once (unlike incidentio, where an
// incident has exactly one type), so there is no single "which scope" value
// to store (§4.1).
func (sc GrafanaIrmScopeConfig) Matches(labels []IncidentLabel) bool {
	if sc.LabelFilterTerm() == "" {
		return true
	}
	for _, label := range labels {
		if label.Key == sc.LabelKey && label.Label == sc.LabelValue {
			return true
		}
	}
	return false
}

func (GrafanaIrmScopeConfig) TableName() string {
	return "_tool_grafana_irm_scope_configs"
}
