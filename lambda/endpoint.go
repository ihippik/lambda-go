package lambda

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type service interface {
	Create(ctx context.Context, name string) error
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

type adminRequest struct {
	Name string `json:"name"`
}

func (e *Endpoint) create(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	// TODO: get tar archive from request body

	//
	//if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
	//	http.Error(w, err.Error(), http.StatusBadRequest)
	//	return
	//}

	e.logger.Debug("got create request", slog.Any("user_name", name))

	err := e.svc.Create(r.Context(), name)
	if err != nil {
		e.logger.Error("create: service error", "err", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = w.Write([]byte("ok"))
	if err != nil {
		e.logger.Error("create: write error", "err", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (e *Endpoint) lambda(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	data, err := io.ReadAll(r.Body)
	if err != nil {
		e.logger.Error("lambda: read body error", "err", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	e.logger.Debug("got lambda request", slog.Any("user_name", name))

	respData, err := e.svc.Invoke(r.Context(), "test", data)
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
	r.HandleFunc("/lambda/{name}/create", e.create)
	r.HandleFunc("/lambda/{name}/run", e.lambda)

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
