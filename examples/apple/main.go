package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/dghubble/gologin/v2"
	"github.com/dghubble/gologin/v2/apple"
	"github.com/dghubble/sessions"
	"golang.org/x/oauth2"
)

var appleEndpoint = oauth2.Endpoint{
	AuthURL:  apple.AppleBaseURL + "/auth/authorize",
	TokenURL: apple.AppleBaseURL + "/auth/token",
}

const (
	sessionName    = "example-apple-app"
	sessionSecret  = "example cookie signing secret"
	sessionUserKey = "appleID"
)

// sessionStore encodes and decodes session data stored in signed cookies
var sessionStore = sessions.NewCookieStore([]byte(sessionSecret), nil)

// Config configures the main ServeMux.
type Config struct {
	AppleClientID      string
	AppleClientKeyPath string
	AppleTeamID        string
	AppleKeyID         string
	AppleClientSecret  string
}

// New returns a new ServeMux with app routes.
func New(config *Config) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", welcomeHandler)
	mux.Handle("/profile", requireLogin(http.HandlerFunc(profileHandler)))
	mux.HandleFunc("/logout", logoutHandler)

	// Register Login and Callback handlers
	oauth2Config := &oauth2.Config{
		ClientID:     config.AppleClientID,
		ClientSecret: config.AppleClientSecret,
		RedirectURL:  "http://localhost.com/apple/callback",
		Endpoint:     appleEndpoint,
		Scopes:       []string{"email"},
	}

	// state param cookies require HTTPS by default; disable for localhost development
	stateConfig := gologin.DebugOnlyCookieConfig
	mux.Handle("/apple/login", apple.StateHandler(stateConfig, apple.LoginHandler(oauth2Config, nil)))
	mux.Handle("/apple/callback", apple.StateHandler(stateConfig, apple.CallbackHandler(oauth2Config, issueSession(), nil)))
	return mux
}

// issueSession issues a cookie session after successful Apple login
func issueSession() http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		appleUser, err := apple.UserFromContext(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// 2. Implement a success handler to issue some form of session
		session := sessionStore.New(sessionName)
		session.Values[sessionUserKey] = appleUser.ID
		session.Save(w)
		http.Redirect(w, req, "/profile", http.StatusFound)
	}
	return http.HandlerFunc(fn)
}

// welcomeHandler shows a welcome message and login button.
func welcomeHandler(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		http.NotFound(w, req)
		return
	}
	if isAuthenticated(req) {
		http.Redirect(w, req, "/profile", http.StatusFound)
		return
	}
	page, _ := ioutil.ReadFile("home.html")
	fmt.Fprintf(w, string(page))
}

// profileHandler shows protected user content.
func profileHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprint(w, `<p>You are logged in!</p><form action="/logout" method="post"><input type="submit" value="Logout"></form>`)
}

// logoutHandler destroys the session on POSTs and redirects to home.
func logoutHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {
		sessionStore.Destroy(w, sessionName)
	}
	http.Redirect(w, req, "/", http.StatusFound)
}

// requireLogin redirects unauthenticated users to the login route.
func requireLogin(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		if !isAuthenticated(req) {
			http.Redirect(w, req, "/", http.StatusFound)
			return
		}
		next.ServeHTTP(w, req)
	}
	return http.HandlerFunc(fn)
}

// isAuthenticated returns true if the user has a signed session cookie.
func isAuthenticated(req *http.Request) bool {
	if _, err := sessionStore.Get(req, sessionName); err == nil {
		return true
	}
	return false
}

// main creates and starts a Server listening.
func main() {
	const address = "localhost.com:80"

	// Read credentials from environment variables if available
	config := &Config{
		AppleClientID:      os.Getenv("APPLE_CLIENT_ID"),
		AppleClientKeyPath: os.Getenv("APPLE_CLIENT_KEY_PATH"),
		AppleTeamID:        os.Getenv("APPLE_TEAM_ID"),
		AppleKeyID:         os.Getenv("APPLE_KEY_ID"),
	}

	// Allow consumer credential flags to override config fields
	clientID := flag.String("client-id", "", "Apple Client ID")
	clientKeyPath := flag.String("client-key-path", "", "Apple ES256 Private Key Path")
	teamID := flag.String("team-id", "", "Apple Team ID")
	keyID := flag.String("key-id", "", "Apple ES256 Private Key ID")
	flag.Parse()
	if *clientID != "" {
		config.AppleClientID = *clientID
	}
	if *clientKeyPath != "" {
		config.AppleClientKeyPath = *clientKeyPath
	}
	if *teamID != "" {
		config.AppleTeamID = *teamID
	}
	if *keyID != "" {
		config.AppleKeyID = *keyID
	}
	if config.AppleClientID == "" {
		log.Fatal("Missing Apple Client ID")
	}
	if config.AppleClientKeyPath == "" {
		log.Fatal("Missing Apple ES256 Private Key Path")
	}
	if config.AppleTeamID == "" {
		log.Fatal("Missing Apple Team ID")
	}
	if config.AppleKeyID == "" {
		log.Fatal("Missing Apple Key ID")
	}

	// Read key file:
	pkey, err := ioutil.ReadFile(config.AppleClientKeyPath)
	if err != nil {
		log.Fatal("Couldn't read ES256 private key file: ", err)
	}

	config.AppleClientSecret, err = apple.ClientSecret(pkey, 86400, config.AppleKeyID, config.AppleTeamID, config.AppleClientID)
	if err != nil {
		log.Fatal("Couldn't generate Apple client secret: ", err)
	}

	// Start server:
	log.Printf("Starting Server listening on %s\n", address)
	err = http.ListenAndServe(address, New(config))
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
