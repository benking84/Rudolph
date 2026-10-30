package machinerules

import (
	"fmt"
	"testing"

	"github.com/airbnb/rudolph/pkg/store"
	"github.com/airbnb/rudolph/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockGetter func(key store.PrimaryKey, consistentRead bool) (*store.GetItemOutput, error)

func (m mockGetter) GetItem(key store.PrimaryKey, consistentRead bool) (*store.GetItemOutput, error) {
	return m(key, consistentRead)
}

type mockUpdater func(key store.PrimaryKey, item interface{}) (*store.UpdateItemOutput, error)

func (m mockUpdater) UpdateItem(key store.PrimaryKey, item interface{}) (*store.UpdateItemOutput, error) {
	return m(key, item)
}

var _ store.GetItemAPI = mockGetter(nil)
var _ store.UpdateItemAPI = mockUpdater(nil)

func Test_RemoveMachineRule_OK(t *testing.T) {
	type test struct {
		globalRuleExists    bool
		expectedFinalPolicy types.Policy
	}

	cases := []test{
		{
			globalRuleExists:    false,
			expectedFinalPolicy: types.Remove,
		},
		{
			globalRuleExists:    true,
			expectedFinalPolicy: types.Blocklist,
		},
	}

	for _, testcase := range cases {
		machineID := "AAAA-BBBB-CCCC"
		var updatecalled bool
		err := RemoveMachineRule(
			mockGetter(
				func(key store.PrimaryKey, consistentRead bool) (*store.GetItemOutput, error) {
					switch key.PartitionKey {
					case "MachineRules#AAAA-BBBB-CCCC":
						if key.SortKey == "AAA#SORTKEY" {
							return &store.GetItemOutput{
								Item: store.Item{
									"PK":     key.PartitionKey,
									"SK":     key.SortKey,
									"Policy": int64(1),
								},
							}, nil
						}

					case "GlobalRules":
						if key.SortKey == "AAA#SORTKEY" {
							if !testcase.globalRuleExists {
								return &store.GetItemOutput{}, nil
							}
							return &store.GetItemOutput{
								Item: store.Item{
									"PK":     key.PartitionKey,
									"SK":     key.SortKey,
									"Policy": int64(2),
								},
							}, nil
						}
					}
					assert.Fail(t, fmt.Sprintf("dynamodb:GetItem call unexpected with %+v", key))
					return &store.GetItemOutput{}, nil
				},
			),
			mockUpdater(
				func(key store.PrimaryKey, item interface{}) (*store.UpdateItemOutput, error) {
					assert.Equal(t, "MachineRules#AAAA-BBBB-CCCC", key.PartitionKey)
					assert.Equal(t, "AAA#SORTKEY", key.SortKey)

					testitem := item.(ruleRemovalRequest)
					assert.Equal(t, testcase.expectedFinalPolicy, testitem.Policy)
					assert.True(t, testitem.DeleteOnNextSync)

					updatecalled = true

					return &store.UpdateItemOutput{}, nil
				},
			),
			machineID,
			"AAA#SORTKEY",
		)

		assert.True(t, updatecalled)
		assert.Empty(t, err)
	}
}

func Test_RemoveMachineRule_NotFound(t *testing.T) {
	err := RemoveMachineRule(
		mockGetter(
			func(key store.PrimaryKey, consistentRead bool) (*store.GetItemOutput, error) {
				return &store.GetItemOutput{}, nil
			},
		),
		mockUpdater(
			func(key store.PrimaryKey, item interface{}) (*store.UpdateItemOutput, error) {
				assert.Fail(t, "UpdateItem should not be called when rule does not exist")
				return &store.UpdateItemOutput{}, nil
			},
		),
		"AAAA-BBBB-CCCC",
		"AAA#SORTKEY",
	)

	assert.ErrorIs(t, err, ErrRuleNotFound)
}

func Test_ConcreteMachineRulesService_Remove_Idempotent(t *testing.T) {
	mocked := &MockDynamodb{}
	mocked.On("GetItem", mock.Anything, mock.Anything).Return(&store.GetItemOutput{}, nil)

	service := ConcreteMachineRulesService{dynamodb: mocked}

	err := service.Remove("AAAA-BBBB-CCCC", "somehash", types.Binary)
	assert.NoError(t, err, "Remove should be a no-op when the rule does not exist")
}

func Test_ConcreteMachineRulesService_RemoveBySortKey_Idempotent(t *testing.T) {
	mocked := &MockDynamodb{}
	mocked.On("GetItem", mock.Anything, mock.Anything).Return(&store.GetItemOutput{}, nil)

	service := ConcreteMachineRulesService{dynamodb: mocked}

	err := service.RemoveBySortKey("AAAA-BBBB-CCCC", "Binary#somehash")
	assert.NoError(t, err, "RemoveBySortKey should be a no-op when the rule does not exist")
}
