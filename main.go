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
	"io/ioutil"
	"regexp"
)

var router = mux.NewRouter()

func listAVDHandler(w http.ResponseWriter, r *http.Request) {
	config, err := config.GetConfig()
	if err != nil {
		log.Panic(err)
	}
	android_binary := filepath.Join(config.SDKLocation, "tools", "android")
	stdout, _, _ := command.GetCommandResponse(android_binary, "list", "avd")

	ms := strings.Split(stdout.String(), "---------")

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
	stdout, _, _ := command.GetCommandResponse(adb_binary, "devices", "-l")

	var adbList []model.ADBDevice = make([]model.ADBDevice, 0)

	parser.UnmarshalADB(&adbList, stdout.String())

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

	wd, err := os.Getwd()
	if err != nil {
		log.Panic(err)
	}

	temp_file_path := filepath.Join(wd, fmt.Sprintf("%d.apk", time.Now().UnixNano()))
	temp_file, err := os.Create(temp_file_path)
	if err != nil {
		log.Panic(err)
	}
	
	defer temp_file.Close()
	defer os.Remove(temp_file_path)

	//Write POST content to file
	_, err = io.Copy(temp_file, file)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	emulator_binary := filepath.Join(config.SDKLocation, "platform-tools", "adb")
	command.RunCommand(emulator_binary, "-s", emu_name, "install", temp_file_path)

	//Run the apk
	//Get the package name using latest build tools' aapt
	build_tools_dir_path := filepath.Join(config.SDKLocation, "build-tools")
	build_tools_dir, err := ioutil.ReadDir(build_tools_dir_path)
	latest_build_tools := build_tools_dir[len(build_tools_dir) - 1]
	aapt_binary := filepath.Join(build_tools_dir_path, latest_build_tools.Name(), "aapt")
	fmt.Println("Running", aapt_binary, "on", temp_file_path)
	stdout, _, err := command.GetCommandResponse(aapt_binary, "dump", "badging", temp_file_path)
	
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	re := regexp.MustCompile("package: name='(.+?)'")
	package_string_matches := re.FindStringSubmatch(stdout.String())
	if len(package_string_matches) < 2 {
		log.Println("Could not match package name")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	package_string := package_string_matches[1]
	fmt.Println("Found package string", package_string)

	//Run the actual app
	adb_binary := filepath.Join(config.SDKLocation, "platform-tools", "adb")
	stdout, stderr, err := command.GetCommandResponse(adb_binary, "-s", emu_name, "shell", "monkey", "-p", package_string, "-c", "android.intent.category.LAUNCHER", "1")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(stdout.String())
	fmt.Println(stderr.String())
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