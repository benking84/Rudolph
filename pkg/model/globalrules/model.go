package globalrules

import (
	"github.com/airbnb/rudolph/pkg/model/rules"
	"github.com/airbnb/rudolph/pkg/store"
	"github.com/airbnb/rudolph/pkg/types"
)

const (
	globalRulesPK = "GlobalRules"
)

type GlobalRuleRow struct {
	store.PrimaryKey
	rules.SantaRule
	Description string `dynamodbav:"Description,omitempty"`
}

type updateRulePolicyRequest struct {
	Policy types.Policy `dynamodbav:"Policy"`
}

func globalRulesSK(identifier string, ruleType types.RuleType) string {
	return rules.RuleSortKeyFromTypeIdentifier(identifier, ruleType)
}
