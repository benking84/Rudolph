// Package store is the Rudolph persistence layer.
//
// Documents keep the same logical primary key Rudolph used on DynamoDB
// (partition key PK + sort key SK). On Firestore Native the keys are stored
// as fields, and each item lives at:
//
//	partitions/{base64url(PK)}/items/{base64url(SK)}
//
// Queries that used to be DynamoDB Query/GSI calls are collection or
// collection-group queries. ExpiresAfter (unix seconds) is mirrored to a
// timestamp field expireAt so a Firestore TTL policy can delete it.
package store

import "time"

// PrimaryKey is the logical key of every Rudolph document.
type PrimaryKey struct {
	PartitionKey string `dynamodbav:"PK"`
	SortKey      string `dynamodbav:"SK"`
}

// Item is a decoded document. Values are Go natives (string, bool, int64,
// float64) plus time.Time for the TTL mirror field.
type Item map[string]interface{}

// Index names preserved from the DynamoDB global secondary indexes so callers
// and operators can recognise the access pattern. DataType_MachineID was used
// by the CLI but was not created by the upstream Terraform; Firestore indexes
// for both patterns are created by this project's Terraform.
const (
	IndexSerialNum = "SerialNum_DataType_MachineID"
	IndexMachineID = "DataType_MachineID"
)

// ExpireAtField is the Firestore timestamp a TTL policy deletes on.
const ExpireAtField = "expireAt"

const (
	defaultTimeout = 10 * time.Second
)

// GetItemOutput is the result of a point read. Item is empty when the
// document does not exist.
type GetItemOutput struct {
	Item Item
}

// PutItemOutput is returned after a successful overwrite.
type PutItemOutput struct{}

// UpdateItemOutput carries the full document after a partial update.
type UpdateItemOutput struct {
	Attributes Item
}

// DeleteItemOutput is returned after a delete. Deleting a missing document
// succeeds.
type DeleteItemOutput struct{}

// QueryInput describes either a partition query or a secondary-index query.
//
// Partition queries return documents with PK = PartitionKey, ordered by SK.
// ExclusiveStartKey is exclusive: the first returned SK is strictly greater.
// When Limit is set and that many documents were returned, LastEvaluatedKey
// is populated so the caller can ask for the next page (an extra empty page
// is possible when the final page is exactly Limit long, matching DynamoDB).
//
// Filters are equality matches applied to non-key fields. Projection, when
// non-empty, limits the fields copied into each result item.
type QueryInput struct {
	ConsistentRead    bool
	PartitionKey      string
	ExclusiveStartKey *PrimaryKey
	Limit             int32
	Filters           map[string]interface{}
	Projection        []string

	// IndexName selects a collection-group query instead of a partition query.
	// IndexSerialNum matches SerialNumber + DataType.
	// IndexMachineID matches DataType, and MachineIDPrefix when it is set
	// (empty prefix returns every machine of that data type).
	IndexName       string
	SerialNumber    string
	DataType        string
	MachineIDPrefix string
}

// QueryOutput is one page of a query.
type QueryOutput struct {
	Items            []Item
	LastEvaluatedKey Item
	Count            int32
	ScannedCount     int32
}

// ScanInput pages across every document. ExclusiveStartKey is exclusive.
type ScanInput struct {
	Limit             int32
	ExclusiveStartKey *PrimaryKey
	ConsistentRead    bool
}

// ScanOutput is one page of a full scan.
type ScanOutput struct {
	Items            []Item
	LastEvaluatedKey Item
	Count            int32
	ScannedCount     int32
}

// TransactWriteItem is a single operation inside a transaction.
// Exactly one of Put, Update, or Delete is set.
type TransactWriteItem struct {
	Put    interface{}
	Update *TransactUpdate
	Delete *PrimaryKey
}

// TransactUpdate merges Fields onto an existing document. Unlike UpdateItem,
// a missing document is created (DynamoDB TransactWrite Update with no
// condition has the same behaviour).
type TransactUpdate struct {
	Key    PrimaryKey
	Fields interface{}
}

// GetItemAPI reads one document.
type GetItemAPI interface {
	GetItem(key PrimaryKey, consistentRead bool) (*GetItemOutput, error)
}

// PutItemAPI overwrites one document.
type PutItemAPI interface {
	PutItem(item interface{}) (*PutItemOutput, error)
}

// UpdateItemAPI merges fields onto an existing document and returns the
// updated document. The document must already exist.
type UpdateItemAPI interface {
	UpdateItem(key PrimaryKey, item interface{}) (*UpdateItemOutput, error)
}

// DeleteItemAPI removes one document.
type DeleteItemAPI interface {
	DeleteItem(key PrimaryKey) (*DeleteItemOutput, error)
}

// QueryAPI runs a partition or index query.
type QueryAPI interface {
	Query(input *QueryInput) (*QueryOutput, error)
}

// ScanAPI pages over the whole collection group.
type ScanAPI interface {
	Scan(in *ScanInput) (*ScanOutput, error)
}

// TransactWriteItemsAPI applies several writes atomically.
type TransactWriteItemsAPI interface {
	TransactWriteItems(items []TransactWriteItem, idempotencyToken *string) error
	CreateTransactPutItem(item interface{}) (TransactWriteItem, error)
	CreateTransactUpdateItem(key PrimaryKey, item interface{}) (TransactWriteItem, error)
	CreateTransactDeleteItem(key PrimaryKey) (TransactWriteItem, error)
}

// Client is the full persistence API used by the sync service and the CLI.
type Client interface {
	GetItemAPI
	PutItemAPI
	UpdateItemAPI
	DeleteItemAPI
	QueryAPI
	ScanAPI
	TransactWriteItemsAPI
}
