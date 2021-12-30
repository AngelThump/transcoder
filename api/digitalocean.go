package api

import (
	"log"

	utils "github.com/angelthump/transcoder/utils"
	"github.com/go-resty/resty/v2"
)

func GetDropletId() string {
	client := resty.New()
	resp, _ := client.R().
		Get(utils.Config.DigitalOcean.Metadata.Hostname + "/metadata/v1/id")

	statusCode := resp.StatusCode()
	if statusCode >= 400 {
		log.Printf("Unexpected status code, got %d %s", statusCode, string(resp.Body()))
		return ""
	}

	return string(resp.Body())
}
