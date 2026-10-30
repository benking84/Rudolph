package postflight

import (
	"testing"

	"github.com/airbnb/rudolph/pkg/clock"
	"github.com/airbnb/rudolph/pkg/model/syncstate"
	"github.com/airbnb/rudolph/pkg/store"
	"github.com/stretchr/testify/assert"
)

type mockUpdater func(key store.PrimaryKey, item interface{}) (*store.UpdateItemOutput, error)

func (m mockUpdater) UpdateItem(key store.PrimaryKey, item interface{}) (*store.UpdateItemOutput, error) {
	return m(key, item)
}

var _ store.UpdateItemAPI = mockUpdater(nil)

func Test_ConcreteSyncStateUpdater_OK(t *testing.T) {
	cur, _ := clock.ParseRFC3339("2000-01-01T00:00:00Z")
	updater := concreteSyncStateUpdater{
		updater: mockUpdater(
			func(key store.PrimaryKey, item interface{}) (*store.UpdateItemOutput, error) {
				assert.Equal(t, "Machine#AAAA-BBBB-CCCC", key.PartitionKey)
				assert.Equal(t, "SyncState", key.SortKey)

				thisItem := item.(syncstate.UpdatePostflightItem)
				assert.Equal(t, "2000-01-01T00:00:00Z", thisItem.PostflightAt)
				assert.Equal(t, "1999-12-31T23:50:00Z", thisItem.FeedSyncCursor) // FIXME (derek.wang) This isn't necessarily "good" nor "correct", just "lazy"
				assert.Equal(t, cur.AddDate(0, 0, 90).Unix(), thisItem.ExpiresAfter)

				return &store.UpdateItemOutput{}, nil
			},
		),
		timeProvider: clock.FrozenTimeProvider{Current: cur},
	}

	err := updater.updatePostflightDate("AAAA-BBBB-CCCC")
	assert.Empty(t, err)
}

type mockQueryer func(input *store.QueryInput) (*store.QueryOutput, error)
type mockDeleter func(key store.PrimaryKey) (*store.DeleteItemOutput, error)

func (m mockQueryer) Query(input *store.QueryInput) (*store.QueryOutput, error) {
	return m(input)
}
func (m mockDeleter) DeleteItem(key store.PrimaryKey) (*store.DeleteItemOutput, error) {
	return m(key)
}

var _ store.QueryAPI = mockQueryer(nil)
var _ store.DeleteItemAPI = mockDeleter(nil)

func Test_RuleDeleter_OK(t *testing.T) {
	destroyer := concreteRuleDestroyer{
		queryer: mockQueryer(
			func(input *store.QueryInput) (*store.QueryOutput, error) {
				assert.Equal(t, "MachineRules#AAAA-BBBB-CCCC", input.PartitionKey)
				assert.Equal(t, true, input.Filters["DeleteOnNextSync"])
				assert.Equal(t, []string{"PK", "SK"}, input.Projection)
				return &store.QueryOutput{}, nil
			},
		),
		deleter: mockDeleter(
			func(key store.PrimaryKey) (*store.DeleteItemOutput, error) {
				assert.Equal(t, "Machine#AAAA-BBBB-CCCC", key.PartitionKey)
				assert.Equal(t, "SyncState", key.SortKey)

				return &store.DeleteItemOutput{}, nil
			},
		),
	}

	err := destroyer.destroyMachineRulesMarkedForDeletion("AAAA-BBBB-CCCC")
	assert.Empty(t, err)
}
