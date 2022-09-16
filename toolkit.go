package youtubetoolkit

import (
	"fmt"
	"io"
	"time"

	"github.com/raffaelecassia/youtubetoolkit/bigg"
	"google.golang.org/api/googleapi"
)

type Toolkit struct {
	service   YoutubeService
	logWriter io.Writer
}

type YoutubeService interface {
	SubscriptionsList(out chan<- *bigg.Sub) error
	SubscriptionInsert(channelId string) (*bigg.Sub, error)
	PlaylistsList(out chan<- *bigg.Playlist) error
	PlaylistInsert(title string) (*bigg.Playlist, error)
	PlaylistItemsList(id string, filter func(*bigg.PlaylistItem) bool, out chan<- *bigg.PlaylistItem) error
	PlaylistItemsInsert(playlistId, videoId string) (*bigg.PlaylistItem, error)
	GetChannelInfo(id string) (*bigg.Channel, error)
}

func New() *Toolkit {
	return &Toolkit{nil, io.Discard}
}

func NewWithService(svc YoutubeService) *Toolkit {
	return &Toolkit{svc, io.Discard}
}

func (tk *Toolkit) SetService(service YoutubeService) {
	tk.service = service
}

// CSVSubscriptions prints to output all the channels from user subscription.
// Output is CSV formatted with columns "channelId, title".
// A true extracols will output also columns "url, thumbnail, subscriptionId".
func (tk *Toolkit) CSVSubscriptions(output io.Writer, extracols bool) error {
	errors, err := multiErrorsHandler()
	subs := make(chan *bigg.Sub)
	go func() {
		errors <- tk.service.SubscriptionsList(subs)
		close(subs)
	}()
	records := sub2record(subs, extracols)
	sinkrecords2csv(errors, records, output)
	close(errors)
	return <-err
}

// Subscribe adds channelId to the user subscriptions.
// Returns the subscription ID or error.
func (tk *Toolkit) Subscribe(channelId string) (string, error) {
	tk.logf("subscribing to %s... ", channelId)
	sub, err := tk.service.SubscriptionInsert(channelId)
	if err != nil {
		if e, ok := err.(*googleapi.Error); ok {
			tk.log("Error:", e.Message)
		} else {
			tk.log(err)
		}
	} else {
		tk.log("channel", sub.Snippet.Title, "added")
		return sub.Id, nil
	}
	return "", err
}

// CSVBulkSubscribe reads from input a list of channel IDs and subscribe to them.
// It writes to output the subscription IDs.
func (tk *Toolkit) CSVBulkSubscribe(input io.Reader, output io.Writer) error {
	errors, err := multiErrorsHandler()
	// read csv
	inputrecords := csv2records(errors, input)
	// first column contains the channel id
	channelIds := recordColumnFilter(inputrecords, 0)
	// subscribe
	for id := range channelIds {
		sid, err := tk.Subscribe(id)
		if err == nil {
			fmt.Fprintln(output, sid)
		}
	}
	close(errors)
	return <-err
}

// CSVPlaylists prints to output all user playlists.
// Output is a CSV with columns "id, title, videoCount".
func (tk *Toolkit) CSVPlaylists(output io.Writer) error {
	errors, err := multiErrorsHandler()
	pls := make(chan *bigg.Playlist)
	go func() {
		errors <- tk.service.PlaylistsList(pls)
		close(pls)
	}()
	records := playlist2record(pls)
	sinkrecords2csv(errors, records, output)
	close(errors)
	return <-err
}

// CSVPlaylist accepts an ID and prints to output the playlist's videos.
// Output is a CSV with columns "videoId, title, channelId, channelTitle".
func (tk *Toolkit) CSVPlaylist(playlistId string, output io.Writer) error {
	errors, err := multiErrorsHandler()
	pls := make(chan *bigg.PlaylistItem)
	go func() {
		errors <- tk.service.PlaylistItemsList(playlistId, allPlaylistItems(), pls)
		close(pls)
	}()
	outputrecords := playlistItem2record(pls)
	sinkrecords2csv(errors, outputrecords, output)
	close(errors)
	return <-err
}

// CreatePlaylist create a new playlist.
// Returns the playlist ID or error.
func (tk *Toolkit) NewPlaylist(title string) (string, error) {
	pl, err := tk.service.PlaylistInsert(title)
	if err != nil {
		return "", err
	}
	return pl.Id, nil
}

// AddVideoToPlaylist adds a video to a playlist.
func (tk *Toolkit) AddVideoToPlaylist(playlistId, videoId string) (string, error) {
	pli, err := tk.service.PlaylistItemsInsert(playlistId, videoId)
	if err != nil {
		return "", err
	}
	return pli.Id, nil
}

func (tk *Toolkit) CSVBulkAddVideoToPlaylist(playlistId string, input io.Reader, output io.Writer) error {
	errors, err := multiErrorsHandler()
	inputrecords := csv2records(errors, input)
	videoIds := recordColumnFilter(inputrecords, 0)
	plitems := tk.videoIds2playlist(errors, playlistId, videoIds)
	// sink to output
	for i := range plitems {
		fmt.Fprintln(output, i.Id)
	}
	close(errors)
	return <-err
}

// CSVLastUploads reads from input a list of channels ID and prints to output a csv
// with the latest channel's video uploads since the time argument.
func (tk *Toolkit) CSVLastUploads(input io.Reader, output io.Writer, since time.Time) error {
	errors, err := multiErrorsHandler()

	filter := sinceDatePlaylistItems(since)

	// read csv
	inputrecords := csv2records(errors, input)
	// first column contains the channel id
	channelIds := recordColumnFilter(inputrecords, 0)

	// fetch all video uploads using three parallel go routines
	playlistItems := tk.channelIds2VideoUploads(errors, channelIds, filter, 3)

	// convert items to records
	outputrecords := playlistItem2record(playlistItems)

	// write csv
	sinkrecords2csv(errors, outputrecords, output)

	close(errors)
	return <-err
}
