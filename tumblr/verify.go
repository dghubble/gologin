package tumblr

import (
	"net/http"

	"github.com/dghubble/sling"
)

const tumblrAPI = "https://api.tumblr.com/v2/"

// userInfoResponse is a Tubmlr user info response.
type userInfoResponse struct {
	Meta     meta `json:"meta"`
	Response struct {
		User `json:"user"`
	} `json:"response"`
}

// User is a Tumblr user.
//
// Note that Tumblr does not provide stable user identifiers.
type User struct {
	Name      string `json:"name"`
	Following int64  `json:"following"`
	Likes     int64  `json:"likes"`
}

// meta is a metadata struct Tumblr includes in responses.
type meta struct {
	Status  int    `json:"status"`
	Message string `json:"msg"`
}

// client is a Tumblr client for obtaining a User.
type client struct {
	sling *sling.Sling
}

func newClient(httpClient *http.Client) *client {
	base := sling.New().Client(httpClient).Base(tumblrAPI)
	return &client{
		sling: base,
	}
}

func (c *client) UserInfo() (*User, *http.Response, error) {
	userResp := new(userInfoResponse)
	resp, err := c.sling.New().Get("user/info").ReceiveSuccess(userResp)
	return &userResp.Response.User, resp, err
}
