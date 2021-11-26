package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

var filename = "config.json"

// Data structures from the config file.
type Config struct {
	Port   string `json:"port"`
	Users  []User `json:"users"`
	Worlds []World
	Scenes []Scene
}

type User struct {
	ID       int    `json:"id"`
	Type     string `json:"type"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type World struct {
	ID      int
	Name    string
	Owner   int
	Players []int
}

type Scene struct {
	ID       int
	Name     string
	world    int
	filepath string
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

	fmt.Printf("%+v\n", config)
	WalkFolders()
	os.Exit(0)

	return config
}

func WalkFolders() {
	visit := func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			fmt.Println("dir:  ", path)
		} else {
			fmt.Println("file: ", path)
		}
		return nil
	}

	err := filepath.Walk("./", visit)
	if err != nil {
		log.Fatal(err)
	}
}
