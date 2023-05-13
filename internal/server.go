package internal

import (
	"bytes"
	"cloud.google.com/go/storage"
	"context"
	"github.com/davidbyttow/govips/v2/vips"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"os"
	"time"
)

const bucketName = "images.sonurai.com"

func Start() error {
	ctx := context.Background()

	storageClient, err := storage.NewClient(ctx)
	if err != nil {
		log.Error().
			Err(err).
			Msg("failed to create storage client")
	}
	defer storageClient.Close()

	b := storageClient.Bucket(bucketName)
	if _, err := b.Attrs(ctx); err != nil {
		return err
	}

	vipsLogFn := func(messageDomain string, level vips.LogLevel, message string) {
		log.Warn().
			Str("domain", messageDomain).
			Msg(message)
	}
	vips.LoggingSettings(vipsLogFn, vips.LogLevelWarning)
	vips.Startup(nil)

	ss := server{bucket: b}

	r := chi.NewRouter()
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(""))
	})
	r.Get("/{img}", ss.GetWallpaperHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	s := http.Server{
		Addr:        ":" + port,
		Handler:     r,
		ReadTimeout: time.Second * 10,
	}

	return s.ListenAndServe()
}

type server struct {
	bucket *storage.BucketHandle
}

func (s *server) GetWallpaperHandler(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "img")

	rr, err := s.bucket.Object(key).NewReader(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	img, err := vips.NewImageFromReader(rr)
	if err != nil {
		log.Error().Err(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := img.Resize(0.417, vips.KernelLanczos3); err != nil {
		log.Error().Err(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	b, _, err := img.ExportJpeg(&vips.JpegExportParams{
		Quality:   70,
		Interlace: true,
	})

	_, err = io.Copy(w, bytes.NewReader(b))
	if err != nil {
		log.Error().Err(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
