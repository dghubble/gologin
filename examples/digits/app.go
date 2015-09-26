// main is an example web app using Login with Digits (phone number).
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/dghubble/ctxh"
	"github.com/dghubble/gologin/digits"
	"github.com/dghubble/sessions"
	"golang.org/x/net/context"
)

const (
	digitsConsumerKey = "YOUR_DIGITS_CONSUMER_KEY"
	sessionName       = "loginapp-session"
	sessionSecret     = "example cookie signing secret"
	sessionUserKey    = "digitsID"
)

// sessionStore encodes and decodes session data stored in signed cookies
var sessionStore = sessions.NewCookieStore([]byte(sessionSecret), nil)

// New returns a new ServeMux with app routes.
func New() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", homeHandler)
	mux.Handle("/profile", requireLogin(http.HandlerFunc(profileHandler)))
	mux.HandleFunc("/logout", logoutHandler)
	// 1. Register a Digits LoginHandler to receive Javascript login POST
	config := &digits.Config{
		ConsumerKey: digitsConsumerKey,
	}
	mux.Handle("/login/digits", ctxh.NewHandler(digits.LoginHandler(config, issueSession(), nil)))
	return mux
}

// issueWebSession issues a cookie session after successful Digits login
func issueSession() ctxh.ContextHandler {
	fn := func(ctx context.Context, w http.ResponseWriter, req *http.Request) {
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
	return ctxh.ContextHandlerFunc(fn)
}

// homeHandler shows a login page or a user profile page if authenticated.
func homeHandler(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		http.NotFound(w, req)
		return
	}
	if isAuthenticated(req) {
		http.Redirect(w, req, "/profile", http.StatusFound)
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
	const address = "localhost:8080"
	log.Printf("Starting Server listening on %s\n", address)
	err := http.ListenAndServe(address, New())
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
