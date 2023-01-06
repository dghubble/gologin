# GitHub Login

Login with GitHub allows users to login to any web app with their GitHub account.

## Setup

Package `gologin` provides Go handlers to perform the GitHub OAuth2 Authorization flow and obtain the GitHub User struct.

```
git clone https://github.com/dghubble/gologin.git
cd gologin/examples/github
```

Obtain a GitHub application client id and secret from [developer settings](https://github.com/settings/developers). Add `http://localhost:8080/github/callback` as a valid OAuth2 Redirect URL.

## Example App

[main.go](main.go) shows an example web app that issues a client-side cookie session. Pass the GitHub client id and secret as arguments or set the `GITHUB_CLIENT_ID` and `GITHUB_CLIENT_SECRET` environment variables.

```
go run main.go -client-id=xx -client-secret=yy
2015/09/25 23:09:13 Starting Server listening on localhost:8080
```

Here's what the flow looks like.

<img src="https://storage.googleapis.com/dghubble/github-web-login.gif">

### Authorization Flow

1. Clicking the "Login with GitHub" link to the login handler directs the user to the GitHub OAuth2 Auth URL to obtain a permission grant.
2. The redirection URI (callback handler) receives the OAuth2 callback, verifies the state parameter, and obtains a Token.
3. The success `http.Handler` is called with a `Context` which contains the GitHub Token and verified GitHub User struct.
4. In this example, that User is read and used to issue a signed cookie session.

