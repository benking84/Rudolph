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

	var ruleSilentCmd = &cobra.Command{
		Use:     "silent",
		Aliases: []string{"silentblock"},
		Short:   "Create a rule that applies the SilentBlocklist policy to the specified file",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// args[0] has already been validated as a file before this
			dynamodbClient, err := gcpclient.FromCommand(cmd)
			if err != nil {
				return err
			}
			time := clock.ConcreteTimeProvider{}

			return applyPolicyForPath(time, dynamodbClient, types.SilentBlocklist, tf, rf)
		},
	}

	tf.AddTargetFlags(ruleSilentCmd)
	rf.AddRuleInfoFlags(ruleSilentCmd)

	RuleCmd.AddCommand(ruleSilentCmd)
}
