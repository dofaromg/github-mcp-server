package utils

import (
	"errors"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewToolResultJSON(t *testing.T) {
	tests := []struct {
		name        string
		input       any
		expectError bool
		expected    string
	}{
		{
			name: "simple map",
			input: map[string]string{
				"key": "value",
			},
			expectError: false,
			expected:    `{"key":"value"}`,
		},
		{
			name: "struct",
			input: struct {
				Name  string `json:"name"`
				Count int    `json:"count"`
			}{
				Name:  "test",
				Count: 42,
			},
			expectError: false,
			expected:    `{"name":"test","count":42}`,
		},
		{
			name:        "simple string",
			input:       "hello",
			expectError: false,
			expected:    `"hello"`,
		},
		{
			name:        "nil value",
			input:       nil,
			expectError: false,
			expected:    "null",
		},
		{
			name: "unmarshalable value",
			input: map[string]any{
				"invalid": make(chan int),
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := NewToolResultJSON(tc.input)

			if tc.expectError {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.False(t, result.IsError)
				require.Len(t, result.Content, 1)
				textContent := result.Content[0].(*mcp.TextContent)
				assert.Equal(t, tc.expected, textContent.Text)
			}
		})
	}
}

func TestNewToolResultText(t *testing.T) {
	result := NewToolResultText("test message")
	require.NotNil(t, result)
	assert.False(t, result.IsError)
	require.Len(t, result.Content, 1)
	textContent := result.Content[0].(*mcp.TextContent)
	assert.Equal(t, "test message", textContent.Text)
}

func TestNewToolResultError(t *testing.T) {
	result := NewToolResultError("error message")
	require.NotNil(t, result)
	assert.True(t, result.IsError)
	require.Len(t, result.Content, 1)
	textContent := result.Content[0].(*mcp.TextContent)
	assert.Equal(t, "error message", textContent.Text)
}

func TestNewToolResultErrorFromErr(t *testing.T) {
	err := errors.New("original error")
	result := NewToolResultErrorFromErr("prefix", err)
	require.NotNil(t, result)
	assert.True(t, result.IsError)
	require.Len(t, result.Content, 1)
	textContent := result.Content[0].(*mcp.TextContent)
	assert.Equal(t, "prefix: original error", textContent.Text)
}
