# Movie Server
- This project currently is a backend only movie server that handles files upload and will create paths to files for streaming.

## Technologies used 
- SQLite for a database to hold information about the movies
- Gin as the Go HTTP framework powering the API
- GORM as the Go ORM framework for simplified database interactions and migrations
- FFmpeg for video streaming

## API Routes

- All endpoints are served at `http://localhost:8080`

### 1. **Create Movie**
- **POST** `/movies`
    - Form Data: title (string), video (file: mp4, mov, or zip), cover_image (optional file)
    - Response: 201 Created with movie details

### 2 **List Movies**
- **GET** `/movies`
    - Response: 200 OK with list of movies

## 3 **Stream Movie**
- **GET** `/movies/:id/stream`
    - Response: 200 OK streaming the video file

### 4 **Update Movie**
- **PUT** `/movies/:id`
    - Form Data: title (string), video (optional file), cover_image (optional file)
    - Response: 200 OK with updated movie details

### 5 **Delete Movie**
- **DELETE** `/movies/:id`
    - Response: 204 No Content

## 6 **Stream Movie**
- **GET** `/movies/:id/HLS`
    - 302 Found  /movies/hls/<timestamp>/index.m3u8
    - The client then streams the video using HLS.

## FFmpeg Requirement

This project uses FFmpeg to generate HLS playlists and video segments for streaming.
FFmpeg **must be installed** on your system for the server to work.

### Install FFmpeg

**macOS (Homebrew)**  
```bash
brew install ffmpeg
```
**Ubuntu / Debian**
sudo apt update
sudo apt install ffmpeg

**Windows** 
Download from: `https://ffmpeg.org`

## Depency checks 
- Check vulnerabilities by running 
```govulncheck ./...```
- If govulncheck not instaled
  ```bash
    go install golang.org/x/vuln/cmd/govulncheck@latest

    echo 'export PATH=$HOME/go/bin:$PATH' >> ~/.zshrc
    source ~/.zshrc

    govulncheck -h
    ```

## Testing
- unit testing coming soon but can run a health check ```go test -v ./testing```

## Coming Soon
- Save state of video so I want to add params to start the video at a certain time
- Network connection so autoscalling the video output