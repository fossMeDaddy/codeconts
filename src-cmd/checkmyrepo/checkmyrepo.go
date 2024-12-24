package checkmyrepoCmd

import (
	"github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"
)

var checkmyrepoCmd = &cobra.Command{
	Use:   "checkmyrepo",
	Short: "command to check the code contribution of all developers",
	Run:   cmdRun,
}

func cmdRun(cmd *cobra.Command, args []string) {
	println("in the check command")
	repo, err := git.PlainOpen(".")
	if err != nil {
		cmd.PrintErrln("Error:this is not a git repo")
	}
}

func Init() *cobra.Command {
	return checkmyrepoCmd
}
