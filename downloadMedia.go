package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"

	"download-youtube/models"

	"github.com/kkdai/youtube/v2"
)

type Download struct {
	JsonFilePath string
	ShowName     string
}

func (d Download) Videos() {
	jsonFile, err := os.Open(d.JsonFilePath)
	if err != nil {
		panic(err)
	}
	defer jsonFile.Close()

	var videos []models.Video

	jsonByte, _ := io.ReadAll(jsonFile)
	json.Unmarshal(jsonByte, &videos)

	for i, video := range videos {
		printVideoTitle(video.Title)

		d.checkSeasonFolderExist(video.Season)
		generateEpisodeNfo(video)

		if !video.ImageSaved {
			err = d.image(video)
			if err != nil {
				log.Print(err)
			} else {
				log.Printf("Successfully Downloaded image to: %s", video.Filepath)
				videos[i].ImageSaved = true
			}
		} else {
			log.Print("Thumbnail already downloaded")
		}

		if !video.Downloaded {
			err = d.video(video)
			if err != nil {
				log.Print(err)
				videos[i].Error = err.Error()
				removeMediaFiles(video)
			} else {
				videos[i].Downloaded = true
				videos[i].Error = ""
			}
		} else {
			log.Print("Video already downloaded")
		}

		videosJSON, _ := json.Marshal(videos)
		err = os.WriteFile(d.JsonFilePath, videosJSON, 0644)
		if err != nil {
			log.Print("Problem with writting JSON", err)
		}

		log.Print("Successfully Downloaded, merged the video and updated the JSON file")
	}
}

// checkSeasonFolderExist creates the season folder if it's missing
func (d Download) checkSeasonFolderExist(season string) error {
	var tvShowName = d.ShowName
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

// DownloadVideo gets the YouTube video, looking for 720p format.
// Have to get both audio and video stream to then merge them into one file
func (d Download) video(v models.Video) error {
	client := youtube.Client{}

	log.Print("Downloading video from: ", v.URL)
	video, err := client.GetVideo(v.URL)
	if err != nil {
		return fmt.Errorf("error fetching video info: %v", err)
	}

	var videoFormat, audioFormat *youtube.Format
	for _, format := range video.Formats {
		if format.QualityLabel == "720p" && format.AudioChannels == 0 {
			videoFormat = &format
		}
		if format.AudioChannels > 0 && format.AudioQuality == "AUDIO_QUALITY_MEDIUM" {
			audioFormat = &format
		}
		if videoFormat != nil && audioFormat != nil {
			break
		}
	}

	if videoFormat == nil || audioFormat == nil {
		log.Fatal("Suitable video or audio format not found")
	}

	videoFileName := v.Filepath + "_video.mp4"
	err = d.stream(client, video, videoFormat, videoFileName)
	if err != nil {
		return err
	}
	audioFileName := v.Filepath + "_audio.mp4"
	err = d.stream(client, video, audioFormat, audioFileName)
	if err != nil {
		return err
	}

	err = mergeAudioVideo(v.Filepath, videoFileName, audioFileName)
	if err != nil {
		return err
	}

	_ = os.Remove(videoFileName)
	_ = os.Remove(audioFileName)

	return nil
}

func removeMediaFiles(v models.Video) {
	audioFileName := v.Filepath + "_audio.mp4"
	videoFileName := v.Filepath + "_video.mp4"

	_ = os.Remove(videoFileName)
	_ = os.Remove(audioFileName)
	_ = os.Remove(v.Filepath + ".mp4")

}

// mergeAudioVideo takes the audio and video file and merges it into one file
func mergeAudioVideo(filePath, videoFileName, audioFileName string) error {
	mergedFileName := filePath + ".mp4"
	ffmpegCmd := exec.Command("ffmpeg", "-i", videoFileName, "-i", audioFileName, "-c:v", "copy", "-c:a", "aac", "-strict", "experimental", mergedFileName)

	log.Print("Merging audio and video...")
	ffmpegCmd.Stdout = os.Stdout
	ffmpegCmd.Stderr = os.Stderr

	if err := ffmpegCmd.Run(); err != nil {
		return fmt.Errorf("error merging audio and video: %s", err)
	}

	log.Print("Merging completed: ", mergedFileName)

	return nil
}

// DownloadStream gets the YouTube audio or video stream and Downloads it
func (d Download) stream(client youtube.Client, video *youtube.Video, format *youtube.Format, filename string) error {
	stream, _, err := client.GetStream(video, format)
	if err != nil {
		return fmt.Errorf("get the video stream - %s", err)
	}
	defer stream.Close()

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("create destination file - %s", err)
	}
	defer file.Close()

	_, err = io.Copy(file, stream)
	if err != nil {
		return fmt.Errorf("copy - Problem streaming the video - %s", err)
	}

	fmt.Printf("Stream Downloaded successfully: %s\n", filename)

	return nil
}

// DownloadImage Downloads an image from the given URL and saves it to the specified file
func (d Download) image(video models.Video) error {
	log.Printf("Downloading Thumbnail to: %s", video.Title)

	response, err := http.Get(video.ThumbnailURL)
	if err != nil {
		return fmt.Errorf("failed to fetch the image: %v", err)
	}
	defer response.Body.Close()

	filePath := fmt.Sprintf("%s-thumb.jpg", video.Filepath)
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	_, err = io.Copy(file, response.Body)
	if err != nil {
		return fmt.Errorf("failed to save the image: %v", err)
	}

	return nil
}

func printVideoTitle(title string) {
	padding := ""
	titleLenght := len(title) + 23

	for i := 1; i < titleLenght; i++ {
		padding = padding + "#"
	}

	log.Printf(`
		%s
		########## %s ##########
		%s`, padding, title, padding)
}
