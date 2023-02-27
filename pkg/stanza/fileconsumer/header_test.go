package fileconsumer

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/helper"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/input/generate"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/parser/regex"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"golang.org/x/text/encoding"
	"golang.org/x/text/transform"
)

func TestHeaderConfig_validate(t *testing.T) {
	regexConf := regex.NewConfig()
	regexConf.Regex = "^#(?P<header_line>.*)"

	invalidRegexConf := regex.NewConfig()
	invalidRegexConf.Regex = "("

	generateConf := generate.NewConfig("")

	defaultMaxHeaderByteSize := helper.ByteSize(defaultMaxHeaderSize)
	negativeMaxHeaderByteSize := helper.ByteSize(-1)

	testCases := []struct {
		name        string
		conf        HeaderConfig
		expectedErr string
	}{
		{
			name: "Valid config",
			conf: HeaderConfig{
				MultilinePattern: "^#",
				MetadataOperators: []operator.Config{
					{
						Builder: regexConf,
					},
				},
				MaxHeaderSize: &defaultMaxHeaderByteSize,
			},
		},
		{
			name: "Valid without specified header size",
			conf: HeaderConfig{
				MultilinePattern: "^#",
				MetadataOperators: []operator.Config{
					{
						Builder: regexConf,
					},
				},
			},
		},
		{
			name: "Invalid pattern",
			conf: HeaderConfig{
				MultilinePattern: "(",
				MetadataOperators: []operator.Config{
					{
						Builder: regexConf,
					},
				},
				MaxHeaderSize: &defaultMaxHeaderByteSize,
			},
			expectedErr: "failed to compile multiline pattern:",
		},
		{
			name: "Negative max header size",
			conf: HeaderConfig{
				MultilinePattern: "^#",
				MetadataOperators: []operator.Config{
					{
						Builder: regexConf,
					},
				},
				MaxHeaderSize: &negativeMaxHeaderByteSize,
			},
			expectedErr: "the maximum size of the header must be greater than 0",
		},
		{
			name: "No operators specified",
			conf: HeaderConfig{
				MultilinePattern:  "^#",
				MetadataOperators: []operator.Config{},
				MaxHeaderSize:     &defaultMaxHeaderByteSize,
			},
			expectedErr: "at least one operator must be specified for `metadata_operators`",
		},
		{
			name: "Invalid operator specified",
			conf: HeaderConfig{
				MultilinePattern: "^#",
				MetadataOperators: []operator.Config{
					{
						Builder: invalidRegexConf,
					},
				},
				MaxHeaderSize: &defaultMaxHeaderByteSize,
			},
			expectedErr: "failed to build pipelines:",
		},
		{
			name: "first operator cannot process",
			conf: HeaderConfig{
				MultilinePattern: "^#",
				MetadataOperators: []operator.Config{
					{
						Builder: generateConf,
					},
				},
				MaxHeaderSize: &defaultMaxHeaderByteSize,
			},
			expectedErr: "first operator must be able to process entries",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.conf.validate()
			if tc.expectedErr != "" {
				require.ErrorContains(t, err, tc.expectedErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestHeaderConfig_buildHeader(t *testing.T) {
	regexConf := regex.NewConfig()
	regexConf.Regex = "^#(?P<header_line>.*)"

	invalidRegexConf := regex.NewConfig()
	invalidRegexConf.Regex = "("

	badEncoding := mockEncoding{
		encodingTransformer: errorTransformer{
			e: errors.New("cannot encode bytes"),
		},
	}

	testCases := []struct {
		name        string
		enc         encoding.Encoding
		conf        HeaderConfig
		expectedErr string
	}{
		{
			name: "valid config",
			enc:  encoding.Nop,
			conf: HeaderConfig{
				MultilinePattern: "^#",
				MetadataOperators: []operator.Config{
					{
						Builder: regexConf,
					},
				},
			},
		},
		{
			name: "invalid regex",
			enc:  encoding.Nop,
			conf: HeaderConfig{
				MultilinePattern: "^(",
				MetadataOperators: []operator.Config{
					{
						Builder: regexConf,
					},
				},
			},
			expectedErr: "failed to compile multiline pattern:",
		},
		{
			name: "invalid operator",
			enc:  encoding.Nop,
			conf: HeaderConfig{
				MultilinePattern: "^#",
				MetadataOperators: []operator.Config{
					{
						Builder: invalidRegexConf,
					},
				},
			},
			expectedErr: "failed to build pipeline:",
		},
		{
			name: "invalid encoding",
			enc:  badEncoding,
			conf: HeaderConfig{
				MultilinePattern: "^#",
				MetadataOperators: []operator.Config{
					{
						Builder: regexConf,
					},
				},
			},
			expectedErr: "failed to create split func",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h, err := tc.conf.buildHeader(tc.enc, zaptest.NewLogger(t).Sugar())
			if tc.expectedErr != "" {
				require.ErrorContains(t, err, tc.expectedErr)
			} else {
				require.NoError(t, err)
				require.NotNil(t, h)
			}

		})
	}
}

func TestHeaderConfig_ReadHeader(t *testing.T) {
	basicRegexConfig := regex.NewConfig()
	basicRegexConfig.Regex = "^#(?P<field_name>[A-z0-9]*): (?P<value>[A-z0-9]*)"

	fullCaptureRegexConfig := regex.NewConfig()
	fullCaptureRegexConfig.Regex = `^(?P<header>[\s\S]*)$`

	generateConf := generate.NewConfig("")

	smallByteSize := helper.ByteSize(28)

	testCases := []struct {
		name               string
		fileContents       string
		expectedAttributes map[string]any
		conf               HeaderConfig
	}{
		{
			name:         "Header + log line",
			fileContents: "#aField: SomeValue\nThis is a non-header line\n",
			expectedAttributes: map[string]any{
				"field_name": "aField",
				"value":      "SomeValue",
			},
			conf: HeaderConfig{
				MultilinePattern: "^#",
				MetadataOperators: []operator.Config{
					{
						Builder: basicRegexConfig,
					},
				},
			},
		},
		{
			name:         "Header truncates when too long",
			fileContents: "#aField: SomeValue\n#aField2: SomeValue2\nThis is a non-header line\n",
			expectedAttributes: map[string]any{
				"header": "#aField: SomeValue\n#aField2:",
			},
			conf: HeaderConfig{
				MultilinePattern: "^#",
				MetadataOperators: []operator.Config{
					{
						Builder: fullCaptureRegexConfig,
					},
				},
				MaxHeaderSize: &smallByteSize,
			},
		},
		{
			name:               "Pipeline starts with non-parser",
			fileContents:       "#aField: SomeValue\nThis is a non-header line\n",
			expectedAttributes: nil,
			conf: HeaderConfig{
				MultilinePattern: "^#",
				MetadataOperators: []operator.Config{
					{
						Builder: generateConf,
					},
					{
						Builder: basicRegexConfig,
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h, err := tc.conf.buildHeader(encoding.Nop, zaptest.NewLogger(t).Sugar())
			require.NoError(t, err)

			r := bytes.NewReader([]byte(tc.fileContents))

			encConf := helper.EncodingConfig{
				Encoding: "nop",
			}

			enc, err := encConf.Build()
			require.NoError(t, err)

			fa := &FileAttributes{}

			h.ReadHeader(context.Background(), r, enc, fa)

			require.Equal(t, tc.expectedAttributes, fa.HeaderAttributes)
		})
	}
}

type mockEncoding struct {
	encodingTransformer transform.Transformer
}

func (m mockEncoding) NewEncoder() *encoding.Encoder {
	return &encoding.Encoder{
		Transformer: m.encodingTransformer,
	}
}

func (m mockEncoding) NewDecoder() *encoding.Decoder {
	// Unimplemented
	return nil
}

// errorTransformer is a mock transform.Transformer that always returns the provided error
type errorTransformer struct {
	e error
}

func (et errorTransformer) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	return 0, 0, et.e
}

func (errorTransformer) Reset() {}
