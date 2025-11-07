package quickenv

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseLine(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantKey string
		wantVal string
		wantErr bool
	}{
		{
			name:    "simple key=value",
			input:   "DB_PORT=8080",
			wantKey: "DB_PORT",
			wantVal: "8080",
			wantErr: false,
		},
		{
			name:    "key with export",
			input:   "export API_KEY=abc123",
			wantKey: "API_KEY",
			wantVal: "abc123",
			wantErr: false,
		},
		{
			name:    "value in double quotes",
			input:   `NAME="Alex Edwards"`,
			wantKey: "NAME",
			wantVal: "Alex Edwards",
			wantErr: false,
		},
		{
			name:    "value in single quotes",
			input:   `CITY='New York'`,
			wantKey: "CITY",
			wantVal: "New York",
			wantErr: false,
		},
		{
			name:    "empty value with =",
			input:   "DB_PORT=",
			wantKey: "DB_PORT",
			wantVal: "",
			wantErr: false,
		},
		{
			name:    "empty value in quotes",
			input:   `EMPTY=""`,
			wantKey: "EMPTY",
			wantVal: "",
			wantErr: false,
		},
		{
			name:    "spaces around =",
			input:   " KEY = value ",
			wantKey: "KEY",
			wantVal: "value",
			wantErr: false,
		},
		{
			name:    "quoted value with spaces",
			input:   `MESSAGE=" hello world "`,
			wantKey: "MESSAGE",
			wantVal: " hello world ",
			wantErr: false,
		},
		{
			name:    "equals inside quotes",
			input:   `CONN_STR="user=pass@host:5432"`,
			wantKey: "CONN_STR",
			wantVal: "user=pass@host:5432",
			wantErr: false,
		},
		{
			name:    "invalid line no equals",
			input:   "justvalue",
			wantErr: true,
		},
		{
			name:    "empty key",
			input:   "=value",
			wantErr: true,
		},
		{
			name:    "comment line",
			input:   "# comment",
			wantErr: true,
		},
		{
			name:    "export with spaces and quotes",
			input:   `export NAME="John Doe"`,
			wantKey: "NAME",
			wantVal: "John Doe",
			wantErr: false,
		},
		{
			name:    "multiple equals",
			input:   "NAME==John=Doe",
			wantKey: "NAME",
			wantVal: "=John=Doe",
			wantErr: false,
		},
		{
			name:    "key in single quotes containing = should be invalid",
			input:   "'NAME='=John=Doe",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, val, err := parseLine(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantKey, key)
				assert.Equal(t, tt.wantVal, val)
			}
		})
	}
}

func TestUnquoteValue(t *testing.T) {
	tests := []struct {
		input  string
		output string
	}{
		{input: `"hello"`, output: "hello"},
		{input: `'world'`, output: "world"},
		{input: `" hello "`, output: " hello "},
		{input: `""`, output: ""},
		{input: `''`, output: ""},
		{input: `"mixed'`, output: `"mixed'`},
		{input: `noquotes`, output: "noquotes"},
		{input: `"with \" escaped"`, output: `with \" escaped`},
		{input: "", output: ""},
		{input: `"a"`, output: "a"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := unquoteValue(tt.input)
			assert.Equal(t, tt.output, got)
		})
	}
}

func TestIsValidEnvKey(t *testing.T) {
	valid := []string{"PORT", "API_KEY", "DEBUG", "A", "_INTERNAL", "Var1"}
	invalid := []string{"", "123", "-", "my-var", "кошка", " ", "a b", ".hidden", ""}

	for _, key := range valid {
		t.Run("valid_"+key, func(t *testing.T) {
			assert.True(t, isValidEnvKey(key))
		})
	}

	for _, key := range invalid {
		t.Run("invalid_"+key, func(t *testing.T) {
			assert.False(t, isValidEnvKey(key))
		})
	}
}
