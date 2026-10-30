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

	var ruleDenyCmd = &cobra.Command{
		Use:     "deny <file-path>",
		Aliases: []string{"block"},
		Short:   "Create a rule that applies the Blocklist policy to the specified file",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// args[0] has already been validated as a file before this
			dynamodbClient, err := gcpclient.FromCommand(cmd)
			if err != nil {
				return err
			}
			time := clock.ConcreteTimeProvider{}

			return applyPolicyForPath(time, dynamodbClient, types.Blocklist, tf, rf)
		},
	}

	tf.AddTargetFlags(ruleDenyCmd)
	rf.AddRuleInfoFlags(ruleDenyCmd)

	RuleCmd.AddCommand(ruleDenyCmd)
}
