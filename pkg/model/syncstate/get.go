package syncstate

import (
	"fmt"

	"github.com/airbnb/rudolph/pkg/store"
)

// Returns the machine's current sync state
func GetByMachineID(client store.GetItemAPI, machineID string) (syncState *SyncStateRow, err error) {
	output, err := client.GetItem(
		store.PrimaryKey{
			PartitionKey: syncStatePK(machineID),
			SortKey:      syncStateSK,
		},
		// The sync state must be retrieved consistently because it is set by a /preflight request and then immediately
		// retrieved again in the /ruledownload step. Unfortunately it cannot be passed from the /preflight step directly
		// to the ruledownload via some sort of "initial cursor" (At least not as far as I can tell).
		true,
	)
	if err != nil {
		return
	}

	if len(output.Item) == 0 {
		return
	}

	err = store.UnmarshalMap(output.Item, &syncState)

	if err != nil {
		err = fmt.Errorf("succeeded GetItem but failed to unmarshalMap into SyncState: %w", err)
		return
	}

	return
}
