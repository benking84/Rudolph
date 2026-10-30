package globalrules

import (
	"errors"
	"fmt"

	"github.com/airbnb/rudolph/pkg/store"
)

func PingDatabase(client store.QueryAPI) (err error) {
	_, _, err = GetPaginatedGlobalRules(client, 1, nil)
	return
}

func GetPaginatedGlobalRules(
	client store.QueryAPI,
	limit int,
	exclusiveStartKey *store.PrimaryKey,
) (
	items []*GlobalRuleRow,
	lastEvaluatedKey *store.PrimaryKey,
	err error,
) {
	partitionKey := globalRulesPK

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
		err = fmt.Errorf("failed to read rules for partitionKey %q: %w", partitionKey, err)
		return
	}

	if result.LastEvaluatedKey != nil {
		err = store.UnmarshalMap(result.LastEvaluatedKey, &lastEvaluatedKey)
		if err != nil {
			err = fmt.Errorf("failed to UnmarshalMap LastEvaluatedKey: %w", err)
			return
		}
	}

	err = store.UnmarshalListOfMaps(result.Items, &items)
	if err != nil {
		err = fmt.Errorf("failed to UnmarshalListOfMaps result: %w", err)
		return
	}

	for _, item := range items {
		if item.SHA256 != "" && item.Identifier == "" {
			item.Identifier = item.SHA256
		}
	}
	return
}
