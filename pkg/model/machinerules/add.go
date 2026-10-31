package machinerules

import (
	"errors"
	"time"

	"github.com/airbnb/rudolph/pkg/model/rules"
	"github.com/airbnb/rudolph/pkg/store"
	"github.com/airbnb/rudolph/pkg/types"
)

func AddNewMachineRule(
	client store.PutItemAPI,
	machineID string,
	identifier string,
	ruleType types.RuleType,
	policy types.Policy,
	description string,
	expires time.Time,
) error {
	// Input Validation
	isValid, err := ruleValidation(
		machineID,
		identifier,
		ruleType,
		policy,
		description,
		expires,
	)
	if err != nil {
		return err
	}
	if !isValid {
		return errors.New("no errors occurred during the rule validation check but the provided rule is not valid")
	}

	rule := MachineRuleRow{
		PrimaryKey: store.PrimaryKey{
			PartitionKey: machineRulePK(machineID),
			SortKey:      machineRuleSK(identifier, ruleType),
		},
		Description: description,
		SantaRule: rules.SantaRule{
			RuleType:   ruleType,
			Policy:     policy,
			Identifier: identifier,
		},
		ExpiresAfter: expires.Unix(),
	}

	_, err = client.PutItem(rule)
	if err != nil {
		return err
	}
	return nil
}

func ruleValidation(
	machineID,
	identifier string,
	ruleType types.RuleType,
	policy types.Policy,
	description string,
	expires time.Time,
) (bool, error) {
	// MachineID validation
	err := types.ValidateMachineID(machineID)
	if err != nil {
		return false, err
	}

	// RuleType validation
	_, err = ruleType.MarshalText()
	if err != nil {
		return false, err
	}

	// Identifier validation based on RuleType
	switch ruleType {
	case types.Binary, types.Certificate:
		if !rules.ValidSha256(identifier) {
			return false, errors.New("identifier is not a valid sha256")
		}
	case types.TeamID:
		if !rules.ValidTeamID(identifier) {
			return false, errors.New("identifier is not a valid TeamID")
		}
	case types.SigningID:
		if !rules.ValidSigningID(identifier) {
			return false, errors.New("identifier is not a valid SigningID")
		}
	}

	// RulePolicy validation
	_, err = policy.MarshalText()
	if err != nil {
		return false, err
	}

	// Expires validation - must be a positive time period
	if expires.IsZero() || expires.Unix() <= 0 {
		err = errors.New("expires time is not a positive time")
		return false, err
	}

	// All validations have passed
	return true, nil
}
