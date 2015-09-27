
# gologin [![Build Status](https://travis-ci.org/dghubble/gologin.png)](https://travis-ci.org/dghubble/gologin) [![GoDoc](http://godoc.org/github.com/dghubble/gologin?status.png)](http://godoc.org/github.com/dghubble/gologin)
<img align="right" src="https://storage.googleapis.com/dghubble/gologin.png">

Gologin provides composable login handlers for Github, Twitter, Digits, Bitbucket, Tumblr, OAuth1, OAuth2, and other authentication providers.

Choose the package for an auth provider. Register the `LoginHandler` and `CallbackHandler` for web logins and the `TokenHandler` for (mobile) token logins. Get the verified User/Account and access token from the `ctx`. Grant any kind of session you like.

See [examples](examples) for tutorials with apps you can run from the command line. Visit [whoam.io](https://whoam.io/) to see a live example site running on some Kubernetes clusters.

### Packages

* Github - [docs](http://godoc.org/github.com/dghubble/gologin/github)
* Bitbucket [docs](http://godoc.org/github.com/dghubble/gologin/bitbucket)
* Twitter - [docs](http://godoc.org/github.com/dghubble/gologin/twitter) &#183; [tutorial](examples/twitter)
* Digits - [docs](http://godoc.org/github.com/dghubble/gologin/digits) &#183; [tutorial](examples/digits)
* Tumblr - [docs](http://godoc.org/github.com/dghubble/gologin/tumblr)
* OAuth2 - [docs](http://godoc.org/github.com/dghubble/gologin/oauth2)
* OAuth1 - [docs](http://godoc.org/github.com/dghubble/gologin/oauth1)

### Features

* `LoginHandler` and `CallbackHandler` support web login flows
* `TokenHandler` supports (mobile) token login flows
* Get the verified User/Account and access token from the `ctx`
* Composable handlers work with any mux accepting an `http.Handler`
* Leverages popular API libraries for models when available (e.g. [go-github](https://github.com/google/go-github) for the Github User)
* Includes OAuth1 and OAuth2 handlers to make it easy to contribute auth providers.

Just as important, is what `gologin` handlers let you choose.

* Agnostic to any sesison library or token library. Login handlers do not make choices about your session system.
* Delegates control of OAuth 2 state parameters to `Stater` implementations.
* Delegates control of OAuth 1 temporary credentials (in-progress)

## Install

    go get github.com/dghubble/gologin

## Docs

Read [GoDoc](https://godoc.org/github.com/dghubble/gologin)

## Intro

Package `gologin` handlers are small, focused `ContextHandler`s which handle requests and use a `ctx` argument to pass data (e.g. tokens, users) to chained handlers.

```go
type ContextHandler interface {
    ServeHTTP(ctx context.Context, w http.ResponseWriter, req *http.Request)
}
```

For example, `oauth1` has `ContextHandler`s for getting request tokens, doing auth redirections, and receiving OAuth1 callbacks. Package `twitter` has `ContextHandler`s which extend this to obtain the Twitter `User` struct needed by most login systems to map from a token to a database user.

[ctxh](https://github.com/dghubble/ctxh) defines the ContextHandler and some convenience functions such as

```go
func NewHandler(h ContextHandler) http.Handler
```

to convert to a handler which plays well with `net/http`.

## Usage

Let's consider Twitter web login as an example.

Add the imports to your app.

```go
import (
    "github.com/dghubble/ctxh"
    "github.com/dghubble/gologin/twitter"
    "github.com/dghubble/oauth1"
    twitterEndpoints "github.com/dghubble/oauth1/twitter"
    "golang.org/x/net/context"
)
```

Configure a config for the `twitter` ContextHandlers `LoginHandler` and `CallbackHandler`.

```go
config := &oauth1.Config{
    ConsumerKey:    "TwitterConsumerKey",
    ConsumerSecret: "TwitterConsumerSecret",
    CallbackURL:    "http://localhost:8080/callback",
    Endpoint:       twitterEndpoints.AuthorizeEndpoint,
}
```

Register the `LoginHandler` and `CallbackHandler` on your `http.ServeMux`.

```go
mux := http.NewServeMux()
mux.Handle("/login", ctxh.NewHandler(twitter.LoginHandler(config, nil)))
mux.Handle("/callback", ctxh.NewHandler(twitter.CallbackHandler(config, success(), nil)))
```

The last argument is the `failure` ContextHandler which will be called in the event of a failure during the authentication flow. Passing nil means the `DefaultFailureHandler` will be used.

Define the success `ContextHandler`. The success handler is the last in the chain and will be called if earlier handlers in the authentication flow succeed. The access token and Twitter `User` can be read from the `ctx`.

```go
func success() ctxh.ContextHandler {
    fn := func(ctx context.Context, w http.ResponseWriter, req *http.Request) {
        accessToken, accessSecret, _ := oauth1Login.AccessTokenFromContext(ctx)
        twitterUser, err := twitter.UserFromContext(ctx)
        // handle errors and grant the visitor a session (cookie, token, etc.)
    }
    return ctxh.ContextHandlerFunc(fn)
}
```

See the [example](examples) apps for more details.

Check out the available auth provider packages. Each has handlers for the web authorization flow and ensures the `ctx` contains the appropriate type of user/account (e.g. a [go-github](https://github.com/google/go-github) User for Github) and the access token.

### Going Further

If you wish to define your own failure ContextHandler, you can get the reason for the error from the `ctx` using `WithError`. See [gologin](http://godoc.org/github.com/dghubble/gologin).

## Mobile

Twitter and Digits include a `TokenHandler` which can be useful for building APIs for mobile devices which use Login with Twitter or Login with Digits.

## Roadmap

* Google
* Facebook
* Soundcloud
* Improve test coverage
* Improve examples and documentation

## Contributing

If an auth provider you'd like is missing, please consider contributing it! The ContextHandlers for oauth1 and oauth2 should make it easy to add a package for OAuth1 and OAuth2 providers.

Also, `gologin` strives to use the defacto standard API libraries for User/Account models and verify endpoints. Tumblr and Bitbucket don't seem to have good ones yet so tiny internal API clients are used.

See the [Contributing Guide](https://gist.github.com/dghubble/be682c123727f70bcfe7).

## Motivations

Package `gologin` is focused on the idea that login should performed with small, composable handlers just like any other sort of middleware.

Chiefly, `gologin` addresses my frustrations with the design of [goth](https://github.com/markbates/goth) and [gomniauth](https://github.com/stretchr/gomniauth).

In my own web apps, I primarily use `http.ServeMux` as a mux and `ContextHandler` or `http.Handler` for simple handlers. `ContextHandler` is the minimalist extension of `http.Handler` to pass a `golang.org/x/net/context` Context. For more info, see the [Go Context blog post](https://blog.golang.org/context), [article](https://joeshaw.org/net-context-and-http-handler/) on handlers, and [Talk on Context Plumbing](https://vimeo.com/115309491).

## License

[MIT License](LICENSE)


