package main

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/maxihafer/paperless-sftp-rest-adapter/pkg/client"
	"log/slog"
	"math"
	"os"
	"sync"
	"time"
)

const (
	PaperproxyWatchDirEnvKey  = "WATCH_DIR"
	PaperproxyWatchDirDefault = "/consume"

	PaperproxyPaperlessHostDefault = "localhost:8000"
	PaperproxyPaperlessHostEnvKey  = "PAPERLESS_HOST"

	PaperproxyPaperlessApiKeyEnvKey = "PAPERLESS_API_KEY"
)

func processFile(paperless *client.Client, filePath string) error {
	id, err := paperless.UploadDocument(filePath)
	if err != nil {
		return fmt.Errorf("failed to upload document: %w", err)
	}
	slog.Info("document processed", "file", filePath, "id", id)
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete file after successful upload: %w", err)
	}
	slog.Debug("cleanup complete", "file", filePath)
	return nil
}

func main() {
	level := LogLevelFromEnv()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))
	slog.SetDefault(logger)

	watchDir, ok := os.LookupEnv(PaperproxyWatchDirEnvKey)
	if !ok {
		watchDir = PaperproxyWatchDirDefault
	}

	paperlessHost, ok := os.LookupEnv(PaperproxyPaperlessHostEnvKey)
	if !ok {
		paperlessHost = PaperproxyPaperlessHostDefault
	}

	paperlessApiKey, ok := os.LookupEnv(PaperproxyPaperlessApiKeyEnvKey)
	if !ok {
		panic("PAPERLESS_API_KEY must be set")
	}

	slog.Info("starting paperless-sftp-rest-adapter", "watchdir", watchDir, "paperless-host", paperlessHost)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	defer watcher.Close()

	paperless := client.New(paperlessHost, paperlessApiKey)

	slog.Info("running startup reconcile")
	files, err := os.ReadDir(watchDir)
	if err != nil {
		slog.Error("failed to list directory for startup reconcile", "err", err)
		os.Exit(1)
	}
	if len(files) > 0 {
		slog.Group("startup-reconcile")
		slog.Warn("cleaning up orphaned files", "count", len(files), "watchdir", watchDir)

		for _, file := range files {
			if err := processFile(paperless, file.Name()); err != nil {
				slog.Error("failed while processing file", "err", err)
			}
		}
	}

	go func() {
		var (
			waitFor = 100 * time.Millisecond
			mu      sync.Mutex
			timers  = make(map[string]*time.Timer)
		)

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				slog.Debug("watched event", "event", event)

				if !event.Has(fsnotify.Create) && !event.Has(fsnotify.Write) {
					continue
				}

				mu.Lock()
				t, ok := timers[event.Name]
				mu.Unlock()

				if !ok {
					t = time.AfterFunc(math.MaxInt64, func() {
						if err := processFile(paperless, event.Name); err != nil {
							slog.Error("failed to process file", "err", err)
							return
						}
					})
					t.Stop()

					mu.Lock()
					timers[event.Name] = t
					mu.Unlock()
				}

				t.Reset(waitFor)

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				slog.Error("watcher encountered an error", "error", err)
			}
		}
	}()

	if err := watcher.Add(watchDir); err != nil {
		slog.Error("failed to add watchdir", "error", err)
		panic(err)
	}

	<-make(chan struct{})
}
