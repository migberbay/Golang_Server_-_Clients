package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Jeffail/gabs"
)

type File struct {
	path         string
	size         float64
	lastModified time.Time
}

// This is a neccesary struct for the diff package to work.
// type Change struct {
// 	Type string      // The type of change detected; can be one of create, update or delete
// 	Path []string    // The path of the detected change; will contain any field name or array index that was part of the traversal
// 	From interface{} // The original value that was present in the "from" structure
// 	To   interface{} // The new value that was detected as a change in the "to" structure
// }

var userfiles = map[int][]File{}
var userfolders = map[int][]string{}
var isInitClient bool = false

var serverfiles = []File{}
var serverfolders = []string{}
var isInitServer bool = false

func CompareAndUpdateClientFiles(conn user_conn, clientFiles string, verbose bool) {
	clientFiles = strings.ReplaceAll(clientFiles, "\\\"", "\"")
	clientFiles = strings.ReplaceAll(clientFiles, "\\r", "")
	clientFiles = strings.TrimPrefix(clientFiles, "\"")
	clientFiles = strings.TrimSuffix(clientFiles, "\"")

	if !isInitClient {
		userfiles = make(map[int][]File)
		userfolders = make(map[int][]string)
		isInitClient = true
	}
	userfiles[conn.user.ID] = make([]File, 0)
	userfolders[conn.user.ID] = make([]string, 0)

	clientFilesJson, _ := gabs.ParseJSON([]byte(clientFiles))
	cMap, _ := clientFilesJson.S("TTData").ChildrenMap()
	for key, child := range cMap {
		ExtractFilesFromDir("TTData", key, child, conn.user.ID)
	}

	if verbose {
		fmt.Print("Client sent the following files.\n")
		for _, f := range userfiles[conn.user.ID] {
			fmt.Printf("path: %v, size: %v, lastMod: %v\n", f.path, f.size, f.lastModified)
		}

		fmt.Print("Client sent the following folders.\n")
		for _, s := range userfolders[conn.user.ID] {
			fmt.Printf("path: %v\n", s)
		}
	}

	if !isInitServer {
		serverfiles = make([]File, 0)
		serverfolders = make([]string, 0)
		isInitClient = true

		var err = filepath.Walk("./TTData", func(path string, info os.FileInfo, err error) error {
			if !info.IsDir() {
				f := File{
					path:         path,
					size:         float64(info.Size()),
					lastModified: info.ModTime().UTC()}
				serverfiles = append(serverfiles, f)

				// scene_data, _ := os.ReadFile(path) // You can now access the data like this info["username"]
			} else {
				if path != "./TTData" {
					serverfolders = append(serverfolders, path)
				}
			}
			return nil
		})
		if err != nil {
			log.Fatal(err)
		}
	}

	if verbose {
		fmt.Print("Server has the following files.\n")
		for _, f := range serverfiles {
			fmt.Printf("path: %v, size: %v, lastMod: %v\n", f.path, f.size, f.lastModified)
		}

		fmt.Print("Server has the following folders.\n")
		for _, s := range serverfolders {
			fmt.Printf("path: %v\n", s)
		}
	}

	foldersToDelete := difference(userfolders[conn.user.ID], serverfolders)
	foldersToCreate := difference(serverfolders, userfolders[conn.user.ID])

	filesToOvewrite := make([]File, 0)
	filesToCreate := make([]File, 0)

	for _, serverfile := range serverfiles {
		existInClient := false
		samefile := false

		for i, clientFile := range userfiles[conn.user.ID] {
			existInClient, samefile = filesEqual(clientFile, serverfile)
			if existInClient {
				// if the file exists on the client we remove from the iterating list cause it won't appear again.
				userfiles[conn.user.ID] = userfiles[conn.user.ID][:i+copy(userfiles[conn.user.ID][i:], userfiles[conn.user.ID][i+1:])]
				break
			}
		}

		if existInClient {
			// update if same
			if samefile {
				filesToOvewrite = append(filesToOvewrite, serverfile)
			}
		} else {
			// create
			filesToCreate = append(filesToCreate, serverfile)
		}
	}

	// Sending file update information:

	current_connection := conn.connection
	infostring := fmt.Sprintf("%v,%v,%v,%v,%v", strconv.Itoa(len(foldersToCreate)),
		strconv.Itoa(len(foldersToDelete)),
		strconv.Itoa(len(filesToCreate)),
		strconv.Itoa(len(filesToOvewrite)),
		strconv.Itoa(len(userfiles[conn.user.ID])))

	SendMessageToClient(current_connection, []byte("005:info;"+infostring))

	fmt.Print("\nFolders to create\n")
	for _, s := range foldersToCreate {
		fmt.Printf("path: %v\n", s)
		time.Sleep(10 * time.Millisecond)
		SendMessageToClient(current_connection, []byte("005:create;folder;"+s))
	}

	fmt.Print("\nFolders to delete\n")
	for _, s := range foldersToDelete {
		fmt.Printf("path: %v\n", s)
		time.Sleep(10 * time.Millisecond)
		SendMessageToClient(current_connection, []byte("005:delete;folder;"+s))
	}

	fmt.Print("\nFiles to create\n")
	for _, s := range filesToCreate {
		fmt.Printf("path: %v\n", s.path)
		dat, err := os.ReadFile(s.path)
		errCheck(err)
		time.Sleep(10 * time.Millisecond)
		SendMessageToClient(current_connection, []byte("005:create;file;"+s.path+";"+s.lastModified.Format(time.RFC3339)+";"+string(dat)))
	}

	fmt.Print("\nFiles to ovewrite\n")
	for _, s := range filesToOvewrite {
		fmt.Printf("path: %v\n", s.path)
		dat, err := os.ReadFile(s.path)
		errCheck(err)
		time.Sleep(10 * time.Millisecond)
		SendMessageToClient(current_connection, []byte("005:update;file;"+s.path+";"+s.lastModified.Format(time.RFC3339)+";"+string(dat)))
	}

	fmt.Print("\nFiles to delete\n")
	for _, s := range userfiles[conn.user.ID] {
		fmt.Printf("path: %v\n", s.path)
		time.Sleep(10 * time.Millisecond)
		SendMessageToClient(current_connection, []byte("005:delete;file;"+s.path))
	}

}

func ExtractFilesFromDir(filepath string, key string, child *gabs.Container, userID int) {
	if key == "files" {
		children, _ := child.Children()
		for _, child := range children {
			p := child.S("name").Data().(string)
			fpath := "TTData" + strings.Split(p, "\\TTData")[1]

			f := File{
				path:         fpath,
				size:         child.S("size").Data().(float64),
				lastModified: StringDateToTime(child.S("lastModified").Data().(string))}

			userfiles[userID] = append(userfiles[userID], f)
		}
	} else {
		f := filepath + "\\" + strings.Replace(key, "folder:", "", 1)
		userfolders[userID] = append(userfolders[userID], f)
		cMap, _ := child.ChildrenMap()
		for k, c := range cMap {
			ExtractFilesFromDir(f, k, c, userID)
		}
	}
}

// returns true if file a is equal enough to b (time variation of 3 seconds between save times is accepted to account for connection delay.)
func filesEqual(a, b File) (bool, bool) {
	pathequal := false
	if a.path != b.path {
		return false, false
	} else {
		pathequal = true
	}

	cond1 := inTimeSpan(b.lastModified.Add(time.Second*-3), b.lastModified.Add(time.Second*3), a.lastModified)
	cond2 := a.size == b.size

	return pathequal, cond1 && cond2
}
