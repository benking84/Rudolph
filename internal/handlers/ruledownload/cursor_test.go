package ruledownload

import (
	"testing"
	"time"

	"github.com/airbnb/rudolph/pkg/clock"
	"github.com/airbnb/rudolph/pkg/store"
	"github.com/stretchr/testify/assert"
)

type mockUpdater func(key store.PrimaryKey, item interface{}) (*store.UpdateItemOutput, error)

func (m mockUpdater) UpdateItem(key store.PrimaryKey, item interface{}) (*store.UpdateItemOutput, error) {
	return m(key, item)
}

type mockGetter func(key store.PrimaryKey, consistentRead bool) (*store.GetItemOutput, error)

func (m mockGetter) GetItem(key store.PrimaryKey, consistentRead bool) (*store.GetItemOutput, error) {
	return m(key, consistentRead)
}

var _ store.GetItemAPI = mockGetter(nil)
var _ store.UpdateItemAPI = mockUpdater(nil)

type mockTimeProvider struct{}

func (m mockTimeProvider) Now() time.Time {
	t, _ := clock.ParseRFC3339("2000-01-01T00:00:00Z")
	return t
}

func Test_ConcreteRuledownloadCursorService_ConstructCursor(t *testing.T) {
	type test struct {
		syncState      store.Item
		expectStrategy ruledownloadStrategy
		expectPK       string
		expectSK       string
	}

	cases := []test{
		{
			syncState: store.Item{
				"PK":        "Whatever",
				"SK":        "Doesnt matter",
				"CleanSync": true,
				"BatchSize": int64(17),
			},
			expectStrategy: ruledownloadStrategyClean,
		},
		{
			syncState: store.Item{
				"PK":             "Whatever",
				"SK":             "Doesnt matter",
				"CleanSync":      false,
				"BatchSize":      int64(17),
				"FeedSyncCursor": "2000-01-01T02:00:00Z",
			},
			expectStrategy: ruledownloadStrategyIncremental,
			expectPK:       "RulesFeed",
			expectSK:       "2000-01-01T02:00:00Z",
		},
	}

	for _, testcase := range cases {
		machineID := "AAAA-BBBB-CCCC-DDDD"
		service := concreteRuledownloadCursorService{
			timer: mockTimeProvider{},
			getter: mockGetter(
				func(key store.PrimaryKey, consistentRead bool) (*store.GetItemOutput, error) {
					assert.True(t, consistentRead)
					assert.Equal(t, key.PartitionKey, "Machine#AAAA-BBBB-CCCC-DDDD")
					assert.Equal(t, key.SortKey, "SyncState")

					return &store.GetItemOutput{Item: testcase.syncState}, nil
				},
			),
			updater: mockUpdater(
				func(key store.PrimaryKey, item interface{}) (*store.UpdateItemOutput, error) {
					assert.Equal(t, key.PartitionKey, "Machine#AAAA-BBBB-CCCC-DDDD")
					assert.Equal(t, key.SortKey, "SyncState")

					// This is tested elsewhere

					return &store.UpdateItemOutput{}, nil
				},
			),
		}

		req := RuledownloadRequest{
			Cursor: nil,
		}
		cursor, err := service.ConstructCursor(req, machineID)

		assert.Empty(t, err)
		assert.Equal(t, testcase.expectStrategy, cursor.Strategy)
		assert.Equal(t, 17, cursor.BatchSize)
		assert.Equal(t, 1, cursor.PageNumber)
		if testcase.expectPK != "" {
			assert.Equal(t, testcase.expectPK, cursor.PartitionKey)
		} else {
			assert.Empty(t, cursor.PartitionKey)
		}
		if testcase.expectSK != "" {
			assert.Equal(t, testcase.expectSK, cursor.SortKey)
		} else {
			assert.Empty(t, cursor.SortKey)
		}
	}
}
