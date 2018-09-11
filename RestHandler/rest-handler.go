package RestHandler

import (
	"fmt"
	"log"
	"net/http"
)

var (
	Mode           bool
	ModeChan       chan bool
	ModeVerifyChan chan bool
	VisitedLinks   map[string]bool
	LastLink       bool
)

func StartHandler() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/reward", rewardHandler)
	ModeChan = make(chan bool, 1)
	ModeVerifyChan = make(chan bool, 1)

	http.ListenAndServe(":8889", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {

	modes, ok := r.URL.Query()["mode"]

	if !ok || len(modes[0]) < 1 {
		log.Println("Url Param 'mode' is missing")
		return
	}

	// Query()["mode"] will return an array of items,
	// we only want the single item.
	mode := modes[0]

	log.Println("Url Param 'mode' is: " + string(mode))
	if mode == "allow" {
		Mode = true
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Mode changed to: \"%s\" \n", "allow")
	} else if mode == "deny" {
		Mode = false
		ModeChan <- false
		<-ModeVerifyChan
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Mode changed to: \"%s\" \n", "deny")
	} else {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Mode not changed. Only supported modes are:\n1. allow - allow all requests\n2. deny - deny all requests\n")
	}

}

func rewardHandler(w http.ResponseWriter, r *http.Request) {

}
