package parser

import (
	"github.com/duncanleo/model"
	"regexp"
	"strings"
)

func UnmarshalADB(adb *[]model.ADBDevice, toparse string) {
	toparse = strings.Replace(toparse, "List of devices attached", "", -1)
	first := regexp.MustCompile("(.+?)\\s+(.+?) (.+)")
	first_matches := first.FindAllStringSubmatch(toparse, -1)
	for _, m := range first_matches {
		device := model.ADBDevice{
			ID: m[1],
			Status: m[2],
			Properties: make(map[string]string),
		}
		second := regexp.MustCompile("(?:(\\w+):(\\w+))+")
		second_matches := second.FindAllStringSubmatch(m[3], -1)
		for _, prop := range second_matches {
			device.Properties[prop[1]] = prop[2]
		}
		*adb = append(*adb, device)
	}
}