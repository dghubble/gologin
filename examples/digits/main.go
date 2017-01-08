// main is an example web app using Login with Digits (phone number).
package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/dghubble/gologin/digits"
	"github.com/dghubble/sessions"
)

const (
	sessionName    = "example-digits-app"
	sessionSecret  = "example cookie signing secret"
	sessionUserKey = "digitsID"
)

// sessionStore encodes and decodes session data stored in signed cookies
var sessionStore = sessions.NewCookieStore([]byte(sessionSecret), nil)

// Config configures the main ServeMux.
type Config struct {
	DigitsConsumerKey string
}

// New returns a new ServeMux with app routes.
func New(c *Config) *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("/", welcomeHandler(c.DigitsConsumerKey))
	mux.Handle("/profile", requireLogin(http.HandlerFunc(profileHandler)))
	mux.HandleFunc("/logout", logoutHandler)
	// 1. Register a Digits LoginHandler to receive Javascript login POST
	config := &digits.Config{
		ConsumerKey: c.DigitsConsumerKey,
	}
	mux.Handle("/login/digits", digits.LoginHandler(config, issueSession(), nil))
	return mux
}

// issueSession issues a cookie session after successful Digits login
func issueSession() http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		digitsAccount, err := digits.AccountFromContext(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// 2. Implement a success handler to issue some form of session
		session := sessionStore.New(sessionName)
		session.Values[sessionUserKey] = digitsAccount.ID
		session.Save(w)
		http.Redirect(w, req, "/profile", http.StatusFound)
	}
	return http.HandlerFunc(fn)
}

// welcomeHandler shows a welcome message and login button.
func welcomeHandler(consumerKey string) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/" {
			http.NotFound(w, req)
			return
		}
		if isAuthenticated(req) {
			http.Redirect(w, req, "/profile", http.StatusFound)
			return
		}
		// using a template purely to inject the consumer key into page js
		tpl, err := template.ParseFiles("home.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = tpl.Execute(w, map[string]string{"digits_consumer_key": consumerKey})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	return http.HandlerFunc(fn)
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
	const address = "localhost:8080"
	// read consumer key from environment variable if available
	config := &Config{
		DigitsConsumerKey: os.Getenv("DIGITS_CONSUMER_KEY"),
	}
	// allow consumer key flag to override config fields
	consumerKey := flag.String("consumer-key", "", "Digits Consumer Key")
	flag.Parse()
	if *consumerKey != "" {
		config.DigitsConsumerKey = *consumerKey
	}
	if config.DigitsConsumerKey == "" {
		log.Fatal("Missing Digits Consumer Key")
	}

	log.Printf("Starting Server listening on %s\n", address)
	err := http.ListenAndServe(address, New(config))
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
