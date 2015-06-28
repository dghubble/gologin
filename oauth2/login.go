// Package oauth2 provides a LoginHandler for OAuth2 login and callback
// requests.
package oauth2

import (
	"errors"
	"net/http"

	"github.com/dghubble/gologin"
	"golang.org/x/oauth2"
)

// TODO: CSRF protection, currently does not obey spec

// Config configures a LoginHandler.
type Config struct {
	OAuth2Config *oauth2.Config
	Success      SuccessHandler
	Failure      gologin.ErrorHandler
}

// LoginHandler handles OAuth2 login and callback requests. If authentication
// succeeds. handling is delegated to a SuccessHandler. Otherwise, an
// ErrorHandler handles responding.
type LoginHandler struct {
	mux          *http.ServeMux
	oauth2Config *oauth2.Config
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
		oauth2Config: config.OAuth2Config,
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

// RequestLoginHandler handles OAuth2 login requests by redirecting to the
// authorization URL.
func (h *LoginHandler) RequestLoginHandler() http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		authorizationURL := h.oauth2Config.AuthCodeURL("vulnerable to csrf!")
		http.Redirect(w, req, authorizationURL, http.StatusFound)
	}
	return http.HandlerFunc(fn)
}

// CallbackHandler handles OAuth2 callback requests by reading the auth code
// and state and obtaining an access token.
func (h *LoginHandler) CallbackHandler() http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		authCode, _, err := validateCallback(req)
		// TODO: verify the state! Vulnerable to CSRF
		if err != nil {
			h.failure.ServeHTTP(w, err, http.StatusBadRequest)
			return
		}
		// use the authorization code to get an access token
		token, err := h.oauth2Config.Exchange(oauth2.NoContext, authCode)
		if err != nil {
			h.failure.ServeHTTP(w, err, http.StatusBadRequest)
			return
		}
		h.success.ServeHTTP(w, req, token.AccessToken)
	}
	return http.HandlerFunc(fn)
}

func validateCallback(req *http.Request) (authCode, state string, err error) {
	// parse the raw query from the URL into req.Form
	err = req.ParseForm()
	if err != nil {
		return "", "", err
	}
	authCode = req.Form.Get("code")
	state = req.Form.Get("state")
	if authCode == "" || state == "" {
		return "", "", errors.New("callback did not receive a code or state")
	}
	return authCode, state, nil
}
