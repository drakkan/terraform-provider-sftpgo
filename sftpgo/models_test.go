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

package sftpgo

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sftpgo/sdk/kms"
	"github.com/stretchr/testify/require"

	"github.com/drakkan/terraform-provider-sftpgo/sftpgo/client"
)

func TestSecretsConversion(t *testing.T) {
	secret := kms.BaseSecret{
		Status:         kms.SecretStatusSecretBox,
		Payload:        "$secret$$payload$",
		Key:            "secret key",
		AdditionalData: "da$ta",
	}
	secretString := getSecretFromSFTPGo(secret)
	require.Equal(t, "$Secretbox$secret key$5$da$ta$secret$$payload$", secretString)
	secretFromString := getSFTPGoSecret(secretString)
	require.Equal(t, secret, secretFromString)
}

func TestRoleSettingsConversion(t *testing.T) {
	data, err := json.Marshal(client.Role{Name: "role"})
	require.NoError(t, err)
	require.NotContains(t, string(data), "settings")

	role := client.Role{
		Name: "role",
		Settings: client.RoleSettings{
			StorageAllowlist: client.RoleStorageAllowlist{
				AllowedProviders: []int{0},
				Local: client.LocalRoleScope{
					AllowedPaths: []string{"/tmp"},
				},
			},
		},
	}
	data, err = json.Marshal(role)
	require.NoError(t, err)
	require.Contains(t, string(data),
		`"settings":{"storage_allowlist":{"allowed_providers":[0],"local":{"allowed_paths":["/tmp"]}}}`)

	// an allowlist granting nothing reads like a role carrying no settings: the
	// WebAdmin stores that envelope whenever a role is saved from its page
	var stored client.Role
	require.NoError(t, json.Unmarshal([]byte(`{"name":"role","settings":{"storage_allowlist":{}}}`), &stored))
	var model roleResourceModel
	diags := model.fromSFTPGo(context.Background(), &stored)
	require.False(t, diags.HasError())
	require.Nil(t, model.Settings)
}

func TestRoleStorageAllowlistIsEmpty(t *testing.T) {
	var allowlist roleStorageAllowlist
	require.True(t, allowlist.isEmpty())

	providers, diags := types.ListValueFrom(context.Background(), types.Int64Type, []int{0})
	require.False(t, diags.HasError())

	// an allowlist granting nothing reads as empty whether its lists are absent
	// or present and empty
	noProviders, diags := types.ListValueFrom(context.Background(), types.Int64Type, []int{})
	require.False(t, diags.HasError())
	require.True(t, (&roleStorageAllowlist{AllowedProviders: noProviders}).isEmpty())

	// isEmpty decides whether a role carries settings, so it must name every
	// field: one left out makes an allowlist holding only that field read as absent
	value := reflect.ValueOf(&allowlist).Elem()
	for idx := range value.NumField() {
		allowlist = roleStorageAllowlist{}
		field := value.Field(idx)
		name := value.Type().Field(idx).Name
		switch {
		case field.Kind() == reflect.Pointer:
			field.Set(reflect.New(field.Type().Elem()))
		case field.Type() == reflect.TypeOf(types.List{}):
			field.Set(reflect.ValueOf(providers))
		default:
			t.Fatalf("unhandled type %s for field %q, extend this test", field.Type(), name)
		}
		require.Falsef(t, allowlist.isEmpty(), "field %q is not checked by isEmpty", name)
	}
}

func TestRoleScopesIsEmpty(t *testing.T) {
	// every field must be named in the predicate: one left out makes a scope
	// carrying only that field read as absent
	for _, scope := range []any{
		&client.LocalRoleScope{},
		&client.S3RoleScope{},
		&client.AzureRoleScope{},
		&client.GCSRoleScope{},
		&client.EndpointRoleScope{},
		&client.SFTPRoleScope{},
	} {
		value := reflect.ValueOf(scope).Elem()
		name := value.Type().Name()
		isEmpty := func() bool {
			return value.Interface().(interface{ IsEmpty() bool }).IsEmpty()
		}
		require.Truef(t, isEmpty(), "the zero %s must be empty", name)

		for idx := range value.NumField() {
			value.Set(reflect.Zero(value.Type()))
			field := value.Field(idx)
			setNonZero(t, field)
			require.Falsef(t, isEmpty(), "field %q of %s is not checked by IsEmpty",
				value.Type().Field(idx).Name, name)
		}
	}
}

func setNonZero(t *testing.T, field reflect.Value) {
	t.Helper()

	switch field.Kind() {
	case reflect.Bool:
		field.SetBool(true)
	case reflect.String:
		field.SetString("value")
	case reflect.Slice:
		field.Set(reflect.MakeSlice(field.Type(), 1, 1))
	case reflect.Struct:
		setNonZero(t, field.Field(0))
	default:
		t.Fatalf("unhandled kind %s, extend this test", field.Kind())
	}
}

func TestSchemaParity(t *testing.T) {
	// the resource and the data source schema are written twice, an attribute
	// missing from one is a runtime conversion error nothing else catches
	for name, schemas := range map[string][2]any{
		"role settings":    {getSchemaForRoleSettings(), getComputedSchemaForRoleSettings()},
		"license features": {getSchemaForLicenseFeatures(), getComputedSchemaForLicenseFeatures()},
	} {
		resourceAttrs := map[string]struct{}{}
		collectSchemaAttributes(reflect.ValueOf(schemas[0]), "", resourceAttrs)
		dataSourceAttrs := map[string]struct{}{}
		collectSchemaAttributes(reflect.ValueOf(schemas[1]), "", dataSourceAttrs)

		require.NotEmptyf(t, resourceAttrs, "no attribute collected for %s", name)
		require.Equalf(t, resourceAttrs, dataSourceAttrs, "the %s schemas differ", name)
	}
}

func collectSchemaAttributes(value reflect.Value, prefix string, out map[string]struct{}) {
	for value.Kind() == reflect.Interface || value.Kind() == reflect.Pointer {
		value = value.Elem()
	}
	if value.Kind() != reflect.Struct {
		return
	}
	if attributes := value.FieldByName("Attributes"); attributes.IsValid() && attributes.Kind() == reflect.Map {
		for _, key := range attributes.MapKeys() {
			name := prefix + key.String()
			out[name] = struct{}{}
			collectSchemaAttributes(attributes.MapIndex(key), name+".", out)
		}
	}
	if nested := value.FieldByName("NestedObject"); nested.IsValid() && nested.Kind() == reflect.Struct {
		collectSchemaAttributes(nested, prefix, out)
	}
}
