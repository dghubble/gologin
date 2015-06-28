package github

import (
	"net/http"

	"github.com/dghubble/gologin"
	oauth2Login "github.com/dghubble/gologin/oauth2"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// LoginHandlerConfig configures a LoginHandler.
type LoginHandlerConfig struct {
	OAuth2Config *oauth2.Config
	Success      SuccessHandler
	Failure      gologin.ErrorHandler
}

// LoginHandler handles Github OAuth2 login and callback requests. If
// authentication succeeds, handling is delegated to a SuccessHandler which
// is provided with the Github user and accessToken. Otherwise, an ErrorHandler
// handles responding.
type LoginHandler struct {
	oauth2Config *oauth2.Config
	success      SuccessHandler
	failure      gologin.ErrorHandler
}

// NewLoginHandler returns a new Handler.
func NewLoginHandler(config *LoginHandlerConfig) http.Handler {
	handler := &LoginHandler{
		oauth2Config: config.OAuth2Config,
		success:      config.Success,
		failure:      config.Failure,
	}
	return oauth2Login.NewLoginHandler(&oauth2Login.Config{
		OAuth2Config: config.OAuth2Config,
		Success:      oauth2Login.SuccessHandlerFunc(handler.successWithUser),
		Failure:      config.Failure,
	})
}

// successWithUser is an oauth2 success handler which wraps the github success
// and failure handlers. It verifies token credentials to obtain the Github
// User object for inclusion in data passed on success.
func (h *LoginHandler) successWithUser(w http.ResponseWriter, req *http.Request, accessToken string) {
	token := &oauth2.Token{AccessToken: accessToken}
	httpClient := h.oauth2Config.Client(oauth2.NoContext, token)
	githubClient := github.NewClient(httpClient)
	user, resp, err := githubClient.Users.Get("")
	err = validateResponse(user, resp, err)
	if err != nil {
		h.failure.ServeHTTP(w, err, http.StatusBadRequest)
		return
	}
	h.success.ServeHTTP(w, req, user, accessToken)
}
