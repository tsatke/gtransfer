package main

import (
	"fmt"
	"os"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/tsatke/gtransfer"
)

const (
	appName = "gtransfer"
)

var (
	rootCmd = &cobra.Command{
		Use: appName,
	}

	serverCmd = &cobra.Command{
		Use:  "serve",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]
			if path == "." {
				wd, err := os.Getwd()
				if err != nil {
					return fmt.Errorf("getwd: %w", err)
				}
				path = wd
			}
			srv := gtransfer.NewServer(":0", afero.NewBasePathFs(afero.NewOsFs(), path))
			defer srv.Stop()
			return srv.Serve()
		},
	}

	clientCmd = &cobra.Command{
		Use:  "dial",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			addr := args[0]
			path := args[1]
			client := gtransfer.Dial(addr)
			return client.DownloadInto(afero.NewBasePathFs(afero.NewOsFs(), path))
		},
	}
)

func init() {
	rootCmd.AddCommand(serverCmd)
	rootCmd.AddCommand(clientCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s: %s", appName, err)
		os.Exit(1)
	}
}
