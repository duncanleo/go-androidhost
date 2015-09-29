package main

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/handlers"
	"github.com/duncanleo/config"
	"github.com/duncanleo/command"
	"github.com/duncanleo/model"
	"github.com/duncanleo/parser"
	"github.com/duncanleo/randomstring"
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
	"net"
)

var router = mux.NewRouter()

var installJobChannel chan InstallJob = make(chan InstallJob)

type InstallJob struct {
	TempFilePath string
	EmuName string
}

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

	temp_file_path := filepath.Join(wd, fmt.Sprintf("%d-%s.apk", time.Now().UnixNano(), randomstring.RandSeq(10)))
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

	installJobChannel <- InstallJob{TempFilePath: temp_file_path, EmuName: emu_name}
}

func installJobWorker() {
	config, err := config.GetConfig()
	if err != nil {
		log.Panic(err)
	}

	for job := range installJobChannel {
		emulator_binary := filepath.Join(config.SDKLocation, "platform-tools", "adb")
		command.RunCommand(emulator_binary, "-s", job.EmuName, "install", "-r", job.TempFilePath)
		// fmt.Println(stdout.String())

		//Run the apk
		//Get the package name using latest build tools' aapt
		build_tools_dir_path := filepath.Join(config.SDKLocation, "build-tools")
		build_tools_dir, err := ioutil.ReadDir(build_tools_dir_path)
		latest_build_tools := build_tools_dir[len(build_tools_dir) - 1]
		aapt_binary := filepath.Join(build_tools_dir_path, latest_build_tools.Name(), "aapt")
		// fmt.Println("Running", aapt_binary, "on", temp_file_path)
		stdout, _, err := command.GetCommandResponse(aapt_binary, "dump", "badging", job.TempFilePath)
		
		if err != nil {
			log.Println(err)
			return
		}

		package_string_regex := regexp.MustCompile("package: name='(.+?)'")
		package_string_matches := package_string_regex.FindStringSubmatch(stdout.String())
		if len(package_string_matches) < 2 {
			log.Println("Could not match package name")
			return
		}
		package_string := package_string_matches[1]

		activity_string_regex := regexp.MustCompile("launchable-activity: name='(.+?)'")
		activity_string_matches := activity_string_regex.FindStringSubmatch(stdout.String())
		if len(package_string_matches) < 2 {
			log.Println("Could not match activity name")
			return
		}
		activity_string := activity_string_matches[1]

		// fmt.Println("Found package string", package_string, "-", activity_string)

		//Run the actual app
		adb_binary := filepath.Join(config.SDKLocation, "platform-tools", "adb")
		stdout, stderr, err := command.GetCommandResponse(adb_binary, "-s", job.EmuName, "shell", "am", "start", fmt.Sprintf("%s/%s", package_string, activity_string))
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(stdout.String())
		fmt.Println(stderr.String())

		os.Remove(job.TempFilePath)
	}
}

//Start discovery service for clients
func startDiscoveryService() {
	socket, err := net.ListenUDP("udp4", &net.UDPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: 8001,
	})
	if err != nil {
		log.Panic(err)
	}
	for {
		data := make([]byte, 4096)
		readCount, remoteAddr, err := socket.ReadFromUDP(data)
		if err == nil {
			fmt.Println("Read", string(data), "length", readCount, "from ip", remoteAddr)
		}
		client_socket, err := net.DialUDP("udp4", nil, remoteAddr)
		if err != nil {
			log.Println(err)
			continue
		}
		client_socket.Write([]byte("ANDROID_HOST_SERVER_RECOGNISED"))
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	router.StrictSlash(true)

	//Start one install job worker
	go installJobWorker()

	//Start discovery service
	go startDiscoveryService()

	router.HandleFunc("/listavd", listAVDHandler)
	router.HandleFunc("/listadb", listADBHandler)
	router.HandleFunc("/install", installHandler)
	router.HandleFunc("/start", startHandler)

	http.Handle("/", router)
	http.ListenAndServe(":8000", handlers.LoggingHandler(os.Stdout, http.DefaultServeMux))
}