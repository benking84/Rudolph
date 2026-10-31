package machinerules

import (
	"fmt"

	"github.com/airbnb/rudolph/pkg/store"
	"github.com/airbnb/rudolph/pkg/types"
)

func getItemAsMachineRule(client store.GetItemAPI, partitionKey string, sortKey string) (rule *MachineRuleRow, err error) {
	output, err := client.GetItem(
		store.PrimaryKey{
			PartitionKey: partitionKey,
			SortKey:      sortKey,
		},
		false,
	)

	if err != nil {
		return
	}

	if len(output.Item) == 0 {
		return
	}

	err = store.UnmarshalMap(output.Item, &rule)

	if err != nil {
		err = fmt.Errorf("succeeded GetItem but failed to unmarshalMap into GlobalRuleRow: %w", err)
		return
	}

	if rule != nil && rule.SHA256 != "" && rule.Identifier == "" {
		rule.Identifier = rule.SHA256
	}

	return
}

// @deprecated
func GetMachineRuleByShaType(client store.GetItemAPI, machineID string, sha256 string, ruleType types.RuleType) (rule *MachineRuleRow, err error) {
	return GetMachineRuleByIdentifierType(client, machineID, sha256, ruleType)
}

func GetMachineRuleByIdentifierType(client store.GetItemAPI, machineID string, identifier string, ruleType types.RuleType) (rule *MachineRuleRow, err error) {
	pk := machineRulePK(machineID)
	sk := machineRuleSK(identifier, ruleType)
	return getItemAsMachineRule(client, pk, sk)
}
