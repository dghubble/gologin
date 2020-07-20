package apple

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"math/big"
	"net/http"
	"time"

	"github.com/dghubble/gologin/v2"
	oauth2Login "github.com/dghubble/gologin/v2/oauth2"

	"github.com/dgrijalva/jwt-go"

	"golang.org/x/oauth2"
)

// Apple base URL
const (
	AppleBaseURL = "https://appleid.apple.com"
)

// User is an Apple user.
//
// Note that user IDs are unique to each app.
type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

// Global cache for Apple public keys
var applePublicKeys map[string]*rsa.PublicKey

// Apple public key description
type applePublicKey struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type applePublicKeyListing struct {
	Keys []applePublicKey `json:"keys"`
}

// Apple id_token claims
type appleIDClaims struct {
	AtHash         string `json:"at_hash"`
	AuthTime       int64  `json:"auth_time"`
	Email          string `json:"email"`
	EmailVerified  string `json:"email_verified"`
	NonceSupported bool   `json:"nonce_supported"`
	jwt.StandardClaims
}

// Apple login errors
var (
	ErrUnableToDecodePEMKey = errors.New("apple: unable to PEM decode client secret key")
	ErrIDTokenNotFound      = errors.New("apple: unable to find id_token in exchanged token extras")
	ErrIDTokenKeyIDNotFound = errors.New("apple: unable to find key ID in id_token headers")
	ErrIDTokenKeyNotFound   = errors.New("apple: unable to find Apple key specified by id_token")
)

// ClientSecret returns the client secret to use for oauth2.Config or
// empty string ("") on failure.
//
// Apple provides a private ES256 key which must be used to sign a JWT token
// which serves as the client secret. Since the JWT token is designed to expire
// (max 6 months), it's suggested to use a short expiry and generate a new
// one each time.
func ClientSecret(pkey []byte, expSecs int64, keyID string, teamID string, clientID string) (string, error) {
	block, _ := pem.Decode(pkey)
	if block == nil {
		return "", ErrUnableToDecodePEMKey
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return "", err
	}

	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.StandardClaims{
		Issuer:    teamID,
		Subject:   clientID,
		Audience:  AppleBaseURL,
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Unix() + expSecs,
	})
	token.Header["kid"] = keyID

	return token.SignedString(key)
}

// CacheAppleKeys fetches and caches Apple's public keys.
// Returns error on failure or nil on success.
// You can call this from e.g., program start to pre-fetch and cache the keys
// for later validation. If you do not pre-fetch, then the fetch will happen
// automatically the first time that a key validation is done.
func CacheAppleKeys() error {
	keys, err := cacheApplePKeys()
	if err == nil {
		applePublicKeys = keys
	}
	return err
}

// StateHandler checks for a state cookie. If found, the state value is read
// and added to the ctx. Otherwise, a non-guessable value is added to the ctx
// and to a (short-lived) state cookie issued to the requester.
//
// Implements OAuth 2 RFC 6749 10.12 CSRF Protection. If you wish to issue
// state params differently, write a http.Handler which sets the ctx state,
// using oauth2 WithState(ctx, state) since it is required by LoginHandler
// and CallbackHandler.
func StateHandler(config gologin.CookieConfig, success http.Handler) http.Handler {
	return oauth2Login.StateHandler(config, success)
}

// LoginHandler handles Apple login requests by reading the state value
// from the ctx and redirecting requests to the AuthURL with that state value.
func LoginHandler(config *oauth2.Config, failure http.Handler) http.Handler {
	return appleLoginHandler(config, failure)
}

// CallbackHandler handles Apple redirection URI requests and adds the
// Apple access token and User to the ctx. If authentication succeeds,
// handling delegates to the success handler, otherwise to the failure
// handler.
func CallbackHandler(config *oauth2.Config, success, failure http.Handler) http.Handler {
	success = appleHandler(config, success, failure)
	return oauth2Login.CallbackHandler(config, success, failure)
}

// appleLoginHandler handles OAuth2 login requests by reading the state value from
// the ctx and redirecting requests to the AuthURL with that state value.
// Very similar to oauth2Login.LoginHandler(), except that it sets the AuthURLParam
// "response_mode" to "form_post", as is required by Apple.
func appleLoginHandler(config *oauth2.Config, failure http.Handler) http.Handler {
	if failure == nil {
		failure = gologin.DefaultFailureHandler
	}
	fn := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		state, err := oauth2Login.StateFromContext(ctx)
		if err != nil {
			ctx = gologin.WithError(ctx, err)
			failure.ServeHTTP(w, req.WithContext(ctx))
			return
		}
		authURL := config.AuthCodeURL(state, oauth2.SetAuthURLParam("response_mode", "form_post"))
		http.Redirect(w, req, authURL, http.StatusFound)
	}
	return http.HandlerFunc(fn)
}

// appleHandler is a http.Handler that gets the OAuth2 Token from the ctx
// to get the corresponding Apple User. If successful, the user is added to
// the ctx and the success handler is called. Otherwise, the failure handler
// is called.
func appleHandler(config *oauth2.Config, success, failure http.Handler) http.Handler {
	if failure == nil {
		failure = gologin.DefaultFailureHandler
	}
	fn := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		token, err := oauth2Login.TokenFromContext(ctx)
		if err != nil {
			ctx = gologin.WithError(ctx, err)
			failure.ServeHTTP(w, req.WithContext(ctx))
			return
		}

		idt, ok := token.Extra("id_token").(string)
		if !ok {
			ctx = gologin.WithError(ctx, ErrIDTokenNotFound)
			failure.ServeHTTP(w, req.WithContext(ctx))
			return
		}

		parser := jwt.Parser{ValidMethods: []string{"RS256"}}
		idToken, err := parser.ParseWithClaims(idt, &appleIDClaims{}, func(t *jwt.Token) (interface{}, error) {
			keyID, ok := t.Header["kid"].(string)
			if !ok {
				return nil, ErrIDTokenKeyIDNotFound
			}
			key := getApplePKey(keyID)
			if key == nil {
				return nil, ErrIDTokenKeyNotFound
			}
			return key, nil
		})
		if (err != nil) || !idToken.Valid {
			ctx = gologin.WithError(ctx, err)
			failure.ServeHTTP(w, req.WithContext(ctx))
			return
		}

		user := new(User)
		if claims, ok := idToken.Claims.(*appleIDClaims); ok {
			user.ID = claims.Subject
			user.Email = claims.Email
		} else {
			ctx = gologin.WithError(ctx, err)
			failure.ServeHTTP(w, req.WithContext(ctx))
			return
		}

		ctx = WithUser(ctx, user)
		success.ServeHTTP(w, req.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}

// getApplePKey returns the Apple public key identified by the provided argument keyID
//
// Checks cache first otherwise fetches public keys into cache and tries again.
func getApplePKey(keyID string) *rsa.PublicKey {
	if pkey, ok := applePublicKeys[keyID]; ok && (pkey != nil) {
		return pkey
	}

	applePublicKeys, err := cacheApplePKeys()
	if err != nil {
		return nil
	}

	if pkey, ok := applePublicKeys[keyID]; ok {
		return pkey
	}

	return nil
}

// cacheApplePKeys retrieves and parses Apple's public keys and returns a map of
// parsed RSA public keys, keyed by ID of public key as specified by Apple
func cacheApplePKeys() (map[string]*rsa.PublicKey, error) {
	client := http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("GET", AppleBaseURL+"/auth/keys", nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var keysContent applePublicKeyListing
	err = json.Unmarshal(body, &keysContent)
	if err != nil {
		return nil, err
	}

	appleKeys := make(map[string]*rsa.PublicKey)
	for _, key := range keysContent.Keys {
		appleKeys[key.Kid] = makeApplePKey(key.N, key.E)
	}

	return appleKeys, nil
}

// makeApplePKey returns an rsa.PublicKey object, constructing it from the specified base
// "n" and exponent "e", which are provided Base64-encoded
func makeApplePKey(b64N string, b64E string) *rsa.PublicKey {
	bytesN, err := base64.RawURLEncoding.DecodeString(b64N)
	if err != nil {
		return nil
	}
	bytesE, err := base64.RawURLEncoding.DecodeString(b64E)
	if err != nil {
		return nil
	}

	intE := int(0)
	for _, v := range bytesE {
		intE = intE << 8
		intE = intE | int(v)
	}

	pkey := new(rsa.PublicKey)
	pkey.N = new(big.Int)
	pkey.N.SetBytes(bytesN)
	pkey.E = intE
	return pkey
}
