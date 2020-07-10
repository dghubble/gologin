# Twitter Login

Login with Twitter allows users to login to any web app with their Twitter account.

## Setup

Package `gologin` provides Go handlers to perform the Twitter OAuth1 Authorization flow and obtain the Twitter User struct.

```
git clone https://github.com/dghubble/gologin.git
cd gologin/examples/github
```

Obtain a Twitter application consumer key/secret from the [developer portal](https://developer.twitter.com).

## Example App

[main.go](main.go) shows an example web app that issues a client-side cookie session. Pass the consumer key/secret as arguments or set the `TWITTER_CONSUMER_KEY` and `TWITTER_CONSUMER_SECRET` environment variables.

```
go run main.go -consumer-key=xx -consumer-secret=yy
2015/09/25 23:09:13 Starting Server listening on localhost:8080
```

Here's what the flow looks like.

<img src="https://storage.googleapis.com/dghubble/twitter-web-login.gif">

### Authorization Flow

1. Clicking the "Login with Twitter" link to the login handler redirects the user to the Twitter OAuth1 Authorization page to obtain a permission grant.
2. The callback handler receives the OAuth1 callback and obtains an access token.
3. The success `http.Handler` is called with a `Context` which contains the Twitter access token and verified Twitter User struct.
4. In this example, that User is read and used to issue a signed cookie session.

