package rule

import (
	"github.com/airbnb/rudolph/internal/cli/flags"
	"github.com/airbnb/rudolph/internal/cli/gcpclient"
	"github.com/airbnb/rudolph/pkg/clock"
	"github.com/airbnb/rudolph/pkg/types"
	"github.com/spf13/cobra"
)

func init() {
	tf := flags.TargetFlags{}
	rf := flags.RuleInfoFlags{}

	var ruleCompilerCmd = &cobra.Command{
		Use:     "compiler  <file-path>",
		Aliases: []string{"allow-complier"},
		Short:   "Create a rule that applies the AllowlistCompiler policy to the specified file",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// args[0] has already been validated as a file before this
			dynamodbClient, err := gcpclient.FromCommand(cmd)
			if err != nil {
				return err
			}
			time := clock.ConcreteTimeProvider{}

			return applyPolicyForPath(time, dynamodbClient, types.AllowlistCompiler, tf, rf)
		},
	}

	tf.AddTargetFlags(ruleCompilerCmd)
	rf.AddRuleInfoFlags(ruleCompilerCmd)

	RuleCmd.AddCommand(ruleCompilerCmd)
}
