package machineconfiguration

import (
	"errors"
	"fmt"
	"testing"

	"github.com/airbnb/rudolph/pkg/clock"
	"github.com/airbnb/rudolph/pkg/store"
	rudolphtypes "github.com/airbnb/rudolph/pkg/types"
	"github.com/stretchr/testify/assert"
)

type getApi func(key store.PrimaryKey, consistentRead bool) (*store.GetItemOutput, error)

func (g getApi) GetItem(key store.PrimaryKey, consistentRead bool) (*store.GetItemOutput, error) {
	return g(key, consistentRead)
}

func Test_GetIntendedConfig(t *testing.T) {
	type test struct {
		machineID        string
		dbError          bool   // Force dynamodb mock to return error
		setGlobal        bool   // Whether or not dynamodb has a "global" machineconfiguration
		expecterror      string // expected error string
		expectbatchsize  int    // expected batch size (in config result)
		expectclientmode rudolphtypes.ClientMode
	}

	cases := []test{
		{
			machineID:        "AAAA", // We've pre-populated the mock db with some values
			setGlobal:        true,
			expectbatchsize:  37,
			expectclientmode: rudolphtypes.Lockdown,
		},
		{
			machineID:        "AAAA",
			setGlobal:        false,
			expectbatchsize:  37,
			expectclientmode: rudolphtypes.Lockdown,
		},
		{
			machineID:        "BBBB", // BBBB doesn't have any pre-existing config
			setGlobal:        true,
			expectbatchsize:  15,
			expectclientmode: rudolphtypes.Monitor,
		},
		{
			machineID:        "BBBB",
			setGlobal:        false,
			expectbatchsize:  50,
			expectclientmode: rudolphtypes.Monitor,
		},
		{
			machineID:   "CCCC",
			dbError:     true,
			expecterror: "failed to get machine config: dynamodb failed yay",
		},
	}

	for _, test := range cases {
		dynamodb := getApi(
			func(key store.PrimaryKey, consistentRead bool) (*store.GetItemOutput, error) {
				if test.dbError {
					return nil, errors.New("dynamodb failed yay")
				}

				// Seed the database with stuff
				switch key.PartitionKey {
				case "Machine#AAAA":
					if key.SortKey == "Config" {
						return &store.GetItemOutput{
							Item: store.Item{
								"ClientMode": int64(2),
								"BatchSize":  int64(37),
							},
						}, nil
					}
				case "GlobalConfig":
					if key.SortKey == "Config" {
						if !test.setGlobal {
							return &store.GetItemOutput{}, nil
						}
						return &store.GetItemOutput{
							Item: store.Item{
								"ClientMode": int64(1),
								"BatchSize":  int64(15),
							},
						}, nil
					}
				}

				return &store.GetItemOutput{}, nil
			},
		)

		fetcher := GetConfigurationFetcher(dynamodb, clock.Y2K{})

		// result, err := GetIntendedConfig(dynamodb, test.machineID) // Deprecated method
		result, err := fetcher.getIntendedConfig(test.machineID)

		if test.expecterror != "" {
			assert.NotEmpty(t, err)
			assert.Equal(t, test.expecterror, err.Error())
		}

		if test.expectbatchsize != 0 {
			assert.NotEmpty(t, result)
			assert.Equal(t, test.expectbatchsize, result.BatchSize)
		}

		if test.expectclientmode != 0 {
			assert.NotEmpty(t, result)
			assert.Equal(t, test.expectclientmode, result.ClientMode)
		}
	}
}

type nullCache struct{}

func (n nullCache) Has(key string) bool {
	return false
}
func (n nullCache) Get(key string) *MachineConfiguration {
	return nil
}
func (n nullCache) Set(key string, config *MachineConfiguration) bool {
	return false
}

func TestGetDesiredConfig_DynamodbReturnsStuff(t *testing.T) {
	// Test that DDB will return a machine specific config

	// Example config
	// {"client_mode":"MONITOR","blacklist_regex":"","whitelist_regex":"","batch_size":50,"enable_bundles":false,"enabled_transitive_whitelisting":false}
	var globalConfig = store.Item{
		"SK":                    "Current",
		"MachineID":             "global",
		"ClientMode":            int64(1),
		"BlockedPathRegex":      "",
		"AllowedPathRegex":      "",
		"BatchSize":             int64(50),
		"EnableBundles":         false,
		"EnableTransitiveRules": false,
	}

	api := getApi(
		func(key store.PrimaryKey, consistentRead bool) (*store.GetItemOutput, error) {
			return &store.GetItemOutput{
				Item: globalConfig,
			}, nil
		},
	)

	client := CachedConcreteGlobalConfigurationFetcher{
		getter: api,
		cache:  nullCache{},
	}
	retrievedConfig, err := client.GetGlobalConfig()
	// retrievedConfig, err := GetGlobalConfig(api) // deprecated method

	assert.Empty(t, err)

	// Test to see if it matches the sample configuration
	// {"client_mode":"MONITOR","blocked_path_regex":"","allowed_path_regex":"","batch_size":50,"enable_bundles":false,"enable_transitive_rules":false}
	assert.Equal(t, rudolphtypes.ClientMode(1), retrievedConfig.ClientMode)
	assert.Equal(t, "", retrievedConfig.BlockedPathRegex)
	assert.Equal(t, "", retrievedConfig.AllowedPathRegex)
	assert.Equal(t, 50, retrievedConfig.BatchSize)
	assert.Equal(t, false, retrievedConfig.EnableBundles)
	assert.Equal(t, false, retrievedConfig.EnabledTransitiveRules)

}

func TestGetMachineConfig_DynamodbReturnsCorrectly(t *testing.T) {
	// Test that DDB will return the global_config
	machineID := "AAAAAAAA-A00A-1234-1234-5864377B4831"
	machinePK := fmt.Sprintf("%s%s", machineConfigurationPKPrefix, machineID)

	// Example config
	// {"client_mode":"MONITOR","blocked_path_regex":"","allowed_path_regex":"","batch_size":50,"enable_bundles":false,"enable_transitive_rules":false}
	var machineConfig = store.Item{
		"PK":                    machinePK,
		"SK":                    currentSK,
		"MachineID":             "AAAAAAAA-A00A-1234-1234-5864377B4831",
		"ClientMode":            int64(1),
		"BlockedPathRegex":      "",
		"AllowedPathRegex":      "",
		"BatchSize":             int64(47),
		"CleanSync":             true,
		"EnableBundles":         true,
		"EnableTransitiveRules": false,
	}

	api := getApi(
		func(key store.PrimaryKey, consistentRead bool) (*store.GetItemOutput, error) {
			return &store.GetItemOutput{
				Item: machineConfig,
			}, nil
		},
	)

	fetcher := GetConfigurationFetcher(api, clock.Y2K{})

	// retrievedConfig, err := GetIntendedConfig(api, machineID) // Deprecated method
	retrievedConfig, err := fetcher.getIntendedConfig(machineID)
	assert.Empty(t, err)

	// Test to see if it matches the sample configuration
	// {"client_mode":"MONITOR","blocked_path_regex":"","allowed_path_regex":"","batch_size":50,"enable_bundles":false,"enable_transitive_rules":false}
	assert.Equal(t, rudolphtypes.ClientMode(1), retrievedConfig.ClientMode)
	assert.Equal(t, "", retrievedConfig.BlockedPathRegex)
	assert.Equal(t, "", retrievedConfig.AllowedPathRegex)
	assert.Equal(t, 47, retrievedConfig.BatchSize)
	assert.Equal(t, true, retrievedConfig.EnableBundles)
	assert.Equal(t, false, retrievedConfig.EnabledTransitiveRules)
	assert.Equal(t, true, retrievedConfig.CleanSync)

}

func TestGetGlobalConfig_DynamodbReturnsCorrectly(t *testing.T) {
	// Test that DDB will return the global_config

	// Example config
	// {"client_mode":"MONITOR","blocked_path_regex":"","allowed_path_regex":"","batch_size":50,"enable_bundles":false,"enable_transitive_rules":false}
	var globalConfig = store.Item{
		"PK":                    globalConfigurationPK,
		"SK":                    currentSK,
		"ClientMode":            int64(2),
		"BlockedPathRegex":      "",
		"AllowedPathRegex":      "",
		"BatchSize":             int64(49),
		"EnableBundles":         false,
		"EnableTransitiveRules": true,
		"UploadLogsUrl":         "https://the-north-pole.santa/api/v1/upload-logs",
	}

	api := getApi(
		func(key store.PrimaryKey, consistentRead bool) (*store.GetItemOutput, error) {
			return &store.GetItemOutput{
				Item: globalConfig,
			}, nil
		},
	)

	client := CachedConcreteGlobalConfigurationFetcher{
		getter: api,
		cache:  nullCache{},
	}
	retrievedConfig, err := client.GetGlobalConfig()
	// retrievedConfig, err := GetGlobalConfig(api) // deprecated method
	assert.Empty(t, err)

	// Test to see if it matches the sample configuration
	// {"client_mode":"MONITOR","blocked_path_regex":"","allowed_path_regex":"","batch_size":50,"enable_bundles":false,"enable_transitive_rules":false}
	assert.Equal(t, rudolphtypes.ClientMode(2), retrievedConfig.ClientMode)
	assert.Equal(t, "", retrievedConfig.BlockedPathRegex)
	assert.Equal(t, "", retrievedConfig.AllowedPathRegex)
	assert.Equal(t, 49, retrievedConfig.BatchSize)
	assert.Equal(t, false, retrievedConfig.EnableBundles)
	assert.Equal(t, true, retrievedConfig.EnabledTransitiveRules)
	assert.Equal(t, "https://the-north-pole.santa/api/v1/upload-logs", retrievedConfig.UploadLogsURL)
}
