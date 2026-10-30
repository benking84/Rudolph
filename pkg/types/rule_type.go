package types

import (
	"fmt"
	"strconv"
)

// RuleType represents a Santa rule type.
type RuleType int

const (
	// @deprecated
	Binary = RuleTypeBinary

	// @deprecated
	Certificate = RuleTypeCertificate
)

const (
	// 	Most Specific                                  Least Specific
	// Binary   -->   Signing ID   -->   Certificate   -->   Team ID

	// Binary rules use the SHA-256 hash of the entire binary as an identifier.
	RuleTypeBinary RuleType = iota + 1

	// Certificate rules are formed from the SHA-256 fingerprint of an X.509 leaf signing certificate.
	// This is a powerful rule type that has a much broader reach than an individual binary rule .
	// A signing certificate can sign any number of binaries.
	RuleTypeCertificate

	// SigningID rules are arbitrary identifiers under developer control that are given to a binary at signing time.
	// Typically, these use reverse domain name notation and include the name of the binary (e.g. com.google.Chrome).
	// Because the signing IDs are arbitrary, the Santa rule identifier must be prefixed with the Team ID associated with the Apple developer certificate used to sign the application.
	// For example, a signing ID rule for Google Chrome would be: EQHXZ8M8AV:com.google.Chrome.
	//For platform binaries (i.e. those binaries shipped by Apple with the OS) which do not have a Team ID, the string platform must be used (e.g. platform:com.apple.curl).
	RuleTypeSigningID

	// TeamID rules are formed from the Apple Developer Program Team ID is a 10-character identifier issued by Apple and tied to developer accounts/organizations.
	// This is distinct from Certificates, as a single developer account can and frequently will request/rotate between multiple different signing certificates and entitlements.
	// This is an even more powerful rule with broader reach than individual certificate rules.
	RuleTypeTeamID
)

// UnmarshalText for JSON marshalling interface
func (r *RuleType) UnmarshalText(text []byte) error {
	switch t := string(text); t {
	case "BINARY":
		*r = RuleTypeBinary
	case "CERTIFICATE":
		*r = RuleTypeCertificate
	case "SIGNINGID":
		*r = RuleTypeSigningID
	case "TEAMID":
		*r = RuleTypeTeamID
	default:
		return fmt.Errorf("unknown rule_type value %q", t)
	}
	return nil
}

// MarshalText for JSON marshalling interface
func (r RuleType) MarshalText() ([]byte, error) {
	switch r {
	case RuleTypeBinary:
		return []byte("BINARY"), nil
	case RuleTypeCertificate:
		return []byte("CERTIFICATE"), nil
	case RuleTypeSigningID:
		return []byte("SIGNINGID"), nil
	case RuleTypeTeamID:
		return []byte("TEAMID"), nil
	default:
		return nil, fmt.Errorf("unknown rule_type %d", r)
	}
}

// MarshalStore stores the rule type as an integer. Numeric strings are accepted
// on read so documents migrated from DynamoDB still decode.
func (r RuleType) MarshalStore() (interface{}, error) {
	switch r {
	case RuleTypeBinary, RuleTypeCertificate, RuleTypeSigningID, RuleTypeTeamID:
		return int64(r), nil
	default:
		return nil, fmt.Errorf("unknown rule_type value %q", r)
	}
}

// UnmarshalStore decodes an integer or a numeric/name string.
func (r *RuleType) UnmarshalStore(v interface{}) error {
	switch t := v.(type) {
	case int64:
		return r.fromInt(t)
	case int:
		return r.fromInt(int64(t))
	case int32:
		return r.fromInt(int64(t))
	case float64:
		return r.fromInt(int64(t))
	case string:
		switch t {
		case "1", "BINARY":
			*r = RuleTypeBinary
		case "2", "CERTIFICATE":
			*r = RuleTypeCertificate
		case "3", "SIGNINGID":
			*r = RuleTypeSigningID
		case "4", "TEAMID":
			*r = RuleTypeTeamID
		default:
			n, err := strconv.ParseInt(t, 10, 64)
			if err != nil {
				return fmt.Errorf("unknown rule_type value %q", t)
			}
			return r.fromInt(n)
		}
		return nil
	default:
		return fmt.Errorf("unexpected rule_type value type %T", v)
	}
}

func (r *RuleType) fromInt(n int64) error {
	switch RuleType(n) {
	case RuleTypeBinary, RuleTypeCertificate, RuleTypeSigningID, RuleTypeTeamID:
		*r = RuleType(n)
		return nil
	default:
		return fmt.Errorf("unknown rule_type value %d", n)
	}
}
