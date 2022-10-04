package youtubetoolkit

import (
	"fmt"
	"sort"
	"sync"

	"github.com/raffaelecassia/youtubetoolkit/bigg"
)

func (tk *Toolkit) channels2newsubscriptions(errors chan<- error, channelIds <-chan string) <-chan *bigg.Sub {
	subs := make(chan *bigg.Sub)
	go func() {
		for channelId := range channelIds {
			tk.logf("subscribing to %s... ", channelId)
			sub, err := tk.service.SubscriptionInsert(channelId)
			if err != nil {
				errors <- fmt.Errorf("channel %s subscribe: %w", channelId, err)
				tk.log("fail!")
			} else {
				subs <- sub
				tk.log("channel", sub.Snippet.Title, "added")
			}
		}
		close(subs)
	}()
	return subs
}

func (tk *Toolkit) channels2channelvideouploads(errors chan<- error, channelIds <-chan string, filter func(*bigg.PlaylistItem) (bool, error), numDigesters int) <-chan *bigg.PlaylistItem {
	output := make(chan *bigg.PlaylistItem, 10)
	var wg sync.WaitGroup
	wg.Add(numDigesters)
	for i := 0; i < numDigesters; i++ {
		go func() {
			for id := range channelIds {
				tk.log("Checking channel", id)
				c, err := tk.service.GetChannelInfo(id)
				if err != nil {
					errors <- err
				} else {
					plid := c.ContentDetails.RelatedPlaylists.Uploads
					err := tk.service.PlaylistItemsList(plid, filter, output)
					if err != nil {
						errors <- err
					}
				}
			}
			wg.Done()
		}()
	}
	go func() {
		wg.Wait()
		close(output)
	}()
	return output
}

func (tk *Toolkit) videos2playlist(errors chan<- error, playlistId string, videoIds <-chan string) <-chan *bigg.PlaylistItem {
	output := make(chan *bigg.PlaylistItem)
	go func() {
		for id := range videoIds {
			tk.log("Adding video", id)
			pli, err := tk.service.PlaylistItemsInsert(playlistId, id)
			if err != nil {
				errors <- err
			} else {
				output <- pli
			}
		}
		close(output)
	}()
	return output
}

func sortPlaylistItemByPublishedAt(input <-chan *bigg.PlaylistItem) <-chan *bigg.PlaylistItem {
	output := make(chan *bigg.PlaylistItem, 10)
	go func() {
		items := []*bigg.PlaylistItem{}
		// buffers all items
		for i := range input {
			items = append(items, i)
		}
		// sorts
		sort.Slice(items, func(i, j int) bool {
			return items[i].Snippet.PublishedAt < items[j].Snippet.PublishedAt
		})
		// sends items to chan
		for _, pi := range items {
			output <- pi
		}
		close(output)
	}()
	return output
}

func sub2item(input <-chan *bigg.Sub) <-chan Item {
	output := make(chan Item, 10)
	go func() {
		for s := range input {
			output <- &sub{
				SubscriptionId:  s.Id,
				ChannelId:       s.Snippet.ResourceId.ChannelId,
				ChannelTitle:    s.Snippet.Title,
				ChannelUrl:      fmt.Sprintf("https://www.youtube.com/channel/%s", s.Snippet.ResourceId.ChannelId),
				ChannelThumbUrl: s.Snippet.Thumbnails.Default.Url,
			}
		}
		close(output)
	}()
	return output
}

func playlist2item(input <-chan *bigg.Playlist) <-chan Item {
	output := make(chan Item, 10)
	go func() {
		for i := range input {
			output <- &playlist{
				PlaylistId:    i.Id,
				PlaylistTitle: i.Snippet.Title,
				VideoCount:    i.ContentDetails.ItemCount,
			}
		}
		close(output)
	}()
	return output
}

func playlistItem2item(input <-chan *bigg.PlaylistItem) <-chan Item {
	output := make(chan Item, 10)
	go func() {
		for i := range input {
			output <- &playlistItem{
				PlaylistItemId: i.Id,
				ChannelId:      i.Snippet.VideoOwnerChannelId,
				ChannelTitle:   i.Snippet.VideoOwnerChannelTitle,
				ChannelUrl:     fmt.Sprintf("https://www.youtube.com/channel/%s", i.Snippet.VideoOwnerChannelId),
				VideoId:        i.Snippet.ResourceId.VideoId,
				VideoTitle:     i.Snippet.Title,
				VideoUrl:       fmt.Sprintf("https://www.youtube.com/watch?v=%s", i.Snippet.ResourceId.VideoId),
				PublishedAt:    i.Snippet.PublishedAt,
			}
		}
		close(output)
	}()
	return output
}
