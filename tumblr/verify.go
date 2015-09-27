package tumblr

import (
	"net/http"

	"github.com/dghubble/ctxh"
	"github.com/dghubble/gologin"
	oauth1Login "github.com/dghubble/gologin/oauth1"
	"github.com/dghubble/oauth1"
	"github.com/dghubble/sling"
	"golang.org/x/net/context"
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

func verifyUser(config *oauth1.Config, success, failure ctxh.ContextHandler) ctxh.ContextHandler {
	if failure == nil {
		failure = gologin.DefaultFailureHandler
	}
	fn := func(ctx context.Context, w http.ResponseWriter, req *http.Request) {
		accessToken, accessSecret, err := oauth1Login.AccessTokenFromContext(ctx)
		if err != nil {
			ctx = gologin.WithError(ctx, err)
			failure.ServeHTTP(ctx, w, req)
			return
		}
		httpClient := config.Client(ctx, oauth1.NewToken(accessToken, accessSecret))
		tumblrClient := newClient(httpClient)
		user, resp, err := tumblrClient.UserInfo()
		err = validateResponse(user, resp, err)
		if err != nil {
			ctx = gologin.WithError(ctx, err)
			failure.ServeHTTP(ctx, w, req)
			return
		}
		ctx = WithUser(ctx, user)
		success.ServeHTTP(ctx, w, req)
	}
	return ctxh.ContextHandlerFunc(fn)
}

// validateResponse returns an error if the given Tumblr User, raw
// http.Response, or error are unexpected. Returns nil if they are valid.
func validateResponse(user *User, resp *http.Response, err error) error {
	if err != nil || resp.StatusCode != http.StatusOK {
		return ErrUnableToGetTumblrUser
	}
	if user == nil || user.Name == "" {
		return ErrUnableToGetTumblrUser
	}
	return nil
}
