package internal

import (
	"net/http"
	"time"

	"github.com/dghubble/gologin"
)

// NewCookie returns a new http.Cookie with the given name, value, and optional
// CookieOptions struct. By default, the DefaultCookieOptions are used.
//
// The MaxAge field is used to determine whether an Expires field should be
// added for Internet Explorer compatability and what its value should be.
func NewCookie(name, value string, opts ...gologin.CookieOptions) *http.Cookie {
	var options gologin.CookieOptions
	if len(opts) > 0 {
		options = opts[0]
	} else {
		options = gologin.DefaultCookieOptions
	}

	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Domain:   options.Domain,
		Path:     options.Path,
		MaxAge:   options.MaxAge,
		HttpOnly: options.HTTPOnly,
		Secure:   options.Secure,
	}
	// IE <9 does not understand MaxAge, set Expires if MaxAge is non-zero.
	if expires, ok := expiresTime(options.MaxAge); ok {
		cookie.Expires = expires
	}
	return cookie
}

// expiresTime converts a maxAge time in seconds to a time.Time in the future
// if the maxAge is positive or the beginning of the epoch if maxAge is
// negative. If maxAge is exactly 0, an empty time and false are returned
// (so the Cookie Expires field should not be set).
// http://golang.org/src/net/http/cookie.go?s=618:801#L23
func expiresTime(maxAge int) (time.Time, bool) {
	if maxAge > 0 {
		d := time.Duration(maxAge) * time.Second
		return time.Now().Add(d), true
	} else if maxAge < 0 {
		return time.Unix(1, 0), true // first second of the epoch
	}
	return time.Time{}, false
}
