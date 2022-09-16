package youtubetoolkit

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/raffaelecassia/youtubetoolkit/bigg"
)

func (tk *Toolkit) SetLogWriter(output io.Writer) {
	tk.logWriter = output
}

func (tk *Toolkit) log(a ...any) {
	fmt.Fprintln(tk.logWriter, a...)
}

func (tk *Toolkit) logf(format string, a ...any) {
	fmt.Fprintf(tk.logWriter, format, a...)
}

func sinceDatePlaylistItems(since time.Time) func(i *bigg.PlaylistItem) bool {
	return func(i *bigg.PlaylistItem) bool {
		// The date and time that the item was added to the playlist.
		// The value is specified in ISO 8601 format.
		// https://developers.google.com/youtube/v3/docs/playlistItems#snippet.publishedAt
		pub, err := time.Parse(bigg.ISO8601_LAYOUT, i.Snippet.PublishedAt)
		if err != nil {
			fmt.Fprintf(os.Stderr, "time error %s: %v\n", i.Snippet.PublishedAt, err) // FIXME must be always shown
			return false
		}
		return pub.After(since)
	}
}

func allPlaylistItems() func(i *bigg.PlaylistItem) bool {
	return func(i *bigg.PlaylistItem) bool {
		return true
	}
}

// MultiErrors is used when one error is not enough
type MultiErrors []error

func (e MultiErrors) Error() string {
	if len(e) == 1 {
		return e[0].Error()
	}
	msg := "multiple errors:"
	for _, err := range e {
		msg += "\n- " + err.Error()
	}
	return msg
}

// multiErrorsHandler handles multiple errors.
// It returns two channels: a send-only for errors and a receive-only to get a MultiErrors.
// Errors received from the first channel are internally stored until the channel is closed.
// Thus it sends a single MultiErrors to the second channel (or nil if no errors).
func multiErrorsHandler() (chan<- error, <-chan error) {
	errs := make(chan error, 1)
	err := make(chan error, 1)
	var out MultiErrors
	go func() {
		for e := range errs {
			if e != nil {
				out = append(out, e)
			}
		}
		if len(out) == 0 {
			err <- nil
		} else {
			err <- out
		}
		close(err)
	}()
	return errs, err
}
