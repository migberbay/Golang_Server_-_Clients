package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/Jeffail/gabs"
)

//https://www.golangprograms.com/dynamic-json-parser-without-struct-in-golang.html --> gabs reference

var filename = "config.json"

// Data structures from the config file.
type Config struct {
	Port   string `json:"port"`
	Users  []User `json:"users"`
	Worlds []World
}

type User struct {
	ID       int    `json:"id"`
	Type     string `json:"type"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type World struct {
	ID      int
	System  string
	Name    string
	Owner   int
	Players []int
	Scenes  []Scene
}

type Scene struct {
	ID       int
	Name     string
	Filepath string
}

// Writes file
func WriteFile() {
	// Read Write Mode
	file, err := os.OpenFile(filename, os.O_RDWR, 0644)

	if err != nil {
		log.Fatalf("failed opening file: %s", err)
	}
	defer file.Close()

	len, err := file.WriteAt([]byte{'G'}, 0) // Write at 0 beginning
	if err != nil {
		log.Fatalf("failed writing to file: %s", err)
	}
	fmt.Printf("\nLength: %d bytes", len)
	fmt.Printf("\nFile Name: %s", file.Name())
}

// Reads Config file loads data into structs
func LoadConfig() Config {
	data, err := ioutil.ReadFile(filename)

	if err != nil {
		log.Panicf("failed reading data from cofiguration file: %s", err)
	}

	var config Config
	json.Unmarshal(data, &config)
	config.Worlds = GetWorlds()

	fmt.Printf("%+v\n", config)

	return config
}

func GetWorlds() []World {
	worlds := make([]World, 0)
	wc := 1
	err := filepath.Walk("./Worlds", func(path string, info os.FileInfo, err error) error {
		path_parts := strings.Split(path, "\\")
		// last := path_parts[len(path_parts)]

		if info.IsDir() && info.Name() != "Worlds" && len(path_parts) <= 2 { //accesing the info.json files in world folder.
			data, _ := os.ReadFile(path + "/info.json") // You can now access the data like this info["username"]

			world_info, err := gabs.ParseJSON(data)
			if err != nil {
				panic(err)
			}

			// array iteration
			a, _ := world_info.S("players").Children()
			players := make([]int, 0)
			for _, player_id := range a {
				players = append(players, int(player_id.Data().(float64)))
			}

			scenes := make([]Scene, 0)

			err = filepath.Walk(path+"/scenes", func(path_ string, info_ os.FileInfo, err error) error {
				if !info_.IsDir() {
					scene_data, _ := os.ReadFile(path_) // You can now access the data like this info["username"]
					scene_info, _ := gabs.ParseJSON(scene_data)

					s := Scene{
						ID:       int(scene_info.S("id").Data().(float64)),
						Name:     info_.Name(),
						Filepath: path_,
					}
					scenes = append(scenes, s)
				}
				return nil
			})
			if err != nil {
				log.Fatal(err)
			}

			w := World{
				ID:      wc,
				System:  world_info.S("system").Data().(string),
				Name:    info.Name(),
				Owner:   int(world_info.S("owner").Data().(float64)),
				Players: players,
				Scenes:  scenes,
			}

			worlds = append(worlds, w)
			wc++
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	return worlds
}
