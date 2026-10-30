package postflight

import (
	"log"

	"github.com/airbnb/rudolph/pkg/clock"
	"github.com/airbnb/rudolph/pkg/model/machinerules"
	"github.com/airbnb/rudolph/pkg/model/syncstate"
	"github.com/airbnb/rudolph/pkg/store"
)

func archiveSyncState(client store.Client, machineID string) (err error) {
	syncState, err := syncstate.GetByMachineID(client, machineID)
	if err != nil {
		return
	}

	err = syncstate.Archive(client, *syncState)
	if err != nil {
		log.Printf("Failed to archive syncState")
		return
	}

	return
}

type syncStateUpdater interface {
	updatePostflightDate(machineID string) error
}

type concreteSyncStateUpdater struct {
	updater      store.UpdateItemAPI
	timeProvider clock.TimeProvider
}

func (c concreteSyncStateUpdater) updatePostflightDate(machineID string) (err error) {
	return syncstate.UpdatePostflightDate(c.timeProvider, c.updater, machineID)
}

type staleRuleDestroyer interface {
	destroyMachineRulesMarkedForDeletion(machineID string) error
}

type concreteRuleDestroyer struct {
	queryer store.QueryAPI
	deleter store.DeleteItemAPI
}

func (c concreteRuleDestroyer) destroyMachineRulesMarkedForDeletion(machineID string) (err error) {
	log.Printf("Now evicting stale machine rules")

	keysToDelete, err := machinerules.GetPrimaryKeysByMachineIDWhereMarkedForDeletion(c.queryer, machineID)
	if err != nil {
		return
	}

	log.Printf("Found %d stale machine rules to delete", len(*keysToDelete))

	for _, keyToDelete := range *keysToDelete {
		_, err = c.deleter.DeleteItem(keyToDelete)
		if err != nil {
			return
		}
	}

	return
}
