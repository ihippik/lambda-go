package lambda

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"time"
)

// Handler is a user function that handles lambda requests.
type Handler func(ctx context.Context, payload []byte) ([]byte, error)

// handle is a wrapper for user Handler.
type handle struct {
	handler Handler
}

// Start starts the lambda handler.
func Start(handler Handler) {
	const serverAddr = ":8080"

	server := http.Server{
		Addr:         serverAddr,
		Handler:      handle{handler: handler},
		WriteTimeout: 2 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		slog.Error(err.Error())
	}
}

// ServeHTTP implements http.Handler interface.
func (h handle) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	data, err := io.ReadAll(request.Body)
	if err != nil {
		errData := NewError(http.StatusBadRequest, err.Error()).Error()

		if _, err := writer.Write(errData); err != nil {
			slog.Error(err.Error())
		}
	}

	respData, err := h.handler(request.Context(), data)
	if err != nil {
		errData := NewError(http.StatusConflict, err.Error()).Error()

		if _, err := writer.Write(errData); err != nil {
			slog.Error(err.Error())
		}
	}

	if _, err = writer.Write(respData); err != nil {
		slog.Error(err.Error())
	}
}
