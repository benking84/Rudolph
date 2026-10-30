package syncstate

import (
	"errors"
	"fmt"
	"testing"

	"github.com/airbnb/rudolph/pkg/store"
	"github.com/airbnb/rudolph/pkg/types"
	"github.com/stretchr/testify/assert"
)

type getSyncState func(key store.PrimaryKey, consistentRead bool) (*store.GetItemOutput, error)

func (get getSyncState) GetItem(key store.PrimaryKey, consistentRead bool) (*store.GetItemOutput, error) {
	return get(key, consistentRead)
}

func Test_GetIntendedConfig(t *testing.T) {
	machineID := "AAAAAAAA-A00A-1234-1234-5864377B4831"
	pkSyncState := fmt.Sprintf("%s%s", "Machine#", machineID)
	type test struct {
		machineID         string
		pk                string
		sk                string
		dbError           bool
		expectedError     string
		expectedBatchSize int
		expectedCleanSync bool
		expectedDataType  types.DataType
	}

	cases := []test{
		{
			machineID:         machineID,
			pk:                pkSyncState,
			sk:                "SyncState",
			dbError:           false,
			expectedBatchSize: 31,
			expectedCleanSync: false,
			expectedDataType:  types.DataTypeSyncState,
		},
		{
			machineID:     machineID,
			pk:            pkSyncState,
			sk:            "SyncState",
			dbError:       true,
			expectedError: "failed to retrieve sync state items",
		},
	}

	for _, test := range cases {
		dynamodb := getSyncState(
			func(key store.PrimaryKey, consistentRead bool) (*store.GetItemOutput, error) {
				if test.dbError {
					return nil, errors.New("failed to retrieve sync state items")
				}

				switch key.PartitionKey {
				case test.pk:
					if key.SortKey == test.sk {
						return &store.GetItemOutput{
							Item: store.Item{
								"CleanSync": false,
								"BatchSize": int64(31),
								"DataType":  string(test.expectedDataType),
							},
						}, nil
					}
				}

				return &store.GetItemOutput{}, nil
			},
		)
		result, err := GetByMachineID(dynamodb, test.machineID)
		if test.expectedError != "" {
			assert.NotEmpty(t, err)
			assert.Equal(t, test.expectedError, err.Error())
		}

		if test.expectedBatchSize != 0 {
			assert.NotEmpty(t, result)
			assert.Equal(t, test.expectedBatchSize, result.BatchSize)
		}

		if test.expectedDataType != "" {
			assert.NotEmpty(t, result)
			assert.Equal(t, test.expectedDataType, result.DataType)
		}

		if test.expectedCleanSync != false {
			assert.NotEmpty(t, result)
			assert.Equal(t, test.expectedCleanSync, result.CleanSync)
		}
	}
}
