package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDataTypes_MarshalTest(t *testing.T) {
	tests := []struct {
		name     string
		dataType DataType
		want     []byte
		wantErr  bool
	}{
		{"SensorData", DataTypeSensorData, []byte(DataTypeSensorData), false},
		{"SyncState", DataTypeSyncState, []byte(DataTypeSyncState), false},
		{"RulesFeed", DataTypeRulesFeed, []byte(DataTypeRulesFeed), false},
		{"MachineConfig", DataTypeMachineConfig, []byte(DataTypeMachineConfig), false},
		{"GlobalConfig", DataTypeGlobalConfig, []byte(DataTypeGlobalConfig), false},
		{"MISSPELLED", DataType(""), []byte(nil), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.dataType.MarshalText()
			if (err != nil) != tt.wantErr {
				t.Errorf("DataType.MarshalText() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDataTypes_UnmarshalText(t *testing.T) {
	tests := []struct {
		name    string
		text    []byte
		want    DataType
		wantErr bool
	}{
		{"SensorData", []byte(DataTypeSensorData), DataTypeSensorData, false},
		{"SyncState", []byte(DataTypeSyncState), DataTypeSyncState, false},
		{"RulesFeed", []byte(DataTypeRulesFeed), DataTypeRulesFeed, false},
		{"MachineConfig", []byte(DataTypeMachineConfig), DataTypeMachineConfig, false},
		{"GlobalConfig", []byte(DataTypeGlobalConfig), DataTypeGlobalConfig, false},
		{"MISSPELLED", []byte(""), DataType(""), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var dt DataType
			err := dt.UnmarshalText(tt.text)
			if (err != nil) != tt.wantErr {
				t.Errorf("DataType.UnmarshalText() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, dt)
		})
	}
}

func TestDataTypes_MarshalStore(t *testing.T) {
	tests := []struct {
		name     string
		dataType DataType
		want     interface{}
		wantErr  bool
	}{
		{"SensorData", DataTypeSensorData, string(DataTypeSensorData), false},
		{"SyncState", DataTypeSyncState, string(DataTypeSyncState), false},
		{"RulesFeed", DataTypeRulesFeed, string(DataTypeRulesFeed), false},
		{"MachineConfig", DataTypeMachineConfig, string(DataTypeMachineConfig), false},
		{"GlobalConfig", DataTypeGlobalConfig, string(DataTypeGlobalConfig), false},
		{"MISSPELLED", DataType(""), nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.dataType.MarshalStore()
			if (err != nil) != tt.wantErr {
				t.Errorf("DataType.MarshalStore() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDataTypes_UnmarshalStore(t *testing.T) {
	tests := []struct {
		name    string
		av      interface{}
		want    DataType
		wantErr bool
	}{
		{"SensorData", string(DataTypeSensorData), DataTypeSensorData, false},
		{"SyncState", string(DataTypeSyncState), DataTypeSyncState, false},
		{"RulesFeed numeric", "5", DataTypeRulesFeed, false},
		{"MachineConfig", string(DataTypeMachineConfig), DataTypeMachineConfig, false},
		{"GlobalConfig", string(DataTypeGlobalConfig), DataTypeGlobalConfig, false},
		{"MISSPELLED", "nope", DataType(""), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DataType("")
			err := got.UnmarshalStore(tt.av)
			if (err != nil) != tt.wantErr {
				t.Errorf("DataType.UnmarshalStore() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
