# gologin Changelog

Notable changes between releases.

## Latest

## v2.1.0

* Add `EnterpriseCallbackHandler` for Github Enterprise ([#33](https://github.com/dghubble/gologin/pull/33))
* Add email address to Facebook Users ([0acc88](https://github.com/dghubble/gologin/commit/0acc881e40b4926bbba0c02944ad5842700a0eab))
* Update Facebook API version to v2.9 ([0acc88](https://github.com/dghubble/gologin/commit/0acc881e40b4926bbba0c02944ad5842700a0eab))
* Fix facebook `CallbackHandler` to pass Facebook errors ([#31](https://github.com/dghubble/gologin/pull/31))
* Fix Github Users.Get call to accomodate a `go-github` [change](https://github.com/google/go-github/pull/529) ([#18](https://github.com/dghubble/gologin/pull/18))
* Remove deprecated `digits` subpackage ([#29](https://github.com/dghubble/gologin/pull/29))

## v2.0.0

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

## v1.0.1

* Use base64.RawURLEncoding for StateHandler's state (#14)
* Fix OAuth1 failure handler's error passing (#13)
* Improve test automation. Validate with Go 1.6 and 1.7.

## v1.0.0

* Official release using the `ContextHandler`
* Support for all OAuth1 and Oauth2 providers
* Convenience handlers for Google, Github, Facebook, Bitbucket, Twitter, Digits, and Tumblr
* Token login handlers for Twitter and Digits

## v0.1.0

* Initial proof of concept
* Web login handlers for Google, Github, Facebook, Bitbucket, Twitter, Digits, and Tumblr
* Token login handlers for Twitter and Digits
* Support for OAuth1 and OAuth2
