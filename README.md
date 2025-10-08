# Download YouTube

My small project to grab all videos from a YouTube channel, then download the videos and add proper Seasons to it, then being able to put them on a server and watch them using Kodi with grabbing thumbnails and setting up all necessery infromation for the seasons and episodes.

## Usage

It usses ffmpeg to combine the video and audio file, had issues finding a stream that had both audio and video.

```
brew install ffmpeg
```

Create a .env file with following values:

```
YT_API_KEY=value
YT_CHANNEL_ID=value
```

Run it with **go run .**

If you run with **go run main.go**, it will not recognize the other files and you get error as **undefined: Video**

## Requierments

[YouTube API Key](https://developers.google.com/youtube/v3/getting-started)

## To Do

- [x] Dynamically check if the season exist, if not, create folder for it
- [x] Create episode .nfo file
- [x] Download and save the thumbnail and save it as follow: S01E01 - 2022 Money Masterclass-thumb
- [x] Right now, when re-running the script, it overwrites the previous downloaded data, we need to check and see if the video has already been added or not
- [x] Re-Download all data from youtube to get the biggest thumbnail possible
- [x] Make Season with Year dynamic, check for the oldest year and put that as seasson 1
- [ ] Delete Audio and Video files after
- [x] Able to specify the Shows Name
- [ ] Check to see what is the latest episode from that season to not overwrite
