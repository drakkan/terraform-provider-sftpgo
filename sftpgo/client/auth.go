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
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	authEndpoint = "/api/v2/token"
)

// signInAdmin returns a new access token for the admin with the specified credentials.
func (c *Client) signInAdmin() (*AuthResponse, error) {
	if c.Auth.Username == "" || c.Auth.Password == "" {
		return nil, fmt.Errorf("define username and password")
	}

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s%s", c.HostURL, authEndpoint), nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.Auth.Username, c.Auth.Password)

	body, err := c.doRequest(req, http.StatusOK)
	if err != nil {
		return nil, err
	}

	ar := AuthResponse{}
	err = json.Unmarshal(body, &ar)
	if err != nil {
		return nil, err
	}

	return &ar, nil
}
