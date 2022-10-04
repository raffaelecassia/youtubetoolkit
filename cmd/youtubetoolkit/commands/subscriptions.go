package commands

import (
	"fmt"
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
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Returns all channels from user subscriptions",
		Long: `Returns all channels from user subscriptions.
Available fields for CSV/Table output: ChannelId*, ChannelTitle*, ChannelUrl*, ChannelThumbUrl*, SubscriptionId.
(* default fields when --fields is not specified.)`,
		Args: cobra.NoArgs,
		Run: func(c *cobra.Command, _ []string) {
			err := tk.Subscriptions(
				outputFromFlags(c, DEFAULT_FIELDS_SUBSCRIPTIONS))
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
			}
		},
	}
	parent.AddCommand(cmd)
	return cmd
}

func Subscribe(parent *cobra.Command, tk *youtubetoolkit.Toolkit) *cobra.Command {
	var print bool
	cmd := &cobra.Command{
		Use:   "add [channel id]",
		Short: "Subscribe to a channel",
		Long: `Subscribe to a channel.
To add multiple channels, send to stdin a list of channel ids (or a CSV with ids in the first column).
If --print-data flag is used, the default fields from command subscriptions-list will apply.`,
		Args: cobra.MaximumNArgs(1),
		Run: func(c *cobra.Command, args []string) {
			var output youtubetoolkit.FlowOption
			if print {
				output = outputFromFlags(c, DEFAULT_FIELDS_SUBSCRIPTIONS)
			} else {
				output = youtubetoolkit.NullSink()
			}
			if len(args) == 1 {
				err := tk.Subscribe(youtubetoolkit.SingleStringSource(args[0]), output)
				if err != nil {
					fmt.Fprintln(os.Stderr, "Error:", err)
				}
			} else {
				if checkStdinInput() {
					err := tk.Subscribe(youtubetoolkit.CSVFirstFieldOnlySource(os.Stdin), output)
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
	cmd.Flags().BoolVarP(&print, "print-data", "p", false, "print to stdout the subscriptions infos of the added channel(s)")
	parent.AddCommand(cmd)
	return cmd
}

func Unsubscribe(parent *cobra.Command, tk *youtubetoolkit.Toolkit) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "del [channel id]",
		Short: "Unsubscribe from a channel",
		Args:  cobra.ExactArgs(1),
		Run: func(c *cobra.Command, args []string) {
			err := tk.Unsubscribe(args[0])
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
			}
		},
	}
	parent.AddCommand(cmd)
	return cmd
}
