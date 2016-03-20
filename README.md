
# gologin [![Build Status](https://travis-ci.org/dghubble/gologin.svg?branch=master)](https://travis-ci.org/dghubble/gologin) [![GoDoc](https://godoc.org/github.com/dghubble/gologin?status.png)](https://godoc.org/github.com/dghubble/gologin)
<img align="right" src="https://storage.googleapis.com/dghubble/gologin.png">

Package `gologin` provides chainable login handlers for Google, Github, Twitter, Digits, Facebook, Bitbucket, Tumblr, OAuth1, OAuth2, and other authentication providers.

Choose an auth provider package. Register the `LoginHandler` and `CallbackHandler` for web logins and the `TokenHandler` for (mobile) token logins. Get the verified User/Account and access token from the `ctx`.

See [examples](examples) for tutorials with apps you can run from the command line. Visit [whoam.io](https://whoam.io/) to see a live site running on some Kubernetes clusters.

**tldr**: Handlers which implement the steps of standard authentication flows to provide access tokens and associated User/Account structs.

### Packages

* Google - [docs](http://godoc.org/github.com/dghubble/gologin/google) &#183; [tutorial](examples/google)
* Github - [docs](http://godoc.org/github.com/dghubble/gologin/github) &#183; [tutorial](examples/github)
* Facebook - [docs](http://godoc.org/github.com/dghubble/gologin/facebook) &#183; [tutorial](examples/facebook)
* Twitter - [docs](http://godoc.org/github.com/dghubble/gologin/twitter) &#183; [tutorial](examples/twitter)
* Digits - [docs](http://godoc.org/github.com/dghubble/gologin/digits) &#183; [tutorial](examples/digits)
* Bitbucket [docs](http://godoc.org/github.com/dghubble/gologin/bitbucket)
* Tumblr - [docs](http://godoc.org/github.com/dghubble/gologin/tumblr)
* OAuth2 - [docs](http://godoc.org/github.com/dghubble/gologin/oauth2)
* OAuth1 - [docs](http://godoc.org/github.com/dghubble/gologin/oauth1)

## Features

* `LoginHandler` and `CallbackHandler` support web login flows
* `TokenHandler` supports native mobile token login flows
* Get the verified User/Account and access token from the `ctx`
* OAuth 2 State Parameter CSRF protection

## Flexibility

* Handlers work with any mux accepting an `http.Handler`
* Does not attempt to be your session system or token system.
* Configurable OAuth 2 state parameter handling
* Configurable OAuth 1 request secret handling

## Install

    go get github.com/dghubble/gologin

## Docs

Read [GoDoc](https://godoc.org/github.com/dghubble/gologin)

## Goals

Create small, chainable handlers to correctly implement the steps of common authentication flows. Handle provider-specific validation requirements.

## Concept

Package `gologin` provides `ContextHandler`'s which can be chained together to implement authorization flows and pass data (e.g. tokens, users) in a `ctx` argument.

```go
type ContextHandler interface {
    ServeHTTP(ctx context.Context, w http.ResponseWriter, req *http.Request)
}
```

For example, `oauth2` has handlers which generate a state parameter, redirect users to an AuthURL, or validate a redirectURL callback to exchange for a Token.

`gologin` handlers generally take `success` and `failure` ContextHandlers to be called next if an authentication step succeeds or fails. They populate the `ctx` with values needed for the next step. If the flow succeeds, the last success ContextHandler `ctx` should include the access token and (optional) associated User/Account.

[ctxh](https://github.com/dghubble/ctxh) defines a ContextHandler and some convenience functions to convert to a handler which plays well with `net/http`.

```go
func NewHandler(h ContextHandler) http.Handler
```

## Usage

Choose an auth provider package such as `github` or `twitter`. These packages chain together lower level `oauth1` and `oauth2` ContextHandlers and fetch the Github or Twitter `User` before calling your success ContextHandler.

Let's walk through Github and Twitter web login examples.

### Github OAuth2

Register the `LoginHandler` and `CallbackHandler` on your `http.ServeMux`.

```go
config := &oauth2.Config{
    ClientID:     "GithubClientID",
    ClientSecret: "GithubClientSecret",
    RedirectURL:  "http://localhost:8080/callback",
    Endpoint:     githubOAuth2.Endpoint,
}
mux := http.NewServeMux()
stateConfig := gologin.DebugOnlyCookieConfig
mux.Handle("/login", ctxh.NewHandler(github.StateHandler(stateConfig, github.LoginHandler(config, nil))))
mux.Handle("/callback", ctxh.NewHandler(github.StateHandler(stateConfig, github.CallbackHandler(config, issueSession(), nil))))
```

The `StateHandler` checks for an OAuth2 state parameter cookie, generates a non-guessable state as a short-lived cookie if missing, and passes the state value in the ctx. The `CookieConfig` allows the cookie name or expiration (default 60 seconds) to be configured. In production, use a config like `gologin.DefaultCookieConfig` which sets *Secure* true to require cookies be sent over HTTPS. If you wish to persist state parameters a different way, you may chain your own `ContextHandler`. ([info](#state-parameters))

The `github` `LoginHandler` reads the state from the ctx and redirects to the AuthURL (at github.com) to prompt the user to grant access. Passing nil for the `failure` ContextHandler just means the `DefaultFailureHandler` should be used, which reports errors. ([info](#failure-handlers))

The `github` `CallbackHandler` receives an auth code and state OAuth2 redirection, validates the state against the state in the ctx, and exchanges the auth code for an OAuth2 Token. The `github` CallbackHandler wraps the lower level `oauth2` `CallbackHandler` to further use the Token to obtain the Github `User` before calling through to the success or failure handlers.

<img src="https://storage.googleapis.com/dghubble/gologin-github.png">

Next, write the success `ContextHandler` to do something with the Token and Github User added to the `ctx`.

```go
func issueSession() ctxh.ContextHandler {
    fn := func(ctx context.Context, w http.ResponseWriter, req *http.Request) {
        token, _ := oauth2Login.TokenFromContext(ctx)
        githubUser, err := github.UserFromContext(ctx)
        // handle errors and grant the visitor a session (cookie, token, etc.)
    }
    return ctxh.ContextHandlerFunc(fn)
}
```

See the [Github tutorial](examples/github) for a web app you can run from the command line.

### Twitter OAuth1

Register the `LoginHandler` and `CallbackHandler` on your `http.ServeMux`.

```go
config := &oauth1.Config{
    ConsumerKey:    "TwitterConsumerKey",
    ConsumerSecret: "TwitterConsumerSecret",
    CallbackURL:    "http://localhost:8080/callback",
    Endpoint:       twitterOAuth1.AuthorizeEndpoint,
}
mux := http.NewServeMux()
mux.Handle("/login", ctxh.NewHandler(twitter.LoginHandler(config, nil)))
mux.Handle("/callback", ctxh.NewHandler(twitter.CallbackHandler(config, issueSession(), nil)))
```

The `twitter` `LoginHandler` obtains a request token and secret, adds them to the ctx, and redirects to the AuthorizeURL to prompt the user to grant access. Passing nil for the `failure` ContextHandler just means the `DefaultFailureHandler` should be used, which reports errors. ([info](#failure-handlers))

The `twitter` `CallbackHandler` receives an OAuth1 token and verifier, reads the request secret from the ctx, and obtains an OAuth1 access token and secret. The `twitter` CallbackHandler wraps the lower level `oauth1` CallbackHandler to further use the access token/secret to obtain the Twitter `User` before calling through to the success or failure handlers.

<img src="https://storage.googleapis.com/dghubble/gologin-twitter.png">

Next, write the success `ContextHandler` to do something with the access token/secret and Twitter User added to the `ctx`.

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

*Note: Some OAuth1 providers (not Twitter), require the request secret be persisted until the callback is received. For this reason, the lower level `oauth1` package splits LoginHandler functionality into a `LoginHandler` and `AuthRedirectHandler`. Provider packages, like `tumblr`, chain these together for you, but the lower level handlers are there if needed.

See the [Twitter tutorial](examples/twitter) for a web app you can run from the command line.

### State Parameters

OAuth2 `StateHandler` implements OAuth 2 [RFC 6749](https://tools.ietf.org/html/rfc6749) 10.12 CSRF Protection using non-guessable values in short-lived HTTPS-only cookies to provide reasonable assurance the user in the login phase and callback phase are the same. If you wish to implement this differently, write a `ContextHandler` which sets a *state* in the ctx, which is expected by LoginHandler and CallbackHandler.

You may use `oauth2.WithState(context.Context, state string)` for this. [docs](https://godoc.org/github.com/dghubble/gologin/oauth2#WithState)

### Failure Handlers

If you wish to define your own failure `ContextHandler`, you can get the error from the `ctx` using `gologin.ErrorFromContext(ctx)`.

### Production Requirements

* Use HTTPS.
* Never put consumer/client secrets in source control.
* Ensure the CookieConfig requires state or temp credential cookies be sent over HTTPS-only.

### Going Further

Check out the available auth provider packages. Each has handlers for the web authorization flow and ensures the `ctx` contains the appropriate type of user/account and the access token.

## Mobile

Twitter and Digits include a `TokenHandler` which can be useful for building APIs for mobile devices which use Login with Twitter or Login with Digits.

## Motivations

Package `gologin` implements authorization flow steps with chained handlers.

* Authentication should be performed with chainable handlers to allow customization, swapping, or adding additional steps easily.
* Authentication should be orthogonal to the session system. Let users choose their session/token library.
* OAuth2 State CSRF should be included out of the box, but easy to customize.
* Packages should import only what is required. OAuth1 and OAuth2 packages are separate.
* ContextHandlers are flexible and useful for more than just data passing.

Projects [goth](https://github.com/markbates/goth) and [gomniauth](https://github.com/stretchr/gomniauth) aim to provide a similar login solution with a different design. Check them out if you decide you don't like the ideas in `gologin`.

## Roadmap

* Improve test coverage
* OTP Handlers
* Per-Provider User types (current) vs one combined gologin User type?

## Contributing

Contributions are welcome. Improving documentation and examples is a good way to start. New auth providers can be implemented by composing the `oauth1` or `oauth2` ContextHandlers.

Also, `gologin` aims to use the defacto standard API libraries for User/Account models and verify endpoints. Tumblr and Bitbucket don't seem to have good ones yet. Tiny internal API clients are used.

See the [Contributing Guide](https://gist.github.com/dghubble/be682c123727f70bcfe7).

### Why Contexts?

I originally thought `gologin` should use only `http.Handler` handlers and `handler(http.Handler) http.Handler` chaining. Passing request data becomes messy with many custom handler types. Global request to context mappings are similarly gross.

A while ago, some great materials like this Go [blog post](https://blog.golang.org/context), Sameer Ajmani's [Gotham talk](https://vimeo.com/115309491), and Joe Shaw's [article](https://joeshaw.org/net-context-and-http-handler/) helped convince me that 

```go
type ContextHandler interface {
    ServeHTTP(ctx context.Context, w http.ResponseWriter, req *http.Request)
}
```

is an excellent choice for more advanced handlers. These days I use it a lot.

## License

[MIT License](LICENSE)


