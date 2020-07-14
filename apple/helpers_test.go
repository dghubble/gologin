package apple

import (
	"crypto/rsa"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/oauth2"
)

var (
	testPrivKey []byte
	testPubKey  []byte
)

var applePrivateKeys map[string]*rsa.PrivateKey

func setupAppleTestKeys() {
	testPrivKey = []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAt+xNTarNOoLJ/LJxeFL/t3rbbvKNZFMLeC8eo2sI9h1LWSOu
6w5ZQk60iIodmSWYxvCLxBWzTOqknL4sDd2QA+cEs9GFT6uS24L2eyb8zY8/AgwI
CJfaqIJZaQ3xQupol+Q5rCPIsjNAKR27oaR/EbkMC1laCfDoXscJg8W+4GvSHWTz
hL9GFhQlHqh6IFn/ta++9UKuWv+9Hs25xMnVXt0WpSjBz0vRatMZ+vrb4pVzqVHL
bRWw/jtBexcwuLFZ/1unY2J4NSg3Jql6jM7KqY8Nf9zFPRihpVeMiX+pxET2l7Pu
8Bu5fNH9k+68vZbQi3XvrmQKHqZs+P/6bM4C7wIDAQABAoIBAF5lZXeLRjHVlp2f
aCV9U8lzwNO8oVzwUl6osGznLn5Cor1pVwlFIAKsKnQ5jt9fMH5KTzGggZnkg//+
itXC9XtLQlqYGne9c24+VQr4A5/s+UWvrx/Z8Fu0KveENGNHs87hT8hNxV/Qdgmk
PPzFVIJgGxJoFZIsltauCPAcuc6sKSaAmbidA3F4Ef/g9/GnuPEh1VzhCqyoNvbE
1d2CyzLxHDmXqJD7vQq86wc7kZvMHziexhBEqP4fZJtn6RhJ5qSuKT8BNDAHASqW
vJJOeT5L8AIFXvmAcwZnSjQHHb1w6rmjyLj3FKfpcCslt2aBQYQRWaERyeV3p34W
pUzxN6ECgYEA3sdTYiOVKNhPEhD96JHKtt/aUflWJma6DUZgO9EXqRlCYqko/0s7
Rza4w0Hznt5dE8Te4Mo4bPUH9/ViZ7NhnvPE6yZRf8DX3CV8vMKs91gn3j88pcVP
xjUQFT5iMTsSJhaDHJQSg4FK2/nxtsz9QT267c/qZKq9zfobSi3vtdMCgYEA01ml
y8LWTXm3h/UDzM9/lhg1wrkU/nVFJEzBDgYEm9WZnQovyInwGh80JDc6QV/u+4ty
eEwj6mtY9J/SpYWjhP0ADi8LKXWJq/EhuiiqrDj6c/qn2RdIaWfb70bXzXBWGOkT
VZiW+WhRNw9NtbBb4YWKzV2sQFr2rb9fKufnAPUCgYEAtrG+BuBpZVqm1YkLwNs2
4+wGDW2tocZi05ogN03M2ob1cxWIonweu9L7iF0gnet7Z0fvA2ezCF+Vzln0/lgU
OZdtqO3+rgcGvuobNm1sDVfFMjSn1sZOGpzPeKx1OCxaQNP7Z8diu2efbXC3MhM/
qW4nSvlUHoQLLczq7lVnnLMCgYBzgZYb8yK9+tx0AFMQVwLKm/adshsoKh0chpon
uOBB7o3ihpOwzLoc/Jq5hDlhSzXH4eEwn6QtVHesUcCE17GTV9X06n72LJeOEd21
6M3GC+nNAttCyPe5K5rGfXgpfdCAErmPWTKBoiJorgNxXa4JZbuDG0OtdElGkcVI
JK9aFQKBgQDRjeYWxbzsdWqJnYtXx6AeM8KPq2x24TrQLr/u34yllS74FUpGPRW7
YKwHlWvWU8oH44NCbLkOXfxp9mZL9wSTt75RH5zOQP/b2IhjVFXAMB3CJiLz9Dbe
UJVSqKWx5yU2Spi6at/VRpYIaC/I5QRviUctuGSYmaToCHoUqf1SpQ==
-----END RSA PRIVATE KEY-----`)

	testPubKey = []byte(`-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAt+xNTarNOoLJ/LJxeFL/
t3rbbvKNZFMLeC8eo2sI9h1LWSOu6w5ZQk60iIodmSWYxvCLxBWzTOqknL4sDd2Q
A+cEs9GFT6uS24L2eyb8zY8/AgwICJfaqIJZaQ3xQupol+Q5rCPIsjNAKR27oaR/
EbkMC1laCfDoXscJg8W+4GvSHWTzhL9GFhQlHqh6IFn/ta++9UKuWv+9Hs25xMnV
Xt0WpSjBz0vRatMZ+vrb4pVzqVHLbRWw/jtBexcwuLFZ/1unY2J4NSg3Jql6jM7K
qY8Nf9zFPRihpVeMiX+pxET2l7Pu8Bu5fNH9k+68vZbQi3XvrmQKHqZs+P/6bM4C
7wIDAQAB
-----END PUBLIC KEY-----`)

	applePublicKeys = make(map[string]*rsa.PublicKey)
	applePublicKeys["testkey"], _ = jwt.ParseRSAPublicKeyFromPEM(testPubKey)

	applePrivateKeys = make(map[string]*rsa.PrivateKey)
	applePrivateKeys["testkey"], _ = jwt.ParseRSAPrivateKeyFromPEM(testPrivKey)
}

func buildTokenWithOIDC(id string, email string, keyID string) *oauth2.Token {
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, appleIDClaims{
		Email:         email,
		EmailVerified: "true",
		StandardClaims: jwt.StandardClaims{
			Issuer:    "https://appleid.apple.com",
			Subject:   id,
			Audience:  "com.example.foo",
			IssuedAt:  now.Unix(),
			ExpiresAt: now.Unix() + (60 * 60),
		},
	})
	token.Header["kid"] = keyID

	oauth2Tok := &oauth2.Token{}
	signedToken, _ := token.SignedString(applePrivateKeys["testkey"])
	return oauth2Tok.WithExtra(map[string]interface{}{
		"id_token": signedToken,
	})
}
