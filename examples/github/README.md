# Github Login

Login with Github allows users to login to any web app with their Github account.

## Setup

Package `gologin` provides Go handlers to perform the Github OAuth2 Authorization flow and obtain the Github User struct.

```
git clone https://github.com/dghubble/gologin.git
cd gologin/examples/github
```

Obtain a Github application client id and secret from [developer settings](https://github.com/settings/developers).

## Example App

[main.go](main.go) shows an example web app that issues a client-side cookie session. Pass the Github client id and secret as arguments or set the `GITHUB_CLIENT_ID` and `GITHUB_CLIENT_SECRET` environment variables.

```
go run main.go -client-id=xx -client-secret=yy
2015/09/25 23:09:13 Starting Server listening on localhost:8080
```

Here's what the flow looks like.

<img src="https://storage.googleapis.com/dghubble/github-web-login.gif">

### Authorization Flow

1. Clicking the "Login with Github" link to the login handler directs the user to the Github OAuth2 Auth URL to obtain a permission grant.
2. The redirection URI (callback handler) receives the OAuth2 callback, verifies the state parameter, and obtains a Token.
3. The success `http.Handler` is called with a `Context` which contains the Github Token and verified Github User struct.
4. In this example, that User is read and used to issue a signed cookie session.

