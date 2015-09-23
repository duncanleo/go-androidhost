package model

type AVD struct {
	Name string `json:"name"`
	Device string `json:"device"`
	Manufacturer string `json:"manufacturer"`
	Target string `json:"target"`
	Tag string `json:"tag"`
	Arch string `json:"arch"`
}