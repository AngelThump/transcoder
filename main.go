package main

import (
	"log"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"

	api "github.com/angelthump/transcoder/api"
	utils "github.com/angelthump/transcoder/utils"
)

func main() {
	cfgPath, err := utils.ParseFlags()
	if err != nil {
		log.Fatal(err)
	}
	err = utils.NewConfig(cfgPath)
	if err != nil {
		log.Fatal(err)
	}

	var mainWg sync.WaitGroup
	mainWg.Add(1)
	dropletId := api.GetDropletId()
	if dropletId == "" {
		log.Fatal("Droplet Id could not be found..")
	}
	check(dropletId)
	mainWg.Wait()
}

func check(dropletId string) {
	transcodeData := api.GetTranscode(dropletId)
	if transcodeData == nil {
		time.AfterFunc(5*time.Second, func() {
			check(dropletId)
		})
		return
	}

	stream := api.GetStream(transcodeData.StreamId)
	if stream == nil {
		time.AfterFunc(5*time.Second, func() {
			check(dropletId)
		})
		return
	}

	if err := api.SetTranscode(transcodeData, true); err != nil {
		log.Printf("Something went wrong trying to patch transcode %s", err)
		time.AfterFunc(5*time.Second, func() {
			check(dropletId)
		})
		return
	}

	var wg sync.WaitGroup
	for _, output := range transcodeData.Outputs {
		wg.Add(1)
		go startTranscode(stream, output)
	}
	wg.Wait()

	if err := api.SetTranscode(transcodeData, false); err != nil {
		log.Printf("Something went wrong trying to patch transcode %s", err)
	}

	time.AfterFunc(5*time.Second, func() {
		check(dropletId)
	})
}

func startTranscode(stream *api.Stream, output api.Output) {
	log.Printf("[%s] Executing ffmpeg: %s", stream.User.Username, output.Variant)
	var cmd *exec.Cmd

	input := "rtmp://" + stream.Ingest.Server + ".angelthump.com/live/" + stream.User.StreamKey + "?key=" + utils.Config.Ingest.AuthKey

	if output.Variant == "src" {
		cmd = exec.Command("ffmpeg", "-hide_banner", "-rtmp_enhanced_codecs", "hvc1,av01", "-i", input,
			"-max_muxing_queue_size", "9999", "-c", "copy",
			"-hls_flags", "append_list+omit_endlist+program_date_time", "-hls_list_size", "6", "-hls_time", "2",
			"-hls_segment_type", "fmp4", "-hls_fmp4_init_filename", "init.mp4", "-hls_fmp4_init_resend", "1",
			"-ignore_io_errors", "1", "-method", "POST", "-headers", "Authorization: Bearer "+utils.Config.Ingest.AuthKey, "-f", "hls",
			"-hls_segment_filename", utils.Config.Cache.Hostname+"/live/"+stream.User.StreamKey+"_"+output.Variant+"/%d.m4s", utils.Config.Cache.Hostname+"/transcode/"+stream.User.StreamKey+"_"+output.Variant+"/index.m3u8")
	} else {
		cmd = exec.Command("ffmpeg", "-hide_banner", "-rtmp_enhanced_codecs", "hvc1,av01", "-i", input,
			"-max_muxing_queue_size", "9999", "-c:v", "libx264", "-x264opts", "no-scenecut", "-preset", "ultrafast", "-s", strconv.Itoa(output.Width)+"x"+strconv.Itoa(output.Height),
			"-b:v", output.VideoBandwidth, "-b:a", output.AudioBandwidth, "-r", strconv.Itoa(int(output.FrameRate)), "-g", strconv.Itoa(int(output.FrameRate*2)),
			"-hls_flags", "append_list+omit_endlist+program_date_time", "-hls_list_size", "6", "-hls_time", "2",
			"-hls_segment_type", "fmp4", "-hls_fmp4_init_filename", "init.mp4", "-hls_fmp4_init_resend", "1",
			"-ignore_io_errors", "1", "-method", "POST", "-headers", "Authorization: Bearer "+utils.Config.Ingest.AuthKey, "-f", "hls",
			"-hls_segment_filename", utils.Config.Cache.Hostname+"/live/"+stream.User.StreamKey+"_"+output.Variant+"/%d.m4s", utils.Config.Cache.Hostname+"/transcode/"+stream.User.StreamKey+"_"+output.Variant+"/index.m3u8")
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}
