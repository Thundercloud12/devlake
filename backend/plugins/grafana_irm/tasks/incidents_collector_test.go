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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/apache/devlake/plugins/grafana_irm/models"
)

// The expected strings below are the exact query shapes verified live against
// a real stack (see grafana_irm_plan.md §10.1/§4.1): `isdrill:false` and
// `field:<key>:<value>` AND together, and the `or(...)` group is one operand
// of that AND.
func TestBuildIncidentsQueryString(t *testing.T) {
	since := time.Date(2026, 9, 19, 11, 2, 0, 0, time.UTC)
	until := time.Date(2026, 9, 19, 23, 59, 59, 0, time.UTC)

	cases := []struct {
		name       string
		filterTerm string
		since      *time.Time
		until      *time.Time
		expected   string
	}{
		{
			name:     "full sync, no scope filter",
			expected: "isdrill:false",
		},
		{
			name:       "full sync, scoped by label",
			filterTerm: "field:team_name:payments",
			expected:   "isdrill:false field:team_name:payments",
		},
		{
			name:     "incremental, no scope filter",
			since:    &since,
			until:    &until,
			expected: "isdrill:false or(declared:2026-09-19T11:02:00Z,2026-09-19T23:59:59Z resolved:2026-09-19T11:02:00Z,2026-09-19T23:59:59Z)",
		},
		{
			name:       "incremental, scoped by label",
			filterTerm: "field:team_name:payments",
			since:      &since,
			until:      &until,
			expected:   "isdrill:false field:team_name:payments or(declared:2026-09-19T11:02:00Z,2026-09-19T23:59:59Z resolved:2026-09-19T11:02:00Z,2026-09-19T23:59:59Z)",
		},
		{
			name:     "a since with no until falls back to full sync",
			since:    &since,
			expected: "isdrill:false",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, buildIncidentsQueryString(tc.filterTerm, tc.since, tc.until))
		})
	}
}

func TestScopeConfigMatches(t *testing.T) {
	payments := []models.IncidentLabel{
		{Key: "service_name", Label: "checkout-api"},
		{Key: "team_name", Label: "payments"},
	}
	unlabelled := []models.IncidentLabel{}

	scoped := &GrafanaIrmOptions{
		ScopeConfig: &models.GrafanaIrmScopeConfig{LabelKey: "team_name", LabelValue: "payments"},
	}
	assert.True(t, scoped.MatchesScope(payments))
	assert.False(t, scoped.MatchesScope(unlabelled))
	assert.False(t, scoped.MatchesScope([]models.IncidentLabel{{Key: "team_name", Label: "frontend"}}))
	// a matching key with the wrong value must not match
	assert.False(t, scoped.MatchesScope([]models.IncidentLabel{{Key: "service_name", Label: "payments"}}))

	// A scope with no filter covers everything, including unlabelled
	// incidents — it means "all incidents", not "whatever no other scope
	// claimed" (grafana_irm_plan.md §4.1).
	unscoped := &GrafanaIrmOptions{ScopeConfig: &models.GrafanaIrmScopeConfig{}}
	assert.True(t, unscoped.MatchesScope(payments))
	assert.True(t, unscoped.MatchesScope(unlabelled))

	// A missing scope config behaves like an unfiltered one.
	missing := &GrafanaIrmOptions{}
	assert.True(t, missing.MatchesScope(unlabelled))
	assert.Equal(t, "", missing.LabelFilterTerm())
}

func TestScopeConfigLabelFilterTerm(t *testing.T) {
	// A half-configured filter is treated as no filter rather than emitting
	// a malformed `field:key:` term the API would reject.
	assert.Equal(t, "", (&models.GrafanaIrmScopeConfig{LabelKey: "team_name"}).LabelFilterTerm())
	assert.Equal(t, "", (&models.GrafanaIrmScopeConfig{LabelValue: "payments"}).LabelFilterTerm())
	assert.Equal(t, "field:team_name:payments",
		(&models.GrafanaIrmScopeConfig{LabelKey: "team_name", LabelValue: "payments"}).LabelFilterTerm())
}
