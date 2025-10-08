package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type App struct {
	download Download
	YT       YouTubeChannel
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	envVar := EnvVar{
		ApiKey:          os.Getenv("YT_API_KEY"),
		ChannelID:       os.Getenv("YT_CHANNEL_ID"),
		ChannelName:     os.Getenv("YT_CHANNEL_NAME"),
		SeasonStartYear: os.Getenv("SEASON_START_YEAR"),
	}

	if err := envVar.Validate(); err != nil {
		log.Fatal(err)
	}

	jsonFilePath := fmt.Sprintf("%s-channel-data.json", envVar.ChannelName)

	var video []Video

	app := &App{
		download: Download{
			jsonFilePath: jsonFilePath,
			showName:     envVar.ChannelName,
		},
		YT: YouTubeChannel{
			evnVar:              envVar,
			jsonFilePath:        jsonFilePath,
			currentVideoData:    video,
			downloadedVideoData: video,
		},
	}

	app.YT.getData()
	app.download.Videos()
}
