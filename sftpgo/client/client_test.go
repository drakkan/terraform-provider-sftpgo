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
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "StatusError with 404",
			err:      &StatusError{StatusCode: http.StatusNotFound},
			expected: true,
		},
		{
			name:     "StatusError with 500",
			err:      &StatusError{StatusCode: http.StatusInternalServerError},
			expected: false,
		},
		{
			name:     "Non-StatusError",
			err:      fmt.Errorf("some other error"),
			expected: false,
		},
		{
			name:     "Nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "StatusError as value (not pointer)",
			err:      StatusError{StatusCode: http.StatusNotFound},
			expected: false, // IsNotFound only matches *StatusError
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, IsNotFound(tt.err))
		})
	}
}
