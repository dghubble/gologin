
# gologin [![Build Status](https://travis-ci.org/dghubble/gologin.png)](https://travis-ci.org/dghubble/gologin) [![GoDoc](http://godoc.org/github.com/dghubble/gologin?status.png)](http://godoc.org/github.com/dghubble/gologin)
<img align="right" src="https://storage.googleapis.com/dghubble/gologin.png">

Package `gologin` provides composable login handlers for Google, Github, Twitter, Digits, Bitbucket, Tumblr, OAuth1, OAuth2, and other authentication providers.

Choose an auth provider package. Register the `LoginHandler` and `CallbackHandler` for web logins and the `TokenHandler` for (mobile) token logins. Get the verified User/Account and access token from the `ctx`.

See [examples](examples) for tutorials with apps you can run from the command line. Visit [whoam.io](https://whoam.io/) to see a live site running on some Kubernetes clusters.

### Packages

* Google - [docs](http://godoc.org/github.com/dghubble/gologin/google)
* Github - [docs](http://godoc.org/github.com/dghubble/gologin/github) &#183; [tutorial](examples/github)
* Twitter - [docs](http://godoc.org/github.com/dghubble/gologin/twitter) &#183; [tutorial](examples/twitter)
* Digits - [docs](http://godoc.org/github.com/dghubble/gologin/digits) &#183; [tutorial](examples/digits)
* Bitbucket [docs](http://godoc.org/github.com/dghubble/gologin/bitbucket)
* Tumblr - [docs](http://godoc.org/github.com/dghubble/gologin/tumblr)
* OAuth2 - [docs](http://godoc.org/github.com/dghubble/gologin/oauth2)
* OAuth1 - [docs](http://godoc.org/github.com/dghubble/gologin/oauth1)

### Features

* `LoginHandler` and `CallbackHandler` support web login flows
* `TokenHandler` supports native mobile token login flows
* Get the verified User/Account and access token from the `ctx`
* Uses popular API libraries for models when available (e.g. [go-github](https://github.com/google/go-github) for the Github User)
* OAuth 2 State Parameter CSRF protection

## Flexibility

* Agnostic to any sesison library or token library. Login handlers should not make choices about your session system.
* Handlers work with any mux accepting an `http.Handler`
* Configurable OAuth 2 state parameter handling (in-progress)
* Configurable OAuth 1 request secret handling (in-progress)

## Install

    go get github.com/dghubble/gologin

## Docs

Read [GoDoc](https://godoc.org/github.com/dghubble/gologin)

## Intro

Package `gologin` handlers are `ContextHandler`s which pass data (e.g. tokens, users) via a `ctx` argument and are easy to chain.

```go
type ContextHandler interface {
    ServeHTTP(ctx context.Context, w http.ResponseWriter, req *http.Request)
}
```

For example, `oauth1` has `ContextHandler`s for getting request tokens, doing auth redirections, and receiving OAuth1 callbacks. Package `twitter`'s `ContextHandler`'s chain these together and add the Twitter `User` struct to the `ctx`.

[ctxh](https://github.com/dghubble/ctxh) defines a ContextHandler and some convenience functions to convert to a handler which plays well with `net/http`.

```go
func NewHandler(h ContextHandler) http.Handler
```

## Usage

To use `gologin`, register a `LoginHandler` and `CallbackHandler` on your `http.ServeMux` and provide a success `ContextHandler` to do something with the verified User/Account and access token added to the `ctx`.

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

Check out the available auth provider packages. Each has handlers for the web authorization flow and ensures the `ctx` contains the appropriate type of user/account and the access token.

### Going Further

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

Please consider contributing additional auth providers, typically by composing the `oauth1` or `oauth2` `ContextHandlers`.

Also, `gologin` aims to use the defacto standard API libraries for User/Account models and verify endpoints. Tumblr and Bitbucket don't seem to have good ones yet. Tiny internal API clients are used.

See the [Contributing Guide](https://gist.github.com/dghubble/be682c123727f70bcfe7).

## Motivations

Package `gologin` is focused on the idea that login should performed with small, composable handlers just like any other sort of middleware. It addresses frustrations with the design of [goth](https://github.com/markbates/goth) and [gomniauth](https://github.com/stretchr/gomniauth).

* Authentication should be performed with chain-able handlers. Its not special.
* Session systems are orthogonal to authentication. Let users choose their session/token library.
* Make it difficult to mess up OAuth 2 CSRF protection, but easy to customize.
* Handlers provide flexibility. For example, if you don't like the OAuth2 StateHandler (cookie-based), easily write another and compose it.
* Use quality existing API libraries and their models, where possible.
* Import only what is needed for the desired authentication providers.
* ContextHandler's are flippin awesome (see below).

### But Why Contexts?

Like you perhaps, I originally wanted `gologin` to use only `http.Handler` handlers and `handler(http.Handler) http.Handler` chaining. As much as I favor using the standard library, passing data becomes messy using this design. Global request to context mappings are similarly gross.

A while ago, some great materials like the [Go Context blog post](https://blog.golang.org/context), Sameer Ajmani's [talk](https://vimeo.com/115309491), and Joe Shaw's [article](https://joeshaw.org/net-context-and-http-handler/) helped convince me that 

```go
type ContextHandler interface {
    ServeHTTP(ctx context.Context, w http.ResponseWriter, req *http.Request)
}
```

is an excellent choice for more advanced handlers. These days I use it a lot.

## License

[MIT License](LICENSE)


