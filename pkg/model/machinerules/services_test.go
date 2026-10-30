package machinerules

import (
	"testing"
	"time"

	"github.com/airbnb/rudolph/pkg/clock"
	"github.com/airbnb/rudolph/pkg/store"
	"github.com/airbnb/rudolph/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockMachineRulesUpdater func(key store.PrimaryKey, item interface{}) (*store.UpdateItemOutput, error)

func (m mockMachineRulesUpdater) UpdateItem(key store.PrimaryKey, item interface{}) (*store.UpdateItemOutput, error) {
	return m(key, item)
}

var _ store.UpdateItemAPI = mockMachineRulesUpdater(nil)

func Test_ConcreteMachineRulesUpdater_OK(t *testing.T) {
	machineID := "858CBF28-5EAA-58A3-A155-BA5E81D5B5DD"
	sha256 := "ed0a9ba83449b5966363e0c20fe7755defcb2d7136657d3880bb462a8d7a7025"

	updater := ConcreteMachineRulesUpdater{
		Updater: mockMachineRulesUpdater(
			func(key store.PrimaryKey, item interface{}) (*store.UpdateItemOutput, error) {
				assert.Equal(t, machineRulePK(machineID), key.PartitionKey)
				assert.Equal(t, machineRuleSK(sha256, types.Binary), key.SortKey)

				thisItem := item.(updateRulePolicyRequest)
				assert.Equal(t, types.Blocklist, thisItem.Policy)

				return &store.UpdateItemOutput{}, nil
			},
		),
		TimeProvider: clock.FrozenTimeProvider{
			Current: time.Now(),
		},
	}

	err := updater.UpdateMachineRulePolicy(machineID, sha256, types.Binary, types.Blocklist)
	assert.Empty(t, err)
}

type MockDynamodb struct {
	store.Client
	mock.Mock
}

func (m *MockDynamodb) GetItem(key store.PrimaryKey, consistentRead bool) (*store.GetItemOutput, error) {
	args := m.Called(key, consistentRead)
	return args.Get(0).(*store.GetItemOutput), args.Error(1)
}
func (m *MockDynamodb) PutItem(item interface{}) (*store.PutItemOutput, error) {
	args := m.Called(item)
	return args.Get(0).(*store.PutItemOutput), args.Error(1)
}

func Test_Service_Get(t *testing.T) {
	machineID := "858CBF28-5EAA-58A3-A155-BA5E81D5B5DD"
	identifier := "ed0a9ba83449b5966363e0c20fe7755defcb2d7136657d3880bb462a8d7a7025"

	t.Run("GetItem returns no item", func(t *testing.T) {
		mocked := &MockDynamodb{}
		mocked.On("GetItem", mock.Anything, mock.Anything).Return(&store.GetItemOutput{}, nil)

		service := ConcreteMachineRulesService{
			dynamodb: mocked,
		}

		item, err := service.Get(machineID, identifier, types.Binary)
		assert.Empty(t, err)
		assert.Empty(t, item)
	})

	t.Run("GetItem returns the item", func(t *testing.T) {
		mocked := &MockDynamodb{}
		mocked.On("GetItem", mock.Anything, mock.Anything).Return(&store.GetItemOutput{
			Item: store.Item{
				"PK":         "MachineRules#858CBF28-5EAA-58A3-A155-BA5E81D5B5DD",
				"SK":         "Binary#ed0a9ba83449b5966363e0c20fe7755defcb2d7136657d3880bb462a8d7a7025",
				"SHA256":     "ed0a9ba83449b5966363e0c20fe7755defcb2d7136657d3880bb462a8d7a7025",
				"Identifier": "ed0a9ba83449b5966363e0c20fe7755defcb2d7136657d3880bb462a8d7a7025",
				"Policy":     int64(1),
			},
		}, nil)

		service := ConcreteMachineRulesService{
			dynamodb: mocked,
		}

		item, err := service.Get(machineID, identifier, types.Binary)
		assert.Empty(t, err)

		assert.Equal(t, item.Identifier, "ed0a9ba83449b5966363e0c20fe7755defcb2d7136657d3880bb462a8d7a7025")
		assert.Equal(t, item.SHA256, "ed0a9ba83449b5966363e0c20fe7755defcb2d7136657d3880bb462a8d7a7025")
		assert.Equal(t, item.Policy, types.Allowlist)
	})
}

func Test_Service_Add_OK(t *testing.T) {
	t.Run("PutItem works with no errors", func(t *testing.T) {
		machineID := "858CBF28-5EAA-58A3-A155-BA5E81D5B5DD"
		identifier := "ed0a9ba83449b5966363e0c20fe7755defcb2d7136657d3880bb462a8d7a7025"
		ruleType := types.Binary
		description := "Description"
		policy := types.AllowlistCompiler
		timeProvider := clock.FrozenTimeProvider{
			Current: time.Now(),
		}
		expires := timeProvider.Now()

		mocked := &MockDynamodb{}
		mocked.On("PutItem", mock.MatchedBy(func(item interface{}) bool {
			rule := item.(MachineRuleRow)
			return rule.Description == description && rule.Policy == policy && rule.Identifier == identifier
		})).Return(&store.PutItemOutput{}, nil)

		service := ConcreteMachineRulesService{
			dynamodb: mocked,
		}

		err := service.Add(machineID, identifier, ruleType, policy, description, expires)
		assert.Empty(t, err)
		mocked.AssertCalled(t, "PutItem", mock.Anything)
	})
}
