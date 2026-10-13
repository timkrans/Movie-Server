package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func ffmpegBin() string {
	ffmpegPath := os.Getenv("FFMPEG_PATH")
	if ffmpegPath == "" {
		return "ffmpeg"
	}
	return ffmpegPath
}

// GenerateHLS creates an ABR HLS ladder (360p / 720p / 1080p) with a master playlist.
func GenerateHLS(inputPath, outputDir string) (string, error) {
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		return "", err
	}

	playlist := filepath.Join(outputDir, "index.m3u8")

	cmd := exec.Command(ffmpegBin(),
		"-y",
		"-i", inputPath,
		"-filter_complex",
		"[0:v]split=3[v360][v720][v1080];"+
			"[v360]scale=-2:360[v360out];"+
			"[v720]scale=-2:720[v720out];"+
			"[v1080]scale=-2:1080[v1080out]",
		"-map", "[v360out]", "-map", "0:a?",
		"-map", "[v720out]", "-map", "0:a?",
		"-map", "[v1080out]", "-map", "0:a?",
		"-c:v", "libx264",
		"-preset", "veryfast",
		"-c:a", "aac",
		"-ac", "2",
		"-ar", "48000",
		"-b:v:0", "800k",
		"-b:v:1", "2500k",
		"-b:v:2", "5000k",
		"-b:a:0", "96k",
		"-b:a:1", "128k",
		"-b:a:2", "192k",
		"-var_stream_map", "v:0,a:0 v:1,a:1 v:2,a:2",
		"-master_pl_name", "index.m3u8",
		"-hls_time", "4",
		"-hls_list_size", "0",
		"-hls_segment_type", "mpegts",
		"-hls_flags", "independent_segments",
		"-f", "hls",
		filepath.Join(outputDir, "stream_%v.m3u8"),
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("FFMPEG ERROR:", string(output))
		return "", err
	}

	return playlist, nil
}
