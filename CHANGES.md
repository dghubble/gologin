
# gologin Changelog

Notable changes between releases.

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