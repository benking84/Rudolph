package types

import "fmt"

// Policy represents the Santa Rule Policy.
type Policy int

const (
	// @deprecated
	Allowlist           = RulePolicyAllowlist
	Blocklist           = RulePolicyBlocklist
	SilentBlocklist     = RulePolicySilentBlocklist
	Remove              = RulePolicyRemove
	AllowlistCompiler   = RulePolicyAllowlistCompiler
	AllowlistTransitive = RulePolicyAllowlistTransitive
)

const (
	RulePolicyAllowlist Policy = iota + 1
	RulePolicyBlocklist
	RulePolicySilentBlocklist
	// Remove is a "special" rule in that, when it is sent by the server, it instructs the sensor
	// to delete any associated rule.
	RulePolicyRemove
	// AllowlistCompiler is a Transitive Allowlist policy which allows binaries created by
	// a specific compiler. EnabledTransitiveRules must be set to true in the Preflight first.
	RulePolicyAllowlistCompiler
	// Transitive rules are created by the santa sensor itself; it is never created by the server.
	// Transitive rules are destroyed upon every clean sync.
	RulePolicyAllowlistTransitive
)

// UnmarshalText for JSON marshalling interface
// Use Santa defined constants
// https://github.com/google/santa/blob/main/Source/santactl/Commands/sync/SNTCommandSyncConstants.m#L98-L109
func (p *Policy) UnmarshalText(text []byte) error {
	switch t := string(text); t {
	case "ALLOWLIST":
		*p = RulePolicyAllowlist
	case "BLOCKLIST":
		*p = RulePolicyBlocklist
	case "SILENT_BLOCKLIST":
		*p = RulePolicySilentBlocklist
	case "REMOVE":
		*p = RulePolicyRemove
	case "ALLOWLIST_COMPILER":
		*p = RulePolicyAllowlistCompiler
	case "ALLOWLIST_TRANSITIVE":
		*p = RulePolicyAllowlistTransitive
	default:
		return fmt.Errorf("unknown policy value %q", t)
	}
	return nil
}

// MarshalText for JSON marshalling interface
func (p Policy) MarshalText() ([]byte, error) {
	switch p {
	case RulePolicyAllowlist:
		return []byte("ALLOWLIST"), nil
	case RulePolicyBlocklist:
		return []byte("BLOCKLIST"), nil
	case RulePolicySilentBlocklist:
		return []byte("SILENT_BLOCKLIST"), nil
	case RulePolicyRemove:
		return []byte("REMOVE"), nil
	case RulePolicyAllowlistCompiler:
		return []byte("ALLOWLIST_COMPILER"), nil
	case RulePolicyAllowlistTransitive:
		return []byte("ALLOWLIST_TRANSITIVE"), nil
	default:
		return nil, fmt.Errorf("unknown policy %d", p)
	}
}

// MarshalStore stores the policy as an integer.
func (p Policy) MarshalStore() (interface{}, error) {
	switch p {
	case RulePolicyAllowlist, RulePolicyBlocklist, RulePolicySilentBlocklist, RulePolicyRemove, RulePolicyAllowlistCompiler, RulePolicyAllowlistTransitive:
		return int64(p), nil
	default:
		return nil, fmt.Errorf("unknown policy value %q", p)
	}
}

// UnmarshalStore decodes an integer or the numeric string DynamoDB used.
func (p *Policy) UnmarshalStore(v interface{}) error {
	var n int64
	switch t := v.(type) {
	case int64:
		n = t
	case int:
		n = int64(t)
	case int32:
		n = int64(t)
	case float64:
		n = int64(t)
	case string:
		switch t {
		case "1":
			n = 1
		case "2":
			n = 2
		case "3":
			n = 3
		case "4":
			n = 4
		case "5":
			n = 5
		case "6":
			n = 6
		default:
			return fmt.Errorf("unknown policy value %q", t)
		}
	default:
		return fmt.Errorf("unexpected policy value type %T", v)
	}
	switch Policy(n) {
	case RulePolicyAllowlist, RulePolicyBlocklist, RulePolicySilentBlocklist, RulePolicyRemove, RulePolicyAllowlistCompiler, RulePolicyAllowlistTransitive:
		*p = Policy(n)
		return nil
	default:
		return fmt.Errorf("unknown policy value %d", n)
	}
}
