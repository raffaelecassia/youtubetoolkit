package youtubetoolkit

import (
	"fmt"
	"reflect"
)

type Item interface {
	// AsRecord returns a string slice with the values of
	// this Item selected by the fields param.
	AsRecord(fields *[]string) []string
}

type sub struct {
	SubscriptionId,
	ChannelId,
	ChannelTitle,
	ChannelUrl,
	ChannelThumbUrl string `json:",omitempty"`
}

func (r *sub) AsRecord(fields *[]string) []string {
	return item2record(r, fields)
}

type playlist struct {
	PlaylistId,
	PlaylistTitle string `json:",omitempty"`
	VideoCount int64 `json:",omitempty"`
}

func (r *playlist) AsRecord(fields *[]string) []string {
	return item2record(r, fields)
}

type playlistItem struct {
	PlaylistItemId,
	ChannelId,
	ChannelTitle,
	ChannelUrl,
	VideoId,
	VideoTitle,
	VideoUrl,
	PublishedAt string `json:",omitempty"`
}

func (r *playlistItem) AsRecord(fields *[]string) []string {
	return item2record(r, fields)
}

func item2record(input Item, fields *[]string) []string {
	out := []string{}
	elem := reflect.ValueOf(input).Elem()
	for _, c := range *fields {
		f := elem.FieldByName(c)
		if f.CanInt() {
			out = append(out, fmt.Sprint(f.Int()))
		} else {
			out = append(out, f.String())
		}
	}
	return out
}
