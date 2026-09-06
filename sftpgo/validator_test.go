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
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestSFTPEndPointValidator(t *testing.T) {
	type testCase struct {
		val         types.String
		expectError bool
	}
	tests := map[string]testCase{
		"unknown": {
			val:         types.StringUnknown(),
			expectError: false,
		},
		"null": {
			val:         types.StringNull(),
			expectError: false,
		},
		"valid": {
			val:         types.StringValue("127.0.0.1:22"),
			expectError: false,
		},
		"missing port": {
			val:         types.StringValue("127.0.0.1"),
			expectError: true,
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			request := validator.StringRequest{
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				ConfigValue:    test.val,
			}
			response := validator.StringResponse{}
			v := sftpEndPointValidator{}
			v.ValidateString(context.TODO(), request, &response)

			if !response.Diagnostics.HasError() && test.expectError {
				t.Fatal("expected error, got no error")
			}

			if response.Diagnostics.HasError() && !test.expectError {
				t.Fatalf("got unexpected error: %s", response.Diagnostics)
			}
		})
	}
}

func TestNonEmptyObjectValidator(t *testing.T) {
	attrTypes := map[string]attr.Type{
		"allowed_endpoints": types.ListType{ElemType: types.StringType},
		"allowed_proxies":   types.ListType{ElemType: types.StringType},
	}
	empty, diags := types.ObjectValue(attrTypes, map[string]attr.Value{
		"allowed_endpoints": types.ListNull(types.StringType),
		"allowed_proxies":   types.ListNull(types.StringType),
	})
	if diags.HasError() {
		t.Fatalf("unable to build the empty object: %s", diags)
	}
	set, diags := types.ObjectValue(attrTypes, map[string]attr.Value{
		"allowed_endpoints": types.ListValueMust(types.StringType, []attr.Value{types.StringValue("host:22")}),
		"allowed_proxies":   types.ListNull(types.StringType),
	})
	if diags.HasError() {
		t.Fatalf("unable to build the object: %s", diags)
	}

	type testCase struct {
		val         types.Object
		expectError bool
	}
	tests := map[string]testCase{
		"unknown": {
			val:         types.ObjectUnknown(attrTypes),
			expectError: false,
		},
		"null": {
			val:         types.ObjectNull(attrTypes),
			expectError: false,
		},
		"every attribute null": {
			val:         empty,
			expectError: true,
		},
		"one attribute set to its zero value": {
			val:         set,
			expectError: false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			request := validator.ObjectRequest{
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				ConfigValue:    test.val,
			}
			response := validator.ObjectResponse{}
			v := nonEmptyObjectValidator{}
			v.ValidateObject(context.TODO(), request, &response)

			if !response.Diagnostics.HasError() && test.expectError {
				t.Fatal("expected error, got no error")
			}

			if response.Diagnostics.HasError() && !test.expectError {
				t.Fatalf("got unexpected error: %s", response.Diagnostics)
			}
		})
	}
}
