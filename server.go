package main

import "net/http"

type Server struct {
	http.Handler
}

func NewServer() *Server {
	server := new(Server)

	router := http.NewServeMux()
	router.Handle("/coffee", http.HandlerFunc(server.getCoffee))

	server.Handler = router

	return server
}

func (s *Server) getCoffee(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusTeapot)
}
