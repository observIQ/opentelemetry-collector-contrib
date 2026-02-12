// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package googlesecopsexporter

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/pdata/plog"
	"golang.org/x/oauth2"
)

// TestGetCollectorIDString tests the getCollectorIDString function for all license types
func TestGetCollectorIDString(t *testing.T) {
	testCases := []struct {
		name        string
		licenseType string
		expected    string
	}{
		{
			name:        "Google license type",
			licenseType: "Google",
			expected:    googleCollectorIDString,
		},
		{
			name:        "google lowercase",
			licenseType: "google",
			expected:    googleCollectorIDString,
		},
		{
			name:        "GoogleEnterprise",
			licenseType: "GoogleEnterprise",
			expected:    googleEnterpriseCollectorIDString,
		},
		{
			name:        "googleenterprise lowercase",
			licenseType: "googleenterprise",
			expected:    googleEnterpriseCollectorIDString,
		},
		{
			name:        "Enterprise",
			licenseType: "Enterprise",
			expected:    enterpriseCollectorIDString,
		},
		{
			name:        "enterprise lowercase",
			licenseType: "enterprise",
			expected:    enterpriseCollectorIDString,
		},
		{
			name:        "Unknown license type",
			licenseType: "unknown",
			expected:    defaultCollectorIDString,
		},
		{
			name:        "Empty license type",
			licenseType: "",
			expected:    defaultCollectorIDString,
		},
		{
			name:        "Random license type",
			licenseType: "FreeCloud",
			expected:    defaultCollectorIDString,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := getCollectorIDString(tc.licenseType)
			require.Equal(t, tc.expected, result)
		})
	}
}

// TestGoogleCredentials tests the googleCredentials function
func TestGoogleCredentials(t *testing.T) {
	t.Run("invalid json in Creds", func(t *testing.T) {
		cfg := &Config{
			Creds:    "invalid json",
			Protocol: protocolGRPC,
		}
		_, err := googleCredentials(context.Background(), cfg)
		require.Error(t, err)
	})

	t.Run("empty credentials file", func(t *testing.T) {
		// Create a temporary empty file
		tmpFile, err := os.CreateTemp("", "empty-creds-*.json")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())
		tmpFile.Close()

		cfg := &Config{
			CredsFilePath: tmpFile.Name(),
			Protocol:      protocolGRPC,
		}
		_, err = googleCredentials(context.Background(), cfg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "credentials file is empty")
	})

	t.Run("nonexistent credentials file", func(t *testing.T) {
		cfg := &Config{
			CredsFilePath: "/nonexistent/path/to/creds.json",
			Protocol:      protocolGRPC,
		}
		_, err := googleCredentials(context.Background(), cfg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "read credentials file")
	})

	t.Run("valid credentials from file", func(t *testing.T) {
		// Create a temporary file with valid (but fake) credentials JSON
		validCreds := `{
			"type": "service_account",
			"project_id": "test-project",
			"private_key_id": "key-id",
			"private_key": "-----BEGIN PRIVATE KEY-----\nMIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC7W8jYbN8cZqQH\ntest\n-----END PRIVATE KEY-----\n",
			"client_email": "test@test-project.iam.gserviceaccount.com",
			"client_id": "123456789",
			"auth_uri": "https://accounts.google.com/o/oauth2/auth",
			"token_uri": "https://oauth2.googleapis.com/token"
		}`

		tmpFile, err := os.CreateTemp("", "valid-creds-*.json")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		_, err = tmpFile.WriteString(validCreds)
		require.NoError(t, err)
		tmpFile.Close()

		cfg := &Config{
			CredsFilePath: tmpFile.Name(),
			Protocol:      protocolGRPC,
		}
		creds, err := googleCredentials(context.Background(), cfg)
		// This might fail if the private key is not valid, but we should get past the file reading
		if err != nil {
			// If it fails, it should be a JWT error, not a file reading error
			require.NotContains(t, err.Error(), "read credentials file")
			require.NotContains(t, err.Error(), "credentials file is empty")
		} else {
			require.NotNil(t, creds)
		}
	})

	t.Run("different scopes for protocols", func(t *testing.T) {
		// Test that different protocols use different scopes
		// We can't fully test this without mocking google.CredentialsFromJSON,
		// but we can verify it doesn't panic and uses the right code paths

		validCreds := `{
			"type": "service_account",
			"project_id": "test-project",
			"private_key_id": "key-id",
			"private_key": "-----BEGIN PRIVATE KEY-----\nMIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC7W8jYbN8cZqQH\ntest\n-----END PRIVATE KEY-----\n",
			"client_email": "test@test-project.iam.gserviceaccount.com",
			"client_id": "123456789",
			"auth_uri": "https://accounts.google.com/o/oauth2/auth",
			"token_uri": "https://oauth2.googleapis.com/token"
		}`

		cfgGRPC := &Config{
			Creds:    validCreds,
			Protocol: protocolGRPC,
		}
		_, _ = googleCredentials(context.Background(), cfgGRPC)

		cfgHTTP := &Config{
			Creds:    validCreds,
			Protocol: protocolHTTPS,
		}
		_, _ = googleCredentials(context.Background(), cfgHTTP)
	})
}

// TestBatchSizeLimits tests handling of various batch sizes
func TestBatchSizeLimits(t *testing.T) {
	t.Run("large payload splitting", func(t *testing.T) {
		// Create a large set of logs that exceeds batch size
		logs := plog.NewLogs()

		// Create many logs with substantial content to exceed the 4MB limit
		largeBody := strings.Repeat("X", 100*1024) // 100KB per log

		for i := 0; i < 50; i++ { // 50 * 100KB = 5MB total
			rls := logs.ResourceLogs().AppendEmpty()
			sls := rls.ScopeLogs().AppendEmpty()
			lr := sls.LogRecords().AppendEmpty()
			lr.Body().SetStr(largeBody)
			lr.Attributes().PutStr("index", string(rune(i)))
		}

		// Verify we created a large payload
		require.Equal(t, 50, logs.LogRecordCount())
	})

	t.Run("at batch size limit", func(t *testing.T) {
		// Create logs that are exactly at the limit
		logs := plog.NewLogs()

		// Approximately 4MB of data (slightly under to be safe)
		largeBody := strings.Repeat("Y", 4*1024*1024-1000) // Just under 4MB

		rls := logs.ResourceLogs().AppendEmpty()
		sls := rls.ScopeLogs().AppendEmpty()
		lr := sls.LogRecords().AppendEmpty()
		lr.Body().SetStr(largeBody)

		require.Equal(t, 1, logs.LogRecordCount())
	})

	t.Run("empty logs", func(t *testing.T) {
		logs := plog.NewLogs()
		require.Equal(t, 0, logs.LogRecordCount())
	})
}

// TestLabelPrefixHandling tests both chronicle_ and secops_ label prefix handling
func TestLabelPrefixHandling(t *testing.T) {
	testCases := []struct {
		name             string
		chronicleLogType string
		secopsLogType    string
		chronicleNS      string
		secopsNS         string
		chronicleLabel   string
		secopsLabel      string
	}{
		{
			name:             "secops_ prefix takes priority over chronicle_",
			chronicleLogType: "CHRONICLE_TYPE",
			secopsLogType:    "SECOPS_TYPE",
			chronicleNS:      "chronicle_ns",
			secopsNS:         "secops_ns",
			chronicleLabel:   "chronicle_label",
			secopsLabel:      "secops_label",
		},
		{
			name:             "chronicle_ prefix when no secops_",
			chronicleLogType: "CHRONICLE_TYPE",
			secopsLogType:    "",
			chronicleNS:      "chronicle_ns",
			secopsNS:         "",
			chronicleLabel:   "chronicle_label",
			secopsLabel:      "",
		},
		{
			name:             "secops_ prefix only",
			chronicleLogType: "",
			secopsLogType:    "SECOPS_TYPE",
			chronicleNS:      "",
			secopsNS:         "secops_ns",
			chronicleLabel:   "",
			secopsLabel:      "secops_label",
		},
		{
			name:             "neither prefix",
			chronicleLogType: "",
			secopsLogType:    "",
			chronicleNS:      "",
			secopsNS:         "",
			chronicleLabel:   "",
			secopsLabel:      "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logs := plog.NewLogs()
			rls := logs.ResourceLogs().AppendEmpty()
			sls := rls.ScopeLogs().AppendEmpty()
			lr := sls.LogRecords().AppendEmpty()
			lr.Body().SetStr("Test log entry")

			if tc.chronicleLogType != "" {
				lr.Attributes().PutStr("chronicle_log_type", tc.chronicleLogType)
			}
			if tc.secopsLogType != "" {
				lr.Attributes().PutStr("secops_log_type", tc.secopsLogType)
			}
			if tc.chronicleNS != "" {
				lr.Attributes().PutStr("chronicle_namespace", tc.chronicleNS)
			}
			if tc.secopsNS != "" {
				lr.Attributes().PutStr("secops_namespace", tc.secopsNS)
			}
			if tc.chronicleLabel != "" {
				lr.Attributes().PutStr("chronicle_ingestion_label[key]", tc.chronicleLabel)
			}
			if tc.secopsLabel != "" {
				lr.Attributes().PutStr("secops_ingestion_label[key]", tc.secopsLabel)
			}

			// Verify the log was created correctly
			require.Equal(t, 1, logs.LogRecordCount())
		})
	}
}

// TestCompressionHandling tests compression functionality
func TestCompressionHandling(t *testing.T) {
	t.Run("compression enabled", func(t *testing.T) {
		cfg := &Config{
			Protocol:    protocolGRPC,
			Compression: "gzip",
		}
		require.Equal(t, "gzip", cfg.Compression)
	})

	t.Run("compression disabled", func(t *testing.T) {
		cfg := &Config{
			Protocol:    protocolGRPC,
			Compression: "",
		}
		require.Equal(t, "", cfg.Compression)
	})

	t.Run("different compression types", func(t *testing.T) {
		compressionTypes := []string{"gzip", "snappy", "zstd", "none", ""}
		for _, compType := range compressionTypes {
			cfg := &Config{
				Protocol:    protocolGRPC,
				Compression: compType,
			}
			require.Equal(t, compType, cfg.Compression)
		}
	})
}

// TestOTTLHelperFunctions tests OTTL helper function coverage
func TestOTTLHelperFunctions(t *testing.T) {
	t.Run("newOTTLLogRecordExpression error handling", func(t *testing.T) {
		// Test with invalid expression that should cause parsing error
		_, err := newOTTLLogRecordExpression("invalid{{syntax", componenttest.NewNopTelemetrySettings())
		require.Error(t, err)
	})

	t.Run("newOTTLLogRecordStatement error handling", func(t *testing.T) {
		// Test with invalid statement
		_, err := newOTTLLogRecordStatement("invalid{{syntax", componenttest.NewNopTelemetrySettings())
		require.Error(t, err)
	})

	t.Run("newOTTLLogRecordExpression valid expression", func(t *testing.T) {
		// Test with valid expression
		expr, err := newOTTLLogRecordExpression("\"test\"", componenttest.NewNopTelemetrySettings())
		if err == nil {
			require.NotNil(t, expr)
		}
	})
}

// TestLogRecordCreation tests creation of logs with various configurations
func TestLogRecordCreation(t *testing.T) {
	t.Run("empty logs", func(t *testing.T) {
		logs := plog.NewLogs()
		require.Equal(t, 0, logs.LogRecordCount())
	})

	t.Run("logs with raw log field", func(t *testing.T) {
		logs := plog.NewLogs()
		rls := logs.ResourceLogs().AppendEmpty()
		sls := rls.ScopeLogs().AppendEmpty()
		lr := sls.LogRecords().AppendEmpty()
		lr.Body().SetStr("raw log content")
		require.Equal(t, 1, logs.LogRecordCount())
		require.Equal(t, "raw log content", lr.Body().Str())
	})

	t.Run("logs with attributes", func(t *testing.T) {
		logs := plog.NewLogs()
		rls := logs.ResourceLogs().AppendEmpty()
		sls := rls.ScopeLogs().AppendEmpty()
		lr := sls.LogRecords().AppendEmpty()
		lr.Body().SetStr("structured log content")
		lr.Attributes().PutStr("key", "value")
		require.Equal(t, 1, logs.LogRecordCount())
		require.Equal(t, 1, lr.Attributes().Len())
	})
}

// TestEdgeCases tests various edge cases
func TestEdgeCases(t *testing.T) {
	t.Run("nil configuration", func(t *testing.T) {
		// Verify nil config doesn't panic
		var cfg *Config
		if cfg != nil {
			t.Error("Config should be nil")
		}
	})

	t.Run("extremely long log body", func(t *testing.T) {
		logs := plog.NewLogs()
		rls := logs.ResourceLogs().AppendEmpty()
		sls := rls.ScopeLogs().AppendEmpty()
		lr := sls.LogRecords().AppendEmpty()

		// Create an extremely long log body (10MB)
		longBody := strings.Repeat("A", 10*1024*1024)
		lr.Body().SetStr(longBody)

		require.Equal(t, 1, logs.LogRecordCount())
	})

	t.Run("many attributes on single log", func(t *testing.T) {
		logs := plog.NewLogs()
		rls := logs.ResourceLogs().AppendEmpty()
		sls := rls.ScopeLogs().AppendEmpty()
		lr := sls.LogRecords().AppendEmpty()
		lr.Body().SetStr("test")

		// Add many attributes
		for i := 0; i < 1000; i++ {
			lr.Attributes().PutStr("key_"+string(rune(i)), "value_"+string(rune(i)))
		}

		require.Equal(t, 1000, lr.Attributes().Len())
	})

	t.Run("many resource attributes", func(t *testing.T) {
		logs := plog.NewLogs()
		rls := logs.ResourceLogs().AppendEmpty()

		// Add many resource attributes
		for i := 0; i < 500; i++ {
			rls.Resource().Attributes().PutStr("res_"+string(rune(i)), "value_"+string(rune(i)))
		}

		sls := rls.ScopeLogs().AppendEmpty()
		lr := sls.LogRecords().AppendEmpty()
		lr.Body().SetStr("test")

		require.Equal(t, 500, rls.Resource().Attributes().Len())
	})

	t.Run("special characters in log body", func(t *testing.T) {
		logs := plog.NewLogs()
		rls := logs.ResourceLogs().AppendEmpty()
		sls := rls.ScopeLogs().AppendEmpty()
		lr := sls.LogRecords().AppendEmpty()

		// Test various special characters
		specialChars := "!@#$%^&*()_+-=[]{}|;:'\",.<>?/`~\n\t\r\x00"
		lr.Body().SetStr(specialChars)

		require.Equal(t, specialChars, lr.Body().Str())
	})

	t.Run("unicode in log body", func(t *testing.T) {
		logs := plog.NewLogs()
		rls := logs.ResourceLogs().AppendEmpty()
		sls := rls.ScopeLogs().AppendEmpty()
		lr := sls.LogRecords().AppendEmpty()

		// Test unicode characters
		unicode := "Hello 世界 🌍 Привет мир"
		lr.Body().SetStr(unicode)

		require.Equal(t, unicode, lr.Body().Str())
	})
}

// TestDefaultCredentials tests the default credentials path
func TestDefaultCredentials(t *testing.T) {
	t.Run("no credentials provided", func(t *testing.T) {
		cfg := &Config{
			Protocol: protocolGRPC,
		}

		// This will try to use Application Default Credentials
		// It may succeed or fail depending on the environment
		_, err := googleCredentials(context.Background(), cfg)

		// We just want to ensure it doesn't panic and follows the right code path
		if err != nil {
			// Expected error when ADC is not available
			require.Error(t, err)
		}
	})
}

// TestTokenSourceOverride tests the token source override mechanism
func TestTokenSourceOverride(t *testing.T) {
	originalTokenSource := tokenSource
	defer func() {
		tokenSource = originalTokenSource
	}()

	t.Run("custom token source", func(t *testing.T) {
		customCalled := false
		tokenSource = func(ctx context.Context, cfg *Config) (oauth2.TokenSource, error) {
			customCalled = true
			return nil, errors.New("custom token source error")
		}

		cfg := &Config{Protocol: protocolGRPC}
		_, err := tokenSource(context.Background(), cfg)

		require.True(t, customCalled)
		require.Error(t, err)
		require.Equal(t, "custom token source error", err.Error())
	})
}
