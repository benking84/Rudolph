package gcpclient

import (
	"context"

	"github.com/airbnb/rudolph/pkg/store"
	"github.com/spf13/cobra"
)

// FromCommand opens the Firestore database selected by the CLI flags.
func FromCommand(cmd *cobra.Command) (store.Client, error) {
	project, err := cmd.Flags().GetString("project")
	if err != nil {
		return nil, err
	}
	database, err := cmd.Flags().GetString("firestore_database")
	if err != nil {
		return nil, err
	}
	if database == "" {
		database = "rudolph"
	}
	return store.New(context.Background(), project, database)
}
