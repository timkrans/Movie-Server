package utils

import (
	"fmt"
	"os"
	"os/exec"
)

func GenerateHLS(inputPath, outputDir string) (string, error) {
	err := os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		return "", err
	}

	playlist := fmt.Sprintf("%s/index.m3u8", outputDir)
	//must be homebrw could be set as env variable
	ffmpegPath := os.Getenv("FFMPEG_PATH")
	if ffmpegPath == "" {
    	ffmpegPath = "ffmpeg" //fallback
	}

	cmd := exec.Command(ffmpegPath,
		"-i", inputPath,
		"-vf", "scale=-2:720",
		"-c:v", "h264",
		"-c:a", "aac",
		"-ac", "2",
		"-ar", "48000",
		"-b:v", "2500k",
		"-hls_time", "4",
		"-hls_list_size", "0",
		"-hls_segment_type", "mpegts",
		"-hls_flags", "independent_segments",
		playlist,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("FFMPEG ERROR:", string(output))
		return "", err
	}

	return playlist, nil
}
