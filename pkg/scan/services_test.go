package scan

import (
	"testing"

	"github.com/airbnb/rudolph/pkg/store"
	"github.com/stretchr/testify/assert"
)

type scanApi func(in *store.ScanInput) (*store.ScanOutput, error)

func (g scanApi) Scan(in *store.ScanInput) (*store.ScanOutput, error) {
	return g(in)
}

type testItem struct {
	store.PrimaryKey
	Field1 string `dynamodbav:"Field1"`
}

func Test_ScanService(t *testing.T) {
	scanner := scanApi(
		func(in *store.ScanInput) (*store.ScanOutput, error) {
			assert.Equal(t, int32(1), in.Limit)

			page1Key := store.PrimaryKey{PartitionKey: "Page1", SortKey: "Yes"}
			page2Key := store.PrimaryKey{PartitionKey: "EndOfPage1", SortKey: "Yes"}
			page3Key := store.PrimaryKey{PartitionKey: "EndOfPage2", SortKey: "Yes"}
			item1in, _ := store.MarshalMap(testItem{PrimaryKey: page1Key, Field1: "value1"})
			item2in, _ := store.MarshalMap(testItem{PrimaryKey: page2Key, Field1: "value2"})
			item3in, _ := store.MarshalMap(testItem{PrimaryKey: page3Key, Field1: "value3"})
			page2KeyIn, _ := store.MarshalMap(page2Key)
			page3KeyIn, _ := store.MarshalMap(page3Key)

			if in.ExclusiveStartKey == nil {
				return &store.ScanOutput{
					LastEvaluatedKey: page2KeyIn,
					Items:            []store.Item{item1in},
					Count:            1,
				}, nil
			}
			if in.ExclusiveStartKey.PartitionKey == "EndOfPage1" {
				return &store.ScanOutput{
					LastEvaluatedKey: page3KeyIn,
					Items:            []store.Item{item2in},
					Count:            1,
				}, nil
			}
			return &store.ScanOutput{
				Items: []store.Item{item3in},
				Count: 1,
			}, nil
		},
	)

	service := ConcreteScanService{scanner: scanner}
	err := service.ScanAll(store.ScanInput{Limit: 1}, func(out *store.ScanOutput) error { return nil }, func(out *store.ScanOutput) (bool, error) {
		return false, nil
	})
	assert.NoError(t, err)
	// Count pages by scanning again with a counter.
	numItems := 0
	err = service.ScanAll(store.ScanInput{Limit: 1}, func(out *store.ScanOutput) error {
		numItems += int(out.Count)
		return nil
	}, func(out *store.ScanOutput) (bool, error) {
		return false, nil
	})
	assert.Empty(t, err)
	assert.Equal(t, 3, numItems)
}
