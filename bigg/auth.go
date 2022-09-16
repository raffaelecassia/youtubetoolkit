package bigg

import (
	"context"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"syscall"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

type Client struct {
	ClientID     string
	ClientSecret string
	TokenFile    string

	context    context.Context
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{
		context: context.Background(),
	}
}

func (a *Client) SetSecretFromFile(file string) error {
	var err error
	a.ClientID, a.ClientSecret, err = getOauth2ClientSecret(file)
	if err != nil {
		return fmt.Errorf("error reading client secret file %s: %w", file, err)
	}
	return nil
}

func (a *Client) EnableLogTransport() {
	a.context = context.WithValue(a.context, oauth2.HTTPClient, &http.Client{
		Transport: &loggingRoundTripper{http.DefaultTransport},
	})
}

// Authorize performs the OAuth authorization to google api
func (a *Client) Authorize() error {

	config := &oauth2.Config{
		ClientID:     a.ClientID,
		ClientSecret: a.ClientSecret,
		Endpoint:     google.Endpoint,
		Scopes:       []string{youtube.YoutubeScope},
	}

	token, tferr := tokenFromFile(a.TokenFile)

	if tferr == errTokenFileNotFound {
		token, err := tokenFromWeb(a.context, config, os.Stderr)
		if err != nil {
			return fmt.Errorf("oauth error: %w", err)
		}
		err = saveToken(a.TokenFile, token)
		if err != nil {
			return fmt.Errorf("save token file error: %w", err)
		}
		a.httpClient = config.Client(a.context, token)
		return nil
	}

	if tferr == nil {
		a.httpClient = config.Client(a.context, token)
		// refresh and save
		if ct, ok := a.httpClient.Transport.(*oauth2.Transport); ok {
			token2, err := ct.Source.Token()
			if err != nil {
				return fmt.Errorf("oauth error: %w", err)
			}
			if token.AccessToken != token2.AccessToken {
				err = saveToken(a.TokenFile, token2)
				if err != nil {
					return fmt.Errorf("save token file error: %w", err)
				}
			}
		}
		return nil
	}

	return fmt.Errorf("read token file error: %v", tferr)
}

func (a *Client) NewYoutubeService() (*Youtube, error) {
	if a.httpClient == nil {
		return nil, errors.New("Client not authorized")
	}
	svc, err := youtube.NewService(a.context, option.WithHTTPClient(a.httpClient))
	return &Youtube{svc, 0}, err
}

//
// support functions
//

func tokenFromWeb(ctx context.Context, config *oauth2.Config, out io.Writer) (*oauth2.Token, error) {
	ch := make(chan string)
	randState := fmt.Sprintf("st%d", time.Now().UnixNano()) // TODO change this
	var localerr bool

	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/favicon.ico" {
			http.Error(rw, "", 404)
			return
		}
		if req.FormValue("state") != randState {
			fmt.Fprintf(out, "Auth error. State doesn't match: req = %#v\n", req)
			localerr = true
			http.Error(rw, "Error", 500)
			return
		}
		if code := req.FormValue("code"); code != "" {
			fmt.Fprintf(rw, "<h1>Success</h1>")
			rw.(http.Flusher).Flush()
			ch <- code
			return
		}
		fmt.Fprintf(out, "Auth error. No code in request\n")
		localerr = true
		ch <- ""
		http.Error(rw, "Error", 500)
	}))
	defer ts.Close()

	config.RedirectURL = ts.URL
	authURL := config.AuthCodeURL(randState)
	go openURL(authURL)
	fmt.Fprintf(out, "Authorize this app at: %s\n", authURL)

	code := <-ch

	if code == "" || localerr {
		return nil, fmt.Errorf("auth error")
	}
	token, err := config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("auth token exchange error: %v", err)
	}
	return token, nil
}

// reads clientSecretFile json file and returns clientid and secret.
func getOauth2ClientSecret(clientSecretFile string) (clientid string, secret string, err error) {
	var data []byte
	data, err = os.ReadFile(clientSecretFile)
	if err != nil {
		return
	}
	var j map[string]interface{}
	err = json.Unmarshal(data, &j)
	if err != nil {
		return
	}
	m := j["installed"].(map[string]interface{})
	clientid = m["client_id"].(string)
	secret = m["client_secret"].(string)
	return
}

var errTokenFileNotFound = errors.New("token file not found or corrupted")

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		if e, ok := err.(*fs.PathError); ok && e.Err == syscall.ENOENT {
			return nil, errTokenFileNotFound
		}
		return nil, err
	}
	t := new(oauth2.Token)
	err = gob.NewDecoder(f).Decode(t)
	if err != nil {
		return nil, errTokenFileNotFound
	}
	return t, nil
}

func saveToken(file string, token *oauth2.Token) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()
	gob.NewEncoder(f).Encode(token)
	return nil
}

func openURL(url string) {
	try := []string{"xdg-open", "google-chrome", "open"}
	for _, bin := range try {
		err := exec.Command(bin, url).Run()
		if err == nil {
			return
		}
	}
	// fmt.Println("Error opening URL in browser.")
}
