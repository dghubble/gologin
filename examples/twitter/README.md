
# Twitter Login

Login with Twitter allows users to login to any web app or mobile app with their Twitter account.

## Web

Package `gologin` provides Go handlers for Twitter which perform the OAuth1 Authorization flow and obtain the Twitter User struct.

### Getting Started

    go get github.com/dghubble/gologin/twitter
    cd $GOPATH/src/github.com/dghubble/gologin/examples/twitter
    go get .

## Example App

[main.go](main.go) shows an example web app which uses `gologin` for Twitter to issue a client-side cookie session. For simplicity, no data is persisted.

Get your Twitter application's consumer key/secret from the [fabric.io](https://fabric.io) dashboard or the [old dashboard](https://apps.twitter.com/). Paste the Consumer Key and Secret in `main.go`.

**Warning** Do not add your credentials to source code for a real application.

Compile and run from the `examples/twitter` directory:

    go run main.go
    2015/09/25 23:09:13 Starting Server listening on localhost:8080

### How it works

1. Clicking the "Login with Twitter" link to the login handler redirects the user to the Twitter OAuth1 Authorization page to obtain a permission grant.
2. The callback handler receives the OAuth1 callback and obtains an access token.
3. The success `ContextHandler` is called with a `Context` which contains the Twitter access token and verified Twitter User struct.
4. In this example, that User is read and used to issue a signed cookie session.

