package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/alexisvisco/gta/pkg/gta"
	"github.com/alexisvisco/gta/pkg/gta/diff"
	"github.com/alexisvisco/gta/pkg/gta/vars"
)

const localRef = ""

func view(_ *cobra.Command, args []string) error {
	previousRef := args[0]
	currentRef := localRef
	if len(args) == 2 {
		currentRef = args[1]
	}

	previousNoder, err := gta.GetTree(vars.Repository, previousRef)
	if err != nil {
		return err
	}

	if currentRef == localRef {
		packages, err := diff.LocalDiff(vars.Repository, previousNoder)
		if err != nil {
			return err
		}

		gta.Output(packages)
	} else {
		currentNoder, err := gta.GetTree(vars.Repository, currentRef)
		if err != nil {
			return err
		}

		packages, err := diff.Diff(vars.Repository, previousNoder, currentNoder)
		if err != nil {
			return err
		}

		gta.Output(packages)
	}

	return nil
}

// viewCmd represents the view command
var viewCmd = &cobra.Command{
	Use:   "view <previous ref> [<current re>]",
	Short: "View packages that have changed between a previous ref and the current ref ",
	Long: "This command accept two arguments.\n" +
		"Each arguments can be either a tag, a branch or a specific commit hash.\n" +
		"If 'new' argument is not specified, gta will use your current uncommitted changes.\n" +
		"The command do nothing except printing the packages which should be tested.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("view called")
	},
	Args: cobra.RangeArgs(1, 2),
	RunE: view,
}

func init() {
	rootCmd.AddCommand(viewCmd)
}
