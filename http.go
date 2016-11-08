package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"

	rice "github.com/GeertJohan/go.rice"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

var (
	fileBox *rice.Box
)

func httpInit() http.Handler {
	m := mux.NewRouter()
	api := m.PathPrefix("/api").Subrouter()
	m.HandleFunc("/stream/{streamID}", httpLoggingMiddleware(httpHandleStream)).Methods("GET").Name("stream")
	m.HandleFunc("/{rest:.*}", httpLoggingMiddleware(httpHandleStatic)).Methods("GET").Name("static")
	api.HandleFunc("/streams", httpLoggingMiddleware(httpJSONHeaderMiddleware(httpHandleAPIStreams(m)))).Name("api.streams")
	fileBox = rice.MustFindBox("public/dist")
	return m
}

func httpHandleStream(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	streamID, err := strconv.Atoi(vars["streamID"])
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte("Bad request"))
		return
	}
	if streamID >= len(config.Streams) {
		w.WriteHeader(400)
		w.Write([]byte("Bad request"))
		return
	}

	conn, _, err := w.(http.Hijacker).Hijack()
	if err != nil {
		logger.Warning("Cannot hijack HTTP session: ", err)
		return
	}

	pipeline := gPipelines[streamID]
	(*pipeline).AddReceiver(conn.(*net.TCPConn))
}

func httpHandleStatic(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	path := vars["rest"]
	b, err := fileBox.Bytes(path)
	if err != nil {
		w.Write(fileBox.MustBytes("index.html"))
	}
	w.Write(b)
}

func httpHandleAPIStreams(router *mux.Router) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		type stream struct {
			ID          string `json:"id"`
			Description string `json:"description"`
			MediaType   string `json:"mediaType"`
			URL         string `json:"url"`
		}
		streams := make([]stream, len(config.Streams))
		for i, v := range config.Streams {
			s := &streams[i]
			s.ID = strconv.Itoa(i)
			s.Description = v.Description
			switch v.Kind.Value {
			case configStreamKindVideoWebM:
				s.MediaType = "video"
			case configStreamKindAudioOpus:
				s.MediaType = "audio"
			case configStreamKindAudioTest1:
				s.MediaType = "audio"
			case configStreamKindVideoTest1:
				s.MediaType = "video"
			default:
				panic(fmt.Errorf("Unhandled media type: %s", v.Kind.Value.String()))
			}
			var streamURL *url.URL
			streamURL, err = router.Get("stream").URL("streamID", strconv.Itoa(i))
			if err != nil {
				panic(err)
			}
			s.URL = streamURL.String()
		}
		j := json.NewEncoder(w)
		j.Encode(streams)
	}
}

func httpJSONHeaderMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.Header().Add("Access-Control-Allow-Origin", "*")
		next.ServeHTTP(w, r)
	})
}

func httpLoggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
		route := mux.CurrentRoute(r)
		fields := log.Fields{
			"host":    r.RemoteAddr,
			"method":  r.Method,
			"handler": route.GetName(),
			"path":    r.URL.String(),
			//"status":  r.Response.StatusCode,
		}
		logger.WithFields(fields).Debug("Serving HTTP")
	})
}
