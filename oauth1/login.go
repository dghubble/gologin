// Package oauth1 provides a LoginHandler for OAuth1 login and callback
// requests.
package oauth1

import (
	"net/http"

	"github.com/dghubble/gologin"
	"github.com/dghubble/oauth1"
)

// TODO: generalize oauth1 to interface
// TODO: remove some Twitter specific aspects

// Config configures a LoginHandler.
type Config struct {
	OAuth1Config *oauth1.Config
	Success      SuccessHandler
	Failure      gologin.ErrorHandler
}

// LoginHandler handles OAuth1 login and callback requests. If authentication
// succeeds, handling is delegated to a SuccessHandler. Otherwise, an
// ErrorHandler handles responding.
type LoginHandler struct {
	mux          *http.ServeMux
	oauth1Config *oauth1.Config
	success      SuccessHandler
	failure      gologin.ErrorHandler
}

// NewLoginHandler returns a new LoginHandler.
func NewLoginHandler(config *Config) *LoginHandler {
	mux := http.NewServeMux()
	failure := config.Failure
	if failure == nil {
		failure = gologin.DefaultErrorHandler
	}
	loginHandler := &LoginHandler{
		mux:          mux,
		oauth1Config: config.OAuth1Config,
		success:      config.Success,
		failure:      failure,
	}
	mux.Handle("/login", loginHandler.RequestLoginHandler())
	mux.Handle("/callback", loginHandler.CallbackHandler())
	return loginHandler
}

func (h *LoginHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	h.mux.ServeHTTP(w, req)
}

// RequestLoginHandler handles OAuth1 login requests by obtaining a request
// token and redirecting to the authorization URL.
func (h *LoginHandler) RequestLoginHandler() http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		requestToken, _, err := h.oauth1Config.RequestToken()
		if err != nil {
			h.failure.ServeHTTP(w, err, http.StatusBadRequest)
			return
		}
		// Twitter does not require the oauth token secret be saved
		authorizationURL, err := h.oauth1Config.AuthorizationURL(requestToken)
		if err != nil {
			h.failure.ServeHTTP(w, err, http.StatusBadRequest)
			return
		}
		http.Redirect(w, req, authorizationURL.String(), http.StatusFound)
	}
	return http.HandlerFunc(fn)
}

// CallbackHandler handles OAuth1 callback requests by reading the token and
// verifier and obtaining an access token.
func (h *LoginHandler) CallbackHandler() http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		requestToken, verifier, err := oauth1.ParseAuthorizationCallback(req)
		if err != nil {
			h.failure.ServeHTTP(w, err, http.StatusBadRequest)
			return
		}
		// Twitter AccessToken endpoint does not require the auth header to be signed
		// No need to lookup the (temporary) RequestToken secret token.
		accessToken, accessSecret, err := h.oauth1Config.AccessToken(requestToken, "", verifier)
		if err != nil {
			h.failure.ServeHTTP(w, err, http.StatusBadRequest)
			return
		}
		h.success.ServeHTTP(w, req, accessToken, accessSecret)
	}
	return http.HandlerFunc(fn)
}
