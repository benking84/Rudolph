package store

import (
	"fmt"
	"sort"
	"sync"
)

// Memory is an in-memory Client. Tests use it as the reference behaviour for
// the Firestore implementation.
type Memory struct {
	mu   sync.Mutex
	docs map[string]Item
	idem map[string]struct{}
}

// NewMemory returns an empty store.
func NewMemory() *Memory {
	return &Memory{
		docs: map[string]Item{},
		idem: map[string]struct{}{},
	}
}

func docKey(pk, sk string) string {
	return pk + "\x00" + sk
}

func (m *Memory) GetItem(key PrimaryKey, _ bool) (*GetItemOutput, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	item, ok := m.docs[docKey(key.PartitionKey, key.SortKey)]
	if !ok {
		return &GetItemOutput{}, nil
	}
	return &GetItemOutput{Item: cloneItem(item)}, nil
}

func (m *Memory) PutItem(item interface{}) (*PutItemOutput, error) {
	encoded, pk, err := encodeStored(item)
	if err != nil {
		return nil, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.docs[docKey(pk.PartitionKey, pk.SortKey)] = encoded
	return &PutItemOutput{}, nil
}

func (m *Memory) UpdateItem(key PrimaryKey, item interface{}) (*UpdateItemOutput, error) {
	fields, err := MarshalMap(item)
	if err != nil {
		return nil, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	existing, ok := m.docs[docKey(key.PartitionKey, key.SortKey)]
	if !ok {
		return nil, fmt.Errorf("store: conditional update failed for %s/%s", key.PartitionKey, key.SortKey)
	}
	merged := cloneItem(existing)
	for k, v := range fields {
		merged[k] = v
	}
	merged["PK"] = key.PartitionKey
	merged["SK"] = key.SortKey
	applyTTL(merged)
	m.docs[docKey(key.PartitionKey, key.SortKey)] = merged
	return &UpdateItemOutput{Attributes: cloneItem(merged)}, nil
}

func (m *Memory) DeleteItem(key PrimaryKey) (*DeleteItemOutput, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.docs, docKey(key.PartitionKey, key.SortKey))
	return &DeleteItemOutput{}, nil
}

func (m *Memory) Query(input *QueryInput) (*QueryOutput, error) {
	if input == nil {
		return nil, fmt.Errorf("store: nil query")
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	var matched []Item
	for _, item := range m.docs {
		if input.IndexName != "" {
			if !indexMatch(item, input) {
				continue
			}
		} else {
			pk, _ := item["PK"].(string)
			if pk != input.PartitionKey {
				continue
			}
			if input.ExclusiveStartKey != nil && input.ExclusiveStartKey.SortKey != "" {
				sk, _ := item["SK"].(string)
				if sk <= input.ExclusiveStartKey.SortKey {
					continue
				}
			}
			if !filtersMatch(item, input.Filters) {
				continue
			}
		}
		matched = append(matched, cloneItem(item))
	}

	sort.Slice(matched, func(i, j int) bool {
		if input.IndexName == IndexMachineID {
			return fmt.Sprint(matched[i]["MachineID"]) < fmt.Sprint(matched[j]["MachineID"])
		}
		return fmt.Sprint(matched[i]["SK"]) < fmt.Sprint(matched[j]["SK"])
	})

	scanned := int32(len(matched))
	if input.Limit > 0 && int32(len(matched)) > input.Limit {
		matched = matched[:input.Limit]
	}
	outItems := make([]Item, len(matched))
	for i, item := range matched {
		outItems[i] = projectItem(item, input.Projection)
	}
	out := &QueryOutput{
		Items:        outItems,
		Count:        int32(len(outItems)),
		ScannedCount: scanned,
	}
	if input.Limit > 0 && scanned > input.Limit && len(matched) > 0 {
		if pk, ok := primaryKeyFromItem(matched[len(matched)-1]); ok {
			out.LastEvaluatedKey = Item{"PK": pk.PartitionKey, "SK": pk.SortKey}
		}
	} else if input.Limit > 0 && int32(len(matched)) == input.Limit && scanned == input.Limit {
		// The page is full. There may be further documents; surface a cursor
		// so the caller continues. The following page is empty when this was
		// the last page.
		if pk, ok := primaryKeyFromItem(matched[len(matched)-1]); ok {
			out.LastEvaluatedKey = Item{"PK": pk.PartitionKey, "SK": pk.SortKey}
		}
	}
	return out, nil
}

func indexMatch(item Item, input *QueryInput) bool {
	switch input.IndexName {
	case IndexSerialNum:
		serial, _ := item["SerialNum"].(string)
		dt, _ := item["DataType"].(string)
		return serial == input.SerialNumber && dt == input.DataType
	case IndexMachineID:
		dt, _ := item["DataType"].(string)
		if dt != input.DataType {
			return false
		}
		if input.MachineIDPrefix == "" {
			return true
		}
		id, _ := item["MachineID"].(string)
		return len(id) >= len(input.MachineIDPrefix) && id[:len(input.MachineIDPrefix)] == input.MachineIDPrefix
	default:
		return false
	}
}

func (m *Memory) Scan(in *ScanInput) (*ScanOutput, error) {
	if in == nil {
		in = &ScanInput{}
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	all := make([]Item, 0, len(m.docs))
	for _, item := range m.docs {
		if in.ExclusiveStartKey != nil {
			pk, _ := item["PK"].(string)
			sk, _ := item["SK"].(string)
			if pk < in.ExclusiveStartKey.PartitionKey {
				continue
			}
			if pk == in.ExclusiveStartKey.PartitionKey && sk <= in.ExclusiveStartKey.SortKey {
				continue
			}
		}
		all = append(all, cloneItem(item))
	}
	sort.Slice(all, func(i, j int) bool {
		pi, _ := all[i]["PK"].(string)
		pj, _ := all[j]["PK"].(string)
		if pi != pj {
			return pi < pj
		}
		si, _ := all[i]["SK"].(string)
		sj, _ := all[j]["SK"].(string)
		return si < sj
	})
	scanned := int32(len(all))
	limit := in.Limit
	if limit <= 0 {
		limit = scanned
	}
	more := int32(len(all)) > limit
	if int32(len(all)) > limit {
		all = all[:limit]
	}
	out := &ScanOutput{
		Items:        all,
		Count:        int32(len(all)),
		ScannedCount: scanned,
	}
	if (more || (in.Limit > 0 && int32(len(all)) == in.Limit && scanned == in.Limit)) && len(all) > 0 {
		if pk, ok := primaryKeyFromItem(all[len(all)-1]); ok {
			out.LastEvaluatedKey = Item{"PK": pk.PartitionKey, "SK": pk.SortKey}
		}
	}
	return out, nil
}

func (m *Memory) TransactWriteItems(items []TransactWriteItem, idempotencyToken *string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if idempotencyToken != nil && *idempotencyToken != "" {
		if _, ok := m.idem[*idempotencyToken]; ok {
			return nil
		}
	}
	// Validate every operation before mutating so a failure leaves the store unchanged.
	type prepared struct {
		key  string
		del  bool
		item Item
	}
	preparedItems := make([]prepared, 0, len(items))
	for _, op := range items {
		switch {
		case op.Put != nil:
			encoded, pk, err := encodeStored(op.Put)
			if err != nil {
				return err
			}
			preparedItems = append(preparedItems, prepared{key: docKey(pk.PartitionKey, pk.SortKey), item: encoded})
		case op.Update != nil:
			fields, err := MarshalMap(op.Update.Fields)
			if err != nil {
				return err
			}
			key := docKey(op.Update.Key.PartitionKey, op.Update.Key.SortKey)
			base := Item{}
			if existing, ok := m.docs[key]; ok {
				base = cloneItem(existing)
			}
			for k, v := range fields {
				base[k] = v
			}
			base["PK"] = op.Update.Key.PartitionKey
			base["SK"] = op.Update.Key.SortKey
			applyTTL(base)
			preparedItems = append(preparedItems, prepared{key: key, item: base})
		case op.Delete != nil:
			preparedItems = append(preparedItems, prepared{
				key: docKey(op.Delete.PartitionKey, op.Delete.SortKey),
				del: true,
			})
		default:
			return fmt.Errorf("store: empty transact write item")
		}
	}
	for _, p := range preparedItems {
		if p.del {
			delete(m.docs, p.key)
			continue
		}
		m.docs[p.key] = p.item
	}
	if idempotencyToken != nil && *idempotencyToken != "" {
		m.idem[*idempotencyToken] = struct{}{}
	}
	return nil
}

func (m *Memory) CreateTransactPutItem(item interface{}) (TransactWriteItem, error) {
	if _, err := MarshalMap(item); err != nil {
		return TransactWriteItem{}, err
	}
	return TransactWriteItem{Put: item}, nil
}

func (m *Memory) CreateTransactUpdateItem(key PrimaryKey, item interface{}) (TransactWriteItem, error) {
	if _, err := MarshalMap(item); err != nil {
		return TransactWriteItem{}, err
	}
	return TransactWriteItem{Update: &TransactUpdate{Key: key, Fields: item}}, nil
}

func (m *Memory) CreateTransactDeleteItem(key PrimaryKey) (TransactWriteItem, error) {
	return TransactWriteItem{Delete: &key}, nil
}

func encodeStored(item interface{}) (Item, PrimaryKey, error) {
	encoded, err := MarshalMap(item)
	if err != nil {
		return nil, PrimaryKey{}, err
	}
	pk, ok := primaryKeyFromItem(encoded)
	if !ok || pk.PartitionKey == "" || pk.SortKey == "" {
		return nil, PrimaryKey{}, fmt.Errorf("store: item is missing PK or SK")
	}
	applyTTL(encoded)
	return encoded, pk, nil
}
