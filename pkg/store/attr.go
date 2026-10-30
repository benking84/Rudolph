package store

import (
	"fmt"
	"reflect"
	"strconv"
	"time"
)

// Marshaler converts a value into a Firestore-friendly scalar.
type Marshaler interface {
	MarshalStore() (interface{}, error)
}

// Unmarshaler decodes a Firestore-friendly scalar into the receiver.
type Unmarshaler interface {
	UnmarshalStore(v interface{}) error
}

// MarshalMap encodes a struct (or pointer to a struct) into an Item using
// `dynamodbav` field names. Anonymous embedded structs are flattened.
// Zero values are omitted when the field tag contains omitempty.
func MarshalMap(in interface{}) (Item, error) {
	if in == nil {
		return Item{}, nil
	}
	v := reflect.ValueOf(in)
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return Item{}, nil
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("store: MarshalMap expects a struct, got %s", v.Kind())
	}
	return marshalStruct(v)
}

// UnmarshalMap decodes an Item into out, which must be a pointer to a struct
// or a pointer to a pointer to a struct (the latter is allocated).
func UnmarshalMap(item Item, out interface{}) error {
	if out == nil {
		return fmt.Errorf("store: UnmarshalMap out is nil")
	}
	rv := reflect.ValueOf(out)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("store: UnmarshalMap out must be a non-nil pointer")
	}
	for rv.Elem().Kind() == reflect.Ptr {
		if rv.Elem().IsNil() {
			rv.Elem().Set(reflect.New(rv.Elem().Type().Elem()))
		}
		rv = rv.Elem()
	}
	if rv.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("store: UnmarshalMap out must point at a struct, got %s", rv.Elem().Kind())
	}
	return unmarshalStruct(item, rv.Elem())
}

// UnmarshalListOfMaps decodes a slice of items into a slice or a pointer to a
// slice. The element type may be a struct or a pointer to a struct.
// An empty input yields an empty (non-nil) slice.
func UnmarshalListOfMaps(items []Item, out interface{}) error {
	if out == nil {
		return fmt.Errorf("store: UnmarshalListOfMaps out is nil")
	}
	rv := reflect.ValueOf(out)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("store: UnmarshalListOfMaps out must be a non-nil pointer")
	}
	for rv.Elem().Kind() == reflect.Ptr {
		if rv.Elem().IsNil() {
			rv.Elem().Set(reflect.New(rv.Elem().Type().Elem()))
		}
		rv = rv.Elem()
	}
	slice := rv.Elem()
	if slice.Kind() != reflect.Slice {
		return fmt.Errorf("store: UnmarshalListOfMaps out must be a slice, got %s", slice.Kind())
	}
	elemType := slice.Type().Elem()
	result := reflect.MakeSlice(slice.Type(), 0, len(items))
	for _, item := range items {
		var elem reflect.Value
		if elemType.Kind() == reflect.Ptr {
			elem = reflect.New(elemType.Elem())
			if err := unmarshalStruct(item, elem.Elem()); err != nil {
				return err
			}
		} else {
			elem = reflect.New(elemType).Elem()
			if err := unmarshalStruct(item, elem); err != nil {
				return err
			}
		}
		result = reflect.Append(result, elem)
	}
	slice.Set(result)
	return nil
}

func marshalStruct(v reflect.Value) (Item, error) {
	item := Item{}
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		fv := v.Field(i)
		if f.Anonymous {
			inner := fv
			if inner.Kind() == reflect.Ptr {
				if inner.IsNil() {
					continue
				}
				inner = inner.Elem()
			}
			if inner.Kind() == reflect.Struct {
				nested, err := marshalStruct(inner)
				if err != nil {
					return nil, err
				}
				for k, val := range nested {
					item[k] = val
				}
				continue
			}
		}
		if f.PkgPath != "" {
			continue
		}
		name, omit := parseTag(f)
		if name == "-" {
			continue
		}
		if omit && fv.IsZero() {
			continue
		}
		val, err := marshalValue(fv)
		if err != nil {
			return nil, fmt.Errorf("store: field %s: %w", name, err)
		}
		if val == nil && omit {
			continue
		}
		item[name] = val
	}
	return item, nil
}

func marshalValue(fv reflect.Value) (interface{}, error) {
	if !fv.IsValid() {
		return nil, nil
	}
	if fv.Kind() == reflect.Ptr {
		if fv.IsNil() {
			return nil, nil
		}
		return marshalValue(fv.Elem())
	}
	if fv.CanInterface() {
		if m, ok := fv.Interface().(Marshaler); ok {
			return m.MarshalStore()
		}
	}
	if fv.CanAddr() {
		if m, ok := fv.Addr().Interface().(Marshaler); ok {
			return m.MarshalStore()
		}
	}
	switch fv.Kind() {
	case reflect.String:
		return fv.String(), nil
	case reflect.Bool:
		return fv.Bool(), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fv.Int(), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return int64(fv.Uint()), nil
	case reflect.Float32, reflect.Float64:
		return fv.Float(), nil
	default:
		return nil, fmt.Errorf("unsupported kind %s", fv.Kind())
	}
}

func unmarshalStruct(item Item, v reflect.Value) error {
	if !v.IsValid() {
		return fmt.Errorf("store: invalid unmarshal target")
	}
	fields := map[string][]int{}
	collectFields(v.Type(), nil, fields)
	for name, raw := range item {
		if name == ExpireAtField {
			continue
		}
		index, ok := fields[name]
		if !ok {
			continue
		}
		fv := fieldByIndex(v, index)
		if err := unmarshalValue(raw, fv); err != nil {
			return fmt.Errorf("store: field %s: %w", name, err)
		}
	}
	return nil
}

func collectFields(t reflect.Type, prefix []int, out map[string][]int) {
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		idx := append(append([]int{}, prefix...), i)
		ft := f.Type
		if f.Anonymous {
			if ft.Kind() == reflect.Ptr {
				ft = ft.Elem()
			}
			if ft.Kind() == reflect.Struct {
				collectFields(ft, idx, out)
				continue
			}
		}
		if f.PkgPath != "" {
			continue
		}
		name, _ := parseTag(f)
		if name == "-" || name == "" {
			continue
		}
		out[name] = idx
	}
}

func fieldByIndex(v reflect.Value, index []int) reflect.Value {
	for _, i := range index {
		for v.Kind() == reflect.Ptr {
			if v.IsNil() {
				v.Set(reflect.New(v.Type().Elem()))
			}
			v = v.Elem()
		}
		v = v.Field(i)
	}
	return v
}

func unmarshalValue(raw interface{}, fv reflect.Value) error {
	if fv.Kind() == reflect.Ptr {
		if raw == nil {
			fv.Set(reflect.Zero(fv.Type()))
			return nil
		}
		if fv.IsNil() {
			fv.Set(reflect.New(fv.Type().Elem()))
		}
		return unmarshalValue(raw, fv.Elem())
	}
	if fv.CanAddr() {
		if u, ok := fv.Addr().Interface().(Unmarshaler); ok {
			return u.UnmarshalStore(raw)
		}
	}
	if raw == nil {
		fv.Set(reflect.Zero(fv.Type()))
		return nil
	}
	switch fv.Kind() {
	case reflect.String:
		s, ok := raw.(string)
		if !ok {
			return fmt.Errorf("expected string, got %T", raw)
		}
		fv.SetString(s)
		return nil
	case reflect.Bool:
		b, ok := raw.(bool)
		if !ok {
			return fmt.Errorf("expected bool, got %T", raw)
		}
		fv.SetBool(b)
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := asInt64(raw)
		if err != nil {
			return err
		}
		if fv.OverflowInt(n) {
			return fmt.Errorf("integer overflow")
		}
		fv.SetInt(n)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := asInt64(raw)
		if err != nil {
			return err
		}
		if n < 0 {
			return fmt.Errorf("negative value for unsigned field")
		}
		fv.SetUint(uint64(n))
		return nil
	case reflect.Float32, reflect.Float64:
		switch n := raw.(type) {
		case float64:
			fv.SetFloat(n)
		case float32:
			fv.SetFloat(float64(n))
		case int64:
			fv.SetFloat(float64(n))
		case int:
			fv.SetFloat(float64(n))
		default:
			return fmt.Errorf("expected float, got %T", raw)
		}
		return nil
	default:
		return fmt.Errorf("unsupported kind %s", fv.Kind())
	}
}

func parseTag(f reflect.StructField) (name string, omitempty bool) {
	tag := f.Tag.Get("dynamodbav")
	if tag == "" {
		return f.Name, false
	}
	name = tag
	if i := indexComma(tag); i >= 0 {
		name = tag[:i]
		omitempty = tag[i+1:] == "omitempty" || hasOption(tag[i+1:], "omitempty")
	}
	if name == "" {
		name = f.Name
	}
	return name, omitempty
}

func indexComma(s string) int {
	for i := 0; i < len(s); i++ {
		if s[i] == ',' {
			return i
		}
	}
	return -1
}

func hasOption(opts, want string) bool {
	for opts != "" {
		opt := opts
		if i := indexComma(opts); i >= 0 {
			opt = opts[:i]
			opts = opts[i+1:]
		} else {
			opts = ""
		}
		if opt == want {
			return true
		}
	}
	return false
}

func asInt64(raw interface{}) (int64, error) {
	switch n := raw.(type) {
	case int:
		return int64(n), nil
	case int8:
		return int64(n), nil
	case int16:
		return int64(n), nil
	case int32:
		return int64(n), nil
	case int64:
		return n, nil
	case uint:
		return int64(n), nil
	case uint32:
		return int64(n), nil
	case uint64:
		return int64(n), nil
	case float32:
		return int64(n), nil
	case float64:
		return int64(n), nil
	case string:
		v, err := strconv.ParseInt(n, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("expected integer, got %q", n)
		}
		return v, nil
	default:
		return 0, fmt.Errorf("expected integer, got %T", raw)
	}
}

func applyTTL(item Item) {
	raw, ok := item["ExpiresAfter"]
	if !ok {
		delete(item, ExpireAtField)
		return
	}
	sec, err := asInt64(raw)
	if err != nil || sec <= 0 {
		delete(item, ExpireAtField)
		return
	}
	item[ExpireAtField] = time.Unix(sec, 0).UTC()
}

func cloneItem(in Item) Item {
	if in == nil {
		return nil
	}
	out := make(Item, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func projectItem(in Item, fields []string) Item {
	if len(fields) == 0 {
		return cloneItem(in)
	}
	out := make(Item, len(fields))
	for _, f := range fields {
		if v, ok := in[f]; ok {
			out[f] = v
		}
	}
	return out
}

func filtersMatch(item Item, filters map[string]interface{}) bool {
	for k, want := range filters {
		got, ok := item[k]
		if !ok {
			return false
		}
		if !valuesEqual(got, want) {
			return false
		}
	}
	return true
}

func valuesEqual(a, b interface{}) bool {
	ai, aerr := asInt64(a)
	bi, berr := asInt64(b)
	if aerr == nil && berr == nil {
		return ai == bi
	}
	return reflect.DeepEqual(a, b)
}

func primaryKeyFromItem(item Item) (PrimaryKey, bool) {
	pk, ok1 := item["PK"].(string)
	sk, ok2 := item["SK"].(string)
	if !ok1 || !ok2 {
		return PrimaryKey{}, false
	}
	return PrimaryKey{PartitionKey: pk, SortKey: sk}, true
}
