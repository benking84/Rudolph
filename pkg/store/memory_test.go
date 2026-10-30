package store

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type row struct {
	PrimaryKey
	Name             string `dynamodbav:"Name"`
	DeleteOnNextSync bool   `dynamodbav:"DeleteOnNextSync,omitempty"`
	ExpiresAfter     int64  `dynamodbav:"ExpiresAfter,omitempty"`
	SerialNum        string `dynamodbav:"SerialNum,omitempty"`
	MachineID        string `dynamodbav:"MachineID,omitempty"`
	DataType         string `dynamodbav:"DataType,omitempty"`
}

func TestMemoryPutGetUpdateQuery(t *testing.T) {
	m := NewMemory()
	err := mustPut(m, row{PrimaryKey: PrimaryKey{PartitionKey: "GlobalRules", SortKey: "Binary#a"}, Name: "a"})
	require.NoError(t, err)
	err = mustPut(m, row{PrimaryKey: PrimaryKey{PartitionKey: "GlobalRules", SortKey: "Binary#b"}, Name: "b"})
	require.NoError(t, err)
	err = mustPut(m, row{PrimaryKey: PrimaryKey{PartitionKey: "GlobalRules", SortKey: "Binary#c"}, Name: "c"})
	require.NoError(t, err)

	got, err := m.GetItem(PrimaryKey{PartitionKey: "GlobalRules", SortKey: "Binary#b"}, true)
	require.NoError(t, err)
	assert.Equal(t, "b", got.Item["Name"])

	missing, err := m.GetItem(PrimaryKey{PartitionKey: "GlobalRules", SortKey: "missing"}, false)
	require.NoError(t, err)
	assert.Empty(t, missing.Item)

	page, err := m.Query(&QueryInput{PartitionKey: "GlobalRules", Limit: 2})
	require.NoError(t, err)
	require.Len(t, page.Items, 2)
	assert.Equal(t, "Binary#a", page.Items[0]["SK"])
	require.NotNil(t, page.LastEvaluatedKey)

	var cursor PrimaryKey
	require.NoError(t, UnmarshalMap(page.LastEvaluatedKey, &cursor))
	page2, err := m.Query(&QueryInput{PartitionKey: "GlobalRules", Limit: 2, ExclusiveStartKey: &cursor})
	require.NoError(t, err)
	require.Len(t, page2.Items, 1)
	assert.Equal(t, "Binary#c", page2.Items[0]["SK"])

	updated, err := m.UpdateItem(PrimaryKey{PartitionKey: "GlobalRules", SortKey: "Binary#a"}, struct {
		Name string `dynamodbav:"Name"`
	}{Name: "a2"})
	require.NoError(t, err)
	assert.Equal(t, "a2", updated.Attributes["Name"])
	_, err = m.UpdateItem(PrimaryKey{PartitionKey: "nope", SortKey: "x"}, struct {
		Name string `dynamodbav:"Name"`
	}{Name: "z"})
	assert.Error(t, err)
}

func TestMemoryFilterTransactTTLAndIndex(t *testing.T) {
	m := NewMemory()
	exp := time.Now().Add(time.Hour).Unix()
	require.NoError(t, mustPut(m, row{
		PrimaryKey:       PrimaryKey{PartitionKey: "MachineRules#m1", SortKey: "Binary#1"},
		DeleteOnNextSync: true,
		ExpiresAfter:     exp,
	}))
	require.NoError(t, mustPut(m, row{
		PrimaryKey: PrimaryKey{PartitionKey: "MachineRules#m1", SortKey: "Binary#2"},
	}))
	filtered, err := m.Query(&QueryInput{
		PartitionKey: "MachineRules#m1",
		Filters:      map[string]interface{}{"DeleteOnNextSync": true},
		Projection:   []string{"PK", "SK"},
	})
	require.NoError(t, err)
	require.Len(t, filtered.Items, 1)
	assert.Equal(t, "Binary#1", filtered.Items[0]["SK"])
	_, hasName := filtered.Items[0]["Name"]
	assert.False(t, hasName)

	stored, err := m.GetItem(PrimaryKey{PartitionKey: "MachineRules#m1", SortKey: "Binary#1"}, false)
	require.NoError(t, err)
	_, ok := stored.Item[ExpireAtField].(time.Time)
	assert.True(t, ok)

	feed := row{PrimaryKey: PrimaryKey{PartitionKey: "RulesFeed", SortKey: "2020#Binary#1"}, Name: "remove"}
	del := PrimaryKey{PartitionKey: "MachineRules#m1", SortKey: "Binary#1"}
	require.NoError(t, m.TransactWriteItems([]TransactWriteItem{
		{Put: feed},
		{Delete: &del},
	}, nil))
	gone, err := m.GetItem(del, false)
	require.NoError(t, err)
	assert.Empty(t, gone.Item)

	token := "idem-1"
	require.NoError(t, m.TransactWriteItems([]TransactWriteItem{{Put: feed}}, &token))
	require.NoError(t, m.TransactWriteItems([]TransactWriteItem{{
		Delete: &feed.PrimaryKey,
	}}, &token))
	still, err := m.GetItem(feed.PrimaryKey, false)
	require.NoError(t, err)
	assert.Equal(t, "remove", still.Item["Name"])

	require.NoError(t, mustPut(m, row{
		PrimaryKey: PrimaryKey{PartitionKey: "Machine#m1", SortKey: "Current"},
		SerialNum:  "C02",
		MachineID:  "m1",
		DataType:   "SensorData",
	}))
	require.NoError(t, mustPut(m, row{
		PrimaryKey: PrimaryKey{PartitionKey: "Machine#m2", SortKey: "Current"},
		SerialNum:  "C03",
		MachineID:  "m2",
		DataType:   "SensorData",
	}))
	bySerial, err := m.Query(&QueryInput{IndexName: IndexSerialNum, SerialNumber: "C02", DataType: "SensorData"})
	require.NoError(t, err)
	require.Len(t, bySerial.Items, 1)
	assert.Equal(t, "m1", bySerial.Items[0]["MachineID"])

	byPrefix, err := m.Query(&QueryInput{IndexName: IndexMachineID, DataType: "SensorData", MachineIDPrefix: "m"})
	require.NoError(t, err)
	assert.Len(t, byPrefix.Items, 2)
}

func TestDocIDRoundTrip(t *testing.T) {
	in := "MachineRules#uuid/with slash"
	out, err := DecodeDocID(EncodeDocID(in))
	require.NoError(t, err)
	assert.Equal(t, in, out)
}

func mustPut(m *Memory, item row) error {
	_, err := m.PutItem(item)
	return err
}
