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
	"sync"
	"time"
)

// StatusError represents an HTTP error with status code
type StatusError struct {
	StatusCode int
	Body       []byte
}

func (e StatusError) Error() string {
	return fmt.Sprintf("status: %d, body: %s", e.StatusCode, e.Body)
}

// HostURL - Default SFTPGo URL
const HostURL string = "http://localhost:8080"

// Client defines the SFTPGo API client
type Client struct {
	HostURL      string
	HTTPClient   *http.Client
	APIKey       string
	Auth         AuthStruct
	Headers      []KeyValue
	Edition      int64
	mu           sync.RWMutex
	authResponse *AuthResponse
}

func (c *Client) IsEnterpriseEdition() bool {
	return c.Edition == 1
}

func (c *Client) setAuthResponse(ar *AuthResponse) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.authResponse = ar
}

func (c *Client) getAccessToken() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.authResponse == nil {
		return ""
	}

	if c.authResponse.ExpiresAt.Before(time.Now().Add(-2 * time.Minute)) {
		return ""
	}

	return c.authResponse.AccessToken
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
	Users        []User              `json:"users"`
	Groups       []Group             `json:"groups"`
	Folders      []BaseVirtualFolder `json:"folders"`
	Admins       []Admin             `json:"admins"`
	EventActions []BaseEventAction   `json:"event_actions"`
	Version      int                 `json:"version"`
}

// NewClient return an SFTPGo API client
func NewClient(host, username, password, apiKey string, headers []KeyValue, edition int64) (*Client, error) {
	c := Client{
		HTTPClient: &http.Client{Timeout: 20 * time.Second},
		// Default SFTPGo URL
		HostURL: HostURL,
		Headers: headers,
		Edition: edition,
	}

	if host != "" {
		c.HostURL = host
	}

	if apiKey != "" {
		c.APIKey = apiKey
		return &c, nil
	}

	// If username or password not provided, return empty client
	if username == "" || password == "" {
		return nil, fmt.Errorf("define username and password")
	}

	c.Auth = AuthStruct{
		Username: username,
		Password: password,
	}

	return &c, nil
}

func (c *Client) setAuthHeader(req *http.Request) error {
	if c.APIKey != "" {
		req.Header.Set("X-SFTPGO-API-KEY", c.APIKey)
		return nil
	}

	accessToken := c.getAccessToken()
	if accessToken == "" {
		ar, err := c.signInAdmin()
		if err != nil {
			return err
		}
		c.setAuthResponse(ar)

		accessToken = ar.AccessToken
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	return nil
}

func (c *Client) doRequestWithAuth(req *http.Request, expectedStatusCode int) ([]byte, error) {
	if err := c.setAuthHeader(req); err != nil {
		return nil, err
	}
	return c.doRequest(req, expectedStatusCode)
}

func (c *Client) doRequest(req *http.Request, expectedStatusCode int) ([]byte, error) {
	for _, h := range c.Headers {
		req.Header.Set(h.Key, h.Value)
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
		return nil, StatusError{
			StatusCode: res.StatusCode,
			Body:       body,
		}
	}

	return body, err
}
