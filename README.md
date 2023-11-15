# QR code generator in Go

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

## Deploy

```sh
gcloud auth login
gcloud auth configure-docker

run deploy go-qr-code-generator --project kubernetes-164514 --image gcr.io/kubernetes-164514/go-qr-code-generator --client-name Cloud Code for VS Code --client-version 2.1.1 --platform managed --region europe-west1 --allow-unauthenticated --port 8080 --cpu 1 --memory 256Mi --concurrency 80 --timeout 300 --clear-env-vars
```