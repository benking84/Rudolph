package store

import (
	"context"
	"fmt"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// FirestoreClient stores Rudolph documents in a Firestore Native database.
type FirestoreClient struct {
	client     *firestore.Client
	timeout    time.Duration
	partitions *firestore.CollectionRef
}

// New opens a Firestore client. databaseID is the Firestore database id
// (for example "rudolph"), not a collection name.
func New(ctx context.Context, projectID, databaseID string) (*FirestoreClient, error) {
	if projectID == "" {
		return nil, fmt.Errorf("store: GCP project id is required")
	}
	if databaseID == "" {
		return nil, fmt.Errorf("store: Firestore database id is required")
	}
	client, err := firestore.NewClientWithDatabase(ctx, projectID, databaseID)
	if err != nil {
		return nil, fmt.Errorf("store: firestore client: %w", err)
	}
	return &FirestoreClient{
		client:     client,
		timeout:    defaultTimeout,
		partitions: client.Collection("partitions"),
	}, nil
}

// FromEnv opens a client using GCP_PROJECT (or GOOGLE_CLOUD_PROJECT) and
// FIRESTORE_DATABASE (default "rudolph").
func FromEnv() (*FirestoreClient, error) {
	project := os.Getenv("GCP_PROJECT")
	if project == "" {
		project = os.Getenv("GOOGLE_CLOUD_PROJECT")
	}
	database := os.Getenv("FIRESTORE_DATABASE")
	if database == "" {
		database = "rudolph"
	}
	return New(context.Background(), project, database)
}

// Close releases the underlying client.
func (c *FirestoreClient) Close() error {
	return c.client.Close()
}

func (c *FirestoreClient) withTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), c.timeout)
}

func (c *FirestoreClient) doc(pk, sk string) *firestore.DocumentRef {
	return c.partitions.Doc(EncodeDocID(pk)).Collection("items").Doc(EncodeDocID(sk))
}

func (c *FirestoreClient) GetItem(key PrimaryKey, _ bool) (*GetItemOutput, error) {
	ctx, cancel := c.withTimeout()
	defer cancel()
	snap, err := c.doc(key.PartitionKey, key.SortKey).Get(ctx)
	if status.Code(err) == codes.NotFound {
		return &GetItemOutput{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("store: get %s/%s: %w", key.PartitionKey, key.SortKey, err)
	}
	return &GetItemOutput{Item: itemFromSnap(snap)}, nil
}

func (c *FirestoreClient) PutItem(item interface{}) (*PutItemOutput, error) {
	encoded, pk, err := encodeStored(item)
	if err != nil {
		return nil, err
	}
	ctx, cancel := c.withTimeout()
	defer cancel()
	if _, err := c.doc(pk.PartitionKey, pk.SortKey).Set(ctx, map[string]interface{}(encoded)); err != nil {
		return nil, fmt.Errorf("store: put %s/%s: %w", pk.PartitionKey, pk.SortKey, err)
	}
	return &PutItemOutput{}, nil
}

func (c *FirestoreClient) UpdateItem(key PrimaryKey, item interface{}) (*UpdateItemOutput, error) {
	fields, err := MarshalMap(item)
	if err != nil {
		return nil, err
	}
	ref := c.doc(key.PartitionKey, key.SortKey)
	ctx, cancel := c.withTimeout()
	defer cancel()
	var updated Item
	err = c.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		snap, err := tx.Get(ref)
		if status.Code(err) == codes.NotFound {
			return fmt.Errorf("store: conditional update failed for %s/%s", key.PartitionKey, key.SortKey)
		}
		if err != nil {
			return err
		}
		merged := itemFromSnap(snap)
		for k, v := range fields {
			merged[k] = v
		}
		merged["PK"] = key.PartitionKey
		merged["SK"] = key.SortKey
		applyTTL(merged)
		updated = cloneItem(merged)
		return tx.Set(ref, map[string]interface{}(merged))
	})
	if err != nil {
		return nil, err
	}
	return &UpdateItemOutput{Attributes: updated}, nil
}

func (c *FirestoreClient) DeleteItem(key PrimaryKey) (*DeleteItemOutput, error) {
	ctx, cancel := c.withTimeout()
	defer cancel()
	_, err := c.doc(key.PartitionKey, key.SortKey).Delete(ctx)
	if err != nil {
		return nil, fmt.Errorf("store: delete %s/%s: %w", key.PartitionKey, key.SortKey, err)
	}
	return &DeleteItemOutput{}, nil
}

func (c *FirestoreClient) Query(input *QueryInput) (*QueryOutput, error) {
	if input == nil {
		return nil, fmt.Errorf("store: nil query")
	}
	ctx, cancel := c.withTimeout()
	defer cancel()

	var q firestore.Query
	if input.IndexName != "" {
		q = c.client.CollectionGroup("items").Query
		switch input.IndexName {
		case IndexSerialNum:
			q = q.Where("SerialNum", "==", input.SerialNumber).Where("DataType", "==", input.DataType)
		case IndexMachineID:
			q = q.Where("DataType", "==", input.DataType).OrderBy("MachineID", firestore.Asc)
			if input.MachineIDPrefix != "" {
				q = q.Where("MachineID", ">=", input.MachineIDPrefix).
					Where("MachineID", "<", prefixUpperBound(input.MachineIDPrefix))
			}
		default:
			return nil, fmt.Errorf("store: unknown index %q", input.IndexName)
		}
	} else {
		q = c.partitions.Doc(EncodeDocID(input.PartitionKey)).Collection("items").OrderBy("SK", firestore.Asc)
		if input.ExclusiveStartKey != nil && input.ExclusiveStartKey.SortKey != "" {
			q = q.Where("SK", ">", input.ExclusiveStartKey.SortKey)
		}
		for field, value := range input.Filters {
			q = q.Where(field, "==", value)
		}
	}
	if input.Limit > 0 {
		q = q.Limit(int(input.Limit))
	}
	snaps, err := q.Documents(ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("store: query: %w", err)
	}
	raw := make([]Item, 0, len(snaps))
	items := make([]Item, 0, len(snaps))
	for _, snap := range snaps {
		full := itemFromSnap(snap)
		raw = append(raw, full)
		items = append(items, projectItem(full, input.Projection))
	}
	out := &QueryOutput{
		Items:        items,
		Count:        int32(len(items)),
		ScannedCount: int32(len(items)),
	}
	if input.Limit > 0 && int32(len(raw)) == input.Limit && len(raw) > 0 {
		if pk, ok := primaryKeyFromItem(raw[len(raw)-1]); ok {
			out.LastEvaluatedKey = Item{"PK": pk.PartitionKey, "SK": pk.SortKey}
		}
	}
	return out, nil
}

func (c *FirestoreClient) Scan(in *ScanInput) (*ScanOutput, error) {
	if in == nil {
		in = &ScanInput{}
	}
	ctx, cancel := c.withTimeout()
	defer cancel()
	q := c.client.CollectionGroup("items").OrderBy("PK", firestore.Asc).OrderBy("SK", firestore.Asc)
	if in.ExclusiveStartKey != nil {
		q = q.StartAfter(in.ExclusiveStartKey.PartitionKey, in.ExclusiveStartKey.SortKey)
	}
	limit := int(in.Limit)
	if limit <= 0 {
		limit = 100
	}
	q = q.Limit(limit)
	snaps, err := q.Documents(ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("store: scan: %w", err)
	}
	items := make([]Item, 0, len(snaps))
	for _, snap := range snaps {
		items = append(items, itemFromSnap(snap))
	}
	out := &ScanOutput{
		Items:        items,
		Count:        int32(len(items)),
		ScannedCount: int32(len(items)),
	}
	if len(items) == limit && len(items) > 0 {
		if pk, ok := primaryKeyFromItem(items[len(items)-1]); ok {
			out.LastEvaluatedKey = Item{"PK": pk.PartitionKey, "SK": pk.SortKey}
		}
	}
	return out, nil
}

func (c *FirestoreClient) TransactWriteItems(items []TransactWriteItem, idempotencyToken *string) error {
	ctx, cancel := c.withTimeout()
	defer cancel()
	return c.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		var idemRef *firestore.DocumentRef
		if idempotencyToken != nil && *idempotencyToken != "" {
			idemRef = c.client.Collection("idempotency").Doc(EncodeDocID(*idempotencyToken))
			_, err := tx.Get(idemRef)
			if err == nil {
				return nil
			}
			if status.Code(err) != codes.NotFound {
				return err
			}
		}

		// Firestore transactions require every read to happen before any write.
		type prepared struct {
			ref  *firestore.DocumentRef
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
				preparedItems = append(preparedItems, prepared{
					ref:  c.doc(pk.PartitionKey, pk.SortKey),
					item: encoded,
				})
			case op.Update != nil:
				fields, err := MarshalMap(op.Update.Fields)
				if err != nil {
					return err
				}
				ref := c.doc(op.Update.Key.PartitionKey, op.Update.Key.SortKey)
				base := Item{}
				snap, err := tx.Get(ref)
				if err == nil {
					base = itemFromSnap(snap)
				} else if status.Code(err) != codes.NotFound {
					return err
				}
				for k, v := range fields {
					base[k] = v
				}
				base["PK"] = op.Update.Key.PartitionKey
				base["SK"] = op.Update.Key.SortKey
				applyTTL(base)
				preparedItems = append(preparedItems, prepared{ref: ref, item: base})
			case op.Delete != nil:
				preparedItems = append(preparedItems, prepared{
					ref: c.doc(op.Delete.PartitionKey, op.Delete.SortKey),
					del: true,
				})
			default:
				return fmt.Errorf("store: empty transact write item")
			}
		}

		if idemRef != nil {
			if err := tx.Set(idemRef, map[string]interface{}{"created": time.Now().UTC()}); err != nil {
				return err
			}
		}
		for _, p := range preparedItems {
			if p.del {
				if err := tx.Delete(p.ref); err != nil {
					return err
				}
				continue
			}
			if err := tx.Set(p.ref, map[string]interface{}(p.item)); err != nil {
				return err
			}
		}
		return nil
	})
}

func (c *FirestoreClient) CreateTransactPutItem(item interface{}) (TransactWriteItem, error) {
	if _, err := MarshalMap(item); err != nil {
		return TransactWriteItem{}, err
	}
	return TransactWriteItem{Put: item}, nil
}

func (c *FirestoreClient) CreateTransactUpdateItem(key PrimaryKey, item interface{}) (TransactWriteItem, error) {
	if _, err := MarshalMap(item); err != nil {
		return TransactWriteItem{}, err
	}
	return TransactWriteItem{Update: &TransactUpdate{Key: key, Fields: item}}, nil
}

func (c *FirestoreClient) CreateTransactDeleteItem(key PrimaryKey) (TransactWriteItem, error) {
	return TransactWriteItem{Delete: &key}, nil
}

func itemFromSnap(snap *firestore.DocumentSnapshot) Item {
	if snap == nil || snap.Data() == nil {
		return Item{}
	}
	raw := snap.Data()
	item := make(Item, len(raw))
	for k, v := range raw {
		item[k] = normalizeFirestoreValue(v)
	}
	return item
}

func normalizeFirestoreValue(v interface{}) interface{} {
	switch t := v.(type) {
	case int:
		return int64(t)
	case time.Time:
		return t.UTC()
	default:
		return v
	}
}
