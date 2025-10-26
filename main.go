package main

import (
	"fmt"
	"log"
	"os"

	"download-youtube/getYTData"
	"download-youtube/models"

	"github.com/joho/godotenv"
)

type App struct {
	Download Download
	YT       getYTData.YouTubeChannel
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	envVar := models.EnvVar{
		ApiKey:          os.Getenv("YT_API_KEY"),
		ChannelID:       os.Getenv("YT_CHANNEL_ID"),
		PlaylistID:      os.Getenv("YT_PLAYLIST_ID"),
		ChannelName:     os.Getenv("YT_CHANNEL_NAME"),
		SeasonStartYear: os.Getenv("SEASON_START_YEAR"),
		SaveLoc:         os.Getenv("SAVE_LOCATION"),
	}

	if err := envVar.Validate(); err != nil {
		log.Fatal(err)
	}

	jsonFilePath := fmt.Sprintf("%s%s-channel-data.json", envVar.SaveLoc, envVar.ChannelName)

	var video []models.Video

	app := &App{
		Download: Download{
			JsonFilePath: jsonFilePath,
			ShowName:     envVar.ChannelName,
			SaveLoc:      envVar.SaveLoc,
		},
		YT: getYTData.YouTubeChannel{
			EnvVar:              envVar,
			JsonFilePath:        jsonFilePath,
			CurrentVideoData:    video,
			DownloadedVideoData: video,
		},
	}

	app.YT.GetData()
	app.Download.Videos()
}
