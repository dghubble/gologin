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
	mux.Handle("/login", RequestLoginHandler(config.OAuth1Config, config.Failure))
	mux.Handle("/callback", CallbackHandler(config.OAuth1Config, config.Success, config.Failure))
	return loginHandler
}

func (h *LoginHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	h.mux.ServeHTTP(w, req)
}

// RequestLoginHandler handles OAuth1 login requests by obtaining a request
// token and redirecting to the authorization URL.
func RequestLoginHandler(config *oauth1.Config, failure gologin.ErrorHandler) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		requestToken, _, err := config.RequestToken()
		if err != nil {
			failure.ServeHTTP(w, err, http.StatusBadRequest)
			return
		}
		// Twitter does not require the oauth token secret be saved
		authorizationURL, err := config.AuthorizationURL(requestToken)
		if err != nil {
			failure.ServeHTTP(w, err, http.StatusBadRequest)
			return
		}
		http.Redirect(w, req, authorizationURL.String(), http.StatusFound)
	}
	return http.HandlerFunc(fn)
}

// CallbackHandler handles OAuth1 callback requests by reading the token and
// verifier and obtaining an access token.
func CallbackHandler(config *oauth1.Config, success SuccessHandler, failure gologin.ErrorHandler) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		requestToken, verifier, err := oauth1.ParseAuthorizationCallback(req)
		if err != nil {
			failure.ServeHTTP(w, err, http.StatusBadRequest)
			return
		}
		// Twitter AccessToken endpoint does not require the auth header to be signed
		// No need to lookup the (temporary) RequestToken secret token.
		accessToken, accessSecret, err := config.AccessToken(requestToken, "", verifier)
		if err != nil {
			failure.ServeHTTP(w, err, http.StatusBadRequest)
			return
		}
		success.ServeHTTP(w, req, accessToken, accessSecret)
	}
	return http.HandlerFunc(fn)
}
