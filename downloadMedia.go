package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/kkdai/youtube/v2"
)

func downloadVideo(videoURL string) error {
	client := youtube.Client{}

	video, err := client.GetVideo(videoURL)
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

	videoFileName := strings.ReplaceAll(video.Title, " ", "_") + "_video.mp4"
	err = downloadStream(client, video, videoFormat, videoFileName)
	if err != nil {
		return err
	}
	audioFileName := strings.ReplaceAll(video.Title, " ", "_") + "_audio.mp4"
	err = downloadStream(client, video, audioFormat, audioFileName)
	if err != nil {
		return err
	}

	err = mergeAudioVideo(video.Title, videoFileName, audioFileName)
	if err != nil {
		return err
	}

	return nil
}

func mergeAudioVideo(title, videoFileName, audioFileName string) error {
	mergedFileName := strings.ReplaceAll(title, " ", "_") + "_merged.mp4"
	ffmpegCmd := exec.Command("ffmpeg", "-i", videoFileName, "-i", audioFileName, "-c:v", "copy", "-c:a", "aac", "-strict", "experimental", mergedFileName)

	log.Print("Merging audio and video...")
	ffmpegCmd.Stdout = os.Stdout
	ffmpegCmd.Stderr = os.Stderr

	if err := ffmpegCmd.Run(); err != nil {
		return fmt.Errorf("error merging audio and video: %s", err)
	}

	log.Print("Merging complete:", mergedFileName)

	return nil
}

func downloadStream(client youtube.Client, video *youtube.Video, format *youtube.Format, filename string) error {
	stream, _, err := client.GetStream(video, format)
	if err != nil {
		return fmt.Errorf("WithAudioChannels - %s", err)
	}
	defer stream.Close()

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("create - %s", err)
	}
	defer file.Close()

	_, err = io.Copy(file, stream)
	if err != nil {
		return fmt.Errorf("copy - %s", err)
	}

	fmt.Printf("Stream downloaded successfully: %s\n", filename)

	return nil
}
