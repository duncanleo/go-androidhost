package main

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/handlers"
	"github.com/duncanleo/config"
	"github.com/duncanleo/command"
	"github.com/duncanleo/model"
	"github.com/duncanleo/parser"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"encoding/json"
)

var router = mux.NewRouter()

func handler(w http.ResponseWriter, r *http.Request) {
	config, err := config.GetConfig()
	if err != nil {
		log.Panic(err)
	}
	android_binary := filepath.Join(config.SDKLocation, "tools", "android")
	s, _ := command.GetCommandResponse(android_binary, "list", "avd")

	ms := strings.Split(s, "---------")

	avd_list := make([]model.AVD, 0)

	for _, m := range ms {
		var avd model.AVD
		parser.Unmarshal(&avd, m)
		avd_list = append(avd_list, avd)
	}

	json, _ := json.MarshalIndent(avd_list, "", "\t")
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, string(json))
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	router.StrictSlash(true)

	router.HandleFunc("/", handler)

	http.Handle("/", router)
	http.ListenAndServe(":8000", handlers.LoggingHandler(os.Stdout, http.DefaultServeMux))
}