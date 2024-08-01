# QR code generator in Go

This is a small repository showing how to generate a QR code with an optional watermark in Go.

Credits to Matthew Setter, his blogpost ["How to Generate a QR Code with Go"][tutorial-url] explains this code. Note: the blogpost mentions `content=` instead of `url=` to use for the content you want to encode in the QR code.

## Prerequisites

To follow along with the tutorial, you don't need much, just the following things:

- [Go][go-url] (a recent version, or the latest, 1.20.5)
- [Curl][curl-url] or [Postman][postman-url]
- A smartphone with a QR code scanner (which, these days, most of them should have)

## Start the API

```sh
go run main.go
```

Test if the API works

```sh
curl -i -X POST http://localhost:8080/generate
```

## Create a QR code

```sh
curl -X POST \
    --form "size=256" \
    --form "url=https://fashionunited.com" \
    --output data/qrcode.png \
    http://localhost:8080/generate
```

You can also watermark the QR code, by uploading a PNG file using the `watermark` POST variable.
Below is an example of how to do so with curl.

```bash
curl -X POST \
    --form "size=256" \
    --form "url=https://fashionunited.com" \
    --form "watermark=@data/twilio-logo.png" \
    --output data/qrcode.png \
    http://localhost:8080/generate
```

## Create multiple QR codes 

Put a list of urls in example.txt. Make sure there is a newline at the end.

```
./generate.sh
```

## Deploy to Google Cloud Run

```sh
gcloud config set project PROJECT_ID
gcloud run deploy --region=europe-west1 --allow-unauthenticated
```

## Get the service URL without deploying

```sh
gcloud run services describe go-qr-code-generator --platform managed --region europe-west1 --format 'value(status.url)'
```

## Test if the API works

```sh
curl -X POST \
    --form "size=256" \
    --form "url=https://fashionunited.nl/modevacatures/werken-bij/ray-ban-vacatures/search/in/rotterdam" \
    --output data/ray-ban-rotterdam.png \
    https://go-qr-code-generator-XXXXXXXXXX.a.run.app/generate
```


### CLI tooling

#### Local development

1. Set Project Id:

    ```bash
    export GOOGLE_CLOUD_PROJECT=<GCP_PROJECT_ID>
    # in this case:
    export GOOGLE_CLOUD_PROJECT=go-qr-code-generator
    ```

<!-- 2. Build and Start the server:

    ```bash
    go build -o server && ./server
    ``` -->
    
[tutorial-url]: https://www.twilio.com/blog/generate-qr-code-with-go
[go-url]: https://go.dev/
[curl-url]: https://curl.se/
[postman-url]: https://www.postman.com/downloads/