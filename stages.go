package youtubetoolkit

import (
	"encoding/csv"
	"fmt"
	"io"
	"sync"

	"github.com/raffaelecassia/youtubetoolkit/bigg"
)

func (tk *Toolkit) channelIds2VideoUploads(errors chan<- error, input <-chan string, filter func(*bigg.PlaylistItem) bool, numDigesters int) <-chan *bigg.PlaylistItem {
	output := make(chan *bigg.PlaylistItem, 10)
	var wg sync.WaitGroup
	wg.Add(numDigesters)
	for i := 0; i < numDigesters; i++ {
		go func() {
			for id := range input {
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

func (tk *Toolkit) videoIds2playlist(errors chan<- error, playlistId string, input <-chan string) <-chan *bigg.PlaylistItem {
	output := make(chan *bigg.PlaylistItem)
	go func() {
		for id := range input {
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

func csv2records(errors chan<- error, input io.Reader) <-chan []string {
	output := make(chan []string)
	go func() {
		reader := csv.NewReader(input)
		for {
			record, err := reader.Read()
			if err != nil {
				if err == io.EOF {
					close(output)
					return
				}
				errors <- fmt.Errorf("csv read error: %w", err)
			} else {
				output <- record
			}
		}
	}()
	return output
}

func sinkrecords2csv(errors chan<- error, input <-chan []string, output io.Writer) {
	w := csv.NewWriter(output)
	// w.Comma = '\t'
	for record := range input {
		err := w.Write(record)
		if err != nil {
			// FIXME fatal?
			errors <- fmt.Errorf("csv write error: %w", err)
		}
	}
	w.Flush()
	if w.Error() != nil {
		errors <- fmt.Errorf("csv write error: %w", w.Error())
	}
}

func recordColumnFilter(input <-chan []string, col uint8) <-chan string {
	output := make(chan string)
	go func() {
		for r := range input {
			output <- r[col]
		}
		close(output)
	}()
	return output
}

func sub2record(input <-chan *bigg.Sub, extracols bool) <-chan []string {
	output := make(chan []string, 10)
	go func() {
		for s := range input {
			if extracols {
				output <- []string{
					s.Snippet.ResourceId.ChannelId,
					s.Snippet.Title,
					fmt.Sprintf("https://www.youtube.com/channel/%s", s.Snippet.ResourceId.ChannelId),
					s.Snippet.Thumbnails.Default.Url,
					s.Id,
				}
			} else {
				output <- []string{
					s.Snippet.ResourceId.ChannelId,
					s.Snippet.Title,
				}
			}
		}
		close(output)
	}()
	return output
}

func playlist2record(input <-chan *bigg.Playlist) <-chan []string {
	output := make(chan []string, 10)
	go func() {
		for i := range input {
			output <- []string{
				i.Id,
				i.Snippet.Title,
				fmt.Sprint(i.ContentDetails.ItemCount),
			}
		}
		close(output)
	}()
	return output
}

func playlistItem2record(input <-chan *bigg.PlaylistItem) <-chan []string {
	output := make(chan []string, 10)
	go func() {
		for i := range input {
			output <- []string{
				i.Snippet.ResourceId.VideoId,
				i.Snippet.Title,
				i.Snippet.VideoOwnerChannelId,
				i.Snippet.VideoOwnerChannelTitle,
			}
		}
		close(output)
	}()
	return output
}

// func merge[T any](cs ...<-chan *T) <-chan *T {
// 	var wg sync.WaitGroup
// 	out := make(chan *T)
// 	// Start an output goroutine for each input channel in cs.  output
// 	// copies values from c to out until c is closed, then calls wg.Done.
// 	output := func(c <-chan *T) {
// 		for n := range c {
// 			out <- n
// 		}
// 		wg.Done()
// 	}
// 	wg.Add(len(cs))
// 	for _, c := range cs {
// 		go output(c)
// 	}
// 	// Start a goroutine to close out once all the output goroutines are
// 	// done.  This must start after the wg.Add call.
// 	go func() {
// 		wg.Wait()
// 		close(out)
// 	}()
// 	return out
// }
