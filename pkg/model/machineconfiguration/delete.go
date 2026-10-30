package machineconfiguration

import (
	"github.com/airbnb/rudolph/pkg/store"
)

func GetConfigurationDeleter(client store.DeleteItemAPI) ConcreteConfigurationDeleter {
	return ConcreteConfigurationDeleter{
		global: ConcreteGlobalConfigurationDeleter{
			deleter: client,
		},
		machine: ConcreteMachineConfigurationDeleter{
			deleter: client,
		},
	}
}

type ConcreteConfigurationDeleter struct {
	global  GlobalConfigurationDeleter
	machine MachineConfigurationDeleter
}

// GlobalConfigurationDeleter
type GlobalConfigurationDeleter interface {
	deleteGlobalConfig() error
}

type ConcreteGlobalConfigurationDeleter struct {
	deleter store.DeleteItemAPI
}

func (d ConcreteGlobalConfigurationDeleter) deleteGlobalConfig() error {
	globalPK := store.PrimaryKey{
		PartitionKey: globalConfigurationPK,
		SortKey:      machineConfigurationSK(),
	}
	_, err := d.deleter.DeleteItem(globalPK)
	return err
}

// GlobalConfigurationDeleter
type MachineConfigurationDeleter interface {
	deleteMachineConfig(machineID string) error
}

type ConcreteMachineConfigurationDeleter struct {
	deleter store.DeleteItemAPI
}

func (d ConcreteMachineConfigurationDeleter) deleteMachineConfig(machineID string) error {
	machinePK := store.PrimaryKey{
		PartitionKey: machineConfigurationPK(machineID),
		SortKey:      machineConfigurationSK(),
	}
	_, err := d.deleter.DeleteItem(machinePK)
	return err
}
