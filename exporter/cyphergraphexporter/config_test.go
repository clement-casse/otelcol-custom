package cyphergraphexporter

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configopaque"
	"go.opentelemetry.io/collector/confmap/confmaptest"
	"go.opentelemetry.io/collector/confmap/xconfmap"

	"github.com/clement-casse/otelcol-custom/exporter/cyphergraphexporter/internal/metadata"
)

func defaultRessourceMapperWith(extensions map[string]*ResourceMapper) map[string]*ResourceMapper {
	rm := defaultResourcesMappers
	for k, v := range extensions {
		rm[k] = v
	}
	return rm
}

func TestLoadConfig(t *testing.T) {
	t.Parallel()

	cm, err := confmaptest.LoadConf(filepath.Join("testdata", "config.yaml"))
	require.NoError(t, err)

	for _, tt := range []struct {
		id                   component.ID
		expected             component.Config
		configValidateAssert assert.ErrorAssertionFunc
	}{
		{
			id:                   component.NewIDWithName(metadata.Type, ""),
			expected:             createDefaultConfig(),
			configValidateAssert: assert.NoError,
		}, {
			id: component.NewIDWithName(metadata.Type, "withnoauth"),
			expected: &Config{
				DatabaseURI:     "bolt://neo4j-host:7687",
				UserAgent:       defaultUserAgent,
				ResourceMappers: defaultResourcesMappers,
			},
			configValidateAssert: assert.NoError,
		}, {
			id: component.NewIDWithName(metadata.Type, "withbasicauth"),
			expected: &Config{
				DatabaseURI:     "bolt://neo4j-host:7687",
				Username:        "neo4j",
				Password:        configopaque.String("password"),
				UserAgent:       defaultUserAgent,
				ResourceMappers: defaultResourcesMappers,
			},
			configValidateAssert: assert.NoError,
		}, {
			id: component.NewIDWithName(metadata.Type, "withbearertoken"),
			expected: &Config{
				DatabaseURI:     "bolt://neo4j-host:7687",
				BearerToken:     configopaque.String("somevalue"),
				UserAgent:       defaultUserAgent,
				ResourceMappers: defaultResourcesMappers,
			},
			configValidateAssert: assert.NoError,
		}, {
			id: component.NewIDWithName(metadata.Type, "withkerberosticket"),
			expected: &Config{
				DatabaseURI:     "bolt://neo4j-host:7687",
				KerberosTicket:  configopaque.String("somevalue"),
				UserAgent:       defaultUserAgent,
				ResourceMappers: defaultResourcesMappers,
			},
			configValidateAssert: assert.NoError,
		}, {
			id: component.NewIDWithName(metadata.Type, "withcustomua"),
			expected: &Config{
				DatabaseURI:     defaultDatabaseURI,
				UserAgent:       "testUserAgent",
				ResourceMappers: defaultResourcesMappers,
			},
			configValidateAssert: assert.NoError,
		}, {
			id: component.NewIDWithName(metadata.Type, "ERRORwithbasicandbearer"),
			expected: &Config{
				DatabaseURI:     defaultDatabaseURI,
				UserAgent:       defaultUserAgent,
				Username:        "neo4j",
				Password:        configopaque.String("password"),
				BearerToken:     configopaque.String("somevalue"),
				ResourceMappers: defaultResourcesMappers,
			},
			configValidateAssert: func(t assert.TestingT, err error, _ ...any) bool {
				return assert.ErrorContains(t, err, errMultipleAuthMethod.Error())
			},
		}, {
			id: component.NewIDWithName(metadata.Type, "ERRORwithabearerandkerb"),
			expected: &Config{
				DatabaseURI:     defaultDatabaseURI,
				UserAgent:       defaultUserAgent,
				BearerToken:     configopaque.String("somevalue"),
				KerberosTicket:  configopaque.String("somevalue"),
				ResourceMappers: defaultResourcesMappers,
			},
			configValidateAssert: func(t assert.TestingT, err error, _ ...any) bool {
				return assert.ErrorContains(t, err, errMultipleAuthMethod.Error())
			},
		}, {
			id: component.NewIDWithName(metadata.Type, "ERRORwithallauthmethods"),
			expected: &Config{
				DatabaseURI:     defaultDatabaseURI,
				UserAgent:       defaultUserAgent,
				Username:        "neo4j",
				Password:        configopaque.String("password"),
				BearerToken:     configopaque.String("somevalue"),
				KerberosTicket:  configopaque.String("somevalue"),
				ResourceMappers: defaultResourcesMappers,
			},
			configValidateAssert: func(t assert.TestingT, err error, _ ...any) bool {
				return assert.ErrorContains(t, err, errMultipleAuthMethod.Error())
			},
		}, {
			id: component.NewIDWithName(metadata.Type, "ERRORbadurl"),
			expected: &Config{
				DatabaseURI:     "://abcdefghijklmno",
				UserAgent:       defaultUserAgent,
				ResourceMappers: defaultResourcesMappers,
			},
			configValidateAssert: assert.Error,
		}, {
			id: component.NewIDWithName(metadata.Type, "withcustomresourcemapper"),
			expected: &Config{
				DatabaseURI: "bolt://neo4j-host:7687",
				UserAgent:   defaultUserAgent,
				ResourceMappers: defaultRessourceMapperWith(map[string]*ResourceMapper{
					"custom.resource1": {
						IdentifiedByKey: "attr.resource1.uid",
						OtherProperties: []string{},
					},
					"custom.resource2": {
						IdentifiedByKey: "attr.resource2.uid",
						OtherProperties: []string{"other.attr.resource2"},
					},
				}),
			},
			configValidateAssert: assert.NoError,
		},
	} {
		t.Run(tt.id.String(), func(t *testing.T) {
			factory := NewFactory()
			cfg := factory.CreateDefaultConfig()

			sub, err := cm.Sub(tt.id.String())
			require.NoError(t, err)
			require.NoError(t, sub.Unmarshal(cfg))

			vv := xconfmap.Validate(cfg)
			tt.configValidateAssert(t, vv)
			assert.Equal(t, tt.expected, cfg)
		})
	}
}
