package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	sc "github.com/zackradisic/soundcloud-api"
)

type TrackInfo struct {
	Id int64 `json:"id"`
}

type Response struct {
	Tracks []TrackInfo `json:"tracks"`
}

const songStationUrl = "https://soundcloud.com/discover/sets/track-stations:78531527"

var clientId string
var api *sc.API

func init() {
	clientId = os.Getenv("CLIENT_ID")

	var err error
	api, err = sc.New(sc.APIOptions{ClientID: clientId})
	if err != nil {
		log.Fatal("Couldn't create SC API")
	}
}

func getSongsIds(stationUrl string) []int64 {
	var encodedUrl = url.QueryEscape(stationUrl)
	var resolveUrl = fmt.Sprintf("https://api-v2.soundcloud.com/resolve?url=%s&client_id=%s&app_version=1711450916&app_locale=en", encodedUrl, clientId)

	resp, err := http.Get(resolveUrl)
	if err != nil {
		log.Fatal("Couldn't get next songs in line")
	}
	defer resp.Body.Close()

	var response Response
	if json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Fatal("Couldn't decode")
	}

	var ids []int64
	for _, track := range response.Tracks {
		ids = append(ids, track.Id)
	}

	return ids
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

func getStationUrlById(id int64) string {
	return fmt.Sprintf("https://soundcloud.com/discover/sets/track-stations:%d", id)
}

func TopUpMusic(songEnded chan struct{}, songName chan string)  {

	nextSongIds := getSongsIds(songStationUrl)
	currentId := 0

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

			nextSongPath := downloadSongById(nextSongIds[currentId])
			currentId++

			songName <- nextSongPath

			// soundcloud gives 40 tracks in a station
			if currentId % 39 == 0 {
				stationUrl := getStationUrlById(nextSongIds[len(nextSongIds)-1])
				nextSongIds = getSongsIds(stationUrl)
				currentId = 1 // do not repeat the same song twice
			}

		}
	}
}
