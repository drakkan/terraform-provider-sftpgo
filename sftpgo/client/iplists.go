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

// IPListEntry defines an entry for the IP addresses list
type IPListEntry struct {
	IPOrNet     string `json:"ipornet"`
	Description string `json:"description,omitempty"`
	Type        int    `json:"type"`
	Mode        int    `json:"mode"`
	// Defines the protocols the entry applies to
	// - 0 all the supported protocols
	// - 1 SSH
	// - 2 FTP
	// - 4 WebDAV
	// - 8 HTTP
	// Protocols can be combined
	Protocols int `json:"protocols"`
	// Creation time as unix timestamp in milliseconds
	CreatedAt int64 `json:"created_at"`
	// last update time as unix timestamp in milliseconds
	UpdatedAt int64 `json:"updated_at"`
}

// GetIPListEntries - Returns entries for the specified IP list type
func (c *Client) GetIPListEntries(listType int) ([]IPListEntry, error) {
	var result []IPListEntry
	limit := 100
	from := ""

	for {
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/v2/iplists/%d?limit=%d&from=%s",
			c.HostURL, listType, limit, url.QueryEscape(from)), nil)
		if err != nil {
			return nil, err
		}

		body, err := c.doRequest(req, http.StatusOK)
		if err != nil {
			return nil, err
		}

		var entries []IPListEntry
		err = json.Unmarshal(body, &entries)
		if err != nil {
			return nil, err
		}
		result = append(result, entries...)
		if len(entries) < limit {
			break
		}
		from = result[len(result)-1].IPOrNet
	}

	return result, nil
}

// CreateIPListEntry - Creates a new IP list entry
func (c *Client) CreateIPListEntry(entry IPListEntry) (*IPListEntry, error) {
	rb, err := json.Marshal(entry)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/v2/iplists/%d", c.HostURL, entry.Type), bytes.NewBuffer(rb))
	if err != nil {
		return nil, err
	}

	_, err = c.doRequest(req, http.StatusCreated)
	if err != nil {
		return nil, err
	}

	return c.GetIPListEntry(entry.Type, entry.IPOrNet)
}

// GetIPListEntry - Returns a specifc IP list entry
func (c *Client) GetIPListEntry(listType int, ipOrNet string) (*IPListEntry, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/v2/iplists/%d/%s", c.HostURL, listType, url.PathEscape(ipOrNet)), nil)
	if err != nil {
		return nil, err
	}
	body, err := c.doRequest(req, http.StatusOK)
	if err != nil {
		return nil, err
	}

	var entry IPListEntry
	err = json.Unmarshal(body, &entry)
	return &entry, err
}

// UpdateIPListEntry - Updates an existing IP list entru
func (c *Client) UpdateIPListEntry(entry IPListEntry) error {
	rb, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/api/v2/iplists/%d/%s",
		c.HostURL, entry.Type, url.PathEscape(entry.IPOrNet)), bytes.NewBuffer(rb))
	if err != nil {
		return err
	}

	_, err = c.doRequest(req, http.StatusOK)
	return err
}

// DeleteIPListEntry - Deletes an IP list entry
func (c *Client) DeleteIPListEntry(listType int, ipOrNet string) error {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/v2/iplists/%d/%s",
		c.HostURL, listType, url.PathEscape(ipOrNet)), nil)
	if err != nil {
		return err
	}
	_, err = c.doRequest(req, http.StatusOK)
	return err
}
