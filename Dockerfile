FROM golang:1.21.0 as builder
WORKDIR /app
RUN go mod init go-qr-code-generator
COPY *.go ./
COPY go.* ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /go-qr-code-generator

FROM gcr.io/distroless/base-debian11
WORKDIR /
COPY --from=builder /go-qr-code-generator /go-qr-code-generator
ENV PORT 8080
USER nonroot:nonroot
CMD ["/go-qr-code-generator"]