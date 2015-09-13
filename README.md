
# gologin [![Build Status](https://travis-ci.org/dghubble/gologin.png)](https://travis-ci.org/dghubble/gologin) [![GoDoc](http://godoc.org/github.com/dghubble/gologin?status.png)](http://godoc.org/github.com/dghubble/gologin)
<img align="right" src="https://storage.googleapis.com/dghubble/gologin.png">

Gologin includes handlers for popular 3rd party authentication providers.

Choose the package for a login provider and use the `LoginHandler` and `CallbackHandler` to power web logins and the `TokenHandler` to power mobile token logins. 

### Features

* `LoginHandler` and `CallbackHandler` support web login flows
* `TokenHandler` supports (mobile) token login flows (avail in most packages)
* Agnostic to any session library or token library.

## Install

    go get github.com/dghubble/gologin

## Docs

Read [GoDoc](https://godoc.org/github.com/dghubble/gologin)

## Packages

#### Github [![GoDoc](http://godoc.org/github.com/dghubble/go-digits/github?status.png)](http://godoc.org/github.com/dghubble/go-digits/github)
#### Twitter [![GoDoc](http://godoc.org/github.com/dghubble/go-digits/twitter?status.png)](http://godoc.org/github.com/dghubble/go-digits/twitter)
#### Digits [![GoDoc](http://godoc.org/github.com/dghubble/go-digits/digits?status.png)](http://godoc.org/github.com/dghubble/go-digits/digits)
#### OAuth2 [![GoDoc](http://godoc.org/github.com/dghubble/go-digits/oauth2?status.png)](http://godoc.org/github.com/dghubble/go-digits/oauth2)
#### OAuth1 [![GoDoc](http://godoc.org/github.com/dghubble/go-digits/oauth1?status.png)](http://godoc.org/github.com/dghubble/go-digits/oauth1)

## Roadmap

* Use context to pass state between composable ContextHandlers
* Improve test coverage and argument checks
* Improve examples and documentation
* Tumblr
* Google
* Soundcloud
* Bitbucket

## Contributing

See the [Contributing Guide](https://gist.github.com/dghubble/be682c123727f70bcfe7).

## License

[MIT License](LICENSE)


