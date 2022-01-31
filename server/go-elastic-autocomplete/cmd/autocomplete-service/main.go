package main

import (
	"context"
	"encoding/json"
	"fmt"
	"go-elastic-autocomplete/pkg/elasticsearch"
	"io"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"
	"syscall"
	"time"

	"github.com/rs/cors"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type ElasticHandler struct {
	esClient *elasticsearch.ElasticClient
}

func main() {
	esInstance, err := elasticsearch.New()
	if err != nil {
		log.Fatalf("failed to create elastic search instance. %v", err)
	}
	eh := ElasticHandler{
		esClient: esInstance,
	}
	router := mux.NewRouter()
	router.HandleFunc("/autocomplete/search", responseHandler(eh.serve)).Methods("GET")

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
	})

	srv := http.Server{
		Addr:    fmt.Sprintf(":%s", "8001"),
		Handler: c.Handler(router),
	}

	quit := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go gracefullyShutdownServer(&srv, quit, done)
	//start http server
	log.WithFields(log.Fields{"port": "8001"}).Info("autocomplete-service server ready.")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("failed to listen:%v", err)
	}

	<-done
	log.Print("profile-api exited gracefully...")
}

func gracefullyShutdownServer(server *http.Server, quit <-chan os.Signal, done chan<- bool) {
	<-quit

	log.Info("autocomplete-service is shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	server.SetKeepAlivesEnabled(false)
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("could not gracefully shutdown autocomplete-service: %v\n", err)
	}

	close(done)
}

func responseHandler(h func(w http.ResponseWriter, r *http.Request) (interface{}, int, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//capture panic call stack in logs
		defer func() {
			if e := recover(); e != nil {
				log.WithFields(
					log.Fields{
						"stack": fmt.Sprintf("%s", debug.Stack())}).Errorf("panic %s", e)
				http.Error(w, fmt.Sprintln(http.StatusBadRequest, "panic - internal server error"), http.StatusBadRequest)
			}

		}()
		data, status, err := h(w, r)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		if err != nil {
			data = fmt.Sprintln(status, err.Error())
		}
		io.WriteString(w, data.(string))
	}
}
func (eh *ElasticHandler) serve(w http.ResponseWriter, r *http.Request) (interface{}, int, error) {
	indexName := r.URL.Query().Get("index_name")
	text := r.URL.Query().Get("text")
	searchBy := r.URL.Query().Get("searchBy")
	searchType := r.URL.Query().Get("searchType")
	fields := strings.Split(r.URL.Query().Get("fields"), ",")
	var response []map[string]interface{}
	if indexName == "" || text == "" || searchBy == "" || searchType == "" {
		response = []map[string]interface{}{
			{"output": "index_name, text, searchBy, searchType are required"},
		}
	} else {
		response = eh.esClient.Search(indexName, text, searchBy, searchType, fields)
	}
	j, _ := json.Marshal(response)

	return string(j), http.StatusOK, nil
}
