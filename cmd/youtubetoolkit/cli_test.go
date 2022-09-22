//go:build integration

// Integration tests
//
// To run these tests a client_secret.json and a valid auth token (see the const values)
// inside the bin/ dir (where make will place the output of go build) are required.
// Tests are run against a build of the toolkit under bin/ dir.
//
// Just a warning. Running these tests will cost you quota.
// Subscriptions test will cost at least 50 + 1 + 51 = 102 units.
// Playlists test will cost at least 50 + 1 + 50 + 1 + 50 = 152 units.
// Subs and playlists pagination may require more quota use.
//
// To execute, launch `make acceptance` from root dir.
//
// TODO cover more cases:
// - subs and playlists pagination fetching (when >50 items)
//
// vscode users should add these fields to workspace settings.json to allow gopls to
// build this file but not execute it during unit testing:
// { "go.buildTags": "integration", "go.testTags": "" }
package main_test

import (
	"fmt"
	"io"
	"math/rand"
	"os/exec"
	"strings"
	"testing"
	"time"
)

const (
	BIN    = "../../bin/youtubetoolkit"
	SECRET = "../../bin/client_secret.json"
	TOKEN  = "../../bin/integrationtests.token"
)

func Test_CLI_Subscriptions(t *testing.T) {
	testChannelID := "UCuAXFkgsw1L7xaCfnd5JJOw"

	// adds a channel to subs
	gotOut, gotErr := run(t, "", "subscriptions", "add", testChannelID)
	want := gotOut == "" &&
		strings.Contains(gotErr, fmt.Sprintf("subscribing to %s...", testChannelID)) &&
		strings.Contains(gotErr, "added") &&
		strings.Contains(gotErr, "Quota cost: 50 units")
	if !want {
		t.Fatalf("should insert channel '%s'.\nOUT=%s\nERR=%s\n", testChannelID, gotOut, gotErr)
	}

	// needed 'cause a fast insert+list may results in a not up-to-date list results...
	time.Sleep(1 * time.Second)

	// checks subs list
	gotOut, gotErr = run(t, "", "subscriptions", "list")
	want = strings.Contains(gotOut, testChannelID)
	if !want {
		// not fatal, it should at least try to delete the subscription
		t.Errorf("should contain channel '%s'.\nOUT=%s\nERR=%s\n", testChannelID, gotOut, gotErr)
	}

	// delete
	gotOut, gotErr = run(t, "", "subscriptions", "del", testChannelID)
	want = gotOut == "" && gotErr == fmt.Sprintf("unsubscribing from %s... done\nQuota cost: 51 units\n", testChannelID)
	if !want {
		t.Fatalf("should delete channel '%s'.\nOUT=%s\nERR=%s\n", testChannelID, gotOut, gotErr)
	}
}

func Test_CLI_Playlists(t *testing.T) {
	testVideoID := "dQw4w9WgXcQ"
	playlistName := "test-" + randSeq(10)

	// adds a new playlist
	gotOut, gotErr := run(t, "", "playlists", "new", playlistName)
	want := gotOut != "" && gotErr == "Quota cost: 50 units\n"
	if !want {
		t.Fatalf("should create playlist '%s'.\nOUT=%s\nERR=%s\n", playlistName, gotOut, gotErr)
	}

	playlistID := gotOut[:len(gotOut)-1] // removes the newline

	// needed 'cause a fast insert+list may results in a not up-to-date list results...
	time.Sleep(1 * time.Second)

	// checks playlists
	gotOut, _ = run(t, "", "playlists")
	want = strings.Contains(gotOut, fmt.Sprintf("%s,%s,0\n", playlistID, playlistName))
	if !want {
		// not fatal, it should at least try to delete the playlist
		t.Errorf("should have a playlist with id '%s'.\nOUT=%s\nERR=%s\n", playlistID, gotOut, gotErr)
	}

	// adds a video to playlist
	gotOut, gotErr = run(t, "", "playlist", "--id", playlistID, "add", testVideoID)
	want = gotOut == "" && gotErr == "Quota cost: 50 units\n"
	if !want {
		t.Errorf("should add video '%s' to playlist '%s'.\nOUT=%s\nERR=%s\n", testVideoID, playlistID, gotOut, gotErr)
	}

	// needed 'cause a fast insert+list may results in a not up-to-date list results...
	time.Sleep(1 * time.Second)

	// check playlist videos
	gotOut, gotErr = run(t, "", "playlist", "--id", playlistID)
	want = strings.Contains(gotOut, testVideoID) && gotErr == "Quota cost: 1 units\n"
	if !want {
		t.Errorf("playlist '%s' should contains video '%s'.\nOUT=%s\nERR=%s\n", playlistID, testVideoID, gotOut, gotErr)
	}

	// deletes
	gotOut, gotErr = run(t, "", "playlists", "del", playlistID)
	want = gotOut == "" && gotErr == "Quota cost: 50 units\n"
	if !want {
		t.Fatalf("should delete playlist '%s'.\nOUT=%s\nERR=%s\n", playlistID, gotOut, gotErr)
	}

}

func run(t *testing.T, stdin string, args ...string) (stdout, stderr string) {
	commonargs := []string{
		"--client-secret", SECRET,
		"--token", TOKEN,
	}

	cmd := exec.Command(BIN, append(commonargs, args...)...)

	// pipes
	if stdin != "" {
		stdinP, err := cmd.StdinPipe()
		if err != nil {
			t.Fatal(err)
		}
		// write
		_, err = io.WriteString(stdinP, stdin)
		if err != nil {
			t.Fatal(err)
		}
		stdinP.Close()
	}
	stdoutP, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	stderrP, err := cmd.StderrPipe()
	if err != nil {
		t.Fatal(err)
	}

	// start cmd
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}

	// read stdout and stderr
	stdoutB, err := io.ReadAll(stdoutP)
	if err != nil {
		t.Fatal(err)
	}
	stderrB, err := io.ReadAll(stderrP)
	if err != nil {
		t.Fatal(err)
	}

	// wait for cmd
	if err := cmd.Wait(); err != nil {
		t.Fatal(err)
	}

	return string(stdoutB), string(stderrB)
}

func randSeq(n int) string {
	rand.Seed(time.Now().UnixNano())
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
