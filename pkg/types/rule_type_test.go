package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRuleType_MarshalText(t *testing.T) {
	tests := []struct {
		name    string
		rule    RuleType
		want    []byte
		wantErr bool
	}{
		{"Binary", RuleTypeBinary, []byte("BINARY"), false},
		{"Certificate", RuleTypeCertificate, []byte("CERTIFICATE"), false},
		{"SigningID", RuleTypeSigningID, []byte("SIGNINGID"), false},
		{"TeamID", RuleTypeTeamID, []byte("TEAMID"), false},
		{"Invalid", RuleType(0), nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.rule.MarshalText()
			if (err != nil) != tt.wantErr {
				t.Errorf("RuleType.MarshalText() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRuleType_UnmarshalText(t *testing.T) {
	tests := []struct {
		name    string
		text    []byte
		want    RuleType
		wantErr bool
	}{
		{"Binary", []byte("BINARY"), RuleTypeBinary, false},
		{"Certificate", []byte("CERTIFICATE"), RuleTypeCertificate, false},
		{"SigningID", []byte("SIGNINGID"), RuleTypeSigningID, false},
		{"TeamID", []byte("TEAMID"), RuleTypeTeamID, false},
		{"Invalid", []byte("INVALID"), RuleType(0), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got RuleType
			err := got.UnmarshalText(tt.text)
			if (err != nil) != tt.wantErr {
				t.Errorf("RuleType.UnmarshalText() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRuleType_MarshalStore(t *testing.T) {
	tests := []struct {
		name     string
		ruleType RuleType
		want     interface{}
		wantErr  bool
	}{
		{"BINARY", RuleTypeBinary, int64(1), false},
		{"CERTIFICATE", RuleTypeCertificate, int64(2), false},
		{"SIGNINGID", RuleTypeSigningID, int64(3), false},
		{"TEAMID", RuleTypeTeamID, int64(4), false},
		{"INVALID", RuleType(0), nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			av, err := tt.ruleType.MarshalStore()
			if (err != nil) != tt.wantErr {
				t.Errorf("RuleType.MarshalStore() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, av)
		})
	}
}

func TestRuleType_UnmarshalStore(t *testing.T) {
	tests := []struct {
		name    string
		av      interface{}
		want    RuleType
		wantErr bool
	}{
		{"BINARY", int64(1), RuleTypeBinary, false},
		{"CERTIFICATE string", "2", RuleTypeCertificate, false},
		{"SIGNINGID", int64(3), RuleTypeSigningID, false},
		{"TEAMID name", "TEAMID", RuleTypeTeamID, false},
		{"INVALID", "nope", RuleType(0), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RuleType(0)
			err := got.UnmarshalStore(tt.av)
			if (err != nil) != tt.wantErr {
				t.Errorf("RuleType.UnmarshalStore() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
