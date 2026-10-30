package scan

import (
	"github.com/airbnb/rudolph/pkg/store"
)

func GetScanService(scanner store.ScanAPI) ScanService {
	return ConcreteScanService{
		scanner: scanner,
	}
}

// ScanService iterates every document. Each page is passed to callback, then
// to stop. stop returns true to halt. A non-nil last evaluated key resumes
// the next page.
type ScanService interface {
	ScanAll(in store.ScanInput, callback func(out *store.ScanOutput) error, stop func(out *store.ScanOutput) (bool, error)) error
}

type ConcreteScanService struct {
	scanner store.ScanAPI
}

func (c ConcreteScanService) ScanAll(in store.ScanInput, callback func(out *store.ScanOutput) error, stop func(out *store.ScanOutput) (bool, error)) (err error) {
	nextInput := &in
	for {
		shouldStop, lastEvaluatedKey, err := scanIterator(c.scanner, nextInput, callback, stop)
		if err != nil {
			return err
		}
		if shouldStop || lastEvaluatedKey == nil {
			return nil
		}
		nextInput.ExclusiveStartKey = lastEvaluatedKey
	}
}

func scanIterator(scanner store.ScanAPI, in *store.ScanInput, callback func(out *store.ScanOutput) error, stop func(out *store.ScanOutput) (bool, error)) (bool, *store.PrimaryKey, error) {
	out, err := scanner.Scan(in)
	if err != nil {
		return true, nil, err
	}
	if err = callback(out); err != nil {
		return true, nil, err
	}
	if out.LastEvaluatedKey == nil {
		return true, nil, nil
	}
	shouldStop, err := stop(out)
	if err != nil {
		return true, nil, err
	}
	pk, ok := primaryKey(out.LastEvaluatedKey)
	if !ok {
		return true, nil, nil
	}
	return shouldStop, &pk, nil
}

func primaryKey(item store.Item) (store.PrimaryKey, bool) {
	pk, ok1 := item["PK"].(string)
	sk, ok2 := item["SK"].(string)
	if !ok1 || !ok2 {
		return store.PrimaryKey{}, false
	}
	return store.PrimaryKey{PartitionKey: pk, SortKey: sk}, true
}
