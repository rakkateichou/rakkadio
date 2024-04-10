package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	sc "github.com/zackradisic/soundcloud-api"
)

var api *sc.API
var tracks TrackPool

func init() {
	initialSongId := os.Getenv("SONG_ID")

	if initialSongId == "" {
		log.Fatal("Env variable SONG_ID must not be empty")
	}

	if err := os.MkdirAll("assets", os.ModeDir); err != nil {
		log.Fatal("Error creating assets folder")
	}

	var err error
	api, err = sc.New(sc.APIOptions{})
	if err != nil {
		log.Fatal("Couldn't create SC API")
	}

	intInitialStringId, err := strconv.Atoi(initialSongId)
	if err != nil {
		log.Fatal("Couldn't convert song id to int")
	}
	tracks = TrackPool{API: api, InitialSongId: int64(intInitialStringId)}
}

func downloadSongById(songId int64) string  {
	trackInfo, err := api.GetTrackInfo(sc.GetTrackInfoOptions{ID: []int64{songId}})
	if err != nil {
		log.Fatal("Couldn't get track info")
	}

	var fileName = fmt.Sprintf("./assets/%s.mp3", fmt.Sprint(time.Now().Unix()))
	newFile, err := os.Create(fileName)
	defer newFile.Close()

	err = api.DownloadTrack(trackInfo[0].Media.Transcodings[0], newFile)
	if err != nil {
		log.Fatal("Couldn't download track")
	}
	return fileName
}

func GetCurrentTrackInfo() TrackInfo {
	return *tracks.CurrentTrack
}

func TopUpMusic(songEnded chan struct{}, songPath chan string){
	for {
		select {
		default:
		case <-songEnded:
			musicFiles, err := os.ReadDir("./assets")
			if err != nil {
				log.Fatal("Couldn't read assets dir")
			}

			if len(musicFiles) >= 5 {
				os.Remove("./assets/" + musicFiles[0].Name())
			}

			path := downloadSongById(tracks.NextTrack().ID)

			log.Println(fmt.Sprintf("Playing %s by %s (%s)", tracks.CurrentTrack.Title, tracks.CurrentTrack.Author, path))

			songPath <- path
		}
	}
}
