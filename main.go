package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
    "encoding/json"
    "strconv"
    "strings"
	"github.com/gorilla/mux"
)

const port = "8080";
const mapFolderPath = "files/maps/";
const deletedMapsFolderPath = "files/deleted_maps/";
const emptyMapPath = "files/empty_map.json";

var mapIsLoaded = false

type MapList struct {
    Maps []string `json: "maps"`
}

func getSavedMaps() []string {
    files, err := ioutil.ReadDir("files/maps/");
    if err != nil {
        log.Fatal(err);
    }

    fileNames := make([]string, len(files));

    for index, file := range files {
        fileNames[index] = file.Name();
    }

    return fileNames;
}

func getMapList(w http.ResponseWriter, r *http.Request) {
    mapList := &MapList{Maps: getSavedMaps()};
    mapListJson, err := json.Marshal(mapList)

    if err != nil {
        log.Println(err);
        http.Error(w, "Error getting map list", http.StatusBadRequest);
    }

    w.Header().Set("Content-Type", "application/json");
    fmt.Fprintf(w, string(mapListJson));
}

func getMap(w http.ResponseWriter, r *http.Request) {
    r.ParseForm();

    mapName := r.FormValue("map");
    mapPath := mapFolderPath + mapName;

	data, err := ioutil.ReadFile(mapPath);
	if err != nil {
        log.Println(err);
        http.Error(w, "Didn't find the map", http.StatusNotFound);
    }

	mapData := string(data)
	fmt.Fprintf(w, mapData)
}

func createMap(w http.ResponseWriter, r *http.Request) {
    r.ParseForm();

    mapName := r.FormValue("map");
    mapPath := mapFolderPath + mapName;

    _, err := os.Stat(mapPath);

    // path exists
    if err == nil {
        http.Error(w, "File already exists", http.StatusForbidden);
        log.Printf("File already exists %s\n", mapName);
    } else {
        emptyMap, err := ioutil.ReadFile(emptyMapPath)
        if err != nil {
            http.Error(w, "Error reading empty map", http.StatusNotModified);
            log.Println(err)
            return
        }

        err = ioutil.WriteFile(mapPath, emptyMap, 0644)
        if err != nil {
            http.Error(w, "Error creating new map", http.StatusNotModified);
            log.Printf("Error creating map %s: %s", mapPath, err)
            return
        }
    }
}

func saveMap(w http.ResponseWriter, r *http.Request) {
	reqBody, err := ioutil.ReadAll(r.Body)

    r.ParseForm();

    mapName := r.FormValue("map");
    mapPath := mapFolderPath + mapName;

	if err != nil {
		log.Println(err)
	} else {
		ioutil.WriteFile(mapPath, reqBody, 0644)
		// mapData := string(reqBody);
		mapIsLoaded = true
	}
}

func deleteMap(w http.ResponseWriter, r *http.Request) {
    r.ParseForm();

    mapName := r.FormValue("map");
    mapPath := mapFolderPath + mapName;

    _, err := os.Stat(mapPath);

    // path exists
    if err == nil {
        // move map instead of deleting it
        newLocation := deletedMapsFolderPath + mapName;
        os.Rename(mapPath, newLocation);
    } else {
        log.Printf("Delete error: file %s doesn't exists\n", mapName);
        http.Error(w, "File not found", http.StatusNotFound);
        return
    }
}

func renameMap(w http.ResponseWriter, r *http.Request) {
    r.ParseForm();

    mapName := r.FormValue("map");
    newMapName := r.FormValue("new_name");
    mapPath := mapFolderPath + mapName;

    // fmt.Printf("Rename request: %s -> %s\n", mapName, newMapName);
    _, err := os.Stat(mapPath);

    // path exists
    if err == nil {
        // move map instead of deleting it
        newLocation := mapFolderPath + newMapName;
        os.Rename(mapPath, newLocation);
    } else {
        log.Printf("Rename error: file %s doesn't exists\n", mapName);
        http.Error(w, "File not found", http.StatusNotFound);
        return
    }
}

func duplicateMap(w http.ResponseWriter, r *http.Request) {
    r.ParseForm();

    mapName := r.FormValue("map");
    mapPath := mapFolderPath + mapName;

    _, err := os.Stat(mapPath);

    // path exists
    if err == nil {
        originalMap, err := ioutil.ReadFile(mapPath)
        if err != nil {
            http.Error(w, "Error reading original map", http.StatusNotModified);
            log.Println("Error reading original map: " + err.Error())
            return
        }

        // search for new map name
        mapNameWithoutJson := strings.Trim(mapName, ".json");
        duplicateMapPath := ""
        duplicateMapName := ""
        for i := 1; i <= 100; i++ {
            duplicatePath := mapFolderPath + mapNameWithoutJson + "(" + strconv.Itoa(i) + ")" + ".json";
            _, err := os.Stat(duplicatePath)


            // file doesn't exist
            if err != nil {
                duplicateMapPath = duplicatePath;
                duplicateMapName = mapNameWithoutJson + "(" + strconv.Itoa(i) + ")" + ".json";
                break;
            }
        }

        if(duplicateMapPath == "") {
            http.Error(w, "Too many duplicates for this map", http.StatusNotModified);
            log.Println("Too many duplicates for this map");
            return
        }

        err = ioutil.WriteFile(duplicateMapPath, originalMap, 0644)
        if err != nil {
            http.Error(w, "Error creating duplicated map", http.StatusNotModified);
            log.Printf("Error duplicating map %s: %s", mapPath, err)
            return
        }

        // send new map name
        fmt.Fprint(w, duplicateMapName)
    } else {
        http.Error(w, "File doesn't exists", http.StatusNotFound);
        log.Printf("File doesn't exists %s\n", mapName);
    }
}

func testServer(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Recv test request\n")
	fmt.Println(w, "Server is up and running")
}

func main() {
	fmt.Printf("Running...\n") 

	// init logger
	f, err := os.OpenFile("log.txt",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	log.SetOutput(f)

	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/test", testServer).Methods("GET")
	router.HandleFunc("/getMap", getMap).Methods("GET")
	router.HandleFunc("/getMapList", getMapList).Methods("GET")
	router.HandleFunc("/createMap", createMap).Methods("POST")
	router.HandleFunc("/saveMap", saveMap).Methods("PUT")
    router.HandleFunc("/deleteMap", deleteMap).Methods("DELETE")
    router.HandleFunc("/renameMap", renameMap).Methods("PUT")
    router.HandleFunc("/duplicateMap", duplicateMap).Methods("POST")

	log.Fatal(http.ListenAndServe(":"+port, router))
}


