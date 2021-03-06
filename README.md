### CFG Parser

A CFG parser compatible with **altV**.


Example CFG file.

```
name: "TestServer",
host: "0.0.0.0",
port: 7788,
players: 1024,
#password: "verysecurepassword", # remove hashtag before password to enable
announce: false, # set to false during development
#token: no-token, # only needed when announce: true
gamemode: "Freeroam",
website: "test.com",
language: "en",
description: "test",
debug: false, # set to true during development
useEarlyAuth: true,
earlyAuthUrl: 'https://login.example.com:PORT',
useCdn: true,
cdnUrl: 'https://cdn.example.com:PORT',
modules: [
  "node-module",
  "csharp-module"
],
resources: [
  "example"
],
tags: [ 
  "customTag1",
  "customTag2",
  "customTag3",
  "customTag4"
],
voice: {
  bitrate: 64000
  #externalSecret: 3499211612
  externalHost: localhost
  externalPort: 7798
  externalPublicHost: 94.19.213.159
  externalPublicPort: 7799
}
```


#### How to use this library
```go
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
```
