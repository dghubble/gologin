
# Apple Login

Login with Apple allows users to login to any web app with their Apple account.

## Web

Package `gologin` provides Go handlers for the Apple OAuth2 Authorization flow and for obtaining the Apple [User struct](https://github.com/dghubble/gologin/blob/master/apple/verify.go).

### Getting Started

    go get github.com/dghubble/gologin/apple
    cd $GOPATH/src/github.com/dghubble/gologin/examples/apple
    go get .

## Example App

[main.go](main.go) shows an example web app which uses `gologin` to issue a client-side cookie session. For simplicity, no data is persisted.

Login to your Apple Developer account and navigate to [Certificates, IDs and Profiles](https://developer.apple.com/account/resources/identifiers/list)

You'll need to create and configure an App ID, Service ID, and Key.

### Find your Team ID

Once logged into the Apple Developer portal, you'll be able to see your *Team ID* in the top right corner beside your name. It has a format similar to `7MD8ZRX9JK`. Take note of it.

### Create an App ID

From the sidebar, choose *Identifiers* and click on the "plus" icon to add one. In the first step, choose `App IDs`. Specify a bundle ID such as `com.foo.app`, then scroll down through the list of capabilities and check the box next to *Sign In with Apple*, then click on *Confirm* to save this step.

### Create a Services ID

The service ID identifies a particular instance of your app, so you'll need to create one for your website. Once again, in *Identifiers*, click on the "plus" icon to add one. In the first step, choose `Services IDs`. In the next step, you'll specify an *Identifier*, such as `com.foo.app.web`. You'll also need to scroll down through the list of capabilities and check the box next to *Sign In with Apple* and also click on the *Configure* button next to it. This is where you'll specify allowed domains and redirect (callback) URLs. *Note that Apple does not accept localhost or 127.0.0.1 as redirect URLs*, so if you are running the example, you may want to point e.g., `localhost.com` to `127.0.0.1` via your `/etc/hosts` file, so that you can use `http://localhost.com/apple/callback` as an authorized redirect URL for running the example locally.

### Create a Private Key

Instead of providing an OAuth client secret, Apple uses a public/private key pair and provides you the private key which you will use to generate a signed JWT token which will serve as the secret. The signed token typically has a short-term expiry (max expiry is 6 months), so it is suggested to re-generate the secret from the private key for every request or periodically. To generate and download a private key, click on the *Keys* section in the Apple Developer portal side nav. Click the "plus" icon to register a new key. Specify a key name and make sure to scroll down and check the *Sign In with Apple* service. Click *Configure* next to it and select the App ID you created earlier as your "Primary App ID". Once you save, you will have one chance to download your secret key, so make sure to save it during this step. The file you download should have a name like `AuthKey_THEKEYID.p8`. Note the key ID which is part of the file name.

### Running the Example

Compile and run `main.go` from `examples/apple`. Pass the client id, private key file path, team id, and private key id as arguments to the executable

    go run main.go -client-id=xx -client-key-path=yy -team-id=zz -key-id=aa
    2015/09/25 23:09:13 Starting Server listening on localhost:8080

or set the `APPLE_CLIENT_ID`, `APPLE_CLIENT_KEY_PATH`, `APPLE_TEAM_ID`, and `APPLE_KEY_ID` environment variables.

### Authorization Flow

1. The "Login with Apple" link to the login handler directs the user to the Apple OAuth2 Auth URL to obtain a permission grant.
2. The redirection URI (callback handler) receives the OAuth2 callback, verifies the state parameter, and obtains an OIDC Token.
3. The success `http.Handler` is called with a `Context` which contains the Apple Token and verified Apple User struct.
4. In this example, that User is read and used to issue a signed cookie session.

