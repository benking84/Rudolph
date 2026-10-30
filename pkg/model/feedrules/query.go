package feedrules

import (
	"errors"
	"fmt"
	"log"

	"github.com/airbnb/rudolph/pkg/store"
)

// GetPaginatedFeedRules returns zero or more rules on the feed, up to the limit.
// If there are more rules to paginate through, lastEvaluatedKey can be passed
// back as exclusiveStartKey. lastEvaluatedKey is nil when this is the last page.
func GetPaginatedFeedRules(
	client store.QueryAPI,
	limit int,
	exclusiveStartKey *store.PrimaryKey,
) (
	items []*FeedRuleRow,
	lastEvaluatedKey *store.PrimaryKey,
	err error,
) {
	partitionKey := feedRulesPK

	if limit <= 0 {
		err = errors.New("invalid limit/batchsize specified")
		return
	}

	result, err := client.Query(&store.QueryInput{
		PartitionKey:      partitionKey,
		ExclusiveStartKey: exclusiveStartKey,
		Limit:             int32(limit),
	})
	if err != nil {
		err = fmt.Errorf("failed to read feed rules for partitionKey %q: %w", partitionKey, err)
		return
	}

	if result.LastEvaluatedKey != nil {
		err = store.UnmarshalMap(result.LastEvaluatedKey, &lastEvaluatedKey)
		if err != nil {
			err = fmt.Errorf("failed to unmarshall LastEvaluatedKey: %w", err)
			return
		}
		log.Printf("    lastEvaluatedKey: %+v", lastEvaluatedKey)
	}

	err = store.UnmarshalListOfMaps(result.Items, &items)
	if err != nil {
		err = fmt.Errorf("failed to unmarshal feed rules: %w", err)
		return
	}

	for _, item := range items {
		if item.SHA256 != "" && item.Identifier == "" {
			item.Identifier = item.SHA256
		}
	}
	return
}
