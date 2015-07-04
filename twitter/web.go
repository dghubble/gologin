package twitter

import (
	"net/http"

	"github.com/dghubble/go-login"
	oauth1Login "github.com/dghubble/go-login/oauth1"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

// LoginHandlerConfig configures a LoginHandler.
type LoginHandlerConfig struct {
	OAuth1Config *oauth1.Config
	Success      SuccessHandler
	Failure      login.ErrorHandler
}

// LoginHandler handles Twitter OAuth1 login and callback requests. If
// authentication succeeds, handling is delegated to a SuccessHandler which
// is provided with the Twitter user, token, and tokenSecret. Otherwise, an
// ErrorHandler handles responding.
type LoginHandler struct {
	oauth1Config *oauth1.Config
	success      SuccessHandler
	failure      login.ErrorHandler
}

// NewLoginHandler returns a new LoginHandler.
func NewLoginHandler(config *LoginHandlerConfig) http.Handler {
	handler := &LoginHandler{
		oauth1Config: config.OAuth1Config,
		success:      config.Success,
		failure:      config.Failure,
	}
	return oauth1Login.NewLoginHandler(&oauth1Login.Config{
		OAuth1Config: config.OAuth1Config,
		Success:      oauth1Login.SuccessHandlerFunc(handler.successWithUser),
		Failure:      config.Failure,
	})
}

// successWithUser is an oauth1 success handler which wraps the twitter success
// and failure handlers. It verifies token credentials to obtain the Twitter
// User object for inclusion in data passed on success.
func (h *LoginHandler) successWithUser(w http.ResponseWriter, req *http.Request, token, tokenSecret string) {
	httpClient := h.oauth1Config.GetClient(token, tokenSecret)
	twitterClient := twitter.NewClient(httpClient)
	accountVerifyParams := &twitter.AccountVerifyParams{
		IncludeEntities: twitter.Bool(false),
		SkipStatus:      twitter.Bool(true),
		IncludeEmail:    twitter.Bool(false),
	}
	user, resp, err := twitterClient.Accounts.VerifyCredentials(accountVerifyParams)
	err = validateResponse(user, resp, err)
	if err != nil {
		h.failure.ServeHTTP(w, err, http.StatusBadRequest)
		return
	}
	h.success.ServeHTTP(w, req, user, token, tokenSecret)
}
