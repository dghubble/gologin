package github

import (
	"net/http"

	"github.com/dghubble/gologin"
	oauth2Login "github.com/dghubble/gologin/oauth2"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// LoginHandler handles Github OAuth2 login requests.
func LoginHandler(config *oauth2.Config, stater oauth2Login.StateSource) http.Handler {
	return oauth2Login.LoginHandler(config, stater)
}

// CallbackHandler handles Github OAuth2 callback requests. If authentication
// succeeds, handling is delegated to the SuccessHandler which is provided
// with the Github user and access token. Otherwise, an ErrorHandler handles
// responding.
func CallbackHandler(config *oauth2.Config, stater oauth2Login.StateSource, success SuccessHandler, failure gologin.ErrorHandler) http.Handler {
	oauth2Success := successWithUser(config, success, failure)
	return oauth2Login.CallbackHandler(config, stater, oauth2Success, failure)
}

// successWithUser returns an oauth2 success handler which wraps the github
// success and failure handlers. It verifies token credentials to obtain the
// Github User object for inclusion in data passed on success.
func successWithUser(config *oauth2.Config, success SuccessHandler, failure gologin.ErrorHandler) oauth2Login.SuccessHandler {
	fn := func(w http.ResponseWriter, req *http.Request, accessToken string) {
		token := &oauth2.Token{AccessToken: accessToken}
		httpClient := config.Client(oauth2.NoContext, token)
		githubClient := github.NewClient(httpClient)
		user, resp, err := githubClient.Users.Get("")
		err = validateResponse(user, resp, err)
		if err != nil {
			failure.ServeHTTP(w, err, http.StatusBadRequest)
			return
		}
		success.ServeHTTP(w, req, user, accessToken)
	}
	return oauth2Login.SuccessHandlerFunc(fn)
}
