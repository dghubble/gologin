
# gologin Changelog

Notable changes between releases.

## v2.0.0 (2016-01-10)

* Support for Go 1.7+ standard `context`
* Change `gologin` handlers to be standard `http.Handler`'s
* Drop requirement for `ctxh.NewHandler` wrapper
* Drop dependency on `github.com/dghubble/ctxh` shim

### Migration

* Update `golang.org/x/net/context` imports to `context`
* Change any `ctxh.ContextHandler` to a `http.Handler`. The `ctx` is passed via the request so the argument is no longer needed.
* Remove any `ctxh.NewHandler(...)` wrap. `gologin` handlers are now standard `http.Handler`'s, conversion is no longer required.
* Use `req.Context()` to obtain the request context within handlers.
* See updated [examples](examples)

## v1.0.1 (2016-10-31)

* Use base64.RawURLEncoding for StateHandler's state (#14)
* Fix OAuth1 failure handler's error passing (#13)
* Improve test automation. Validate with Go 1.6 and 1.7.

## v1.0.0 (2016-03-09)

* Official release using the `ContextHandler`
* Support for all OAuth1 and Oauth2 providers
* Convenience handlers for Google, Github, Facebook, Bitbucket, Twitter, Digits, and Tumblr
* Token login handlers for Twitter and Digits

## v0.1.0 (2015-10-09)

* Initial proof of concept
* Web login handlers for Google, Github, Facebook, Bitbucket, Twitter, Digits, and Tumblr
* Token login handlers for Twitter and Digits
* Support for OAuth1 and OAuth2