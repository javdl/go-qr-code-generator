# QR code generator in Go
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fjavdl%2Fgo-qr-code-generator.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Fjavdl%2Fgo-qr-code-generator?ref=badge_shield)


This is a small repository showing how to generate a QR code with an optional watermark in Go.

Credits to Matthew Setter, his blogpost [https://www.twilio.com/blog/generate-qr-code-with-go](https://www.twilio.com/blog/generate-qr-code-with-go) explains this code. Note: the blogpost mentions `content=` instead of `url=` to use for the content you want to encode in the QR code.

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

## License
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fjavdl%2Fgo-qr-code-generator.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2Fjavdl%2Fgo-qr-code-generator?ref=badge_large)