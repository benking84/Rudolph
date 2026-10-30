package repair

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/airbnb/rudolph/internal/cli/gcpclient"
	"github.com/airbnb/rudolph/pkg/scan"
	"github.com/airbnb/rudolph/pkg/store"
)

var (
	RepairCmd *cobra.Command
)

func init() {
	RepairCmd = &cobra.Command{
		Use:   "repair",
		Short: "Scans Firestore and reports legacy machine records",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := gcpclient.FromCommand(cmd)
			if err != nil {
				return err
			}
			return repair(scan.GetScanService(client), client)
		},
	}
}

func repair(scanService scan.ScanService, deleter store.DeleteItemAPI) error {

	// Configuration
	pageSize := int32(10)
	delete := false

	input := store.ScanInput{
		Limit: pageSize,
	}

	page := 0
	callback := func(out *store.ScanOutput) (err error) {
		page += 1
		fmt.Println("==========")
		fmt.Printf("Page #%d\n", page)
		fmt.Printf("  Scanned %d items\n", out.ScannedCount)
		fmt.Printf("  Discovered %d items\n", out.Count)

		var items []dynamodbItem
		err = store.UnmarshalListOfMaps(out.Items, &items)
		if err != nil {
			err = fmt.Errorf("failed to unmarshal output items: %w", err)
			return
		}

		for _, item := range items {
			if strings.HasPrefix(item.PartitionKey, "MachineInfo#") || strings.HasPrefix(item.PartitionKey, "MachineConfig#") {
				fmt.Printf("    ++ LEGACY ITEM LOCATED: (%+v | %+v) ++\n", item.PartitionKey, item.SortKey)
				if delete {
					_, err := deleter.DeleteItem(item.PrimaryKey)
					if err != nil {
						return fmt.Errorf("failed to delete item: %w", err)
					}
					fmt.Printf("      Item Deleted\n")
				}
			}
		}

		fmt.Println("")
		return nil
	}

	stop := func(out *store.ScanOutput) (bool, error) {
		// Don't stop until done.
		return false, nil
	}

	err := scanService.ScanAll(input, callback, stop)
	if err != nil {
		return err
	}
	return nil
}

type dynamodbItem struct {
	store.PrimaryKey
}
