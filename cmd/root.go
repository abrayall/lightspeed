package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"lightspeed/internal/ui"
)

// Version is set by ldflags during build
var Version = "dev"

var rootCmd = &cobra.Command{
	Use:   "lightspeed",
	Short: "Lightweight rapid development tool for PHP websites",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Long = ui.Divider() + "\n" + ui.Banner() + "\n" + ui.VersionLine(Version) + "\n\n" + ui.Divider() + "\n\nA lightweight, rapid development tool for small PHP websites"
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("lightspeed %s\n", Version)
	},
}
