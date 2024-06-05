package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

type HttpTaskCoordinator struct {
	pool TaskManagerPool
	httpServer *http.Server
	done chan error
}

func NewHttpTaskCoordinator(port int, pool TaskManagerPool) *HttpTaskCoordinator {
	mux := http.NewServeMux()

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	server := &HttpTaskCoordinator{
		httpServer: srv,
		pool: pool,
		done: make(chan error),
	}


	mux.HandleFunc("GET /task/{token}", func(w http.ResponseWriter, r *http.Request) {
		taskToken := r.PathValue("token")
		
		client_token := r.Header.Get("Authorization")
		if client_token == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return 
		}

		manager, err := pool.GetTaskManager(taskToken)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		payload, err := manager.GetPayload(client_token)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", payload.MimeType)
		io.Copy(w, payload.Data)
	})
	
	mux.HandleFunc("POST /task/{token}", func(w http.ResponseWriter, r *http.Request) {
		taskToken := r.PathValue("token")
		
		client_token := r.Header.Get("Authorization")
		if client_token == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return 
		}


		manager, err := pool.GetTaskManager(taskToken)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		ctype := r.Header.Get("Content-type")

		if !manager.IsValidResponseType(ctype) {
			http.Error(w, "Invalid content type", http.StatusBadRequest)
			return
		}	

		manager.RegisterResponse(client_token, r.Body)
	})

	return server
}

func (s *HttpTaskCoordinator) Start(ctx context.Context) error {

	go func() {
		<-ctx.Done()
		timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err := s.httpServer.Shutdown(timeoutCtx)

		// check if error is context or server closed
		// if not, send error to done channel (The server failed to shutdown properly)
		if err != nil && err != http.ErrServerClosed && err != context.DeadlineExceeded {
			s.done <- err
		}

		cancel()
	}()

	go func() {
		err := s.httpServer.ListenAndServe()
		if err != nil {
			s.done <- err
		}
	}()

	return nil
}

func (s *HttpTaskCoordinator) Wait() error {
	return <-s.done
}
