package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	_ "github.com/zackradisic/soundcloud-api"
)

type TrackInfo struct {
	Id int `json:"id"`
}

type Response struct {
	Tracks []TrackInfo `json:"tracks"`
}

const songStationUrl = "https://soundcloud.com/discover/sets/track-stations:718571584"

func GetSongsIds(stationUrl string) []string {
	var clientId = os.Getenv("CLIENT_ID")
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

	var ids []string
	for _, track := range response.Tracks {
		ids = append(ids, fmt.Sprint(track.Id))
	}

	return ids
}
