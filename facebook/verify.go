package facebook

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/dghubble/sling"
)

const facebookAPI = "https://graph.facebook.com/v2.9/"

// User is a Facebook user.
//
// Note that user ids are unique to each app.
type User struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Verified bool   `json:"verified"`
	Picture  struct {
		Data struct {
			Height     int    `json:"height"`
			Width      int    `json:"width"`
			URL        string `json:"url"`
			Silhouette bool   `json:"is_silhouette"`
		}
	} `json:"picture"`
}

// client is a Facebook client for obtaining the current User.
type client struct {
	c     *http.Client
	sling *sling.Sling
}

func newClient(httpClient *http.Client) *client {
	base := sling.New().Client(httpClient).Base(facebookAPI)
	return &client{
		c:     httpClient,
		sling: base,
	}
}

func (c *client) Me() (*User, *http.Response, error) {
	return c.Get([]string{"name", "email"})
}

func (c *client) Get(fields []string) (*User, *http.Response, error) {
	user := new(User)
	path := fmt.Sprintf("me?fields=%s", strings.Join(fields[:], ","))
	// Facebook returns JSON as Content-Type text/javascript :(
	// Set Accept header to receive proper Content-Type application/json
	// so Sling will decode into the struct
	resp, err := c.sling.New().Set("Accept", "application/json").Get(path).ReceiveSuccess(user)
	return user, resp, err
}
