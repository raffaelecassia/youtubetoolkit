package commands

import (
	"fmt"
	"io"
	"os"

	"github.com/raffaelecassia/youtubetoolkit"
	"github.com/spf13/cobra"
)

func Subscriptions(parent *cobra.Command, tk *youtubetoolkit.Toolkit) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "subscriptions",
		Short: "Manage user subscriptions",
		Args:  cobra.NoArgs,
	}
	parent.AddCommand(cmd)
	return cmd
}

func SubscriptionsList(parent *cobra.Command, tk *youtubetoolkit.Toolkit) *cobra.Command {
	var extracols bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Returns all channels from user subscriptions",
		Long: `Returns all channels from user subscriptions.
Data printed to stdout is a CSV with columns "channelId, channelTitle".`,
		Args: cobra.NoArgs,
		Run: func(_ *cobra.Command, _ []string) {
			err := tk.CSVSubscriptions(os.Stdout, extracols)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
			}
		},
	}
	cmd.Flags().BoolVarP(&extracols, "extracols", "e", false, "output additional columns: \"url, thumbnail, subscriptionId\"")
	parent.AddCommand(cmd)
	return cmd
}

func Subscribe(parent *cobra.Command, tk *youtubetoolkit.Toolkit) *cobra.Command {
	var printid bool
	cmd := &cobra.Command{
		Use:   "add [channel id]",
		Short: "Subscribe to a channel",
		Long: `Subscribe to a channel.
To add multiple channels, send to stdin a list of channel ids (or a CSV with ids in the first column)`,
		Args: cobra.MaximumNArgs(1),
		Run: func(c *cobra.Command, args []string) {
			if len(args) == 1 {
				id, err := tk.Subscribe(args[0])
				if err != nil {
					return // errors are already printed to stderr
				}
				if printid {
					fmt.Fprintln(os.Stdout, id)
				}
			} else {
				if checkStdinInput() {
					out := io.Discard
					if printid {
						out = os.Stdout
					}
					err := tk.CSVBulkSubscribe(os.Stdin, out)
					if err != nil {
						fmt.Fprintln(os.Stderr, "Error:", err)
					}
				} else {
					err := c.Help()
					if err != nil {
						fmt.Fprintln(os.Stderr, err)
					}
					os.Exit(0)
				}
			}
		},
	}
	cmd.Flags().BoolVarP(&printid, "print-id", "i", false, "print to stdout the subscription id of the added channel(s)")
	parent.AddCommand(cmd)
	return cmd
}

func Unsubscribe(parent *cobra.Command, tk *youtubetoolkit.Toolkit) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "del [channel id]",
		Short: "Unsubscribe from a channel",
		// Long:  "Unsubscribe from a channel.",
		Args: cobra.ExactArgs(1),
		Run: func(c *cobra.Command, args []string) {
			_ = tk.Unsubscribe(args[0]) // errors are already printed to stderr
		},
	}
	parent.AddCommand(cmd)
	return cmd
}
