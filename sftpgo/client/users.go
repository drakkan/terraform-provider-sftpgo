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

// GetUsers - Returns list of users
func (c *Client) GetUsers() ([]User, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/v2/dumpdata?output-data=1&scopes=users", c.HostURL), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithAuth(req, http.StatusOK)
	if err != nil {
		return nil, err
	}

	var data backupData
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}

	return data.Users, nil
}

// CreateUser - creates a new user
func (c *Client) CreateUser(user User) (*User, error) {
	rb, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/v2/users?confidential_data=1", c.HostURL),
		bytes.NewBuffer(rb))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithAuth(req, http.StatusCreated)
	if err != nil {
		return nil, err
	}

	var newUser User
	err = json.Unmarshal(body, &newUser)
	return &newUser, err
}

// GetUser - Returns a specifc user
func (c *Client) GetUser(username string) (*User, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/v2/users/%s?confidential_data=1", c.HostURL,
		url.PathEscape(username)), nil)
	if err != nil {
		return nil, err
	}
	body, err := c.doRequestWithAuth(req, http.StatusOK)
	if err != nil {
		return nil, err
	}

	var user User
	err = json.Unmarshal(body, &user)
	return &user, err
}

// UpdateUser - Updates an existing user. When password is nil the field is
// omitted from the JSON payload and the server keeps the current password;
// pass a pointer to an empty string to explicitly clear it, or a pointer to a
// non-empty string to set a new value. See User.Password for the rationale
// behind the string-not-omitempty tag on the base struct.
func (c *Client) UpdateUser(user User, password *string) error {
	payload := struct {
		User
		Password *string `json:"password,omitempty"`
	}{User: user, Password: password}
	rb, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/api/v2/users/%s", c.HostURL, url.PathEscape(user.Username)),
		bytes.NewBuffer(rb))
	if err != nil {
		return err
	}

	_, err = c.doRequestWithAuth(req, http.StatusOK)
	return err
}

// DeleteUser - Deletes a user
func (c *Client) DeleteUser(username string) error {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/v2/users/%s", c.HostURL, url.PathEscape(username)), nil)
	if err != nil {
		return err
	}
	_, err = c.doRequestWithAuth(req, http.StatusOK)
	return err
}
