package api

import (
	"encoding/json"
	"errors"
	"log"

	utils "github.com/angelthump/transcoder/utils"
	"github.com/go-resty/resty/v2"
)

type Stream struct {
	Ingest struct {
		Server   string `json:"server"`
		Url      string `json:"url"`
		Mediamtx bool   `json:"mediamtx"`
	} `json:"ingest"`
	User struct {
		UserId   string `json:"id"`
		Username string `json:"username"`
	} `json:"user"`
	Created_at string `json:"createdAt"`
}

type Transcode struct {
	Data []TranscodeData `json:"data"`
}

type TranscodeData struct {
	StreamId    string   `json:"streamId"`
	Outputs     []Output `json:"outputs"`
	DropletId   string   `json:"droplet_id"`
	Transcoding bool     `json:"transcoding"`
}

type Output struct {
	Name           string  `json:"name"`
	Variant        string  `json:"variant"`
	Bandwidth      int     `json:"bandwidth"`
	AudioBandwidth string  `json:"audio_bandwidth"`
	VideoBandwidth string  `json:"video_bandwidth"`
	Width          int     `json:"width"`
	Height         int     `json:"height"`
	FrameRate      float64 `json:"framerate"`
}

func GetStream(streamId string) *Stream {
	client := resty.New()
	resp, _ := client.R().
		SetHeader("X-Api-Key", utils.Config.StreamsAPI.AuthKey).
		Get(utils.Config.StreamsAPI.Hostname + "/streams?id=" + streamId)

	statusCode := resp.StatusCode()
	if statusCode >= 400 {
		log.Printf("Unexpected status code, got %d %s", statusCode, string(resp.Body()))
		return nil
	}

	var streams []Stream
	err := json.Unmarshal(resp.Body(), &streams)
	if err != nil {
		log.Printf("Unmarshal Error %v", err)
		return nil
	}

	if len(streams) == 0 {
		log.Printf("Could not find stream: %s", streamId)
		return nil
	}

	return &streams[0]
}

func GetTranscode(dropletId string) *TranscodeData {
	client := resty.New()
	resp, _ := client.R().
		SetHeader("X-Api-Key", utils.Config.StreamsAPI.AuthKey).
		Get(utils.Config.StreamsAPI.Hostname + "/transcodes?droplet_id=" + dropletId)

	statusCode := resp.StatusCode()
	if statusCode >= 400 {
		log.Printf("Unexpected status code, got %d %s", statusCode, string(resp.Body()))
		return nil
	}

	var transcode Transcode
	err := json.Unmarshal(resp.Body(), &transcode)
	if err != nil {
		log.Printf("Unmarshal Error %v", err)
		return nil
	}

	if len(transcode.Data) == 0 {
		log.Printf("Could not find Transcode Data for %v", dropletId)
		return nil
	}

	return &transcode.Data[0]
}

type TranscodePatch struct {
	Transcoding bool `json:"transcoding"`
}

func SetTranscode(transcodeData *TranscodeData, transcoding bool) error {
	transcodeData.Transcoding = transcoding

	client := resty.New()
	resp, _ := client.R().
		SetHeader("X-Api-Key", utils.Config.StreamsAPI.AuthKey).
		SetHeader("Content-Type", "application/json").
		SetBody(transcodeData).
		Put(utils.Config.StreamsAPI.Hostname + "/transcodes/" + transcodeData.StreamId)

	statusCode := resp.StatusCode()
	if statusCode >= 400 {
		return errors.New(string(resp.Body()))
	}

	return nil
}
