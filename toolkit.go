package youtubetoolkit

import (
	"io"
	"time"

	"github.com/raffaelecassia/youtubetoolkit/bigg"
)

type Toolkit struct {
	service   YoutubeService
	logWriter io.Writer
}

type YoutubeService interface {
	SubscriptionsList(out chan<- *bigg.Sub) error
	SubscriptionInsert(channelId string) (*bigg.Sub, error)
	SubscriptionDelete(channelId string) error
	PlaylistsList(out chan<- *bigg.Playlist) error
	PlaylistInsert(title string) (*bigg.Playlist, error)
	PlaylistDelete(playlistId string) error
	PlaylistItemsList(id string, filter func(*bigg.PlaylistItem) (bool, error), out chan<- *bigg.PlaylistItem) error
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

// Subscriptions gets all channels from user subscription.
// Flow: only sink is required
func (tk *Toolkit) Subscriptions(opts ...FlowOption) error {
	flow := options2flowconfig(opts...)
	errors, err := multiErrorsHandler()
	subs := make(chan *bigg.Sub)
	go func() {
		errors <- tk.service.SubscriptionsList(subs)
		close(subs)
	}()
	items := sub2item(subs)
	flow.itemSink(errors, items)
	close(errors)
	return <-err
}

// Subscribe adds channels to user subscriptions.
// Flow: source and sink are required
func (tk *Toolkit) Subscribe(opts ...FlowOption) error {
	flow := options2flowconfig(opts...)
	errors, err := multiErrorsHandler()
	channelIds := flow.stringSource(errors)
	subs := tk.channels2newsubscriptions(errors, channelIds)
	items := sub2item(subs)
	flow.itemSink(errors, items)
	close(errors)
	return <-err
}

// Unsubscribe removes channel from user subscriptions.
func (tk *Toolkit) Unsubscribe(channelId string) error {
	tk.logf("unsubscribing from %s... ", channelId)
	err := tk.service.SubscriptionDelete(channelId)
	if err != nil {
		tk.log("fail!")
		return err
	}
	tk.log("done")
	return nil
}

// Playlists gets all user playlists.
// Flow: only sink is required
func (tk *Toolkit) Playlists(opts ...FlowOption) error {
	flow := options2flowconfig(opts...)
	errors, err := multiErrorsHandler()
	pls := make(chan *bigg.Playlist)
	go func() {
		errors <- tk.service.PlaylistsList(pls)
		close(pls)
	}()
	items := playlist2item(pls)
	flow.itemSink(errors, items)
	close(errors)
	return <-err
}

// Playlist gets a playlist's videos.
// Flow: only sink is required
func (tk *Toolkit) Playlist(playlistId string, opts ...FlowOption) error {
	flow := options2flowconfig(opts...)
	errors, err := multiErrorsHandler()
	pls := make(chan *bigg.PlaylistItem)
	go func() {
		errors <- tk.service.PlaylistItemsList(playlistId, allPlaylistItems(), pls)
		close(pls)
	}()
	items := playlistItem2item(pls)
	flow.itemSink(errors, items)
	close(errors)
	return <-err
}

// NewPlaylist creates a new private playlist.
// Returns the playlist ID or error.
func (tk *Toolkit) NewPlaylist(title string) (string, error) {
	pl, err := tk.service.PlaylistInsert(title)
	if err != nil {
		return "", err
	}
	return pl.Id, nil
}

// DeletePlaylist deletes a user playlist.
func (tk *Toolkit) DeletePlaylist(playlistId string) error {
	return tk.service.PlaylistDelete(playlistId)
}

// AddVideoToPlaylist adds videos to a playlist.
// Flow: source and sink are required
func (tk *Toolkit) AddVideoToPlaylist(playlistId string, opts ...FlowOption) error {
	flow := options2flowconfig(opts...)
	errors, err := multiErrorsHandler()
	videoIds := flow.stringSource(errors)
	plitems := tk.videos2playlist(errors, playlistId, videoIds)
	items := playlistItem2item(plitems)
	flow.itemSink(errors, items)
	close(errors)
	return <-err
}

// CSVLastUploads gets the latest channels' video uploads since the time argument.
// Videos are sorted by the published date (oldest first).
// Flow: source and sink are required
func (tk *Toolkit) LastUploads(since time.Time, opts ...FlowOption) error {
	flow := options2flowconfig(opts...)
	errors, err := multiErrorsHandler()

	filter := sinceDatePlaylistItems(since)

	channelIds := flow.stringSource(errors)
	// fetch all video uploads using three parallel go routines
	playlistItems := tk.channels2channelvideouploads(errors, channelIds, filter, 3)
	sorted := sortPlaylistItemByPublishedAt(playlistItems)

	items := playlistItem2item(sorted)
	flow.itemSink(errors, items)

	close(errors)
	return <-err
}
