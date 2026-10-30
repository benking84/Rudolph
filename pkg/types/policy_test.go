package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPolicyTypes_MarshallText(t *testing.T) {
	tests := []struct {
		name    string
		policy  Policy
		want    []byte
		wantErr bool
	}{
		{"ALLOWLIST", RulePolicyAllowlist, []byte("ALLOWLIST"), false},
		{"BLOCKLIST", RulePolicyBlocklist, []byte("BLOCKLIST"), false},
		{"SILENT_BLOCKLIST", RulePolicySilentBlocklist, []byte("SILENT_BLOCKLIST"), false},
		{"REMOVE", RulePolicyRemove, []byte("REMOVE"), false},
		{"ALLOWLIST_COMPILER", RulePolicyAllowlistCompiler, []byte("ALLOWLIST_COMPILER"), false},
		{"ALLOWLIST_TRANSITIVE", RulePolicyAllowlistTransitive, []byte("ALLOWLIST_TRANSITIVE"), false},
		{"MISSPELLED", Policy(99999), nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.policy.MarshalText()
			if (err != nil) != tt.wantErr {
				t.Errorf("Policy.MarshalText() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPolicyTypes_UnmarshallText(t *testing.T) {
	tests := []struct {
		name    string
		text    []byte
		want    Policy
		wantErr bool
	}{
		{"ALLOWLIST", []byte("ALLOWLIST"), RulePolicyAllowlist, false},
		{"BLOCKLIST", []byte("BLOCKLIST"), RulePolicyBlocklist, false},
		{"SILENT_BLOCKLIST", []byte("SILENT_BLOCKLIST"), RulePolicySilentBlocklist, false},
		{"REMOVE", []byte("REMOVE"), RulePolicyRemove, false},
		{"ALLOWLIST_COMPILER", []byte("ALLOWLIST_COMPILER"), RulePolicyAllowlistCompiler, false},
		{"ALLOWLIST_TRANSITIVE", []byte("ALLOWLIST_TRANSITIVE"), RulePolicyAllowlistTransitive, false},
		{"MISSPELLED", []byte("MISSPELLED"), Policy(0), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got Policy
			err := got.UnmarshalText(tt.text)
			if (err != nil) != tt.wantErr {
				t.Errorf("Policy.UnmarshalText() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPolicyTypes_MarshalStore(t *testing.T) {
	tests := []struct {
		name    string
		policy  Policy
		want    interface{}
		wantErr bool
	}{
		{"ALLOWLIST", RulePolicyAllowlist, int64(1), false},
		{"BLOCKLIST", RulePolicyBlocklist, int64(2), false},
		{"SILENT_BLOCKLIST", RulePolicySilentBlocklist, int64(3), false},
		{"REMOVE", RulePolicyRemove, int64(4), false},
		{"ALLOWLIST_COMPILER", RulePolicyAllowlistCompiler, int64(5), false},
		{"ALLOWLIST_TRANSITIVE", RulePolicyAllowlistTransitive, int64(6), false},
		{"MISSPELLED", Policy(0), nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			av, err := tt.policy.MarshalStore()
			if (err != nil) != tt.wantErr {
				t.Errorf("Policy.MarshalStore() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, av)
		})
	}
}

func TestPolicyType_UnmarshalStore(t *testing.T) {
	tests := []struct {
		name    string
		av      interface{}
		want    Policy
		wantErr bool
	}{
		{"ALLOWLIST", "1", RulePolicyAllowlist, false},
		{"BLOCKLIST", int64(2), RulePolicyBlocklist, false},
		{"SILENT_BLOCKLIST", int64(3), RulePolicySilentBlocklist, false},
		{"REMOVE", int64(4), RulePolicyRemove, false},
		{"ALLOWLIST_COMPILER", int64(5), RulePolicyAllowlistCompiler, false},
		{"ALLOWLIST_TRANSITIVE", int64(6), RulePolicyAllowlistTransitive, false},
		{"MISSPELLED", "nope", Policy(0), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Policy(0)
			err := got.UnmarshalStore(tt.av)
			if (err != nil) != tt.wantErr {
				t.Errorf("PolicyType.UnmarshalStore() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
