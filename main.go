package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"
)

func main() {
	server := URLShortenerServer{
		store: &URLStore{
			endpoints: make(map[string]string),
			kg: KeyGenerator{
				taken: make(map[string]bool),
			},
		},
	}

	server.store.Set("google.com")
	for i, j := range server.store.endpoints {
		fmt.Println(i, j)
	}

	tcpPort := ":1337"
	listener, err := net.Listen("tcp", tcpPort)
	if err != nil {
		panic(err)
	}

	go server.acceptNetcats(listener)
	fmt.Println("listening to tcp on", tcpPort)

	mux := http.NewServeMux()
	mux.HandleFunc("/", server.ServeHTTP)

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20, //1MB
	}

	log.Printf("starting http server on %s", srv.Addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

type URLShortenerServer struct {
	store *URLStore
}

func (s *URLShortenerServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	url := s.store.Get(path.Base(r.URL.Path))
	if url != "" {
		http.Redirect(w, r, url, http.StatusFound)
		return
	}
	http.NotFound(w, r)
}

func (s *URLShortenerServer) acceptNetcats(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go s.handleNetcats(conn)
	}
}

func (s *URLShortenerServer) handleNetcats(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSuffix(input, "\n")

	if strings.TrimSpace(input) == "" {
		_, _ = conn.Write([]byte("Don't send empty spaces!"))
		return
	}

	if _, err := url.ParseRequestURI(input); err != nil {
		_, _ = conn.Write([]byte("Not a valid URL! "))
		_, _ = conn.Write([]byte(input))
		return
	}

	key := s.store.Set(input)
	_, _ = conn.Write(fmt.Appendln(nil, "key is", key))
}

type URLStore struct {
	endpoints map[string]string
	kg        KeyGenerator
	mu        sync.Mutex
}

func (store *URLStore) Set(url string) string {
	store.mu.Lock()
	defer store.mu.Unlock()

	key := store.kg.GenKey()
	store.endpoints[key] = url
	return key
}

func (store *URLStore) Get(key string) string {
	store.mu.Lock()
	defer store.mu.Unlock()

	if url, ok := store.endpoints[key]; ok {
		return url
	}
	return ""
}

type KeyGenerator struct {
	taken map[string]bool
}

const letters = "abcdefghijklmnopqrstuvwxyz123456789"

func (kg *KeyGenerator) GenKey() string {
	for {
		b := make([]byte, 6)
		for i := range b {
			b[i] = letters[rand.Intn(len(letters))]
		}

		key := string(b)
		if _, ok := kg.taken[key]; !ok {
			kg.taken[key] = true
			return key
		}
	}
}
