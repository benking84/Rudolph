package types

import "fmt"

// DataType identifies the current DynamoDB data model
type DataType string

const (
	DataTypeSensorData    DataType = "SensorData"
	DataTypeSyncState     DataType = "SyncState"
	DataTypeGlobalConfig  DataType = "GlobalConfig"
	DataTypeMachineConfig DataType = "MachineConfig"
	DataTypeRulesFeed     DataType = "RulesFeed"
)

// UnmarshalText
func (dt *DataType) UnmarshalText(text []byte) error {
	switch mode := string(text); mode {
	case "SENSOR_DATA":
		fallthrough
	case "SENSORDATA":
		fallthrough
	case "SensorData":
		*dt = DataTypeSensorData
	case "RULES_FEED":
		fallthrough
	case "RULESFEED":
		fallthrough
	case "RulesFeed":
		*dt = DataTypeRulesFeed
	case "SYNC_STATE":
		fallthrough
	case "SYNCSTATE":
		fallthrough
	case "SyncState":
		*dt = DataTypeSyncState
	case "MACHINE_CONFIG":
		fallthrough
	case "MACHINECONFIG":
		fallthrough
	case "MachineConfig":
		*dt = DataTypeMachineConfig
	case "GLOBAL_CONFIG":
		fallthrough
	case "GLOBALCONFIG":
		fallthrough
	case "GlobalConfig":
		*dt = DataTypeGlobalConfig
	default:
		return fmt.Errorf("unknown data_type value %q", mode)
	}
	return nil
}

// MarshalText
func (dt DataType) MarshalText() ([]byte, error) {
	switch dt {
	case DataTypeSensorData:
		return []byte("SensorData"), nil
	case DataTypeSyncState:
		return []byte("SyncState"), nil
	case DataTypeMachineConfig:
		return []byte("MachineConfig"), nil
	case DataTypeGlobalConfig:
		return []byte("GlobalConfig"), nil
	case DataTypeRulesFeed:
		return []byte("RulesFeed"), nil
	default:
		return nil, fmt.Errorf("unknown data_type %s", dt)
	}
}

// MarshalStore stores the data type as its canonical string.
func (dt DataType) MarshalStore() (interface{}, error) {
	switch dt {
	case DataTypeSensorData, DataTypeSyncState, DataTypeMachineConfig, DataTypeGlobalConfig, DataTypeRulesFeed:
		return string(dt), nil
	default:
		return nil, fmt.Errorf("unknown data_type value %q", dt)
	}
}

// UnmarshalStore accepts the canonical name and the legacy numeric aliases.
func (dt *DataType) UnmarshalStore(v interface{}) error {
	s, ok := v.(string)
	if !ok {
		return fmt.Errorf("unexpected data_type value type: %T", v)
	}
	switch s {
	case "1", "SENSOR_DATA", "SENSORDATA", "SensorData":
		*dt = DataTypeSensorData
	case "2", "SYNC_STATE", "SYNCSTATE", "SyncState":
		*dt = DataTypeSyncState
	case "3", "GLOBAL_CONFIG", "GLOBALCONFIG", "GlobalConfig":
		*dt = DataTypeGlobalConfig
	case "4", "MACHINE_CONFIG", "MACHINECONFIG", "MachineConfig":
		*dt = DataTypeMachineConfig
	case "5", "RULES_FEED", "RULESFEED", "RulesFeed":
		*dt = DataTypeRulesFeed
	default:
		return fmt.Errorf("unknown data_type value %q", s)
	}
	return nil
}
