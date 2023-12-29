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
	"fmt"
	"net"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type sftpEndPointValidator struct{}

// Description describes the validation in plain text formatting.
func (sftpEndPointValidator) Description(_ context.Context) string {
	return "string must be in the format host:port"
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v sftpEndPointValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// Validate performs the validation.
func (v sftpEndPointValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	value := request.ConfigValue.ValueString()

	_, _, err := net.SplitHostPort(value)
	if err != nil {
		response.Diagnostics.Append(invalidAttributeSFTPEndPointDiagnostic(
			request.Path,
			v.Description(ctx),
			fmt.Sprintf("%v", err),
		))
		return
	}
}

func invalidAttributeSFTPEndPointDiagnostic(path path.Path, description string, value string) diag.Diagnostic {
	return diag.NewAttributeErrorDiagnostic(
		path,
		"Invalid Attribute Value SFTP Endpoint",
		fmt.Sprintf("Attribute %s %s, got: %s", path, description, value),
	)
}
