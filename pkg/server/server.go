package server

import (
	"bytes"
	"fmt"
	"github.com/flant/promicher/pkg/promicher"
	"github.com/romana/rlog"
	"io/ioutil"
	"net/http"
)

type Server struct {
	ListenHost     string
	DestinationURL string
	Promicher      *promicher.Promicher
}

func NewServer(listenHost, destinationURL string, promicher *promicher.Promicher) *Server {
	return &Server{
		Promicher:      promicher,
		ListenHost:     listenHost,
		DestinationURL: destinationURL,
	}
}

func (server *Server) Run() error {
	http.HandleFunc("/healthz", server.HandleHealth)
	http.HandleFunc("/api/v1/alerts", server.HandleAlerts)
	return http.ListenAndServe(server.ListenHost, nil)
}

func (server *Server) HandleHealth(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (server *Server) HandleAlerts(w http.ResponseWriter, r *http.Request) {
	dataBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Promicher internal server error: cannot read request data: %s", err)))
		return
	}

	rlog.Debugf("Received request %s, body:\n%s", r.URL.Path, dataBytes)

	client := &http.Client{}

	proxyRequest, err := http.NewRequest(r.Method, server.DestinationURL, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Promicher internal server error: %s", err)))
		return
	}
	for k, v := range r.Header {
		proxyRequest.Header[k] = v
	}

	newDataBytes, err := server.Promicher.ProcessData(dataBytes)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Promicher internal server error: cannot enrich request data: %s", err)))
		return
	}

	rlog.Debugf("Request %s, enriched body:\n%s", r.URL.Path, newDataBytes)
	proxyRequest.Body = ioutil.NopCloser(bytes.NewReader(newDataBytes))

	rlog.Debugf("Proxying to %s", server.DestinationURL)

	response, err := client.Do(proxyRequest)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Promicher internal server error: %s", err)))
		return
	}

	dataBytes, err = ioutil.ReadAll(response.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Promicher internal server error: error reading response from proxy destination %s: %s", server.DestinationURL, err)))
		return
	}

	rlog.Debugf("Received response from %s: %s\n%s", server.DestinationURL, response.Status, string(dataBytes))

	for k, v := range response.Header {
		w.Header()[k] = v
	}
	w.WriteHeader(response.StatusCode)
	w.Write(dataBytes)
}
