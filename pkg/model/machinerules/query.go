package machinerules

import (
	"fmt"

	"github.com/airbnb/rudolph/pkg/store"
)

// GetPrimaryKeysByMachineIDWhereMarkedForDeletion returns machine rules that
// postflight should delete.
func GetPrimaryKeysByMachineIDWhereMarkedForDeletion(client store.QueryAPI, machineID string) (keys *[]store.PrimaryKey, err error) {
	pk := machineRulePK(machineID)

	output, err := client.Query(&store.QueryInput{
		PartitionKey: pk,
		Filters:      map[string]interface{}{"DeleteOnNextSync": true},
		Projection:   []string{"PK", "SK"},
	})
	if err != nil {
		return
	}
	err = store.UnmarshalListOfMaps(output.Items, &keys)
	return
}

// GetMachineRules returns every rule stored for one machine.
func GetMachineRules(client store.QueryAPI, machineID string) (items *[]MachineRuleRow, err error) {
	partitionKey := machineRulePK(machineID)

	result, err := client.Query(&store.QueryInput{
		PartitionKey: partitionKey,
	})
	if err != nil {
		err = fmt.Errorf("failed to read rules for partitionKey %q: %w", partitionKey, err)
		return
	}

	err = store.UnmarshalListOfMaps(result.Items, &items)
	if err != nil {
		err = fmt.Errorf("failed to unmarshal machine rules: %w", err)
		return
	}
	if items != nil {
		for i := range *items {
			if (*items)[i].SHA256 != "" && (*items)[i].Identifier == "" {
				(*items)[i].Identifier = (*items)[i].SHA256
			}
		}
	}
	return
}
