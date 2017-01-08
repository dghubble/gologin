
# Digits Login

Twitter's [Digits](https://get.digits.com/) service provides phone number login via SMS confirmation codes for any web app or mobile app.

## Digits for Web

The Digits [Javascipt SDK](https://cdn.digits.com/1/sdk.js) launches a web popup to prompt a user to enter a phone number and an SMS confirmation code.

The page receives OAuth Echo headers which can be posted to a Go server, validated, and used to fetch the user's Digits account. Package `gologin` provides Go handlers for Digits which perform these steps correctly.

### Getting Started

    go get github.com/dghubble/gologin/digits
    cd $GOPATH/src/github.com/dghubble/gologin/examples/digits
    go get .

## Example App

[main.go](main.go) shows an example web app which uses `gologin` for Digits to issue a client-side cookie session. For simplicity, no data is persisted.

Get your Digits application's consumer key from the [fabric.io](https://fabric.io) dashboard.

Note: Currently a Digits application must be created by making a dummy iOS or Android app via the Fabric [iOS Mac App](https://fabric.io/downloads/xcode) or [Android Studio Plugin](https://fabric.io/downloads).

Compile and run `main.go` from `examples/digits`. Pass the key as an argument to the executable

    go run main.go -consumer-key=xx
    2015/09/25 21:44:48 Starting Server listening on localhost:8080

or set the `DIGITS_CONSUMER_KEY` environment variable.

Here is what the web login flow looks like (showing a mobile browser).

<img src="https://storage.googleapis.com/dghubble/digits-web-login.gif">

### Authorization Flow

<img align="right" src="https://storage.googleapis.com/dghubble/digits-phone-number.png">

1. Clicking the "Login with Digits" button launches the Digits web login popup.

2. User enters a phone number and receives an SMS confirmation number to enter. The page receives OAuth Echo fields and POSTs them to the Go server.

3. The Echo fields are validated, used to obtain the Digits `Account`, and provided to the specified success `http.Handler`.

4. In this example, that account is read and used to issue the user a signed cookie session.

