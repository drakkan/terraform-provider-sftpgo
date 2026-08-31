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
	// Resource isolation level: 0 disabled, 1 enabled. Enterprise edition only
	ResourceIsolation int `json:"resource_isolation,omitempty"`
	// Additional settings. Enterprise edition only
	Settings RoleSettings `json:"settings,omitzero"`
}

// RoleSettings defines the additional settings of a role.
// Settings granting nothing are omitted: the server reads an absent envelope
// as an empty one
type RoleSettings struct {
	StorageAllowlist RoleStorageAllowlist `json:"storage_allowlist,omitzero"`
}

// RoleStorageAllowlist defines the storage the resources carrying the role may name.
type RoleStorageAllowlist struct {
	AllowedProviders []int             `json:"allowed_providers,omitempty"`
	Local            LocalRoleScope    `json:"local,omitzero"`
	S3               S3RoleScope       `json:"s3,omitzero"`
	Azure            AzureRoleScope    `json:"azure,omitzero"`
	GCS              GCSRoleScope      `json:"gcs,omitzero"`
	SFTP             SFTPRoleScope     `json:"sftp,omitzero"`
	FTP              EndpointRoleScope `json:"ftp,omitzero"`
	HTTP             EndpointRoleScope `json:"http,omitzero"`
}

// LocalRoleScope defines the local paths the resources carrying the role may name.
type LocalRoleScope struct {
	AllowedPaths []string `json:"allowed_paths,omitempty"`
	UsersBaseDir string   `json:"users_base_dir,omitempty"`
}

// IsEmpty returns true if the scope grants nothing.
func (s LocalRoleScope) IsEmpty() bool {
	return len(s.AllowedPaths) == 0 && s.UsersBaseDir == ""
}

// S3BucketRef identifies an S3 resource.
type S3BucketRef struct {
	Bucket    string `json:"bucket"`
	KeyPrefix string `json:"key_prefix,omitempty"`
	Endpoint  string `json:"endpoint,omitempty"`
}

// S3RoleScope defines the S3 resources the resources carrying the role may name.
type S3RoleScope struct {
	DefaultAllow   bool          `json:"default_allow,omitempty"`
	AllowedBuckets []S3BucketRef `json:"allowed_buckets,omitempty"`
	DeniedBuckets  []S3BucketRef `json:"denied_buckets,omitempty"`
}

// IsEmpty returns true if the scope grants nothing.
func (s S3RoleScope) IsEmpty() bool {
	return !s.DefaultAllow && len(s.AllowedBuckets) == 0 && len(s.DeniedBuckets) == 0
}

// AzureContainerRef identifies an Azure Blob resource.
type AzureContainerRef struct {
	Account   string `json:"account"`
	Container string `json:"container"`
	KeyPrefix string `json:"key_prefix,omitempty"`
	Endpoint  string `json:"endpoint,omitempty"`
}

// AzureRoleScope defines the Azure Blob resources the resources carrying the role may name.
type AzureRoleScope struct {
	DefaultAllow      bool                `json:"default_allow,omitempty"`
	AllowedContainers []AzureContainerRef `json:"allowed_containers,omitempty"`
	DeniedContainers  []AzureContainerRef `json:"denied_containers,omitempty"`
}

// IsEmpty returns true if the scope grants nothing.
func (s AzureRoleScope) IsEmpty() bool {
	return !s.DefaultAllow && len(s.AllowedContainers) == 0 && len(s.DeniedContainers) == 0
}

// GCSBucketRef identifies a GCS resource.
type GCSBucketRef struct {
	Bucket         string `json:"bucket"`
	KeyPrefix      string `json:"key_prefix,omitempty"`
	UniverseDomain string `json:"universe_domain,omitempty"`
}

// GCSRoleScope defines the GCS resources the resources carrying the role may name.
type GCSRoleScope struct {
	DefaultAllow   bool           `json:"default_allow,omitempty"`
	AllowedBuckets []GCSBucketRef `json:"allowed_buckets,omitempty"`
	DeniedBuckets  []GCSBucketRef `json:"denied_buckets,omitempty"`
}

// IsEmpty returns true if the scope grants nothing.
func (s GCSRoleScope) IsEmpty() bool {
	return !s.DefaultAllow && len(s.AllowedBuckets) == 0 && len(s.DeniedBuckets) == 0
}

// EndpointRoleScope defines the remote endpoints the resources carrying the role may name.
type EndpointRoleScope struct {
	DefaultAllow     bool     `json:"default_allow,omitempty"`
	AllowedEndpoints []string `json:"allowed_endpoints,omitempty"`
	DeniedEndpoints  []string `json:"denied_endpoints,omitempty"`
}

// IsEmpty returns true if the scope grants nothing.
func (s EndpointRoleScope) IsEmpty() bool {
	return !s.DefaultAllow && len(s.AllowedEndpoints) == 0 && len(s.DeniedEndpoints) == 0
}

// SFTPRoleScope defines the SFTP endpoints and SOCKS proxies the resources
// carrying the role may name.
type SFTPRoleScope struct {
	EndpointRoleScope
	AllowedProxies []string `json:"allowed_proxies,omitempty"`
}

// IsEmpty returns true if the scope grants nothing. It shadows the promoted
// method, which answers for the endpoints alone. Renaming either of them to
// IsZero makes it the omitzero predicate
func (s SFTPRoleScope) IsEmpty() bool {
	return s.EndpointRoleScope.IsEmpty() && len(s.AllowedProxies) == 0
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
