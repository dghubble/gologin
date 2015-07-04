// Package oauth1 provides a LoginHandler for OAuth1 login and callback
// requests.
package oauth1

import (
	"net/http"

	"github.com/dghubble/go-login"
	"github.com/dghubble/oauth1"
)

// TODO: generalize oauth1 to interface
// TODO: remove some Twitter specific aspects

// Config configures a LoginHandler.
type Config struct {
	OAuth1Config *oauth1.Config
	Success      SuccessHandler
	Failure      login.ErrorHandler
}

// LoginHandler handles OAuth1 login and callback requests. If authentication
// succeeds, handling is delegated to a SuccessHandler. Otherwise, an
// ErrorHandler handles responding.
type LoginHandler struct {
	*http.ServeMux
	oauth1Config *oauth1.Config
	success      SuccessHandler
	failure      login.ErrorHandler
}

// NewLoginHandler returns a new LoginHandler.
func NewLoginHandler(config *Config) *LoginHandler {
	mux := http.NewServeMux()
	loginMux := &LoginHandler{
		ServeMux:     mux,
		oauth1Config: config.OAuth1Config,
		success:      config.Success,
		failure:      config.Failure,
	}
	loginMux.Handle("/login", loginMux.RequestLoginHandler())
	loginMux.Handle("/callback", loginMux.CallbackHandler())
	return loginMux
}

// RequestLoginHandler handles OAuth1 login requests by obtaining a request
// token and redirecting to the authorization URL.
func (h *LoginHandler) RequestLoginHandler() http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		requestToken, err := h.oauth1Config.GetRequestToken()
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
		tokenKey, verifier, _ := h.oauth1Config.HandleAuthorizationCallback(req)
		// Twitter AccessToken endpoint does not require the auth header to be signed
		// No need to lookup the (temporary) RequestToken secret token.
		requestToken := &oauth1.RequestToken{Token: tokenKey, TokenSecret: ""}
		accessToken, err := h.oauth1Config.GetAccessToken(requestToken, verifier)
		if err != nil {
			h.failure.ServeHTTP(w, err, http.StatusBadRequest)
			return
		}
		h.success.ServeHTTP(w, req, accessToken.Token, accessToken.TokenSecret)
	}
	return http.HandlerFunc(fn)
}

// SuccessHandler is called when OAuth1 authentication succeeds.
type SuccessHandler interface {
	ServeHTTP(w http.ResponseWriter, req *http.Request, token, secret string)
}

// SuccessHandlerFunc is an adapter to allow an ordinary function to be used as
// a SuccessHandler.
type SuccessHandlerFunc func(w http.ResponseWriter, req *http.Request, token, secret string)

func (f SuccessHandlerFunc) ServeHTTP(w http.ResponseWriter, req *http.Request, token, secret string) {
	f(w, req, token, secret)
}
