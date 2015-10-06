package gologin

// CookieOptions configures http.Cookie creation.
type CookieOptions struct {
	// Domain sets the cookie domain. Defaults to the host name of the responding
	// server when left zero valued.
	Domain string
	// Path sets the cookie path. Defaults to the path of the URL responding to
	// the request when left zero valued.
	Path string
	// MaxAge=0 means no 'Max-Age' attribute should be set.
	// MaxAge<0 means delete cookie now, equivalently 'Max-Age: 0'.
	// MaxAge>0 means Max-Age attribute present and given in seconds.
	// Cookie 'Expires' will be set (or left unset) according to MaxAge
	MaxAge int
	// HTTPOnly indicates whether the browser should prohibit a cookie from
	// being accessible via Javascript. Recommended true.
	HTTPOnly bool
	// Secure flag indicating to the browser that the cookie should only be
	// transmitted over a TLS HTTPS connection. Recommended true in production.
	Secure bool
}

// DefaultCookieOptions configures http.Cookie creation for short-lived state
// parameter cookies.
var DefaultCookieOptions = CookieOptions{
	Path:     "/",
	MaxAge:   60, // 60 seconds
	HTTPOnly: true,
	Secure:   true, // HTTPS only
}

// DebugOnlyCookieOptions configures http.Cookie create for short-lived state
// parameter cookies, but does NOT require cookies be sent over HTTPS! It
// may be used for development, but should NEVER be used for production.
var DebugOnlyCookieOptions = CookieOptions{
	Path:     "/",
	MaxAge:   60, // 60 seconds
	HTTPOnly: true,
	Secure:   false, // allows cookies to be send over HTTP
}
