package rules

import (
	"encoding/json"
	"fmt"
	"github.com/airbnb/rudolph/internal/cli/gcpclient"
	"os"

	"github.com/spf13/cobra"

	"github.com/airbnb/rudolph/internal/csv"
	"github.com/airbnb/rudolph/pkg/model/globalrules"
	"github.com/airbnb/rudolph/pkg/store"
	"github.com/airbnb/rudolph/pkg/types"
)

func addRuleExportCommand() {
	var filename string
	var format string
	var ruleExportCmd = &cobra.Command{
		Use:     "export  <file-name>",
		Aliases: []string{"rules-export"},
		Short:   "Export all rules into a csv file",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// args[0] has already been validated as a file before this
			dynamodbClient, err := gcpclient.FromCommand(cmd)
			if err != nil {
				return err
			}

			return runExport(dynamodbClient, filename, format)
		},
	}

	ruleExportCmd.Flags().StringVarP(&filename, "filename", "f", "", "The filename")
	_ = ruleExportCmd.MarkFlagRequired("filename")

	ruleExportCmd.Flags().StringVarP(&format, "fileformat", "t", "csv", "File format (one of: [json|csv])")

	RulesCmd.AddCommand(ruleExportCmd)
}

func runExport(
	client store.QueryAPI,
	filename string,
	format string,
) (err error) {
	switch format {
	case "json":
		return runJsonExport(client, filename)
	case "csv":
		return runCsvExport(client, filename)
	}
	return
}

type fileRule struct {
	RuleType      types.RuleType `json:"type"`
	Policy        types.Policy   `json:"policy"`
	Identifier    string         `json:"identifier"`
	CustomMessage string         `json:"custom_msg,omitempty"`
	Description   string         `json:"description"`
}

func runJsonExport(client store.QueryAPI, filename string) (err error) {
	var jsonRules []fileRule
	fmt.Println("Querying rules from DynamoDB...")
	total, err := getRules(client, func(rule globalrules.GlobalRuleRow) (err error) {
		jsonRules = append(jsonRules, fileRule{
			Identifier:    rule.Identifier,
			RuleType:      rule.RuleType,
			Policy:        rule.Policy,
			CustomMessage: rule.CustomMessage,
			Description:   rule.Description,
		})
		return
	})
	if err != nil {
		return
	}

	jsondata, err := json.MarshalIndent(jsonRules, "", "  ")
	if err != nil {
		return
	}
	err = os.WriteFile(filename, jsondata, 0644)
	if err != nil {
		return
	}

	fmt.Printf("rules discovered: %d, rules written: %d\n", total, len(jsonRules))

	return
}

func runCsvExport(
	client store.QueryAPI,
	filename string,
) (err error) {
	csvRules := make(chan []string)

	header := []string{
		"sha256",
		"type",
		"policy",
		"custom_msg",
		"description",
	}

	wg, err := csv.WriteCsvFile(filename, header, csvRules)
	if err != nil {
		panic(err)
	}

	fmt.Println("Querying rules from DynamoDB...")
	var totalWritten int64
	total, err := getRules(client, func(rule globalrules.GlobalRuleRow) (err error) {
		ruleType, err := rule.RuleType.MarshalText()
		if err != nil {
			return
		}
		policy, err := rule.Policy.MarshalText()
		if err != nil {
			return
		}
		record := []string{
			rule.Identifier,
			string(ruleType),
			string(policy),
			rule.CustomMessage,
			rule.Description,
		}
		if err != nil {
			return
		}

		totalWritten += 1
		csvRules <- record
		return
	})
	if err != nil {
		return
	}

	close(csvRules)
	wg.Wait()

	fmt.Printf("rules discovered: %d, rules written: %d\n", total, totalWritten)

	return
}

func getRules(client store.QueryAPI, callback func(globalrules.GlobalRuleRow) error) (total int64, err error) {
	var key *store.PrimaryKey
	for {
		rules, nextkey, inerr := globalrules.GetPaginatedGlobalRules(client, 50, key)
		if inerr != nil {
			err = fmt.Errorf("something went wrong querying global rules: %w", inerr)
			return
		}
		if len(rules) == 0 {
			break
		}

		for _, rule := range rules {
			total += 1
			err = callback(*rule)
			if err != nil {
				return
			}
		}

		if nextkey == nil {
			break
		}
		key = nextkey
	}
	return
}
