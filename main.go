package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/joho/godotenv"
)

var JSON_FILE_NAME = ""

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	apiKey := os.Getenv("YT_API_KEY")
	channelID := os.Getenv("YT_CHANNEL_ID")
	channelName := os.Getenv("YT_SHOW_NAME")
	seasonStartYear := os.Getenv("SEASON_START_YEAR")
	if apiKey == "" || channelID == "" || channelName == "" || seasonStartYear == "" {
		log.Fatal("missing YT_API_KEY or YT_CHANNEL_ID environment variables")
	}

	JSON_FILE_NAME = fmt.Sprintf("%s-channel-data.json", channelName)

	// updateYTChannelData()
	downloadAndOrganizeVideos()
}

func downloadAndOrganizeVideos() {
	jsonFile, err := os.Open(JSON_FILE_NAME)
	if err != nil {
		panic(err)
	}
	defer jsonFile.Close()

	var videos []Video

	jsonByte, _ := io.ReadAll(jsonFile)
	json.Unmarshal(jsonByte, &videos)

	for i, video := range videos {
		if video.Downloaded {
			log.Printf("Video %s has already been downloaded", video.Title)
			continue
		}

		log.Printf("Downloading Video %s", video.Title)

		checkSeasonFolderExist(video.Season)
		generateEpisodeNfo(video)
		err = downloadImage(video)
		if err != nil {
			log.Print(err)
		} else {
			log.Printf("Successfully downloaded image to: %s", video.Filepath)
			videos[i].ImageSaved = true
		}

		err = downloadVideo(video)
		if err != nil {
			log.Print(err)
		} else {
			videos[i].Downloaded = true
		}

		videosJSON, _ := json.Marshal(videos)
		err = os.WriteFile(JSON_FILE_NAME, videosJSON, 0644)
		if err != nil {
			log.Print("Problem with writting JSON", err)
		}

		log.Print("Successfully downloaded, merged the video and updated the JSON file")
	}
}

// checkSeasonFolderExist creates the season folder if it's missing
func checkSeasonFolderExist(season string) error {
	var tvShowName = os.Getenv("YT_SHOW_NAME")
	folderPath := fmt.Sprintf("%s/Season %s", tvShowName, season)

	_, err := os.Stat(folderPath)

	if os.IsNotExist(err) {
		// Create the folder (and any necessary parent directories)
		err := os.MkdirAll(folderPath, 0755) // 0755 is the permission mode
		if err != nil {
			return fmt.Errorf("error creating folder: %s", err)

		}
		log.Print("Folder created successfully: ", folderPath)
	} else {
		log.Print("Folder already exists:", folderPath)
	}

	return nil
}
