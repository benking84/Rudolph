package globalrules

import (
	"fmt"

	"github.com/airbnb/rudolph/pkg/clock"
	"github.com/airbnb/rudolph/pkg/model/feedrules"
	"github.com/airbnb/rudolph/pkg/store"
	"github.com/airbnb/rudolph/pkg/types"
	"github.com/google/uuid"
)

// RemoveGlobalRule will remove the rule from the global repository of rules.
// It also creates a new rule entry in the feed that explicitly tells future syncs
// to remove the rule too.
func RemoveGlobalRule(
	timeProvider clock.TimeProvider,
	getter store.GetItemAPI,
	transacter store.TransactWriteItemsAPI,
	ruleSortKey string,
	txnIdempotencyKey string,
) error {
	rule, err := GetGlobalRuleBySortKey(getter, ruleSortKey)
	if err != nil {
		return fmt.Errorf("query to retrieve existing rule failed: %w", err)
	}
	if rule == nil {
		return fmt.Errorf("no such rule with sk (%s) exists", ruleSortKey)
	}

	// Delete the global rule
	txnDeleteItem, err := transacter.CreateTransactDeleteItem(rule.PrimaryKey)
	if err != nil {
		return fmt.Errorf("failed to create txn item to delete original rule: %w", err)
	}

	// In order to get non-clean sync clients to pick up the new rule diff, add it to the feed as a "remove"
	// Subsequent non-clean syncs will use the feed cursor to cross the date
	feedrule := feedrules.ConstructFeedRuleFromBaseRule(timeProvider, rule.SantaRule)
	feedrule.Policy = types.Remove
	txnPutItem, err := transacter.CreateTransactPutItem(feedrule)
	if err != nil {
		return fmt.Errorf("failed to create txn item to add removal to feed: %w", err)
	}

	if len(txnIdempotencyKey) == 0 {
		txnIdempotencyKey = uuid.NewString()
	}
	err = transacter.TransactWriteItems([]store.TransactWriteItem{txnDeleteItem, txnPutItem}, &txnIdempotencyKey)
	if err != nil {
		return fmt.Errorf("transaction delete failed: %w", err)
	}

	return nil
}

type RuleRemovalService interface {
	RemoveGlobalRule(ruleSortKey string, idempotencyKey string) (err error)
}
type ConcreteRuleRemovalService struct {
	TimeProvider clock.TimeProvider
	Getter       store.GetItemAPI
	Transacter   store.TransactWriteItemsAPI
}

func (c ConcreteRuleRemovalService) RemoveGlobalRule(ruleSortKey string, idempotencyKey string) (err error) {
	return RemoveGlobalRule(c.TimeProvider, c.Getter, c.Transacter, ruleSortKey, idempotencyKey)
}
