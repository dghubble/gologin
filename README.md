
# gologin [![Build Status](https://travis-ci.org/dghubble/gologin.png)](https://travis-ci.org/dghubble/gologin) [![GoDoc](http://godoc.org/github.com/dghubble/gologin?status.png)](http://godoc.org/github.com/dghubble/gologin)
<img align="right" src="https://storage.googleapis.com/dghubble/gologin.png">

Gologin provides composable login handlers for Github, Twitter, Digits, Tumblr and any other OAuth1 or OAuth2 based authentication providers.

Choose a provider package. Use the `LoginHandler` and `CallbackHandler` for web logins and the `TokenHandler` for (mobile) token logins.

See the [examples](examples) for tutorials and example apps you can run from the command line.

### Packages

* Github - [docs](http://godoc.org/github.com/dghubble/gologin/github)
* Twitter - [docs](http://godoc.org/github.com/dghubble/gologin/twitter) &#183; [tutorial](examples/twitter)
* Digits - [docs](http://godoc.org/github.com/dghubble/gologin/digits) &#183; [tutorial](examples/digits)
* Tumblr - [docs](http://godoc.org/github.com/dghubble/gologin/tumblr)
* OAuth2 - [docs](http://godoc.org/github.com/dghubble/gologin/oauth2)
* OAuth1 - [docs](http://godoc.org/github.com/dghubble/gologin/oauth1)

### Features

* `LoginHandler` and `CallbackHandler` support web login flows
* `TokenHandler` supports (mobile) token login flows
* Agnostic to any session library or token library

## Install

    go get github.com/dghubble/gologin

## Docs

Read [GoDoc](https://godoc.org/github.com/dghubble/gologin)

## Roadmap

* Soundcloud
* Bitbucket
* Google
* Improve test coverage
* Improve examples and documentation

## Contributing

See the [Contributing Guide](https://gist.github.com/dghubble/be682c123727f70bcfe7).

## License

[MIT License](LICENSE)


