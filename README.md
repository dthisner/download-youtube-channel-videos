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
