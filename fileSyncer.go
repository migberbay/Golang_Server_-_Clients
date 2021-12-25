package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/Jeffail/gabs"
)

type File struct {
	ID       int
	Name     string
	Filepath string
}

func CompareAndUpdateClientFiles(clientFiles string) {

	clientFilesJson, _ := gabs.ParseJSON([]byte(clientFiles))

	cMap, _ := clientFilesJson.S("TTData").ChildrenMap()

	for key, child := range cMap {
		fmt.Printf("key: %v, value: %v\n", key, child.Data().(float64))
	}

	var err = filepath.Walk("./TTData/Worlds", func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			// scene_data, _ := os.ReadFile(path) // You can now access the data like this info["username"]
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

}
