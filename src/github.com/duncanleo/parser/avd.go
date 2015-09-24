package parser

import (
	"github.com/duncanleo/model"
	"regexp"
	"strings"
)

func UnmarshalAVD(avd *model.AVD, toparse string) {
	toparse = strings.Replace(toparse, "Available Android Virtual Devices:", "", -1)
	re := regexp.MustCompile("\\s+(.*): (.*)")
	matches := re.FindAllStringSubmatch(toparse, -1)
	for _, m := range matches {
		if m[1] == "Name" {
			avd.Name = m[2]
		} else if m[1] == "Device" {
			r := regexp.MustCompile("(.+) \\((.+)\\)")
			mm := r.FindStringSubmatch(m[2])
			avd.Device = mm[1]
			avd.Manufacturer = mm[2]
		} else if m[1] == "Target" {
			avd.Target = m[2]
		} else if m[1] == "Tag/ABI" {
			parts := strings.Split(m[2], "/")
			avd.Tag = parts[0]
			avd.Arch = parts[1]
		}
	}
}