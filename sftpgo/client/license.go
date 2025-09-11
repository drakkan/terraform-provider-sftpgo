// Copyright (C) 2025 Nicola Murino
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
)

type LicenseFeatures struct {
	MaxConcurrentTransfers int   `json:"max_concurrent_transfers"`
	FSProviders            []int `json:"fs_providers"`
	EventActions           []int `json:"event_actions"`
	FSActions              []int `json:"fs_actions"`
	Plugins                int   `json:"plugins"`
	Metering               int   `json:"metering"`
	WOPIUsers              int   `json:"wopi_users"`
	HA                     []int `json:"ha"`
}

type License struct {
	Key       string          `json:"key"`
	Type      int             `json:"type"`
	ValidFrom int64           `json:"valid_from"`
	ValidTo   int64           `json:"valid_to"`
	Features  LicenseFeatures `json:"features"`
}

// GetLicense - Returns the current license
func (c *Client) GetLicense() (*License, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/v2/license", c.HostURL), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithAuth(req, http.StatusOK)
	if err != nil {
		return nil, err
	}

	var license License
	err = json.Unmarshal(body, &license)
	return &license, err
}

// AddLicense - Adds or update a license key
func (c *Client) AddLicense(key string) (*License, error) {
	lic := map[string]string{"key": key}

	rb, err := json.Marshal(lic)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/v2/license", c.HostURL), bytes.NewBuffer(rb))
	if err != nil {
		return nil, err
	}

	_, err = c.doRequestWithAuth(req, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return c.GetLicense()
}
