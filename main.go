// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/sndurkin/game-night-in/models"
	"github.com/sndurkin/game-night-in/api"
)

const (
	defaultPort = "3000"
)

func serveHome(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "public/index.html")
}

func remoteAddr(r *http.Request) string {
	via := r.RemoteAddr
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		return fmt.Sprintf("%v via %v", xff, r.RemoteAddr)
	}
	return via
}

func logRequest(r *http.Request) {
	log.Printf("%v %v %v %v", remoteAddr(r), r.Method, r.URL, r.Proto)
}

func logRoute(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logRequest(r)
		f(w, r)
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	api.Init()

	h := newHub()
	go h.run()
	go h.runRoomCleanup()

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	http.HandleFunc("/", logRoute(serveHome))

	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/public/", http.StripPrefix("/public/", fs))

	http.HandleFunc("/ws", logRoute(func(w http.ResponseWriter, r *http.Request) {
		serveWs(h, w, r)
	}))

	addr := fmt.Sprintf(":%s", port)
	log.Printf("Server listening on %s\n", addr)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
