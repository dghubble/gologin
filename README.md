
# gologin [![Build Status](https://travis-ci.org/dghubble/gologin.png)](https://travis-ci.org/dghubble/gologin) [![GoDoc](http://godoc.org/github.com/dghubble/gologin?status.png)](http://godoc.org/github.com/dghubble/gologin)
<img align="right" src="https://storage.googleapis.com/dghubble/gologin.png">

Package `gologin` provides composable login handlers for Google, Github, Twitter, Digits, Facebook, Bitbucket, Tumblr, OAuth1, OAuth2, and other authentication providers.

Choose an auth provider package. Register the `LoginHandler` and `CallbackHandler` for web logins and the `TokenHandler` for (mobile) token logins. Get the verified User/Account and access token from the `ctx`.

See [examples](examples) for tutorials with apps you can run from the command line. Visit [whoam.io](https://whoam.io/) to see a live site running on some Kubernetes clusters.

tldr: Chained ContextHandlers which implement the steps of auth flows to provide access tokens and (optional) associated User/Account structs.

### Packages

* Google - [docs](http://godoc.org/github.com/dghubble/gologin/google)
* Github - [docs](http://godoc.org/github.com/dghubble/gologin/github) &#183; [tutorial](examples/github)
* Facebook - [docs](http://godoc.org/github.com/dghubble/gologin/facebook)
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
* Uses popular API libraries for models when available (e.g. [go-github](https://github.com/google/go-github) for the Github User)
* OAuth 2 State Parameter CSRF protection

## Flexibility

* Handlers work with any mux accepting an `http.Handler`
* Does not attempt to be your session system or token system.
* Configurable OAuth 2 state parameter handling (in-progress)
* Configurable OAuth 1 request secret handling (in-progress)

## Install

    go get github.com/dghubble/gologin

## Docs

Read [GoDoc](https://godoc.org/github.com/dghubble/gologin)

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

Choose an auth provider package such as `github` or `twitter`. These packages chain together the lower level `oauth1` and `oauth2` ContextHandlers and fetch the Github or Twitter `User` before calling your success ContextHandler.

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
mux.Handle("/login", ctxh.NewHandler(github.StateHandler(github.LoginHandler(config, nil))))
mux.Handle("/callback", ctxh.NewHandler(github.StateHandler(github.CallbackHandler(config, issueSession(), nil))))
```

Passing nil for the `failure` ContextHandler just means the `DefaultFailureHandler` should be used.

Next, write the success `ContextHandler` to do something with the access token and Github User added to the `ctx` (e.g. issue a cookie session).

```go
func issueSession() ctxh.ContextHandler {
    fn := func(ctx context.Context, w http.ResponseWriter, req *http.Request) {
        token, _ := oauth2Login.AccessTokenFromContext(ctx)
        githubUser, err := github.UserFromContext(ctx)
        // handle errors and grant the visitor a session (cookie, token, etc.)
    }
    return ctxh.ContextHandlerFunc(fn)
}
```

See the [Github tutorial](examples/github) for a web app you can run from the command line.

#### In Depth

If you're curious how this works, `github` ContextHandlers chain together the right sequence of `oauth2` (green) ContextHandlers.

<img src="https://storage.googleapis.com/dghubble/gologin-github.png">

The `StateHandler` grants a temproary cookie with a random state value. The `LoginHandler` uses the state value and redirects to the AuthURL to ask the user to grant access. Later, an OAuth2 redirect is sent to the `CallbackHandler`. It validates the temporary state value against the OAuth2 state parameter and exchanges the auth code for an access Token. Github fetches the User and adds it to the `ctx`. Then your success handler is called.

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

Passing nil for the `failure` ContextHandler just means the `DefaultFailureHandler` should be used.

Next, write the success `ContextHandler` to do something with the access token and Twitter User added to the `ctx` (e.g. issue a cookie session).

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

See the [Twitter tutorial](examples/twitter) for a web app you can run from the command line.

#### In Depth

If you're curious how this works, `twitter` ContextHandlers chain together the right sequence of `oauth1` (purple) ContextHandlers.

<img src="https://storage.googleapis.com/dghubble/gologin-twitter.png">

The `LoginHandler` obtains a request token and secret and adds them to the `ctx`.* The `AuthRedirectHandler` redirects to the AuthorizeURL to ask the user to grant access. Later, an OAuth1 callback is sent to the `CallbackHandler` which gets the request token secret from the `ctx`, validates parameters, and gets an access token and secret. Twitter fetches the User and adds it to the `ctx`. Then your success handler is called.

*Note, if the OAuth1 provider requires the request token secret for later steps (e.g. Tumblr) it can be intercepted and persisted. Package `tumblr` does this already.

### Going Further

Check out the available auth provider packages. Each has handlers for the web authorization flow and ensures the `ctx` contains the appropriate type of user/account and the access token.

If you wish to define your own failure `ContextHandler`, you can get the error from the `ctx` using `ErrorFromContext(ctx)`.

## Production

* Always use HTTPS
* Keep your OAuth Consumer/Client secret out of source control

## Mobile

Twitter and Digits include a `TokenHandler` which can be useful for building APIs for mobile devices which use Login with Twitter or Login with Digits.

## Roadmap

* Improve examples and documentation
* Improve test coverage
* Facebook
* Soundcloud

## Contributing

Please consider contributing! Improving documentation and examples is a good way to start. New auth providers can be implemented by composing the `oauth1` or `oauth2` ContextHandlers.

Also, `gologin` aims to use the defacto standard API libraries for User/Account models and verify endpoints. Tumblr and Bitbucket don't seem to have good ones yet. Tiny internal API clients are used.

See the [Contributing Guide](https://gist.github.com/dghubble/be682c123727f70bcfe7).

## Motivations

Package `gologin` is focused on the idea that login should performed with small, chainable handlers just like any other sort of middleware. It addresses my own frustrations with [goth](https://github.com/markbates/goth) and [gomniauth](https://github.com/stretchr/gomniauth).

* Authentication should be performed with chainable handlers. Its not special.
* Session systems are orthogonal to authentication. Let users choose their session/token library.
* Make it difficult to mess up OAuth 2 CSRF protection, but easy to customize.
* Handlers provide flexibility. Swap OAuth2 StateHandler (cookie-based) for something else if you like.
* Use quality existing API libraries and their models, where possible.
* Import only what is needed for the desired authentication providers.
* ContextHandler's are flippin awesome (see below).

### But Why Contexts?

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


