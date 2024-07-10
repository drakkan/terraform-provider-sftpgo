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

// Role defines an SFTPGo role.
type Role struct {
	// Role name
	Name string `json:"name"`
	// optional description
	Description string `json:"description,omitempty"`
	// Creation time as unix timestamp in milliseconds
	CreatedAt int64 `json:"created_at"`
	// last update time as unix timestamp in milliseconds
	UpdatedAt int64 `json:"updated_at"`
}

// GetRoles - Returns list of roles
func (c *Client) GetRoles() ([]Role, error) {
	var result []Role
	limit := 100

	for {
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/v2/roles?limit=%d&offset=%d", c.HostURL, limit, len(result)), nil)
		if err != nil {
			return nil, err
		}

		body, err := c.doRequestWithAuth(req, http.StatusOK)
		if err != nil {
			return nil, err
		}

		var roles []Role
		err = json.Unmarshal(body, &roles)
		if err != nil {
			return nil, err
		}
		result = append(result, roles...)
		if len(roles) < limit {
			break
		}
	}

	return result, nil
}

// CreateRole - Creates a new role
func (c *Client) CreateRole(role Role) (*Role, error) {
	rb, err := json.Marshal(role)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/v2/roles", c.HostURL), bytes.NewBuffer(rb))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithAuth(req, http.StatusCreated)
	if err != nil {
		return nil, err
	}

	var newRole Role
	err = json.Unmarshal(body, &newRole)
	return &newRole, err
}

// GetRole - Returns a specifc role
func (c *Client) GetRole(name string) (*Role, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/v2/roles/%s", c.HostURL, url.PathEscape(name)), nil)
	if err != nil {
		return nil, err
	}
	body, err := c.doRequestWithAuth(req, http.StatusOK)
	if err != nil {
		return nil, err
	}

	var role Role
	err = json.Unmarshal(body, &role)
	return &role, err
}

// UpdateRole - Updates an existing role
func (c *Client) UpdateRole(role Role) error {
	rb, err := json.Marshal(role)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/api/v2/roles/%s", c.HostURL, url.PathEscape(role.Name)),
		bytes.NewBuffer(rb))
	if err != nil {
		return err
	}

	_, err = c.doRequestWithAuth(req, http.StatusOK)
	return err
}

// DeleteRole - Deletes a role
func (c *Client) DeleteRole(name string) error {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/v2/roles/%s", c.HostURL, url.PathEscape(name)), nil)
	if err != nil {
		return err
	}
	_, err = c.doRequestWithAuth(req, http.StatusOK)
	return err
}
