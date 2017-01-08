
# Github Login

Login with Github allows users to login to any web app with their Github account.

## Web

Package `gologin` provides Go handlers for Github which perform the OAuth2 Authorization flow and obtain the Github User struct.

### Getting Started

    go get github.com/dghubble/gologin/github
    cd $GOPATH/src/github.com/dghubble/gologin/examples/github
    go get .

## Example App

[main.go](main.go) shows an example web app which uses `gologin` for Github to issue a client-side cookie session. For simplicity, no data is persisted.

Get your Github application's client id and secret from [developer settings](https://github.com/settings/developers).

Compile and run `main.go` from `examples/github`. Pass the client id and secret as arguments to the executable

    go run main.go -client-id=xx -client-secret=yy
    2015/09/25 23:09:13 Starting Server listening on localhost:8080

or set the `GITHUB_CLIENT_ID` and `GITHUB_CLIENT_SECRET` environment variables.

Here's what the flow looks like.

<img src="https://storage.googleapis.com/dghubble/github-web-login.gif">

### Authorization Flow

1. Clicking the "Login with Github" link to the login handler directs the user to the Github OAuth2 Auth URL to obtain a permission grant.
2. The redirection URI (callback handler) receives the OAuth2 callback, verifies the state parameter, and obtains a Token.
3. The success `http.Handler` is called with a `Context` which contains the Github Token and verified Github User struct.
4. In this example, that User is read and used to issue a signed cookie session.

