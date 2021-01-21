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
		Use:   "serve",
		Short: "Serves a directory on any free port. Only arg is the directory.",
		Long: `Serves the given directory on port :0, which will make the system choose a free port
for you. The port is written to stdout in a formatted message. An example call would be
the following.

$ gtransfer serve /Users/jdoe/Desktop/localFolder

This would serve the localFolder folder on a free port, available for a gtransfer client to download.
See 'gtransfer help dial' for help on how to do that.`,
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
		Use:   "dial",
		Short: "Download the served directory. First arg is the remote address, second is the path to download to.",
		Long: `Dials the remote address, which is the first argument. The server serves a certain directory,
which is written under the target path on the local system. This target path is the second argument.
An example call would be the following.

$ gtransfer dial 10.10.10.10 /Users/jdoe/Desktop/remoteSystem

This would download the served directory on the host into the remoteSystem folder.
see 'gtransfer help serve' for help on how to setup a server.`,
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
