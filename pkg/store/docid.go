package store

import "encoding/base64"

// EncodeDocID returns a Firestore document id that round-trips any PK or SK.
// Sort order of document ids is not the sort order of the original string;
// range queries order by the SK field instead.
func EncodeDocID(s string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(s))
}

// DecodeDocID reverses EncodeDocID.
func DecodeDocID(s string) (string, error) {
	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// prefixUpperBound is the exclusive end of a string prefix match.
// Machine IDs are UUIDs, which do not contain U+F8FF.
func prefixUpperBound(prefix string) string {
	return prefix + "\uf8ff"
}
