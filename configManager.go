package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

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
	wids := 1

	AllWorldsInfo, _ := ioutil.ReadDir("./TTData/Worlds")

	for _, f := range AllWorldsInfo {
		if f.IsDir() {
			worldinfo, _ := ioutil.ReadDir("./TTData/Worlds/" + f.Name())
			for _, w := range worldinfo {
				//SINGLE WORLD INFO RECOPILATION
				id := wids
				name := f.Name()
				system := ""
				owner := 0
				players := make([]int, 0)
				scenes := make([]Scene, 0)

				if w.IsDir() && w.Name() == "scenes" { // Scene directory
					scenespath := "./TTData/Worlds/" + f.Name() + "/scenes"
					scenesinfo, _ := ioutil.ReadDir(scenespath)
					for _, s := range scenesinfo {
						scene_data, _ := os.ReadFile(scenespath + "/" + s.Name()) // You can now access the data like this info["username"]
						scene_info, _ := gabs.ParseJSON(scene_data)

						s := Scene{
							ID:       int(scene_info.S("id").Data().(float64)),
							Name:     s.Name(),
							Filepath: scenespath + "/" + s.Name(),
						}
						scenes = append(scenes, s)
					}
				}

				if !w.IsDir() && w.Name() == "info.json" { // info file
					data, _ := os.ReadFile("./TTData/Worlds/" + f.Name() + "/info.json") // You can now access the data like this info["username"]

					world_info, err := gabs.ParseJSON(data)
					errCheck(err)

					// array iteration
					a, _ := world_info.S("players").Children()
					for _, player_id := range a {
						players = append(players, int(player_id.Data().(float64)))
					}
					system = world_info.S("system").Data().(string)
					owner = int(world_info.S("owner").Data().(float64))

				}

				w := World{
					ID:      id,     //wc,
					System:  system, //world_info.S("system").Data().(string),
					Name:    name,   //info.Name(),
					Owner:   owner,  //int(world_info.S("owner").Data().(float64)),
					Players: players,
					Scenes:  scenes,
				}

				worlds = append(worlds, w)

				wids++
			}
		}
	}

	return worlds
}
