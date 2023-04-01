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
)

// HostURL - Default SFTPGo URL
const HostURL string = "http://localhost:8080"

// Client defines the SFTPGo API client
type Client struct {
	HostURL     string
	HTTPClient  *http.Client
	AccessToken string
	Auth        AuthStruct
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

// NewClient return an SFTPGo API client
func NewClient(host, username, password *string) (*Client, error) {
	c := Client{
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
		// Default SFTPGo URL
		HostURL: HostURL,
	}

	if host != nil {
		c.HostURL = *host
	}

	// If username or password not provided, return empty client
	if username == nil || password == nil {
		return nil, fmt.Errorf("define username and password")
	}

	c.Auth = AuthStruct{
		Username: *username,
		Password: *password,
	}

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
