package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/Kshitij29/go-proxy/RestHandler"
)

var (
	AllConnections chan net.Conn
)

func handleTunneling(w http.ResponseWriter, r *http.Request) {
	fmt.Println("called: ", r.Host)
	if !RestHandler.Mode {
		w.WriteHeader(204)
	} else {
		dest_conn, err := net.DialTimeout("tcp", r.Host, 10*time.Second)
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		hijacker, ok := w.(http.Hijacker)
		if !ok {
			http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
			return
		}
		client_conn, _, err := hijacker.Hijack()
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
		AllConnections <- dest_conn
		AllConnections <- client_conn

		go transfer(dest_conn, client_conn)
		go transfer(client_conn, dest_conn)
	}
}

func transfer(destination io.WriteCloser, source io.ReadCloser) {
	defer destination.Close()
	defer source.Close()
	io.Copy(destination, source)
}

func handleHTTP(w http.ResponseWriter, req *http.Request) {
	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()
	copyHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}
func main() {
	server := &http.Server{
		Addr: ":8888",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			contents, _ := ioutil.ReadAll(r.Body)
			fmt.Println(r.URL.Host)
			hdr := r.Header
			for key, value := range hdr {
				fmt.Println("   ", key, ":", value)
			}
			fmt.Println(string(contents))
			if r.Method == http.MethodConnect {
				handleTunneling(w, r)
			} else {
				handleHTTP(w, r)
			}
		}),
		// Disable HTTP/2 as it does not support Hijacking
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}

	//setting default mode
	RestHandler.Mode = true
	AllConnections = make(chan net.Conn, 1000)

	//To handle the change in mode
	go func() {
		for {
			fmt.Println("This is reachable.")
			if !<-RestHandler.ModeChan {
				fmt.Println("Here I am.")
				CloseAllConnections()
				RestHandler.Mode = false
			}
		}
	}()

	//Starting handler server
	go RestHandler.StartHandler()

	//Starting main server
	log.Fatal(server.ListenAndServe())
}

func CloseAllConnections() {
	fmt.Println(len(AllConnections))
	for i := 0; i < len(AllConnections); i++ {
		connection := <-AllConnections
		(connection).Close()
	}
	RestHandler.ModeVerifyChan <- true
}
