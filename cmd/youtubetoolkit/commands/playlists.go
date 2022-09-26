package commands

import (
	"fmt"
	"io"
	"os"

	"github.com/raffaelecassia/youtubetoolkit"
	"github.com/spf13/cobra"
)

func Playlists(parent *cobra.Command, tk *youtubetoolkit.Toolkit) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "playlists",
		Short: "Manage user playlists",
		Long: `Returns all user playlists.
Data printed to stdout is a CSV with columns "id, title, videoCount".`,
		Args: cobra.NoArgs,
		Run: func(_ *cobra.Command, _ []string) {
			err := tk.CSVPlaylists(os.Stdout)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
			}
		},
	}
	parent.AddCommand(cmd)
	return cmd
}

func Playlist(parent *cobra.Command, tk *youtubetoolkit.Toolkit) *cobra.Command {
	var id string
	cmd := &cobra.Command{
		Use:   "playlist",
		Short: "Manage a playlist",
		Long: `Returns all videos of a playlist.
Data printed to stdout is a CSV with columns "videoId, title, channelId, channelTitle".`,
		Args: cobra.NoArgs,
		Run: func(c *cobra.Command, _ []string) {
			err := tk.CSVPlaylist(id, os.Stdout)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
			}
		},
	}
	cmd.PersistentFlags().StringVarP(&id, "id", "", "", "playlist id, mandatory")
	err := cmd.MarkPersistentFlagRequired("id")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	parent.AddCommand(cmd)
	return cmd
}

func NewPlaylist(parent *cobra.Command, tk *youtubetoolkit.Toolkit) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "new [playlist name]",
		Short: "Creates a new playlist",
		Long: `Creates a new playlist.
To also add videos to the newly created playlist, send to stdin a list of 
video ids (or a CSV with ids in the first column).
Prints to stdout the playlist id.`,
		Args: cobra.ExactArgs(1),
		Run: func(_ *cobra.Command, args []string) {
			id, err := tk.NewPlaylist(args[0])
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
			} else {
				fmt.Fprintln(os.Stdout, id)
				if checkStdinInput() {
					err := tk.CSVBulkAddVideoToPlaylist(id, os.Stdin, io.Discard)
					if err != nil {
						fmt.Fprintln(os.Stderr, "Error:", err)
					}
				}
			}
		},
	}
	parent.AddCommand(cmd)
	return cmd
}

func DelPlaylist(parent *cobra.Command, tk *youtubetoolkit.Toolkit) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "del [playlist id]",
		Short: "Deletes a playlist",
		Args:  cobra.ExactArgs(1),
		Run: func(_ *cobra.Command, args []string) {
			err := tk.DeletePlaylist(args[0])
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
			}
		},
	}
	parent.AddCommand(cmd)
	return cmd
}

func AddToPlaylist(parent *cobra.Command, tk *youtubetoolkit.Toolkit) *cobra.Command {
	var printid bool
	cmd := &cobra.Command{
		Use:   "add [video id]",
		Short: "Adds a video to a playlist",
		Long: `Adds a video to a playlist.
To add multiple videos, send to stdin a list of video ids (or a CSV with ids in the first column).
The flag --id is mandatory.`,
		Args: cobra.MaximumNArgs(1),
		Run: func(c *cobra.Command, args []string) {
			plid := c.Flag("id").Value.String()
			if len(args) == 1 {
				id, err := tk.AddVideoToPlaylist(plid, args[0])
				if err != nil {
					fmt.Fprintln(os.Stderr, "Error:", err)
				} else if printid {
					fmt.Fprintln(os.Stdout, id)
				}
			} else {
				if checkStdinInput() {
					out := io.Discard
					if printid {
						out = os.Stdout
					}
					err := tk.CSVBulkAddVideoToPlaylist(plid, os.Stdin, out)
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
	cmd.Flags().BoolVarP(&printid, "print-id", "i", false, "print to stdout the playlist item id of the added video(s)")
	parent.AddCommand(cmd)
	return cmd
}
