package youtubetoolkit_test

import (
	"bytes"
	"io"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/raffaelecassia/youtubetoolkit"
	"github.com/raffaelecassia/youtubetoolkit/bigg"
	"google.golang.org/api/youtube/v3"
)

func TestCSVSubscriptions(t *testing.T) {
	t.Run("write a csv with 3 subs", func(t *testing.T) {
		f := newFakeService()
		f.subslist = []bigg.Sub{newSub("A", "TA"), newSub("B", "TB"), newSub("C", "TC")}
		s := youtubetoolkit.NewWithService(f)
		w := &bytes.Buffer{}
		err := s.CSVSubscriptions(w, false)
		if err != nil {
			t.Error(err)
		}
		want := "A,TA\nB,TB\nC,TC\n"
		got := w.String()
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("mismatch (-want +got):\n%s", diff)
		}
	})
}

func TestSubscribe(t *testing.T) {
	t.Run("subscribe to a channel", func(t *testing.T) {
		f := newFakeService()
		s := youtubetoolkit.NewWithService(f)

		_, err := s.Subscribe("CH1")
		if err != nil {
			t.Error(err)
		}
		_, err = s.Subscribe("CH6")
		if err != nil {
			t.Error(err)
		}

		want := []string{"CH1", "CH6"}
		got := f.subinsert
		if !reflect.DeepEqual(want, got) {
			t.Errorf("want: %s got: %s", want, got)
		}
	})
}

func TestCSVBulkSubscribe(t *testing.T) {
	t.Run("subscribe to channels from a csv", func(t *testing.T) {
		f := newFakeService()
		s := youtubetoolkit.NewWithService(f)

		in := "A,TA\nB,TB\nC,TC\nD,TD\nE,TE\n"
		r := strings.NewReader(in)

		err := s.CSVBulkSubscribe(r, io.Discard)
		if err != nil {
			t.Error(err)
		}

		want := []string{"A", "B", "C", "D", "E"}
		got := f.subinsert
		if !reflect.DeepEqual(want, got) {
			t.Errorf("want: %s got: %s", want, got)
		}
	})
}

func TestCSVPlaylists(t *testing.T) {
	t.Run("write a csv with 2 playlists", func(t *testing.T) {
		f := newFakeService()
		f.playlists = []bigg.Playlist{newPlaylist("aaa", "AAA", 3), newPlaylist("bbb", "BBB", 33)}
		s := youtubetoolkit.NewWithService(f)
		w := &bytes.Buffer{}
		err := s.CSVPlaylists(w)
		if err != nil {
			t.Error(err)
		}
		want := "aaa,AAA,3\nbbb,BBB,33\n"
		got := w.String()
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("mismatch (-want +got):\n%s", diff)
		}
	})

}

func TestCSVLastUploads(t *testing.T) {
	t.Run("write a csv with 4 videos from 2 channels", func(t *testing.T) {
		f := newFakeService()
		f.channels = map[string]bigg.Channel{
			"CH1": newChannel("PL1"),
			"CH2": newChannel("PL2"),
		}

		since := time.Now().Add(-time.Hour * 5)

		f.playlistitems = map[string][]bigg.PlaylistItem{
			"PL1": {
				newPlaylistItem("VIDEO1", "T1", "", "", time.Now().Add(-time.Hour*1).Format(bigg.ISO8601_LAYOUT)),
				newPlaylistItem("VIDEO2", "T2", "", "", time.Now().Add(-time.Hour*2).Format(bigg.ISO8601_LAYOUT)),
				newPlaylistItem("VIDEO5", "T5", "", "", time.Now().Add(-time.Hour*6).Format(bigg.ISO8601_LAYOUT)),
				newPlaylistItem("VIDEO6", "T6", "", "", time.Now().Add(-time.Hour*6).Format(bigg.ISO8601_LAYOUT)),
			},
			"PL2": {
				newPlaylistItem("VIDEO3", "T3", "", "", time.Now().Add(-time.Hour*2).Format(bigg.ISO8601_LAYOUT)),
				newPlaylistItem("VIDEO4", "T4", "", "", time.Now().Add(-time.Hour*3).Format(bigg.ISO8601_LAYOUT)),
				newPlaylistItem("VIDEO7", "T7", "", "", time.Now().Add(-time.Hour*7).Format(bigg.ISO8601_LAYOUT)),
				newPlaylistItem("VIDEO8", "T8", "", "", time.Now().Add(-time.Hour*8).Format(bigg.ISO8601_LAYOUT)),
			},
		}
		s := youtubetoolkit.NewWithService(f)

		in := "CH2\nCH1\n"
		inR := strings.NewReader(in)
		oR := &bytes.Buffer{}

		err := s.CSVLastUploads(inR, oR, since)
		if err != nil {
			t.Error(err)
		}

		got := strings.Split(oR.String(), "\n")
		ok := len(got) == 5 && // the last \n counts
			SliceContains(got, "VIDEO1,") && SliceContains(got, "VIDEO2,") &&
			SliceContains(got, "VIDEO3,") && SliceContains(got, "VIDEO4,")

		if !ok {
			t.Errorf("want 4 videos, got: %s", got)
		}
	})
}

//
// fakes
//

func newFakeService() *fakeService {
	return &fakeService{}
}

type fakeService struct {
	subslist      []bigg.Sub
	subinsert     []string
	playlists     []bigg.Playlist
	playlistitems map[string][]bigg.PlaylistItem
	channels      map[string]bigg.Channel
}

// PlaylistDelete implements youtubetoolkit.YoutubeService
func (*fakeService) PlaylistDelete(playlistId string) error {
	panic("unimplemented")
}

// SubscriptionDelete implements youtubetoolkit.YoutubeService
func (*fakeService) SubscriptionDelete(channelId string) error {
	panic("unimplemented")
}

// GetChannelInfo implements youtubetoolkit.YoutubeService
func (s *fakeService) GetChannelInfo(id string) (*bigg.Channel, error) {
	o := s.channels[id]
	return &o, nil
}

// PlaylistItemsListFiltered implements youtubetoolkit.YoutubeService
func (s *fakeService) PlaylistItemsList(playlistId string, filter func(*bigg.PlaylistItem) (bool, error), out chan<- *bigg.PlaylistItem) error {
	for _, v := range s.playlistitems[playlistId] {
		o := v
		ok, err := filter(&o)
		if err != nil {
			return err
		} else if ok {
			out <- &o
		} else {
			return nil
		}
	}
	return nil
}

// PlaylistItemsInsert implements youtubetoolkit.YoutubeService
func (*fakeService) PlaylistItemsInsert(playlistId string, videoId string) (*bigg.PlaylistItem, error) {
	panic("unimplemented")
}

// PlaylistInsert implements youtubetoolkit.YoutubeService
func (*fakeService) PlaylistInsert(title string) (*bigg.Playlist, error) {
	panic("unimplemented")
}

// PlaylistsList implements youtubetoolkit.YoutubeService
func (s *fakeService) PlaylistsList(out chan<- *bigg.Playlist) error {
	for _, v := range s.playlists {
		o := v
		out <- &o
	}
	return nil
}

func (s *fakeService) SubscriptionsList(out chan<- *bigg.Sub) error {
	for _, v := range s.subslist {
		o := v
		out <- &o
	}
	return nil
}

func (s *fakeService) SubscriptionInsert(chanid string) (*bigg.Sub, error) {
	s.subinsert = append(s.subinsert, chanid)
	n := newSub(chanid, chanid)
	return &n, nil
}

//
// support functions
//

func newSub(id, title string) bigg.Sub {
	return bigg.Sub{
		Subscription: &youtube.Subscription{
			Kind: "youtube#subscription",
			Snippet: &youtube.SubscriptionSnippet{
				Title: title,
				// Description: "Channel Description",
				ResourceId: &youtube.ResourceId{
					Kind:      "youtube#channel",
					ChannelId: id,
				},
			},
		},
	}
}

func newPlaylist(id, title string, count int64) bigg.Playlist {
	return bigg.Playlist{
		Playlist: &youtube.Playlist{
			ContentDetails: &youtube.PlaylistContentDetails{
				ItemCount: count,
			},
			Id:   id,
			Kind: "youtube#playlist",
			Snippet: &youtube.PlaylistSnippet{
				// ChannelId:        "",
				// ChannelTitle:     "",
				// DefaultLanguage:  "",
				// Description:      "",
				// Localized:        &youtube.PlaylistLocalization{},
				// PublishedAt:      "",
				// Tags:             []string{},
				// ThumbnailVideoId: "",
				// Thumbnails:       &youtube.ThumbnailDetails{},
				Title: title,
			},
		},
	}
}

func newPlaylistItem(videoid, title, chaId, chaTitle, publishedAt string) bigg.PlaylistItem {
	return bigg.PlaylistItem{
		PlaylistItem: &youtube.PlaylistItem{
			Snippet: &youtube.PlaylistItemSnippet{
				ResourceId: &youtube.ResourceId{
					VideoId: videoid,
				},
				Title:                  title,
				VideoOwnerChannelId:    chaId,
				VideoOwnerChannelTitle: chaTitle,
				PublishedAt:            publishedAt,
			},
		},
	}
}

func newChannel(uploadsPl string) bigg.Channel {
	return bigg.Channel{
		Channel: &youtube.Channel{
			ContentDetails: &youtube.ChannelContentDetails{
				RelatedPlaylists: &youtube.ChannelContentDetailsRelatedPlaylists{
					Uploads: uploadsPl,
				},
			},
		},
	}
}

func SliceContains(s []string, m string) bool {
	for _, v := range s {
		if strings.Contains(v, m) {
			return true
		}
	}
	return false
}
