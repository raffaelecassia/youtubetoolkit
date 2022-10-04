package commands

import (
	"fmt"
	"os"
	"time"

	"github.com/raffaelecassia/youtubetoolkit"
	"github.com/spf13/cobra"
)

func LastUploads(parent *cobra.Command, tk *youtubetoolkit.Toolkit) *cobra.Command {
	var days uint16
	cmd := &cobra.Command{
		Use:   "lastuploads [channel id]",
		Short: "Returns channels' last video uploads",
		Long: `Returns channels' last video uploads sorted by the published date (oldest first).
Multiple channel IDs are received from stdin (one per line, or a csv with ids in the first column).
Available fields for CSV/Table output: VideoId*, VideoTitle*, VideoUrl, PublishedAt*, ChannelId*, ChannelTitle*, ChannelUrl.
(* default fields when --fields is not specified)`,
		Args: cobra.MaximumNArgs(1),
		Run: func(c *cobra.Command, args []string) {
			since := time.Now().Add(-time.Hour * time.Duration(24*int64(days)))
			output := outputFromFlags(c, DEFAULT_FIELDS_UPLOADS_PLAYLIST)
			if len(args) == 1 {
				err := tk.LastUploads(since, youtubetoolkit.SingleStringSource(args[0]), output)
				if err != nil {
					fmt.Fprintln(os.Stderr, "Error:", err)
				}
			} else if checkStdinInput() {
				err := tk.LastUploads(since, youtubetoolkit.CSVFirstFieldOnlySource(os.Stdin), output)
				if err != nil {
					fmt.Fprintln(os.Stderr, "Error:", err)
				}
			} else {
				err := c.Help()
				if err != nil {
					fmt.Fprintln(os.Stderr, "Error:", err)
				}
				os.Exit(0)
			}
		},
	}
	cmd.Flags().Uint16VarP(&days, "days", "", 7, "days since")
	parent.AddCommand(cmd)
	return cmd
}
