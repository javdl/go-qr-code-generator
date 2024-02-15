package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/nfnt/resize"
	"github.com/settermjq/go-qr-code-generator/metadata"
	qrcode "github.com/skip2/go-qrcode"

	"cloud.google.com/go/logging"
	"github.com/gorilla/mux"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const MAX_UPLOAD_SIZE = 1024 * 1024 // 1MB
const WATERMARK_WIDTH = 64

type simpleQRCode struct {
	Content string
	Size    int
}
type App struct {
	*http.Server
	projectID string
	log       *logging.Logger
}

func main() {
	ctx := context.Background()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("listening on port %s", port)
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	app, err := newApp(ctx, port, projectID)
	if err != nil {
		log.Fatalf("unable to initialize application: %v", err)
	}
	log.Println("starting HTTP server")
	go func() {
		if err := app.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server closed: %v", err)
		}
	}()

	// Listen for SIGINT to gracefully shutdown.
	nctx, stop := signal.NotifyContext(ctx, os.Interrupt, os.Kill)
	defer stop()
	<-nctx.Done()
	log.Println("shutdown initiated")

	// Cloud Run gives apps 10 seconds to shutdown. See
	// https://cloud.google.com/blog/topics/developers-practitioners/graceful-shutdowns-cloud-run-deep-dive
	// for more details.
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	app.Shutdown(ctx)
	log.Println("shutdown")

}

func newApp(ctx context.Context, port, projectID string) (*App, error) {
	app := &App{
		Server: &http.Server{
			Addr: ":" + port,
			// Add some defaults, should be changed to suit your use case.
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
		},
	}

	if projectID == "" {
		projID, err := metadata.ProjectID()
		if err != nil {
			return nil, fmt.Errorf("unable to detect Project ID from GOOGLE_CLOUD_PROJECT or metadata server: %w", err)
		}
		projectID = projID
	}
	app.projectID = projectID

	client, err := logging.NewClient(ctx, fmt.Sprintf("projects/%s", app.projectID),
		// We don't need to make any requests when logging to stderr.
		option.WithoutAuthentication(),
		option.WithGRPCDialOption(
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		))
	if err != nil {
		return nil, fmt.Errorf("unable to initialize logging client: %v", err)
	}
	app.log = client.Logger("test-log", logging.RedirectAsJSON(os.Stderr))

	// Setup request router.
	r := mux.NewRouter()
	r.HandleFunc("/generate", handleRequest).
		Methods("GET")
	app.Server.Handler = r

	http.ListenAndServe(":8080", r)

	return app, nil
}

// Generate generates a QR code using the value of simpleQRCode.Content
func (code *simpleQRCode) Generate() ([]byte, error) {
	qrCode, err := qrcode.Encode(code.Content, qrcode.Medium, code.Size)
	if err != nil {
		return nil, fmt.Errorf("could not generate a QR code: %v", err)
	}
	return qrCode, nil
}

// GenerateWithWatermark generates a QR code using the value of simpleQRCode.Content
// and adds a watermark to it, centered in the middle of the QR code, using the
// supplied watermark image data
func (code *simpleQRCode) GenerateWithWatermark(watermark []byte) ([]byte, error) {
	qrCode, err := code.Generate()
	if err != nil {
		return nil, err
	}

	qrCode, err = code.addWatermark(qrCode, watermark, code.Size)
	if err != nil {
		return nil, fmt.Errorf("could not add watermark to QR code: %v", err)
	}

	return qrCode, nil
}

// addWatermark adds a watermark to a QR code, centered in the middle of the QR code
func (code *simpleQRCode) addWatermark(qrCode []byte, watermarkData []byte, size int) ([]byte, error) {
	qrCodeData, err := png.Decode(bytes.NewBuffer(qrCode))
	if err != nil {
		return nil, fmt.Errorf("could not decode QR code: %v", err)
	}

	watermarkImage, err := png.Decode(bytes.NewBuffer(watermarkData))
	if err != nil {
		return nil, fmt.Errorf("could not decode watermark: %v", err)
	}

	// Determine the offset to center the watermark on the QR code
	offset := image.Pt(((size / 2) - 32), ((size / 2) - 32))

	watermarkImageBounds := qrCodeData.Bounds()
	m := image.NewRGBA(watermarkImageBounds)

	// Center the watermark over the QR code
	draw.Draw(m, watermarkImageBounds, qrCodeData, image.Point{}, draw.Src)
	draw.Draw(
		m,
		watermarkImage.Bounds().Add(offset),
		watermarkImage,
		image.Point{},
		draw.Over,
	)

	watermarkedQRCode := bytes.NewBuffer(nil)
	png.Encode(watermarkedQRCode, m)

	return watermarkedQRCode.Bytes(), nil
}

// resizeWatermark resizes a watermark image to the desired width and height
func resizeWatermark(watermark io.Reader, width uint) ([]byte, error) {
	decodedImage, err := png.Decode(watermark)
	if err != nil {
		return nil, fmt.Errorf("could not decode watermark image: %v", err)
	}

	m := resize.Resize(width, 0, decodedImage, resize.Lanczos3)
	resized := bytes.NewBuffer(nil)
	png.Encode(resized, m)

	return resized.Bytes(), nil
}

// uploadFile uploads an image file to be used as a watermark for a QR code
func uploadFile(file multipart.File) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, file); err != nil {
		return nil, fmt.Errorf("could not upload file. %v", err)
	}

	return buf.Bytes(), nil
}

// buildErrorResponse is a small utility function to simplify returning a JSON response
// to be returned to the user when an error has occurred
func buildErrorResponse(message string) []byte {
	responseData := make(map[string]string)
	responseData["error"] = message

	response, err := json.Marshal(responseData)
	if err != nil {
		log.Fatalln("Could not generate error message.")
	}

	return response
}

func handleRequest(writer http.ResponseWriter, request *http.Request) {
	request.ParseMultipartForm(10 << 20)
	var size, url string = request.FormValue("size"), request.FormValue("url")
	var codeData []byte

	if url == "" {
		writer.Write(buildErrorResponse("Could not determine the desired QR code content."))
		writer.WriteHeader(400)
		return
	}

	qrCodeSize, err := strconv.Atoi(size)
	if err != nil || size == "" {
		writer.Write(buildErrorResponse(fmt.Sprint("Could not determine the desired QR code size:", err)))
		writer.WriteHeader(400)
		return
	}

	qrCode := simpleQRCode{Content: url, Size: qrCodeSize}

	watermarkFile, _, err := request.FormFile("watermark")
	if err != nil && errors.Is(err, http.ErrMissingFile) {
		fmt.Println("Watermark image was not uploaded or could not be retrieved. Reason: ", err)
		codeData, err = qrCode.Generate()
		if err != nil {
			writer.Write(buildErrorResponse(fmt.Sprintf("Could not generate QR code. %v", err)))
			writer.WriteHeader(400)
			return
		}
		writer.Header().Add("Content-Type", "image/png")
		writer.Write(codeData)
		return
	}

	watermark, err := uploadFile(watermarkFile)
	if err != nil {
		writer.Write(buildErrorResponse(fmt.Sprint("Could not upload the watermark image.", err)))
		writer.WriteHeader(400)
		return
	}

	contentType := http.DetectContentType(watermark)
	if contentType != "image/png" {
		response := buildErrorResponse(fmt.Sprintf("Provided watermark image is a %s not a PNG. %v.", err, contentType))
		writer.Write(response)
		writer.WriteHeader(400)
		return
	}

	watermark, err = resizeWatermark(bytes.NewBuffer(watermark), WATERMARK_WIDTH)
	if err != nil {
		writer.Write(buildErrorResponse("Could not resize the watermark image."))
		writer.WriteHeader(400)
		return
	}

	codeData, err = qrCode.GenerateWithWatermark(watermark)
	if err != nil {
		response := buildErrorResponse(fmt.Sprintf("Could not generate QR code with the watermark image. %v", err))
		writer.Write(response)
		writer.WriteHeader(400)
		return
	}

	writer.Header().Add("Content-Type", "image/png")
	writer.Write(codeData)
}

func main() {
	addr := flag.String("addr", ":8080", "HTTP network address")
	flag.Parse()

	mux := http.NewServeMux()
	mux.HandleFunc("/generate", handleRequest)

	log.Printf("Starting server on %s", *addr)
	err := http.ListenAndServe(*addr, mux)
	if err != nil {
		log.Fatalln(err)
	}
}
