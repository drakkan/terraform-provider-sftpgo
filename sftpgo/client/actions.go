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

	"github.com/sftpgo/sdk/kms"
)

// Supported event actions
const (
	ActionTypeHTTP = iota + 1
	ActionTypeCommand
	ActionTypeEmail
	ActionTypeBackup
	ActionTypeUserQuotaReset
	ActionTypeFolderQuotaReset
	ActionTypeTransferQuotaReset
	ActionTypeDataRetentionCheck
	ActionTypeFilesystem
	actionTypeReserved
	ActionTypePasswordExpirationCheck
	ActionTypeUserExpirationCheck
	ActionTypeIDPAccountCheck
	ActionTypeUserInactivityCheck
)

// Supported filesystem actions
const (
	FilesystemActionRename = iota + 1
	FilesystemActionDelete
	FilesystemActionMkdirs
	FilesystemActionExist
	FilesystemActionCompress
	FilesystemActionCopy
	FilesystemActionPGP // Enterprise
)

// KeyValue defines a key/value pair
type KeyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// HTTPPart defines a part for HTTP multipart requests
type HTTPPart struct {
	Name     string     `json:"name,omitempty"`
	Filepath string     `json:"filepath,omitempty"`
	Headers  []KeyValue `json:"headers,omitempty"`
	Body     string     `json:"body,omitempty"`
}

// EventActionHTTPConfig defines the configuration for an HTTP event target
type EventActionHTTPConfig struct {
	Endpoint        string         `json:"endpoint,omitempty"`
	Username        string         `json:"username,omitempty"`
	Password        kms.BaseSecret `json:"password,omitempty"`
	Headers         []KeyValue     `json:"headers,omitempty"`
	Timeout         int            `json:"timeout,omitempty"`
	SkipTLSVerify   bool           `json:"skip_tls_verify,omitempty"`
	Method          string         `json:"method,omitempty"`
	QueryParameters []KeyValue     `json:"query_parameters,omitempty"`
	Body            string         `json:"body,omitempty"`
	Parts           []HTTPPart     `json:"parts,omitempty"`
}

// EventActionCommandConfig defines the configuration for a command event target
type EventActionCommandConfig struct {
	Cmd     string     `json:"cmd,omitempty"`
	Args    []string   `json:"args,omitempty"`
	Timeout int        `json:"timeout,omitempty"`
	EnvVars []KeyValue `json:"env_vars,omitempty"`
}

// EventActionEmailConfig defines the configuration options for SMTP event actions
type EventActionEmailConfig struct {
	Recipients  []string `json:"recipients,omitempty"`
	Bcc         []string `json:"bcc,omitempty"`
	Subject     string   `json:"subject,omitempty"`
	Body        string   `json:"body,omitempty"`
	Attachments []string `json:"attachments,omitempty"`
	ContentType int      `json:"content_type,omitempty"`
}

// FolderRetention defines a folder retention configuration
type FolderRetention struct {
	// Path is the virtual directory path, if no other specific retention is defined,
	// the retention applies for sub directories too. For example if retention is defined
	// for the paths "/" and "/sub" then the retention for "/" is applied for any file outside
	// the "/sub" directory
	Path string `json:"path"`
	// Retention time in hours. 0 means exclude this path
	Retention int `json:"retention"`
	// DeleteEmptyDirs defines if empty directories will be deleted.
	// The user need the delete permission
	DeleteEmptyDirs bool `json:"delete_empty_dirs,omitempty"`
}

// EventActionDataRetentionConfig defines the configuration for a data retention check
type EventActionDataRetentionConfig struct {
	Folders       []FolderRetention `json:"folders,omitempty"`
	ArchiveFolder string            `json:"archive_folder,omitempty"`
	ArchivePath   string            `json:"archive_path,omitempty"`
}

type EventActionPGP struct {
	// 1 Encrypt, 2 Decrypt
	Mode int `json:"mode,omitempty"`
	// 0 Default, 1 RFC 4880, 2 RFC 9580
	Profile    int            `json:"profile,omitempty"`
	Paths      []KeyValue     `json:"paths,omitempty"`
	Password   kms.BaseSecret `json:"password,omitempty"`
	PrivateKey kms.BaseSecret `json:"private_key,omitempty"`
	Passphrase kms.BaseSecret `json:"passphrase,omitempty"`
	PublicKey  string         `json:"public_key,omitempty"`
}

// EventActionFsCompress defines the configuration for the compress filesystem action
type EventActionFsCompress struct {
	// Archive path
	Name string `json:"name,omitempty"`
	// Paths to compress
	Paths []string `json:"paths,omitempty"`
}

// RenameConfig defines the configuration for a filesystem rename
type RenameConfig struct {
	// key is the source and target the value
	KeyValue
	// This setting only applies to storage providers that support
	// changing modification times.
	UpdateModTime bool `json:"update_modtime,omitempty"`
}

// EventActionFilesystemConfig defines the configuration for filesystem actions
type EventActionFilesystemConfig struct {
	// Filesystem actions, see the above enum
	Type int `json:"type,omitempty"`
	// files/dirs to rename, key is the source and target the value
	Renames []RenameConfig `json:"renames,omitempty"`
	// directories to create
	MkDirs []string `json:"mkdirs,omitempty"`
	// files/dirs to delete
	Deletes []string `json:"deletes,omitempty"`
	// file/dirs to check for existence
	Exist []string `json:"exist,omitempty"`
	// files/dirs to copy, key is the source and target the value
	Copy []KeyValue `json:"copy,omitempty"`
	// paths to compress and archive name
	Compress EventActionFsCompress `json:"compress"`
	// PGP encryption or decryption
	PGP          EventActionPGP `json:"pgp"`
	Folder       string         `json:"folder,omitempty"`
	TargetFolder string         `json:"target_folder,omitempty"`
}

// EventActionPasswordExpiration defines the configuration for password expiration actions
type EventActionPasswordExpiration struct {
	// An email notification will be generated for users whose password expires in a number
	// of days less than or equal to this threshold
	Threshold int `json:"threshold,omitempty"`
}

// EventActionUserInactivity defines the configuration for user inactivity checks.
type EventActionUserInactivity struct {
	// DisableThreshold defines inactivity in days, since the last login before disabling the account
	DisableThreshold int `json:"disable_threshold,omitempty"`
	// DeleteThreshold defines inactivity in days, since the last login before deleting the account
	DeleteThreshold int `json:"delete_threshold,omitempty"`
}

// EventActionIDPAccountCheck defines the check to execute after a successful IDP login
type EventActionIDPAccountCheck struct {
	// 0 create/update, 1 create the account if it doesn't exist
	Mode          int    `json:"mode,omitempty"`
	TemplateUser  string `json:"template_user,omitempty"`
	TemplateAdmin string `json:"template_admin,omitempty"`
}

// EventActionOptions defines the supported configuration options for event actions
type EventActionOptions struct {
	HTTPConfig           EventActionHTTPConfig          `json:"http_config"`
	CmdConfig            EventActionCommandConfig       `json:"cmd_config"`
	EmailConfig          EventActionEmailConfig         `json:"email_config"`
	RetentionConfig      EventActionDataRetentionConfig `json:"retention_config"`
	FsConfig             EventActionFilesystemConfig    `json:"fs_config"`
	PwdExpirationConfig  EventActionPasswordExpiration  `json:"pwd_expiration_config"`
	UserInactivityConfig EventActionUserInactivity      `json:"user_inactivity_config"`
	IDPConfig            EventActionIDPAccountCheck     `json:"idp_config"`
}

// BaseEventAction defines the common fields for an event action
type BaseEventAction struct {
	// Action name
	Name string `json:"name"`
	// optional description
	Description string `json:"description,omitempty"`
	// ActionType, see the above enum
	Type int `json:"type"`
	// Configuration options specific for the action type
	Options EventActionOptions `json:"options"`
}

// GetActions - Returns list of actions
func (c *Client) GetActions() ([]BaseEventAction, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/v2/dumpdata?output-data=1&scopes=actions", c.HostURL), nil)
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

	return data.EventActions, nil
}

// CreateAction - creates a new action
func (c *Client) CreateAction(action BaseEventAction) (*BaseEventAction, error) {
	rb, err := json.Marshal(action)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/v2/eventactions?confidential_data=1", c.HostURL),
		bytes.NewBuffer(rb))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithAuth(req, http.StatusCreated)
	if err != nil {
		return nil, err
	}

	var newAction BaseEventAction
	err = json.Unmarshal(body, &newAction)
	return &newAction, err
}

// GetAction - Returns a specifc action
func (c *Client) GetAction(name string) (*BaseEventAction, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/v2/eventactions/%s?confidential_data=1", c.HostURL,
		url.PathEscape(name)), nil)
	if err != nil {
		return nil, err
	}
	body, err := c.doRequestWithAuth(req, http.StatusOK)
	if err != nil {
		return nil, err
	}

	var action BaseEventAction
	err = json.Unmarshal(body, &action)
	return &action, err
}

// UpdateAction - Updates an existing action
func (c *Client) UpdateAction(action BaseEventAction) error {
	rb, err := json.Marshal(action)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/api/v2/eventactions/%s", c.HostURL, url.PathEscape(action.Name)),
		bytes.NewBuffer(rb))
	if err != nil {
		return err
	}

	_, err = c.doRequestWithAuth(req, http.StatusOK)
	return err
}

// DeleteAction - Deletes a action
func (c *Client) DeleteAction(name string) error {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/v2/eventactions/%s", c.HostURL, url.PathEscape(name)), nil)
	if err != nil {
		return err
	}
	_, err = c.doRequestWithAuth(req, http.StatusOK)
	return err
}
