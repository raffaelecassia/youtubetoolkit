package commands

import (
	"fmt"
	"os"

	"github.com/raffaelecassia/youtubetoolkit"
	"github.com/raffaelecassia/youtubetoolkit/bigg"
	"github.com/spf13/cobra"
)

func Root(tk *youtubetoolkit.Toolkit) *cobra.Command {
	var clientSecretFile string
	var tokenFile string
	var debug bool

	var ytsvc *bigg.Youtube

	cmd := &cobra.Command{
		Use:   "youtubetoolkit",
		Short: "A toolkit for Youtube",
		PersistentPreRun: func(c *cobra.Command, _ []string) {
			// skip for help and completion commands (and hidden ones like __complete)
			if c.Use[:4] == "help" || (c.HasParent() && c.Parent().Use == "completion") || c.Use[:2] == "__" {
				return
			}
			//
			// login
			//
			client := bigg.NewClient()
			client.TokenFile = tokenFile

			if debug {
				client.EnableLogTransport()
			}

			if err := client.SetSecretFromFile(clientSecretFile); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			if err := client.Authorize(); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			svc, err := client.NewYoutubeService()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			tk.SetService(svc)
			ytsvc = svc

			tk.SetLogWriter(os.Stderr)
		},
		PersistentPostRun: func(c *cobra.Command, _ []string) {
			// skip for help and completion commands (and hidden ones like __complete)
			if c.Use[:4] == "help" || (c.HasParent() && c.Parent().Use == "completion") || c.Use[:2] == "__" {
				return
			}
			fmt.Fprintln(os.Stderr, "Quota cost:", ytsvc.GetCost(), "units")
		},
	}

	// global flags
	cmd.PersistentFlags().StringVarP(&clientSecretFile, "client-secret", "s", "client_secret.json", "OAuth2 client secret JSON file")
	cmd.PersistentFlags().StringVarP(&tokenFile, "token", "t", "goauth.token", "login token filename")
	cmd.PersistentFlags().BoolVarP(&debug, "debug-http", "d", false, "logs to stdout each http request/response")

	cmd.PersistentFlags().Bool("csv", true, "CSV output")
	cmd.PersistentFlags().Bool("table", false, "Table output")
	cmd.PersistentFlags().Bool("jsonl", false, "JSON Lines output")
	cmd.PersistentFlags().StringSlice("fields", []string{}, "Fields for CSV/Table output (fields separated by commas). See each command for the fields list.")

	cmd.MarkFlagsMutuallyExclusive("csv", "table", "jsonl")

	return cmd
}

func Execute() error {
	tk := youtubetoolkit.New()

	root := Root(tk)

	subs := Subscriptions(root, tk)
	_ = SubscriptionsList(subs, tk)
	_ = Subscribe(subs, tk)
	_ = Unsubscribe(subs, tk)

	pls := Playlists(root, tk)
	_ = NewPlaylist(pls, tk)
	_ = DelPlaylist(pls, tk)

	pl := Playlist(root, tk)
	_ = AddToPlaylist(pl, tk)

	_ = LastUploads(root, tk)

	return root.Execute()
}

//
// utils
//

func checkStdinInput() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		// fatal...
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
	// checks if data is being piped to stdin
	return (stat.Mode() & os.ModeCharDevice) == 0
}

var (
	DEFAULT_FIELDS_SUBSCRIPTIONS    = &[]string{"ChannelId", "ChannelTitle", "ChannelUrl", "ChannelThumbUrl"}
	DEFAULT_FIELDS_UPLOADS_PLAYLIST = &[]string{"VideoId", "VideoTitle", "PublishedAt", "ChannelId", "ChannelTitle"}
	DEFAULT_FIELDS_PLAYLISTS        = &[]string{"PlaylistId", "PlaylistTitle", "VideoCount"}
	DEFAULT_FIELDS_PLAYLIST         = &[]string{"VideoId", "VideoTitle", "VideoUrl", "ChannelId", "ChannelTitle", "ChannelUrl"}
)

func outputFromFlags(c *cobra.Command, defaultfields *[]string) youtubetoolkit.FlowOption {
	fields, err := c.Flags().GetStringSlice("fields")
	if err != nil || len(fields) == 0 {
		fields = *defaultfields
	}
	if val, err := c.Flags().GetBool("table"); val && err == nil {
		return youtubetoolkit.TableSink(os.Stdout, &fields)
	} else if val, err := c.Flags().GetBool("jsonl"); val && err == nil {
		return youtubetoolkit.JSONLinesSink(os.Stdout)
	}
	return youtubetoolkit.CSVSink(os.Stdout, &fields)
}
