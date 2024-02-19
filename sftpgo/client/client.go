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
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sftpgo/sdk"
)

// HostURL - Default SFTPGo URL
const HostURL string = "http://localhost:8080"

// Client defines the SFTPGo API client
type Client struct {
	HostURL     string
	HTTPClient  *http.Client
	AccessToken string
	APIKey      string
	Auth        AuthStruct
	Headers     map[string]string
}

// AuthStruct defines th SFTPGo API auth
type AuthStruct struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// AuthResponse defines the SFTPGo API auth response
type AuthResponse struct {
	AccessToken string    `json:"access_token"`
	ExpiresAt   time.Time `json:"expires_at"`
}

type backupData struct {
	Users        []User                  `json:"users"`
	Groups       []sdk.Group             `json:"groups"`
	Folders      []sdk.BaseVirtualFolder `json:"folders"`
	Admins       []Admin                 `json:"admins"`
	EventActions []BaseEventAction       `json:"event_actions"`
	Version      int                     `json:"version"`
}

// NewClient return an SFTPGo API client
func NewClient(host, username, password, apiKey *string, headers map[string]string) (*Client, error) {
	c := Client{
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
		// Default SFTPGo URL
		HostURL: HostURL,
	}

	if host != nil {
		c.HostURL = *host
	}

	if getStringFromPointer(apiKey) != "" {
		c.APIKey = *apiKey
		return &c, nil
	}

	// If username or password not provided, return empty client
	if getStringFromPointer(username) == "" || getStringFromPointer(password) == "" {
		return nil, fmt.Errorf("define username and password")
	}

	c.Auth = AuthStruct{
		Username: *username,
		Password: *password,
	}

	c.Headers = headers

	ar, err := c.SignInAdmin()
	if err != nil {
		return nil, err
	}

	c.AccessToken = ar.AccessToken

	return &c, nil
}

func (c *Client) doRequest(req *http.Request, expectedStatusCode int) ([]byte, error) {
	if c.AccessToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.AccessToken))
	} else if c.APIKey != "" {
		req.Header.Set("X-SFTPGO-API-KEY", c.APIKey)
	}

	if c.Headers != nil {
		for k, v := range c.Headers {
			req.Header.Set(k, v)
		}
	}

	if req.Body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != expectedStatusCode {
		return nil, fmt.Errorf("status: %d, body: %s", res.StatusCode, body)
	}

	return body, err
}

func getStringFromPointer(val *string) string {
	if val == nil {
		return ""
	}
	return *val
}
