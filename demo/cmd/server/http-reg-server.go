package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type HttpRegistryServer struct {	
	registry Registry
	httpServer *http.Server
	done chan error
}

func NewHttpRegistryServer(reg Registry, port int) *HttpRegistryServer {
	mux := http.NewServeMux()

	srv := &http.Server{
		Addr: fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	server := &HttpRegistryServer{
		registry: reg,
		httpServer: srv,
		done: make(chan error),
	}


	mux.HandleFunc("GET /task", clinet_middleware(reg, getTask))
	mux.HandleFunc("POST /register", registerSubscriber(reg))
	mux.HandleFunc("POST /unregister", clinet_middleware(reg, unregisterSubscriber))
	mux.HandleFunc("POST /subscribe", clinet_middleware(reg, subscribe))
	mux.HandleFunc("POST /unsubscribe", clinet_middleware(reg, unsubscribe))

	return server
}

func (s *HttpRegistryServer) Start(ctx context.Context) error {
	
	go func() {
		<-ctx.Done()
		timeoutCtx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
		err := s.httpServer.Shutdown(timeoutCtx)

		// check if error is context or server closed
		// if not, send error to done channel (The server failed to shutdown properly)
		if err != nil && err != http.ErrServerClosed && err != context.DeadlineExceeded {
			s.done <- err
		}

		cancel()
	}()	

	go func () {
		err := s.httpServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			s.done <- err
		}

		s.done <- nil
	}()

	return nil
}

func (s *HttpRegistryServer) Wait() error {
	return <-s.done
}


// ---------------------------------------------
// ---- Routes
// ---------------------------------------------

type HttpRegServerRoute = func (reg Registry) http.HandlerFunc

func getClientParams(r *http.Request) (string, string) {
	token := r.Header.Get("Authorization")
	group := r.Header.Get("Group")
	return token, group
}

func clinet_middleware(reg Registry, route HttpRegServerRoute) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		token, group := getClientParams(r)
		if token == "" || group == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return 
		}

		route(reg)(w, r)
	}
}

func registerSubscriber (reg Registry) http.HandlerFunc {
 return func (w http.ResponseWriter, r *http.Request) {
		type Body struct {
			Params []Param 
		}

		_, group := getClientParams(r)
		if group == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		var body Body
		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		token, err := reg.RegisterSubscriber(group, body.Params)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		fmt.Println("Registered: ", token)


		resp := Json{"token": token}
		resp.Ok(w)
	}
}

func unregisterSubscriber (reg Registry) http.HandlerFunc {
 return func (w http.ResponseWriter, r *http.Request) {
		token, group := getClientParams(r)
		err := reg.UnregisterSubscriber(group, token)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func subscribe (reg Registry) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
			token, group := getClientParams(r)
		
			type Body struct { 
				Tasks []TaskSubscription `json:"tasks"`
			}

			var body Body
			err := json.NewDecoder(r.Body).Decode(&body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(err.Error()))
				return
			}

			if len(body.Tasks) == 0 {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("No tasks to subscribe"))
			}

			err = reg.Subscribe(group, token, body.Tasks)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}

			w.WriteHeader(http.StatusOK)
	 }
}

func unsubscribe (reg Registry) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
			token, group := getClientParams(r)
		
			type Body struct { 
				Tasks []Task `json:"tasks"`
			}

			var body Body
			err := json.NewDecoder(r.Body).Decode(&body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			if len(body.Tasks) == 0 {
				err = reg.UnsubscribeAll(group, token)
			} else {
				err = reg.Unsubscribe(group, token, body.Tasks)
			}

			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}

			w.WriteHeader(http.StatusOK)
	 }
}

func getTask (reg Registry) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
			token, group := getClientParams(r)

			task, err := reg.GetTask(group, token)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}

			if task == nil {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			err = json.NewEncoder(w).Encode(task)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
	 }
}


 
