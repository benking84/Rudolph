package sensordata

import (
	"fmt"
	"testing"

	"github.com/airbnb/rudolph/pkg/clock"
	"github.com/airbnb/rudolph/pkg/store"
	rudolphtypes "github.com/airbnb/rudolph/pkg/types"
	"github.com/stretchr/testify/assert"
)

type getSensorData func(key store.PrimaryKey, consistentRead bool) (*store.GetItemOutput, error)

func (getter getSensorData) GetItem(key store.PrimaryKey, consistentRead bool) (*store.GetItemOutput, error) {
	return getter(key, consistentRead)
}

var (
	machineID = "AAAAAAAA-A00A-1234-1234-5864377B4831"
	pk, sk    = MachineIDSensorDataPKSK(machineID)
)

func Test_GetSensorData(t *testing.T) {
	type test struct {
		comment              string
		pk                   string
		sk                   string
		hostname             string
		machineID            string
		serialNumber         string
		requestCleanSync     bool
		osBuild              string
		osVersion            string
		santaVersion         string
		clientMode           rudolphtypes.ClientMode
		binaryRuleCount      int
		certRuleCount        int
		cdHashRuleCount      int
		teamIDRuleCount      int
		signingIDRuleCount   int
		transitiveRuleCount  int
		compilerRuleCount    int
		ruleCount            int
		primaryUser          string
		expectedTime         string
		expectedExpiresAfter int64
		expectedError        string
		expectedDataType     rudolphtypes.DataType
	}

	expected := test{
		comment:              fmt.Sprintf("%s %s", "Testing", machineID),
		hostname:             "macbook.pro.localhost",
		pk:                   pk,
		sk:                   sk,
		machineID:            machineID,
		osBuild:              "20A21",
		osVersion:            "12.34",
		santaVersion:         "2021.1",
		clientMode:           rudolphtypes.Monitor,
		serialNumber:         "123456789ABC",
		requestCleanSync:     false,
		binaryRuleCount:      4,
		certRuleCount:        3,
		cdHashRuleCount:      1,
		teamIDRuleCount:      1,
		signingIDRuleCount:   1,
		transitiveRuleCount:  2,
		compilerRuleCount:    1,
		ruleCount:            12,
		primaryUser:          "john_doe",
		expectedTime:         clock.RFC3339(timeProvider.Now()),
		expectedExpiresAfter: clock.Unixtimestamp(timeProvider.Now().UTC().AddDate(0, 0, 90)),
		expectedDataType:     rudolphtypes.DataTypeSensorData,
	}

	// Marshal the DynamoDB DataType
	dataType, _ := expected.expectedDataType.MarshalStore()

	dynamodb := getSensorData(
		func(key store.PrimaryKey, consistentRead bool) (*store.GetItemOutput, error) {
			switch key.PartitionKey {
			case expected.pk:
				if key.SortKey == expected.sk {
					return &store.GetItemOutput{
						Item: store.Item{
							"PK":                   expected.pk,
							"SK":                   expected.sk,
							"MachineID":            expected.machineID,
							"SerialNum":            expected.serialNumber,
							"OSVersion":            expected.osVersion,
							"OSBuild":              expected.osBuild,
							"SantaVersion":         expected.santaVersion,
							"ClientMode":           int64(expected.clientMode),
							"RequestCleanSync":     expected.requestCleanSync,
							"PrimaryUser":          expected.primaryUser,
							"RuleCount":            int64(expected.ruleCount),
							"CertificateRuleCount": int64(expected.certRuleCount),
							"BinaryRuleCount":      int64(expected.binaryRuleCount),
							"CDHashRuleCount":      int64(expected.cdHashRuleCount),
							"TeamIDRuleCount":      int64(expected.teamIDRuleCount),
							"SigningIDRuleCount":   int64(expected.signingIDRuleCount),
							"CompilerRuleCount":    int64(expected.compilerRuleCount),
							"TransitiveRuleCount":  int64(expected.transitiveRuleCount),
							"Time":                 expected.expectedTime,
							"ExpiresAfter":         int64(expected.expectedExpiresAfter),
							"DataType":             dataType,
						},
					}, nil
				}
			}
			return &store.GetItemOutput{}, nil
		},
	)
	sensorData, err := GetSensorData(dynamodb, expected.machineID)
	if expected.expectedError != "" {
		assert.NotEmpty(t, err)
		assert.Equal(t, expected.expectedError, err.Error())
	}

	assert.Equal(t, expected.machineID, sensorData.MachineID)
	assert.Equal(t, expected.serialNumber, sensorData.SerialNum)
	assert.Equal(t, expected.requestCleanSync, sensorData.RequestCleanSync)
	assert.Equal(t, expected.osBuild, sensorData.OSBuild)
	assert.Equal(t, expected.osVersion, sensorData.OSVersion)
	assert.Equal(t, expected.santaVersion, sensorData.SantaVersion)
	assert.Equal(t, expected.clientMode, sensorData.ClientMode)
	assert.Equal(t, expected.binaryRuleCount, sensorData.BinaryRuleCount)
	assert.Equal(t, expected.certRuleCount, sensorData.CertificateRuleCount)
	assert.Equal(t, expected.cdHashRuleCount, sensorData.CDHashRuleCount)
	assert.Equal(t, expected.teamIDRuleCount, sensorData.TeamIDRuleCount)
	assert.Equal(t, expected.signingIDRuleCount, sensorData.SigningIDRuleCount)
	assert.Equal(t, expected.compilerRuleCount, sensorData.CompilerRuleCount)
	assert.Equal(t, expected.transitiveRuleCount, sensorData.TransitiveRuleCount)
	assert.Equal(t, expected.ruleCount, sensorData.RuleCount)
	assert.Equal(t, expected.primaryUser, sensorData.PrimaryUser)
	assert.Equal(t, pk, sensorData.PartitionKey)
	assert.Equal(t, sk, sensorData.SortKey)
	assert.Equal(t, expected.expectedDataType, sensorData.DataType)
}
