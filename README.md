
# gologin [![Build Status](https://travis-ci.org/dghubble/gologin.png)](https://travis-ci.org/dghubble/gologin) [![GoDoc](http://godoc.org/github.com/dghubble/gologin?status.png)](http://godoc.org/github.com/dghubble/gologin)

gologin provides boilerplate `http.Handler`s (3-legged OAuth, token login) for popular 3rd party login providers (3-legged OAuth, mobile token login).

## Status

Alpha

### Features

* Register a `LoginHandler` to handle web login and OAuth callbacks.
* Register a `TokenHandler` to handle mobile token logins.
* Works with any session library, web token library, and context library.

Login handlers for:

* Twitter
* Digits (Phone SMS)
* Github
* OAuth1
* OAuth2

Token handlers for:

* Twitter
* Digits

## Install

    go get github.com/dghubble/gologin

## Docs

Read [GoDoc](https://godoc.org/github.com/dghubble/gologin)

## Roadmap

* Improve test coverage and argument checks.
* Abstract the dependency on an OAuth1 implementation.
* Google
* Improve examples and documentation
* Soundcloud
* Tumblr
* Dropbox
* Bitbucket

## License

[MIT License](LICENSE)


