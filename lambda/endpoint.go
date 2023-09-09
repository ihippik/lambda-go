package lambda

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type service interface {
	Create(ctx context.Context, name string, file io.ReadCloser) error
	Invoke(ctx context.Context, name string, data []byte) ([]byte, error)
}

// Endpoint represent http-service endpoints.
type Endpoint struct {
	svc        service
	serverAddr string
	logger     *slog.Logger
}

// NewEndpoint returns new Endpoint instance.
func NewEndpoint(svc service, logger *slog.Logger, addr string) *Endpoint {
	return &Endpoint{svc: svc, logger: logger, serverAddr: addr}
}

type createResponse struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// create http endpoint for create lambda function.
func (e *Endpoint) create(w http.ResponseWriter, r *http.Request) {
	const gzHeader = "application/gzip"
	vars := mux.Vars(r)
	name := vars["name"]

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		fmt.Println(err)
		return
	}

	if fileHeader.Header.Get("Content-Type") != gzHeader {
		http.Error(w, "invalid file type", http.StatusBadRequest)
		return
	}

	e.logger.Info("got create request", slog.Any("func_name", name))

	if err := e.svc.Create(r.Context(), name, file); err != nil {
		e.logger.Error("create: service error", "err", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := createResponse{Name: name, Description: "lambda function was created"}
	data, err := json.Marshal(resp)
	if err != nil {
		e.logger.Error("create: marshal error", "err", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)

	_, err = w.Write(data)
	if err != nil {
		e.logger.Error("create: write error", "err", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// invoke http endpoint for invoke lambda function.
func (e *Endpoint) invoke(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	data, err := io.ReadAll(r.Body)
	if err != nil {
		e.logger.Error("lambda: read body error", "err", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	e.logger.Info("got lambda request", slog.Any("func_name", name))

	respData, err := e.svc.Invoke(r.Context(), name, data)
	if err != nil {
		e.logger.Error("lambda: invoke service error", "err", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if _, err := w.Write(respData); err != nil {
		e.logger.Error("lambda: write error", "err", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// StartServer starts http-server.
func (e *Endpoint) StartServer(ctx context.Context) error {
	r := mux.NewRouter()
	r.HandleFunc("/lambda/{name}/create", e.create).Methods(http.MethodPost)
	r.HandleFunc("/lambda/{name}/invoke", e.invoke).Methods(http.MethodPost)

	srv := &http.Server{
		Handler:      r,
		Addr:         e.serverAddr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	go func() {
		<-ctx.Done()
		if err := srv.Shutdown(ctx); err != nil {
			e.logger.Error(err.Error())
		}
	}()

	e.logger.Info("server started", slog.String("addr", e.serverAddr))

	return srv.ListenAndServe()
}
