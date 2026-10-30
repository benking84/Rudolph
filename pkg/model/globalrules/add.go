package globalrules

import (
	"errors"

	"github.com/airbnb/rudolph/pkg/clock"
	"github.com/airbnb/rudolph/pkg/model/feedrules"
	"github.com/airbnb/rudolph/pkg/model/rules"
	"github.com/airbnb/rudolph/pkg/store"
	"github.com/airbnb/rudolph/pkg/types"
)

func AddNewGlobalRule(
	time clock.TimeProvider,
	client store.TransactWriteItemsAPI,
	identifier string,
	ruleType types.RuleType,
	policy types.Policy,
	description string,
) error {
	rule := &GlobalRuleRow{
		PrimaryKey: store.PrimaryKey{
			PartitionKey: globalRulesPK,
			SortKey:      globalRulesSK(identifier, ruleType),
		},
		Description: description,
		SantaRule: rules.SantaRule{
			RuleType:   ruleType,
			Policy:     policy,
			Identifier: identifier,
		},
	}

	// Input Validation
	isValid, err := rule.globalRuleValidation()
	if err != nil {
		return err
	}
	if !isValid {
		return errors.New("no errors occurred during the rule validation check but the provided rule is not valid")
	}

	feedRule := feedrules.ConstructFeedRuleFromBaseRule(time, rule.SantaRule)

	putItem1, err := client.CreateTransactPutItem(rule)
	if err != nil {
		return err
	}
	putItem2, err := client.CreateTransactPutItem(feedRule)
	if err != nil {
		return err
	}

	err = client.TransactWriteItems([]store.TransactWriteItem{putItem1, putItem2}, nil)
	return err
}
