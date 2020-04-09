package main

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/crossworth/cfg"
)

func main() {
	example := struct {
		Name         string   `cfg:"name"`
		Host         string   `cfg:"host"`
		Port         int      `cfg:"port"`
		Players      int      `cfg:"players"`
		Announce     bool     `cfg:"announce"`
		GameMode     string   `cfg:"gamemode"`
		WebSite      string   `cfg:"website"`
		Language     string   `cfg:"language"`
		Description  string   `cfg:"description"`
		Debug        bool     `cfg:"debug"`
		UseEarlyAuth bool     `cfg:"useEarlyAuth"`
		EarlyAuthURL string   `cfg:"earlyAuthUrl"`
		UseCDN       bool     `cfg:"useCdn"`
		CDNUrl       string   `cfg:"cdnUrl"`
		Modules      []string `cfg:"modules"`
		Resources    []string `cfg:"resources"`
		Tags         []string `cfg:"tags"`
		Voice        struct {
			BitRate            int    `cfg:"bitrate"`
			ExternalHost       string `cfg:"externalHost"`
			ExternalPort       int    `cfg:"externalPort"`
			ExternalPublicHost string `cfg:"externalPublicHost"`
			ExternalPublicPort int    `cfg:"externalPublicPort"`
		} `cfg:"voice"`
	}{}

	data, err := ioutil.ReadFile("./example/example.cfg")
	if err != nil {
		log.Fatal(err)
	}

	err = cfg.Unmarshal(data, &example)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%+v", example)
}
