package main

import (
	"fmt"
	"log"
	"os"
	"text/template"

	"download-youtube/models"
)

// generateEpisodeNfo created an .nfo file with all data based on the extracted data from youtube
func generateEpisodeNfo(video models.Video) {
	// Template string
	xmlTemplate := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<!--created on {{.CreationDate}} - tinyMediaManager {{.Version}}-->
<episodedetails>
  <title>{{.Title}}</title>
  <originaltitle>{{.OriginalTitle}}</originaltitle>
  <showtitle>{{.ShowTitle}}</showtitle>
  <season>{{.Season}}</season>
  <episode>{{.Episode}}</episode>
  <displayseason>{{.DisplaySeason}}</displayseason>
  <displayepisode>{{.DisplayEpisode}}</displayepisode>
  <id>{{.ID}}</id>
  <ratings>{{.Ratings}}</ratings>
  <userrating>{{.UserRating}}</userrating>
  <plot>{{.Plot}}</plot>
  <runtime>{{.Runtime}}</runtime>
  <mpaa>{{.MPAA}}</mpaa>
  <premiered>{{.Premiered}}</premiered>
  <aired>{{.Aired}}</aired>
  <watched>{{.Watched}}</watched>
  <playcount>{{.PlayCount}}</playcount>
  <trailer>{{.Trailer}}</trailer>
  <dateadded>{{.DateAdded}}</dateadded>
  <epbookmark>{{.EpBookmark}}</epbookmark>
  <code>{{.Code}}</code>
  <fileinfo>
    <streamdetails>
      <video>
        <codec>{{.VideoCodec}}</codec>
        <aspect>{{.VideoAspect}}</aspect>
        <width>{{.VideoWidth}}</width>
        <height>{{.VideoHeight}}</height>
        <durationinseconds>{{.VideoDuration}}</durationinseconds>
        <stereomode>{{.StereoMode}}</stereomode>
      </video>
    </streamdetails>
  </fileinfo>
  <!--tinyMediaManager meta data-->
  <source>{{.Source}}</source>
  <original_filename>{{.OriginalFilename}}</original_filename>
  <user_note>{{.UserNote}}</user_note>
  <episode_groups>
    <group episode="{{.GroupEpisode}}" id="{{.GroupID}}" name="{{.GroupName}}" season="{{.GroupSeason}}"/>
  </episode_groups>
</episodedetails>`

	// Define the episode details
	episode := models.NFOEpisodeDetails{
		CreationDate:     "2024-07-25 15:06:07",
		Title:            video.Title,
		OriginalTitle:    video.Title,
		ShowTitle:        video.Title,
		Season:           video.Season,
		Episode:          video.Episode,
		Plot:             video.Description,
		Runtime:          "0",
		MPAA:             "",
		Premiered:        video.PublishedAt,
		Aired:            video.PublishedAt,
		Watched:          "false",
		PlayCount:        "0",
		Trailer:          "",
		EpBookmark:       "",
		Code:             "",
		VideoCodec:       "",
		VideoAspect:      "0.0",
		VideoWidth:       "1280",
		VideoHeight:      "720",
		VideoDuration:    "0",
		StereoMode:       "",
		Source:           video.URL,
		OriginalFilename: video.Filename,
		UserNote:         "",
	}

	filename := fmt.Sprintf("%s.nfo", video.Filepath)

	if _, err := os.Stat(filename); err == nil {
		log.Print("NFO file already exist, skipping")
		return
	}

	// Parse the template
	tmpl, err := template.New("episodedetails").Parse(xmlTemplate)
	if err != nil {
		log.Print("Error parsing template:", err)
		return
	}

	file, err := os.Create(filename)
	if err != nil {
		log.Print("Error creating file:", err)
		return
	}
	defer file.Close()

	// Execute the template and write to the file
	err = tmpl.Execute(file, episode)
	if err != nil {
		log.Print("Error executing template:", err)
		return
	}

	log.Printf(".nfi file created successfully: %s", filename)
}
