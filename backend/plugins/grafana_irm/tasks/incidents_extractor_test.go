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
	"github.com/stretchr/testify/require"

	"github.com/apache/devlake/plugins/grafana_irm/models"
)

// These fixtures are not hand-authored: they're the exact (compacted) JSON
// captured live from a real Grafana Cloud dev stack on 2026-09-19 — incident
// `2` (active, real labels, one real role assignment among the placeholder
// slots) and incident `5` (resolved, no labels, no assignments). See
// grafana_irm_plan.md §9.

const activeWithAssignmentJSON = `{"incidentID": "2", "refs": [], "severity": "Major", "labels": [{"key": "service_name", "label": "checkout-api", "description": "", "colorHex": ""}, {"key": "team_name", "label": "checkout", "description": "", "colorHex": ""}], "isDrill": true, "incidentType": "internal", "createdTime": "2026-09-19T11:02:02.688373Z", "modifiedTime": "2026-09-19T11:07:35.404041Z", "createdByUser": {"userID": "grafana-incident:user-6aae6b80244e1a8626a0a593", "name": "Service Account: grafana-irm-plugin-dev", "photoURL": "https://www.gravatar.com/avatar/e244d010798c52b18bb2009c179dac20?s=512&d=retro"}, "closedTime": "", "durationSeconds": 17876, "status": "active", "title": "[DevLake seed] Checkout service 5xx spike", "overviewURL": "/a/grafana-irm-app/incidents/2/devlake-seed-checkout-service-5xx-spike", "incidentMembership": {"assignments": [{"user": {"userID": "grafana-incident:user-6aae6b80244e1a8626a0a593", "name": "Service Account: grafana-irm-plugin-dev", "photoURL": "https://www.gravatar.com/avatar/e244d010798c52b18bb2009c179dac20?s=512&d=retro"}, "role": {"roleID": 9576, "orgID": "1836253", "name": "commander", "description": "Owns the incident (has their full-time attention)", "important": true, "mandatory": true, "archived": false, "createdAt": "2026-09-19T11:00:21Z", "updatedAt": ""}, "roleID": 9576}, {"user": {"userID": "", "name": "", "photoURL": ""}, "role": {"roleID": 0, "orgID": "", "name": "", "description": "", "important": false, "mandatory": false, "archived": false, "createdAt": "", "updatedAt": ""}, "roleID": 0}, {"user": {"userID": "", "name": "", "photoURL": ""}, "role": {"roleID": 0, "orgID": "", "name": "", "description": "", "important": false, "mandatory": false, "archived": false, "createdAt": "", "updatedAt": ""}, "roleID": 0}], "totalAssignments": 1, "totalParticipants": 0}, "taskList": {"tasks": [], "todoCount": 0, "doneCount": 0}, "summary": "", "incidentStart": "2026-09-19T11:02:02Z", "incidentEnd": "", "incidentChannels": []}`

const resolvedNoExtrasJSON = `{"incidentID": "5", "refs": [], "severity": "Minor", "labels": [], "isDrill": false, "incidentType": "internal", "createdTime": "2026-09-19T11:33:10.948822Z", "modifiedTime": "2026-09-19T11:33:27.648602Z", "createdByUser": {"userID": "grafana-incident:user-6aae6b80244e1a8626a0a593", "name": "Service Account: grafana-irm-plugin-dev", "photoURL": "https://www.gravatar.com/avatar/e244d010798c52b18bb2009c179dac20?s=512&d=retro"}, "closedTime": "2026-09-19T11:33:27.17626Z", "durationSeconds": 17, "status": "resolved", "title": "[DevLake seed] Cache cluster node failure", "overviewURL": "/a/grafana-irm-app/incidents/5/devlake-seed-cache-cluster-node-failure", "incidentMembership": {"assignments": [{"user": {"userID": "", "name": "", "photoURL": ""}, "role": {"roleID": 0, "orgID": "", "name": "", "description": "", "important": false, "mandatory": false, "archived": false, "createdAt": "", "updatedAt": ""}, "roleID": 0}], "totalAssignments": 0, "totalParticipants": 0}, "taskList": {"tasks": [], "todoCount": 0, "doneCount": 0}, "summary": "", "incidentStart": "2026-09-19T11:33:10Z", "incidentEnd": "2026-09-19T11:33:27.17626Z", "incidentChannels": []}`

func newTestOptions() *GrafanaIrmOptions {
	return &GrafanaIrmOptions{ConnectionId: 1}
}

func TestExtractIncident_ActiveWithLabelsAndAssignment(t *testing.T) {
	op := newTestOptions()
	results, err := extractIncident([]byte(activeWithAssignmentJSON), op, "https://mystack.grafana.net")
	require.NoError(t, err)
	// 1 incident + 2 labels + 1 assignment (9 empty placeholder slots filtered out)
	require.Len(t, results, 4)

	incident := results[0].(*models.Incident)
	assert.Equal(t, uint64(1), incident.ConnectionId)
	assert.Equal(t, "2", incident.Id)
	assert.Equal(t, "[DevLake seed] Checkout service 5xx spike", incident.Title)
	assert.Equal(t, "https://mystack.grafana.net/a/grafana-irm-app/incidents/2/devlake-seed-checkout-service-5xx-spike", incident.Url)
	assert.Equal(t, "active", incident.Status)
	assert.Equal(t, "Major", incident.Severity)
	assert.True(t, incident.CreatedDate.Equal(mustParseTime(t, "2026-09-19T11:02:02.688373Z")))
	assert.True(t, incident.UpdatedDate.Equal(mustParseTime(t, "2026-09-19T11:07:35.404041Z")))
	assert.Nil(t, incident.ResolvedDate)

	label1 := results[1].(*models.IncidentLabel)
	assert.Equal(t, "service_name", label1.Key)
	assert.Equal(t, "checkout-api", label1.Label)
	label2 := results[2].(*models.IncidentLabel)
	assert.Equal(t, "team_name", label2.Key)
	assert.Equal(t, "checkout", label2.Label)

	assignment := results[3].(*models.IncidentAssignment)
	assert.Equal(t, "grafana-incident:user-6aae6b80244e1a8626a0a593", assignment.UserId)
	assert.Equal(t, "Service Account: grafana-irm-plugin-dev", assignment.UserName)
	assert.Equal(t, "commander", assignment.RoleName)
}

func TestExtractIncident_ResolvedNoLabelsNoAssignments(t *testing.T) {
	op := newTestOptions()
	results, err := extractIncident([]byte(resolvedNoExtrasJSON), op, "https://mystack.grafana.net")
	require.NoError(t, err)
	require.Len(t, results, 1)

	incident := results[0].(*models.Incident)
	assert.Equal(t, "5", incident.Id)
	assert.Equal(t, "resolved", incident.Status)
	require.NotNil(t, incident.ResolvedDate)
	assert.True(t, incident.ResolvedDate.Equal(mustParseTime(t, "2026-09-19T11:33:27.17626Z")))
}

func mustParseTime(t *testing.T, value string) time.Time {
	t.Helper()
	parsed, err := time.Parse(time.RFC3339, value)
	require.NoError(t, err)
	return parsed
}
