package globalrules

// generate a unit test for the GetPaginatedGlobalRules function

import (
	"testing"

	"github.com/airbnb/rudolph/pkg/model/rules"
	"github.com/airbnb/rudolph/pkg/store"
	"github.com/airbnb/rudolph/pkg/types"
	"github.com/stretchr/testify/assert"
)

type mockQuery func(input *store.QueryInput) (*store.QueryOutput, error)

func (m mockQuery) Query(input *store.QueryInput) (*store.QueryOutput, error) {
	return m(input)
}

func TestGetPaginatedGlobalRules(t *testing.T) {
	rules := []rules.SantaRule{
		{
			RuleType:   types.RuleTypeBinary,
			Policy:     types.RulePolicyAllowlist,
			Identifier: "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9820",
		},
		{
			RuleType: types.RuleTypeBinary,
			Policy:   types.RulePolicyAllowlist,
			SHA256:   "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9821",
		},
		{
			RuleType:   types.RuleTypeCertificate,
			Policy:     types.RulePolicyAllowlist,
			Identifier: "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9822",
		},
		{
			RuleType: types.RuleTypeCertificate,
			Policy:   types.RulePolicyAllowlist,
			SHA256:   "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9823",
		},
		{
			RuleType:   types.RuleTypeTeamID,
			Policy:     types.RulePolicyAllowlist,
			Identifier: "EQHXZ8M8AV",
		},
		{
			RuleType:   types.RuleTypeSigningID,
			Policy:     types.RulePolicyAllowlist,
			Identifier: "platform:com.apple.Safari",
		},
		{
			RuleType:   types.RuleTypeSigningID,
			Policy:     types.RulePolicyAllowlist,
			Identifier: "EQHXZ8M8AV:com.google.Chrome",
		},
		{
			RuleType:   types.RuleTypeSigningID,
			Policy:     types.RulePolicySilentBlocklist,
			Identifier: "BJ4HAAB9B3:us.zoom.xos",
		},
	}

	mockQueryClient := mockQuery(
		func(input *store.QueryInput) (*store.QueryOutput, error) {
			items := make([]store.Item, len(rules))

			for i, rule := range rules {
				ruleType, err := rule.RuleType.MarshalStore()
				if err != nil {
					return nil, err
				}
				policy, err := rule.Policy.MarshalStore()
				if err != nil {
					return nil, err
				}
				effectiveID := rule.Identifier
				if effectiveID == "" {
					effectiveID = rule.SHA256
				}
				items[i] = store.Item{
					"PK":         globalRulesPK,
					"SK":         globalRulesSK(effectiveID, rule.RuleType),
					"Identifier": rule.Identifier,
					"SHA256":     rule.SHA256,
					"RuleType":   ruleType,
					"Policy":     policy,
				}
			}

			return &store.QueryOutput{
				Items: items,
				Count: int32(len(items)),
			}, nil
		},
	)

	globalRules, lastEvaluatedKey, err := GetPaginatedGlobalRules(
		mockQueryClient,
		1000,
		nil,
	)
	assert.Nil(t, err)
	assert.Nil(t, lastEvaluatedKey)
	if len(rules) != len(globalRules) {
		t.Fatalf("expected %d rules, got %d", len(rules), len(globalRules))
	}

	for i, globalRule := range globalRules {
		expectedIdentifier := rules[i].Identifier
		if expectedIdentifier == "" {
			expectedIdentifier = rules[i].SHA256
		}
		assert.Equal(t, globalRulesPK, globalRule.PrimaryKey.PartitionKey)
		assert.Equal(t, globalRulesSK(expectedIdentifier, rules[i].RuleType), globalRule.PrimaryKey.SortKey)
		assert.Equal(t, expectedIdentifier, globalRule.SantaRule.Identifier)
		assert.Equal(t, rules[i].RuleType, globalRule.RuleType)
		assert.Equal(t, rules[i].Policy, globalRule.SantaRule.Policy)
	}
}
