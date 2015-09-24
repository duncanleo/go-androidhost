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
	"time"
	"io"
)

var router = mux.NewRouter()

func listAVDHandler(w http.ResponseWriter, r *http.Request) {
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
		parser.UnmarshalAVD(&avd, m)
		avd_list = append(avd_list, avd)
	}

	json, _ := json.MarshalIndent(avd_list, "", "\t")
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, string(json))
}

func listADBHandler(w http.ResponseWriter, r *http.Request) {
	config, err := config.GetConfig()
	if err != nil {
		log.Panic(err)
	}
	adb_binary := filepath.Join(config.SDKLocation, "platform-tools", "adb")
	s, _ := command.GetCommandResponse(adb_binary, "devices", "-l")

	var adbList []model.ADBDevice = make([]model.ADBDevice, 0)

	parser.UnmarshalADB(&adbList, s)

	json, _ := json.MarshalIndent(adbList, "", "\t")
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, string(json))
}

func startHandler(w http.ResponseWriter, r *http.Request) {
	config, err := config.GetConfig()
	if err != nil {
		log.Panic(err)
	}
	emu_name := r.URL.Query().Get("name")
	if len(emu_name) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	emulator_binary := filepath.Join(config.SDKLocation, "tools", "emulator")
	go command.RunCommand(emulator_binary, "-avd", emu_name)
}

func installHandler(w http.ResponseWriter, r *http.Request) {
	config, err := config.GetConfig()
	if err != nil {
		log.Panic(err)
	}
	emu_name := r.URL.Query().Get("name")
	if len(emu_name) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	//Current dir
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(dir)

	temp_file_path := filepath.Join(dir, fmt.Sprintf("%d.apk", time.Now().UnixNano()))
	temp_file, err := os.Create(temp_file_path)
	if err != nil {
		log.Panic(err)
	}
	defer temp_file.Close()

	//Write POST content to file
	_, err = io.Copy(temp_file, file)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	emulator_binary := filepath.Join(config.SDKLocation, "platform-tools", "adb")
	s, _ := command.GetCommandResponse(emulator_binary, "-s", emu_name, "install", temp_file_path)
	fmt.Println(s)
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	router.StrictSlash(true)

	router.HandleFunc("/listavd", listAVDHandler)
	router.HandleFunc("/listadb", listADBHandler)
	router.HandleFunc("/install", installHandler)
	router.HandleFunc("/start", startHandler)

	http.Handle("/", router)
	http.ListenAndServe(":8000", handlers.LoggingHandler(os.Stdout, http.DefaultServeMux))
}