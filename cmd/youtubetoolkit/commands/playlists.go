package commands

import (
	"fmt"
	"os"

	"github.com/raffaelecassia/youtubetoolkit"
	"github.com/spf13/cobra"
)

func Playlists(parent *cobra.Command, tk *youtubetoolkit.Toolkit) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "playlists",
		Short: "Manage user playlists",
		Long: `Returns all user playlists.
Available fields for CSV/Table output: PlaylistId, PlaylistTitle, VideoCount.`,
		Args: cobra.NoArgs,
		Run: func(c *cobra.Command, _ []string) {
			err := tk.Playlists(outputFromFlags(c, DEFAULT_FIELDS_PLAYLISTS))
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
Fields for CSV/Table output: VideoId*, VideoTitle*, VideoUrl*, ChannelId*, ChannelTitle*, ChannelUrl*, PlaylistItemId.
(* default fields when --fields is not specified)`,
		Args: cobra.NoArgs,
		Run: func(c *cobra.Command, _ []string) {
			err := tk.Playlist(id, outputFromFlags(c, DEFAULT_FIELDS_PLAYLIST))
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
					err := tk.AddVideoToPlaylist(id,
						youtubetoolkit.CSVFirstFieldOnlySource(os.Stdin),
						youtubetoolkit.NullSink())
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
	var print bool
	cmd := &cobra.Command{
		Use:   "add [video id]",
		Short: "Adds a video to a playlist",
		Long: `Adds a video to a playlist.
To add multiple videos, send to stdin a list of video ids (or a CSV with ids in the first column).
The flag --id is mandatory.
If --print-data flag is used, the default fields from the playlist command will apply.`,
		Args: cobra.MaximumNArgs(1),
		Run: func(c *cobra.Command, args []string) {
			var output youtubetoolkit.FlowOption
			if print {
				output = outputFromFlags(c, DEFAULT_FIELDS_PLAYLIST)
			} else {
				output = youtubetoolkit.NullSink()
			}

			playlistId := c.Flag("id").Value.String()
			if len(args) == 1 {
				err := tk.AddVideoToPlaylist(playlistId, youtubetoolkit.SingleStringSource(args[0]), output)
				if err != nil {
					fmt.Fprintln(os.Stderr, "Error:", err)
				}
			} else {
				if checkStdinInput() {
					err := tk.AddVideoToPlaylist(playlistId, youtubetoolkit.CSVFirstFieldOnlySource(os.Stdin), output)
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
	cmd.Flags().BoolVarP(&print, "print-data", "p", false, "print to stdout the playlist/video infos of the added video(s)")
	parent.AddCommand(cmd)
	return cmd
}
