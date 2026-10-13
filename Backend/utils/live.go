package utils

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
)

// LiveManager tracks active FFmpeg RTMP listeners (one at a time on :1935).
type LiveManager struct {
	mu      sync.Mutex
	active  uint // live stream ID currently listening
	cmd     *exec.Cmd
	hlsDir  string
}

var Live = &LiveManager{}

// OnLiveEnded is called when the FFmpeg listener exits (OBS disconnect or kill).
// Handlers wire this to mark the DB stream as ended.
var OnLiveEnded func(streamID uint)

func (m *LiveManager) IsBusy() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.cmd != nil && m.cmd.Process != nil
}

func (m *LiveManager) ActiveID() uint {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.active
}

// StartListen spawns FFmpeg to accept one OBS RTMP publish and write live HLS.
func (m *LiveManager) StartListen(streamID uint, streamKey, hlsDir string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.cmd != nil && m.cmd.Process != nil {
		return fmt.Errorf("another live session is already active (id=%d)", m.active)
	}

	if err := os.MkdirAll(hlsDir, os.ModePerm); err != nil {
		return err
	}

	rtmpURL := fmt.Sprintf("rtmp://0.0.0.0:1935/live/%s", streamKey)
	playlist := fmt.Sprintf("%s/index.m3u8", hlsDir)

	cmd := exec.Command(ffmpegBin(),
		"-y",
		"-listen", "1",
		"-i", rtmpURL,
		"-c:v", "libx264",
		"-preset", "veryfast",
		"-tune", "zerolatency",
		"-c:a", "aac",
		"-ac", "2",
		"-ar", "48000",
		"-f", "hls",
		"-hls_time", "2",
		"-hls_list_size", "6",
		"-hls_flags", "delete_segments+append_list",
		playlist,
	)

	if err := cmd.Start(); err != nil {
		return err
	}

	m.cmd = cmd
	m.active = streamID
	m.hlsDir = hlsDir

	go func(id uint, process *exec.Cmd) {
		_ = process.Wait()
		ended := false
		m.mu.Lock()
		if m.active == id {
			m.cmd = nil
			m.active = 0
			m.hlsDir = ""
			ended = true
		}
		m.mu.Unlock()
		if ended && OnLiveEnded != nil {
			OnLiveEnded(id)
		}
	}(streamID, cmd)

	return nil
}

func (m *LiveManager) Stop(streamID uint) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.cmd == nil || m.cmd.Process == nil {
		return nil
	}
	if m.active != streamID {
		return fmt.Errorf("stream %d is not the active session", streamID)
	}

	_ = m.cmd.Process.Kill()
	m.cmd = nil
	m.active = 0
	m.hlsDir = ""
	return nil
}
