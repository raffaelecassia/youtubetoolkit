package commands

import (
	"fmt"
	"os"
	"time"

	"github.com/raffaelecassia/youtubetoolkit"
	"github.com/spf13/cobra"
)

func LastUploads(parent *cobra.Command, tk *youtubetoolkit.Toolkit) *cobra.Command {
	var days uint8
	cmd := &cobra.Command{
		Use:   "lastuploads",
		Short: "Returns channels' last video uploads",
		Long: `Returns channels' last video uploads sorted by the published date (oldest first).
Channels IDs are received from stdin (one per line, or a csv with ids in the first column).
Data printed to stdout is a CSV with columns "videoId, published_date, title, channelId, channelTitle".`,
		Args: cobra.NoArgs,
		Run: func(c *cobra.Command, _ []string) {
			if checkStdinInput() {
				since := time.Now().Add(-time.Hour * time.Duration(24*int64(days)))
				tk.SetLogWriter(os.Stderr) // FIXME
				err := tk.CSVLastUploads(os.Stdin, os.Stdout, since)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
				}
			} else {
				c.Help()
				os.Exit(0)
			}
		},
	}
	cmd.Flags().Uint8VarP(&days, "days", "", 7, "days since")
	// cmd.MarkFlagRequired("days")
	parent.AddCommand(cmd)
	return cmd
}
