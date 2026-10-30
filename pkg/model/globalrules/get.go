package globalrules

import (
	"fmt"

	"github.com/airbnb/rudolph/pkg/store"
	"github.com/airbnb/rudolph/pkg/types"
)

func GetGlobalRuleBySortKey(client store.GetItemAPI, ruleSortKey string) (*GlobalRuleRow, error) {
	return getItemAsGlobalRule(client, globalRulesPK, ruleSortKey)
}

// @deprecated Use GetGlobalRuleByIdentifier
func GetGlobalRuleByShaType(client store.GetItemAPI, sha256 string, ruleType types.RuleType) (*GlobalRuleRow, error) {
	return GetGlobalRuleByIdentifier(client, sha256, ruleType)
}

func GetGlobalRuleByIdentifier(client store.GetItemAPI, identifier string, ruleType types.RuleType) (*GlobalRuleRow, error) {
	pk := globalRulesPK
	sk := globalRulesSK(identifier, ruleType)
	return getItemAsGlobalRule(client, pk, sk)
}

func getItemAsGlobalRule(
	client store.GetItemAPI,
	partitionKey string,
	sortKey string,
) (rule *GlobalRuleRow, err error) {
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

	if rule.SHA256 != "" && rule.Identifier == "" {
		rule.Identifier = rule.SHA256
	}

	return
}
