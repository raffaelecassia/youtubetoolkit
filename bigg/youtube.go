package bigg

import (
	"fmt"
	"sync/atomic"

	"google.golang.org/api/youtube/v3"
)

type Youtube struct {
	svc  *youtube.Service
	cost uint32
}

type Sub struct {
	*youtube.Subscription
}

type Playlist struct {
	*youtube.Playlist
}

type PlaylistItem struct {
	*youtube.PlaylistItem
}

type Channel struct {
	*youtube.Channel
}

const ISO8601_LAYOUT string = "2006-01-02T15:04:05Z0700"

func (s *Youtube) addcost(q uint32) {
	atomic.AddUint32(&s.cost, q)
}

func (s *Youtube) GetCost() uint32 {
	return atomic.LoadUint32(&s.cost)
}

//
// SUBSCRIPTIONS
//

// SubscriptionsList sends to out all the user subscriptions.
// Items will contain only the "snippet" resource property (https://developers.google.com/youtube/v3/docs/subscriptions#snippet).
// The GCloud quota impact is 1 unit every 50 items fetched.
func (s *Youtube) SubscriptionsList(out chan<- *Sub) error {
	call := s.svc.Subscriptions.List([]string{"snippet"})
	call.Mine(true)
	call.MaxResults(50)
	call.Order("alphabetical")
	t := "-"
	for t != "" {
		s.addcost(1)
		r, err := call.Do()
		if err != nil {
			return fmt.Errorf("subs list error (page %s): %w", t, err)
		}
		for _, v := range r.Items {
			out <- &Sub{v}
		}
		t = r.NextPageToken
		call.PageToken(t)
	}
	return nil
}

// SubscriptionInsert adds a subscription for the authenticated user's channel.
// The GCloud quota impact is 50 units.
func (s *Youtube) SubscriptionInsert(channelId string) (*Sub, error) {
	sub := &youtube.Subscription{
		Kind: "youtube#subscription",
		Snippet: &youtube.SubscriptionSnippet{
			ResourceId: &youtube.ResourceId{
				Kind:      "youtube#channel",
				ChannelId: channelId,
			},
		},
	}
	call := s.svc.Subscriptions.Insert([]string{"snippet"}, sub)
	s.addcost(50)
	r, err := call.Do()
	return &Sub{r}, err
}

// SubscriptionDelete delete a channel from subscriptions for the authenticated user's channel.
// This command will results in two api requests (subscriptionId from the channelId and the delete op).
// The GCloud quota impact is 51 units.
func (s *Youtube) SubscriptionDelete(channelId string) error {
	// find the sub id from channel id
	lcall := s.svc.Subscriptions.List([]string{"snippet"})
	lcall.Mine(true)
	lcall.MaxResults(1)
	lcall.ForChannelId(channelId)
	s.addcost(1)
	list, err := lcall.Do()
	if err != nil {
		return fmt.Errorf("subs delete error (list.ForChannelId '%s'): %w", channelId, err)
	}
	if len(list.Items) != 1 {
		return fmt.Errorf("subs delete: channelId '%s' not in subscriptions", channelId)
	}
	subid := list.Items[0].Id
	// delete sub
	call := s.svc.Subscriptions.Delete(subid)
	s.addcost(50)
	err = call.Do()
	if err != nil {
		return fmt.Errorf("subs delete error (channelId '%s', subId '%s'): %w", channelId, subid, err)
	}
	return nil
}

//
// PLAYLISTS
//

// PlaylistsList returns all user playlists.
// Items will contain the "snippet" (https://developers.google.com/youtube/v3/docs/playlists#snippet)
// and the "contentDetails" (https://developers.google.com/youtube/v3/docs/playlists#contentDetails) resource properties
// The GCloud quota impact is 1 unit every 50 items fetched.
func (s *Youtube) PlaylistsList(out chan<- *Playlist) error {
	call := s.svc.Playlists.List([]string{"snippet", "contentDetails"})
	call.Mine(true)
	call.MaxResults(50)
	t := "-"
	for t != "" {
		s.addcost(1)
		r, err := call.Do()
		if err != nil {
			return fmt.Errorf("playlist list error (page %s): %w", t, err)
		}
		for _, v := range r.Items {
			out <- &Playlist{v}
		}
		t = r.NextPageToken
		call.PageToken(t)
	}
	return nil
}

// PlaylistInsert creates a new private playlist for the authenticated user.
// The GCloud quota impact is 50 units.
func (s *Youtube) PlaylistInsert(title string) (*Playlist, error) {
	pl := &youtube.Playlist{
		Snippet: &youtube.PlaylistSnippet{
			Title: title,
			// Description:      "",
			// DefaultLanguage:  "",
			// Tags: []string{},
		},
		Status: &youtube.PlaylistStatus{
			PrivacyStatus: "private",
		},
		// Localizations: map[string]youtube.PlaylistLocalization{},
	}
	call := s.svc.Playlists.Insert([]string{"snippet", "status"}, pl)
	s.addcost(50)
	pl, err := call.Do()
	return &Playlist{pl}, err
}

// PlaylistItemsList sends to out the items of a playlist until the filter function returns false.
// Playlist id can be a user own playlist or a public playlist.
// Items will contain only the "snippet" resource property (https://developers.google.com/youtube/v3/docs/playlistItems#snippet).
// The GCloud quota impact is 1 unit every 50 items fetched.
func (s *Youtube) PlaylistItemsList(playlistId string, filter func(*PlaylistItem) bool, out chan<- *PlaylistItem) error {
	call := s.svc.PlaylistItems.List([]string{"snippet"})
	call.PlaylistId(playlistId)
	call.MaxResults(50)
	t := "-"
	for t != "" {
		s.addcost(1)
		res, err := call.Do()
		if err != nil {
			return fmt.Errorf("playlist items list error (id=\"%s\" and page=\"%s\"): %w", playlistId, t, err)
		}
		for _, pli := range res.Items {
			o := &PlaylistItem{pli}
			if filter(o) {
				out <- o
			} else {
				return nil
			}
		}
		t = res.NextPageToken
		call.PageToken(t)
	}
	return nil
}

// PlaylistItemsInsert adds a video to a playlist.
// The GCloud quota impact is 50 units.
func (s *Youtube) PlaylistItemsInsert(playlistId, videoId string) (*PlaylistItem, error) {
	pli := &youtube.PlaylistItem{
		// ContentDetails: &youtube.PlaylistItemContentDetails{
		// 	EndAt:            "",
		// 	Note:             "",
		// 	StartAt:          "",
		// },
		Snippet: &youtube.PlaylistItemSnippet{
			PlaylistId: playlistId,
			Position:   0,
			ResourceId: &youtube.ResourceId{
				Kind:    "youtube#video",
				VideoId: videoId,
			},
		},
	}
	call := s.svc.PlaylistItems.Insert([]string{"snippet"}, pli)
	s.addcost(50)
	pli, err := call.Do()
	return &PlaylistItem{pli}, err
}

// GetChannelInfo returns channel info from ID.
// Returned value will contain only the "contentDetails" resource property (https://developers.google.com/youtube/v3/docs/channels#contentDetails).
// The GCloud quota impact is 1 unit
func (s *Youtube) GetChannelInfo(id string) (*Channel, error) {
	call := s.svc.Channels.List([]string{"contentDetails"})
	// call.ForUsername("username")
	call.Id(id)
	s.addcost(1)
	res, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("channel list error for id=\"%s\": %w", id, err)
	} else if len(res.Items) == 0 {
		return nil, fmt.Errorf("channel not found for id=\"%s\"", id)
	} else if len(res.Items) > 1 {
		// expecting a single channel only... should not happen
		return nil, fmt.Errorf("channel WTF? for id=\"%s\"", id)
	}
	return &Channel{res.Items[0]}, nil
}
