// Copyright (C) 2023 Nicola Murino
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// Schedule defines an event schedule
type Schedule struct {
	Minute     string `json:"minute"`
	Hours      string `json:"hour"`
	DayOfWeek  string `json:"day_of_week"`
	DayOfMonth string `json:"day_of_month"`
	Month      string `json:"month"`
}

// ConditionPattern defines a pattern for condition filters
type ConditionPattern struct {
	Pattern      string `json:"pattern,omitempty"`
	InverseMatch bool   `json:"inverse_match,omitempty"`
}

// ConditionOptions defines options for event conditions
type ConditionOptions struct {
	// Usernames or folder names
	Names []ConditionPattern `json:"names,omitempty"`
	// Group names
	GroupNames []ConditionPattern `json:"group_names,omitempty"`
	// Role names
	RoleNames []ConditionPattern `json:"role_names,omitempty"`
	// Virtual paths
	FsPaths         []ConditionPattern `json:"fs_paths,omitempty"`
	Protocols       []string           `json:"protocols,omitempty"`
	ProviderObjects []string           `json:"provider_objects,omitempty"`
	MinFileSize     int64              `json:"min_size,omitempty"`
	MaxFileSize     int64              `json:"max_size,omitempty"`
	EventStatuses   []int              `json:"event_statuses,omitempty"`
	// allow to execute scheduled tasks concurrently from multiple instances
	ConcurrentExecution bool `json:"concurrent_execution,omitempty"`
}

// EventRuleConditions defines the conditions for an event rule
type EventRuleConditions struct {
	// Only one between FsEvents, ProviderEvents and Schedule is allowed
	FsEvents       []string   `json:"fs_events,omitempty"`
	ProviderEvents []string   `json:"provider_events,omitempty"`
	Schedules      []Schedule `json:"schedules,omitempty"`
	// 0 any, 1 user, 2 admin
	IDPLoginEvent int              `json:"idp_login_event,omitempty"`
	Options       ConditionOptions `json:"options"`
}

// EventActionRelationOptions defines the supported relation options for an event action
type EventActionRelationOptions struct {
	IsFailureAction bool `json:"is_failure_action"`
	StopOnFailure   bool `json:"stop_on_failure"`
	ExecuteSync     bool `json:"execute_sync"`
}

// EventAction defines an event action
type EventAction struct {
	Name string `json:"name"`
	// Order defines the execution order
	Order   int                        `json:"order,omitempty"`
	Options EventActionRelationOptions `json:"relation_options"`
}

// EventRule defines the trigger, conditions and actions for an event
type EventRule struct {
	// Rule name
	Name string `json:"name"`
	// 1 enabled, 0 disabled
	Status int `json:"status"`
	// optional description
	Description string `json:"description,omitempty"`
	// Creation time as unix timestamp in milliseconds
	CreatedAt int64 `json:"created_at"`
	// last update time as unix timestamp in milliseconds
	UpdatedAt int64 `json:"updated_at"`
	// Event trigger
	Trigger int `json:"trigger"`
	// Event conditions
	Conditions EventRuleConditions `json:"conditions"`
	// actions to execute
	Actions []EventAction `json:"actions"`
}

// GetActions - Returns list of actions
func (c *Client) GetRules() ([]EventRule, error) {
	var result []EventRule
	limit := 100

	for {
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/v2/eventrules?limit=%d&offset=%d",
			c.HostURL, limit, len(result)), nil)
		if err != nil {
			return nil, err
		}

		body, err := c.doRequestWithAuth(req, http.StatusOK)
		if err != nil {
			return nil, err
		}

		var rules []EventRule
		err = json.Unmarshal(body, &rules)
		if err != nil {
			return nil, err
		}
		result = append(result, rules...)
		if len(rules) < limit {
			break
		}
	}

	return result, nil
}

// CreateRule - Creates a new rule
func (c *Client) CreateRule(rule EventRule) (*EventRule, error) {
	rb, err := json.Marshal(rule)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/v2/eventrules", c.HostURL), bytes.NewBuffer(rb))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithAuth(req, http.StatusCreated)
	if err != nil {
		return nil, err
	}

	var newRule EventRule
	err = json.Unmarshal(body, &newRule)
	return &newRule, err
}

// GetRule - Returns a specifc role
func (c *Client) GetRule(name string) (*EventRule, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/v2/eventrules/%s", c.HostURL, url.PathEscape(name)), nil)
	if err != nil {
		return nil, err
	}
	body, err := c.doRequestWithAuth(req, http.StatusOK)
	if err != nil {
		return nil, err
	}

	var rule EventRule
	err = json.Unmarshal(body, &rule)
	return &rule, err
}

// UpdateRule - Updates an existing rule
func (c *Client) UpdateRule(rule EventRule) error {
	rb, err := json.Marshal(rule)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/api/v2/eventrules/%s", c.HostURL, url.PathEscape(rule.Name)),
		bytes.NewBuffer(rb))
	if err != nil {
		return err
	}

	_, err = c.doRequestWithAuth(req, http.StatusOK)
	return err
}

// DeleteRule - Deletes a rule
func (c *Client) DeleteRule(name string) error {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/v2/eventrules/%s", c.HostURL, url.PathEscape(name)), nil)
	if err != nil {
		return err
	}
	_, err = c.doRequestWithAuth(req, http.StatusOK)
	return err
}
