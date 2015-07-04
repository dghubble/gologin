/*
Package login provides a WebHandler for implementing Digits phone number login.

Package login provides a Digits WebHandler which receives POSTed OAuth Echo
headers, validates them, and fetches the Digits Account. Handling is delegated
to the SuccessHandler or ErrorHandler to issue a session.

Web Login

Get started with the example app https://github.com/dghubble/go-digits/tree/master/examples/login.
Paste in your Digits consumer key and run it locally to see phone number login
in action.

Alternately, add Login with Digits to your existing web app:

1. Follow the Digits for Web instructions to add a "Use My Phone Number"
button and Digits JS snippet to your login page. https://dev.twitter.com/twitter-kit/web/digits

2. Add the go-digits imports

  import (
      "github.com/dghubble/go-digits/digits"
      "github.com/dghubble/go-digits/login"
  )

3. Register a WebHandler to receive POST's from your login page.

  handlerConfig := login.Config{
      ConsumerKey: "YOUR_DIGITS_CONSUMER_KEY",
      Success: login.SuccessHandlerFunc(issueWebSession),
      Failure: login.DefaultErrorHandler,
  }
  http.Handle("/digits_login", login.NewWebHandler(handlerConfig))

4. Receive the validated Digits.Account in a SuccessHandler. Issue a session
your backend supports.

  func issueWebSession(w http.ResponseWriter, r *http.Request, account *digits.Account) {
      session := sessionStore.New(sessionName)
      session.Values["digitsID"] = account.ID
      session.Values["phoneNumer"] = account.PhoneNumber
      session.Save(w)
      http.Redirect(w, r, "/profile", http.StatusFound)
  }

*/
package login
