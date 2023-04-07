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

	"github.com/sftpgo/sdk"
)

// User defines a SFTPGo user
type User struct {
	sdk.User
	// we remote the omitempty attribute from the password
	// otherwise setting an empty password will preserve the current one
	// on updated
	Password string `json:"password"`
}

// GetUsers - Returns list of users
func (c *Client) GetUsers() ([]User, error) {
	var result []User
	limit := 100

	for {
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/v2/users?limit=%d&offset=%d", c.HostURL, limit, len(result)), nil)
		if err != nil {
			return nil, err
		}

		body, err := c.doRequest(req, http.StatusOK)
		if err != nil {
			return nil, err
		}

		var users []User
		err = json.Unmarshal(body, &users)
		if err != nil {
			return nil, err
		}
		result = append(result, users...)
		if len(users) < limit {
			break
		}
	}

	return result, nil
}

// CreateUser - creates a new user
func (c *Client) CreateUser(user User) (*User, error) {
	rb, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/v2/users", c.HostURL), bytes.NewBuffer(rb))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req, http.StatusCreated)
	if err != nil {
		return nil, err
	}

	var newUser User
	err = json.Unmarshal(body, &newUser)
	return &newUser, err
}

// GetUser - Returns a specifc user
func (c *Client) GetUser(username string) (*User, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/v2/users/%s", c.HostURL, url.PathEscape(username)), nil)
	if err != nil {
		return nil, err
	}
	body, err := c.doRequest(req, http.StatusOK)
	if err != nil {
		return nil, err
	}

	var user User
	err = json.Unmarshal(body, &user)
	return &user, err
}

// UpdateUser - Updates an existing user
func (c *Client) UpdateUser(user User) error {
	rb, err := json.Marshal(user)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/api/v2/users/%s", c.HostURL, url.PathEscape(user.Username)),
		bytes.NewBuffer(rb))
	if err != nil {
		return err
	}

	_, err = c.doRequest(req, http.StatusOK)
	return err
}

// DeleteUser - Deletes a user
func (c *Client) DeleteUser(username string) error {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/v2/users/%s", c.HostURL, url.PathEscape(username)), nil)
	if err != nil {
		return err
	}
	_, err = c.doRequest(req, http.StatusOK)
	return err
}
