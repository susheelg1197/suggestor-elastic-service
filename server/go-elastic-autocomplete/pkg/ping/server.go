package ping

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	es "github.com/elastic/go-elasticsearch/v7"
	log "github.com/sirupsen/logrus"
)

type server struct {
	name     string
	port     string
	srv      *http.Server
	esClient *es.Client
}

func NewServer(name, port string, esClient *es.Client) *server {
	return &server{
		name:     name,
		port:     ":" + port,
		srv:      &http.Server{Addr: ":" + port},
		esClient: esClient,
	}
}

type PingResponse struct {
	Status string `json:"status"`
}

func (s *server) Start() {

	http.HandleFunc(fmt.Sprintf("/%s/v1/ping", s.name), func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(PingResponse{"success"})
	})

	http.HandleFunc(fmt.Sprintf("/%s/v1/search", s.name), func(w http.ResponseWriter, r *http.Request) {
		if r := recover(); r != nil {
			responseData, _ := json.Marshal(r)
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-Type", "application/json")
			w.Write(responseData)
		}
		switch r.Method {
		case "POST":
			decoder := json.NewDecoder(r.Body)
			var reqBody map[string]interface{}
			if err := decoder.Decode(&reqBody); err != nil {
				fmt.Fprintf(w, `{"status":"error","message":%q}`, err.Error())
				panic(err)
			}
			fmt.Println(reqBody, r.URL.Query().Get("index_name"))
			defer r.Body.Close()

		}
	})
	go func() {
		log.Infof("ping server started.")
		// returns ErrServerClosed on graceful close
		if err := s.srv.ListenAndServe(); err != nil {
			log.Errorf("ping server closed. err:%v", err)
			return
		}
	}()
}

func (s *server) Stop() {
	log.Infof("ping server stop initiated.")
	if err := s.srv.Shutdown(context.TODO()); err != nil {
		log.Errorf("Failed to shutdown ping server. err:%v", err)
		return
	}
}
