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

// AdminPreferences defines the admin preferences
type AdminPreferences struct {
	// Allow to hide some sections from the user page.
	// These are not security settings and are not enforced server side
	// in any way. They are only intended to simplify the user page in
	// the WebAdmin UI.
	//
	// 1 means hide groups section
	// 2 means hide filesystem section, "users_base_dir" must be set in the config file otherwise this setting is ignored
	// 4 means hide virtual folders section
	// 8 means hide profile section
	// 16 means hide ACLs section
	// 32 means hide disk and bandwidth quota limits section
	// 64 means hide advanced settings section
	//
	// The settings can be combined
	HideUserPageSections int `json:"hide_user_page_sections,omitempty"`
	// Defines the default expiration for newly created users as number of days.
	// 0 means no expiration
	DefaultUsersExpiration int `json:"default_users_expiration,omitempty"`
}

// AdminFilters defines additional restrictions for SFTPGo admins
type AdminFilters struct {
	// only clients connecting from these IP/Mask are allowed.
	// IP/Mask must be in CIDR notation as defined in RFC 4632 and RFC 4291
	// for example "192.0.2.0/24" or "2001:db8::/32"
	AllowList []string `json:"allow_list,omitempty"`
	// API key auth allows to impersonate this administrator with an API key
	AllowAPIKeyAuth bool             `json:"allow_api_key_auth,omitempty"`
	Preferences     AdminPreferences `json:"preferences"`
}

// AdminGroupMappingOptions defines the options for admin/group mapping
type AdminGroupMappingOptions struct {
	AddToUsersAs int `json:"add_to_users_as,omitempty"`
}

// AdminGroupMapping defines the mapping between an SFTPGo admin and a group
type AdminGroupMapping struct {
	Name    string                   `json:"name"`
	Options AdminGroupMappingOptions `json:"options"`
}

// Admin defines a SFTPGo admin
type Admin struct {
	// 1 enabled, 0 disabled (login is not allowed)
	Status int `json:"status"`
	// Username
	Username       string       `json:"username"`
	Password       string       `json:"password,omitempty"`
	Email          string       `json:"email,omitempty"`
	Permissions    []string     `json:"permissions"`
	Filters        AdminFilters `json:"filters,omitempty"`
	Description    string       `json:"description,omitempty"`
	AdditionalInfo string       `json:"additional_info,omitempty"`
	// Groups membership
	Groups []AdminGroupMapping `json:"groups,omitempty"`
	// Creation time as unix timestamp in milliseconds
	CreatedAt int64 `json:"created_at"`
	// last update time as unix timestamp in milliseconds
	UpdatedAt int64 `json:"updated_at"`
	// Last login as unix timestamp in milliseconds
	LastLogin int64 `json:"last_login"`
	// Role name. If set the admin can only administer users with the same role.
	// Role admins cannot have the following permissions:
	// - manage_admins
	// - manage_apikeys
	// - manage_system
	// - manage_event_rules
	// - manage_roles
	Role string `json:"role,omitempty"`
}

// GetAdmins - Returns list of admin
func (c *Client) GetAdmins() ([]Admin, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/v2/dumpdata?output-data=1&scopes=admins", c.HostURL), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req, http.StatusOK)
	if err != nil {
		return nil, err
	}

	var data backupData
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}
	return data.Admins, nil
}

// CreateAdmin - creates a new admin
func (c *Client) CreateAdmin(admin Admin) (*Admin, error) {
	rb, err := json.Marshal(admin)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/v2/admins?confidential_data=1", c.HostURL),
		bytes.NewBuffer(rb))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req, http.StatusCreated)
	if err != nil {
		return nil, err
	}

	var newAdmin Admin
	err = json.Unmarshal(body, &newAdmin)
	return &newAdmin, err
}

// GetAdmin - Returns a specifc admin
func (c *Client) GetAdmin(username string) (*Admin, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/v2/admins/%s?confidential_data=1", c.HostURL,
		url.PathEscape(username)), nil)
	if err != nil {
		return nil, err
	}
	body, err := c.doRequest(req, http.StatusOK)
	if err != nil {
		return nil, err
	}

	var admin Admin
	err = json.Unmarshal(body, &admin)
	return &admin, err
}

// UpdateAdmin - Updates an existing admin
func (c *Client) UpdateAdmin(admin Admin) error {
	rb, err := json.Marshal(admin)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/api/v2/admins/%s", c.HostURL, url.PathEscape(admin.Username)),
		bytes.NewBuffer(rb))
	if err != nil {
		return err
	}

	_, err = c.doRequest(req, http.StatusOK)
	return err
}

// DeleteAdmin - Deletes an admin
func (c *Client) DeleteAdmin(username string) error {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/v2/admins/%s", c.HostURL, url.PathEscape(username)), nil)
	if err != nil {
		return err
	}
	_, err = c.doRequest(req, http.StatusOK)
	return err
}
