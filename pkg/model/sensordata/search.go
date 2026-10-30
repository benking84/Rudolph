package sensordata

import (
	"github.com/airbnb/rudolph/pkg/store"
	"github.com/airbnb/rudolph/pkg/types"
)

const (
	MachineID_DataType_GSI           string = store.IndexMachineID
	SerialNum_DataType_MachineID_GSI string = store.IndexSerialNum
)

func GetSensorDataFinder(api store.QueryAPI) SensorDataFinder {
	return ConcreteSensorDataFinder{
		queryapi: api,
	}
}

// SensorDataFinder searches recent preflight sensor documents for machine IDs.
type SensorDataFinder interface {
	GetMachineIDsStartingWith(prefix string, limit int32) ([]string, error)
	GetMachineIDsFromSerialNumber(serialNumber string, limit int32) ([]string, error)
}

type ConcreteSensorDataFinder struct {
	queryapi store.QueryAPI
}

func (f ConcreteSensorDataFinder) GetMachineIDsStartingWith(prefix string, limit int32) (machineIDs []string, err error) {
	output, err := f.queryapi.Query(&store.QueryInput{
		IndexName:       MachineID_DataType_GSI,
		DataType:        string(GetDataType()),
		MachineIDPrefix: prefix,
		Projection:      []string{"MachineID"},
		Limit:           limit,
	})
	if err != nil {
		return
	}
	return machineIDsFromItems(output.Items)
}

func (f ConcreteSensorDataFinder) GetMachineIDsFromSerialNumber(serialNumber string, limit int32) (machineIDs []string, err error) {
	output, err := f.queryapi.Query(&store.QueryInput{
		IndexName:    SerialNum_DataType_MachineID_GSI,
		SerialNumber: serialNumber,
		DataType:     string(GetDataType()),
		Projection:   []string{"MachineID"},
		Limit:        limit,
	})
	if err != nil {
		return
	}
	return machineIDsFromItems(output.Items)
}

func machineIDsFromItems(items []store.Item) ([]string, error) {
	machineIDs := make([]string, len(items))
	gsiItems := make([]dataTypeMachineIDGSIItem, len(items))
	if err := store.UnmarshalListOfMaps(items, &gsiItems); err != nil {
		return nil, err
	}
	for index, item := range gsiItems {
		machineIDs[index] = item.MachineID
	}
	return machineIDs, nil
}

type SerialNumDataTypeMachineIdGSIItem struct {
	PrimaryKey store.PrimaryKey
	SerialNumb string         `dynamodbav:"SerialNum"`
	MachineID  string         `dynamodbav:"MachineID"`
	DataType   types.DataType `dynamodbav:"DataType"`
}

type dataTypeMachineIDGSIItem struct {
	PrimaryKey store.PrimaryKey
	DataType   types.DataType `dynamodbav:"DataType"`
	MachineID  string         `dynamodbav:"MachineID"`
}
