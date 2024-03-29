# Google Login

Login with Google allows users to login to any web app with their Google account.

## Setup

Package `gologin` provides Go handlers to perform the Google OAuth2 Authorization flow and for obtaining the Google [Userinfo struct](https://godoc.org/google.golang.org/api/oauth2/v2#Userinfo).

```
git clone https://github.com/dghubble/gologin.git
cd gologin/examples/google
```

Obtain a Google OAuth2 application client id and secret from [Google Developer Console](https://console.cloud.google.com). Navigate to APIs & Services, then Credentials. Add `http://localhost:8080/google/callback` as a valid OAuth2 Redirect URL.

<img src="https://storage.googleapis.com/dghubble/google-valid-callback.png">

## Example App

[main.go](main.go) shows an example web app that issues a client-side session cookie. Pass the client id and secret as arguments or set the `GOOGLE_CLIENT_ID` and `GOOGLE_CLIENT_SECRET` environment variables.

```
go run main.go -client-id=xx -client-secret=yy
2015/09/25 23:09:13 Starting Server listening on localhost:8080
```

Here's what the flow looks like.

<img src="https://storage.googleapis.com/dghubble/google-web-login.gif">

### Authorization Flow

1. The "Login with Google" link to the login handler directs the user to the Google OAuth2 Auth URL to obtain a permission grant.
2. The redirection URI (callback handler) receives the OAuth2 callback, verifies the state parameter, and obtains a Token.
3. The success `http.Handler` is called with a `Context` which contains the Google Token and verified Google Userinfoplus struct.
4. In this example, that User is read and used to issue a signed cookie session.

