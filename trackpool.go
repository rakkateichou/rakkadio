package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sync"

	sc "github.com/zackradisic/soundcloud-api"
)

type TrackInfo struct {
	ID int64 `json:"id"`
	Title string `json:"title"`
	Author string `json:"author"`
	Url string `json:"url"`
}

type TrackPool struct {
	API *sc.API
	InitialSongId int64

	CurrentTrack *TrackInfo
	CurrentIndex int
	List []TrackInfo
	mu sync.Mutex
}

type TrackInfoResponse struct {
	ID int64 `json:"id"`
}

type Response struct {
	Tracks []TrackInfoResponse `json:"tracks"`
}

func (tp *TrackPool) AddTrack(ti TrackInfo) {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	tp.List = append(tp.List, ti)

	if len(tp.List) == 1 {
		tp.CurrentTrack = &tp.List[0]
	}

}

func (tp *TrackPool) NextTrack() *TrackInfo {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	if len(tp.List) == 1 {
		return tp.CurrentTrack
	}

	if tp.CurrentIndex >= len(tp.List) - 1 {
		tp.fetchNextPage()
	} else {
		tp.CurrentIndex++
	}

	log.Printf("Playing track [%d/%d]\n", tp.CurrentIndex + 1, len(tp.List))

	tp.CurrentTrack = &tp.List[tp.CurrentIndex]
	
	return tp.CurrentTrack
}

func (tp *TrackPool) clearList() {
	tp.List = []TrackInfo{}
	tp.CurrentTrack = &TrackInfo{}
	tp.CurrentIndex = 0
}

func (tp *TrackPool) GetCurrentTrack() *TrackInfo {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	return tp.CurrentTrack
}

func (tp *TrackPool) fetchNextPage() {

	log.Println("Fetching next page")

	var stationTrackID int64
	if len(tp.List) == 0 {
		stationTrackID = tp.InitialSongId
	} else {
		stationTrackID = tp.List[len(tp.List)-1].ID
	}

	var stationUrl = fmt.Sprintf("https://soundcloud.com/discover/sets/track-stations:%d", stationTrackID)
	var encodedUrl = url.QueryEscape(stationUrl)
	var resolveUrl = fmt.Sprintf("https://api-v2.soundcloud.com/resolve?url=%s&client_id=%s&app_version=1711450916&app_locale=en", encodedUrl, tp.API.ClientID())

	resp, err := http.Get(resolveUrl)
	if err != nil {
		log.Fatal("Couldn't get next songs in line")
	}
	defer resp.Body.Close()

	var response Response
	if json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Fatal("Couldn't decode")
	}

	tp.clearList()

	// TODO: make it async
	for i, track := range response.Tracks {
		trackInfo, err := api.GetTrackInfo(sc.GetTrackInfoOptions{ID: []int64{track.ID}})
		if err != nil {
			log.Fatal("Couldn't get track info")
		}

		log.Printf("Fetched [%d/%d]\n", i+1, len(response.Tracks))

		tp.List = append(tp.List, TrackInfo{
			ID: track.ID,
			Title: trackInfo[0].Title,
			Author: trackInfo[0].User.Username,
			Url: trackInfo[0].PermalinkURL,
		})
	}

	if len(tp.List) > 1 && stationTrackID != tp.InitialSongId {
		tp.CurrentIndex = 1
	}
}
