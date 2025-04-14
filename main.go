package main

import (
	"github.com/fsnotify/fsnotify"
	"github.com/maxihafer/paperless-sftp-rest-adapter/pkg/client"
	"log/slog"
	"os"
)

const (
	PaperproxyWatchDirEnvKey  = "WATCH_DIR"
	PaperproxyWatchDirDefault = "/consume"

	PaperproxyPaperlessHostDefault = "localhost:8000"
	PaperproxyPaperlessHostEnvKey  = "PAPERLESS_HOST"

	PaperproxyPaperlessApiKeyEnvKey = "PAPERLESS_API_KEY"
)

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

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				slog.Debug("watched event", "event", event)

				if event.Has(fsnotify.Create) {
					slog.Info("processing file", "file", event.Name)
					if id, err := paperless.UploadDocument(event.Name); err != nil {
						slog.Error("failed to upload document", "file", event.Name, "error", err)
					} else {
						slog.Info("uploaded document", "file", event.Name, "id", id)
						if err := os.Remove(event.Name); err != nil {
							slog.Error("failed to remove file after upload", "file", event.Name, "error", err)
						} else {
							slog.Info("removed file after upload", "file", event.Name)
						}
					}
				}
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
