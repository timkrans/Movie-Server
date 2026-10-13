import React, { useCallback, useEffect, useRef, useState } from "react";
import Hls from "hls.js";
import { api } from "../api";
import "./VideoPlayer.css";

type VideoPlayerProps = {
  src: string;
  poster?: string;
  title?: string;
  live?: boolean;
  contentType?: "movie" | "episode";
  contentId?: number;
  initialPosition?: number;
  onEnded?: () => void;
};

function formatTime(seconds: number) {
  if (!Number.isFinite(seconds) || seconds < 0) return "0:00";
  const h = Math.floor(seconds / 3600);
  const m = Math.floor((seconds % 3600) / 60);
  const s = Math.floor(seconds % 60);
  const pad = (n: number) => n.toString().padStart(2, "0");
  return h > 0 ? `${h}:${pad(m)}:${pad(s)}` : `${m}:${pad(s)}`;
}

async function isTauriRuntime() {
  try {
    const { isTauri } = await import("@tauri-apps/api/core");
    return isTauri();
  } catch {
    return false;
  }
}

const VideoPlayer: React.FC<VideoPlayerProps> = ({
  src,
  poster,
  title,
  live = false,
  contentType,
  contentId,
  initialPosition = 0,
  onEnded,
}) => {
  const containerRef = useRef<HTMLDivElement | null>(null);
  const videoRef = useRef<HTMLVideoElement | null>(null);
  const hlsRef = useRef<Hls | null>(null);
  const hideTimer = useRef<number | null>(null);
  const lastSave = useRef(0);
  const didSeek = useRef(false);
  const onEndedRef = useRef(onEnded);
  onEndedRef.current = onEnded;

  const [playing, setPlaying] = useState(false);
  const [current, setCurrent] = useState(0);
  const [duration, setDuration] = useState(0);
  const [volume, setVolume] = useState(1);
  const [muted, setMuted] = useState(false);
  const [fullscreen, setFullscreen] = useState(false);
  const [showControls, setShowControls] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [buffering, setBuffering] = useState(true);

  const saveProgress = useCallback(
    (position: number, dur: number, force = false) => {
      if (live || !contentType || contentId == null) return;
      const now = Date.now();
      if (!force && now - lastSave.current < 5000) return;
      lastSave.current = now;
      void api
        .put("/progress", {
          content_type: contentType,
          content_id: contentId,
          position_seconds: position,
          duration_seconds: dur,
        })
        .catch(() => undefined);
    },
    [live, contentType, contentId],
  );

  useEffect(() => {
    const video = videoRef.current;
    if (!video || !src) return;

    setError(null);
    setBuffering(true);
    didSeek.current = false;

    if (hlsRef.current) {
      hlsRef.current.destroy();
      hlsRef.current = null;
    }

    if (Hls.isSupported()) {
      const hls = new Hls({
        enableWorker: true,
        lowLatencyMode: live,
        liveSyncDurationCount: live ? 3 : undefined,
      });
      hlsRef.current = hls;
      hls.loadSource(src);
      hls.attachMedia(video);
      hls.on(Hls.Events.MANIFEST_PARSED, () => {
        setBuffering(false);
        void video.play().catch(() => undefined);
      });
      hls.on(Hls.Events.ERROR, (_event, data) => {
        if (data.fatal) {
          setError(live ? "Waiting for live feed…" : "Failed to load stream");
          setBuffering(false);
        }
      });
    } else if (video.canPlayType("application/vnd.apple.mpegurl")) {
      video.src = src;
      video.addEventListener(
        "loadedmetadata",
        () => {
          setBuffering(false);
          void video.play().catch(() => undefined);
        },
        { once: true },
      );
    } else {
      setError("HLS is not supported in this browser");
      setBuffering(false);
    }

    return () => {
      if (hlsRef.current) {
        hlsRef.current.destroy();
        hlsRef.current = null;
      }
    };
  }, [src, live]);

  const bumpControls = useCallback(() => {
    setShowControls(true);
    if (hideTimer.current) window.clearTimeout(hideTimer.current);
    hideTimer.current = window.setTimeout(() => {
      if (videoRef.current && !videoRef.current.paused) {
        setShowControls(false);
      }
    }, 2500);
  }, []);

  useEffect(() => {
    return () => {
      if (hideTimer.current) window.clearTimeout(hideTimer.current);
    };
  }, []);

  useEffect(() => {
    const video = videoRef.current;
    if (!video) return;

    const onTime = () => {
      setCurrent(video.currentTime);
      if (!live) saveProgress(video.currentTime, video.duration || 0);
    };
    const onMeta = () => {
      setDuration(video.duration || 0);
      if (
        !live &&
        !didSeek.current &&
        initialPosition > 5 &&
        video.duration > 0 &&
        initialPosition < video.duration * 0.9
      ) {
        video.currentTime = initialPosition;
        didSeek.current = true;
      }
    };
    const onPlay = () => setPlaying(true);
    const onPause = () => {
      setPlaying(false);
      if (!live) saveProgress(video.currentTime, video.duration || 0, true);
    };
    const onWaiting = () => setBuffering(true);
    const onPlaying = () => setBuffering(false);
    const onVideoEnded = () => {
      if (!live) {
        saveProgress(video.duration || 0, video.duration || 0, true);
      }
      onEndedRef.current?.();
    };

    video.addEventListener("timeupdate", onTime);
    video.addEventListener("loadedmetadata", onMeta);
    video.addEventListener("durationchange", onMeta);
    video.addEventListener("play", onPlay);
    video.addEventListener("pause", onPause);
    video.addEventListener("waiting", onWaiting);
    video.addEventListener("playing", onPlaying);
    video.addEventListener("ended", onVideoEnded);

    return () => {
      video.removeEventListener("timeupdate", onTime);
      video.removeEventListener("loadedmetadata", onMeta);
      video.removeEventListener("durationchange", onMeta);
      video.removeEventListener("play", onPlay);
      video.removeEventListener("pause", onPause);
      video.removeEventListener("waiting", onWaiting);
      video.removeEventListener("playing", onPlaying);
      video.removeEventListener("ended", onVideoEnded);
      if (!live && video.currentTime > 0) {
        saveProgress(video.currentTime, video.duration || 0, true);
      }
    };
  }, [live, initialPosition, saveProgress]);

  useEffect(() => {
    const onUnload = () => {
      const video = videoRef.current;
      if (!video || live || !contentType || contentId == null) return;
      saveProgress(video.currentTime, video.duration || 0, true);
    };
    window.addEventListener("beforeunload", onUnload);
    return () => window.removeEventListener("beforeunload", onUnload);
  }, [live, contentType, contentId, saveProgress]);

  const togglePlay = useCallback(() => {
    const video = videoRef.current;
    if (!video) return;
    if (video.paused) void video.play();
    else video.pause();
    bumpControls();
  }, [bumpControls]);

  const seekBy = useCallback(
    (delta: number) => {
      const video = videoRef.current;
      if (!video || live) return;
      video.currentTime = Math.max(0, Math.min(video.duration || 0, video.currentTime + delta));
      bumpControls();
    },
    [bumpControls, live],
  );

  const onSeek = (value: number) => {
    const video = videoRef.current;
    if (!video || live) return;
    video.currentTime = value;
    setCurrent(value);
  };

  const onVolume = (value: number) => {
    const video = videoRef.current;
    if (!video) return;
    video.volume = value;
    video.muted = value === 0;
    setVolume(value);
    setMuted(value === 0);
  };

  const toggleMute = () => {
    const video = videoRef.current;
    if (!video) return;
    video.muted = !video.muted;
    setMuted(video.muted);
  };

  const toggleFullscreen = useCallback(async () => {
    const next = !fullscreen;
    const tauri = await isTauriRuntime();

    if (tauri) {
      try {
        const { getCurrentWindow } = await import("@tauri-apps/api/window");
        await getCurrentWindow().setFullscreen(next);
        setFullscreen(next);
        containerRef.current?.classList.toggle("is-theater", next);
        bumpControls();
        return;
      } catch {
        // fall through to DOM fullscreen
      }
    }

    const el = containerRef.current;
    if (!el) return;
    if (!document.fullscreenElement) {
      await el.requestFullscreen();
      setFullscreen(true);
    } else {
      await document.exitFullscreen();
      setFullscreen(false);
    }
    bumpControls();
  }, [fullscreen, bumpControls]);

  useEffect(() => {
    const onFsChange = () => {
      setFullscreen(Boolean(document.fullscreenElement));
    };
    document.addEventListener("fullscreenchange", onFsChange);
    return () => document.removeEventListener("fullscreenchange", onFsChange);
  }, []);

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      const tag = (e.target as HTMLElement)?.tagName;
      if (tag === "INPUT" || tag === "TEXTAREA") return;

      switch (e.key.toLowerCase()) {
        case " ":
        case "k":
          e.preventDefault();
          togglePlay();
          break;
        case "f":
          e.preventDefault();
          void toggleFullscreen();
          break;
        case "m":
          e.preventDefault();
          toggleMute();
          break;
        case "arrowleft":
          e.preventDefault();
          seekBy(-5);
          break;
        case "arrowright":
          e.preventDefault();
          seekBy(5);
          break;
        case "arrowup":
          e.preventDefault();
          onVolume(Math.min(1, volume + 0.05));
          break;
        case "arrowdown":
          e.preventDefault();
          onVolume(Math.max(0, volume - 0.05));
          break;
        default:
          break;
      }
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [togglePlay, toggleFullscreen, seekBy, volume]);

  const progress = duration > 0 ? (current / duration) * 100 : 0;

  return (
    <div
      className={`video-player ${fullscreen ? "is-theater" : ""} ${showControls ? "show-controls" : ""}`}
      ref={containerRef}
      onMouseMove={bumpControls}
      onMouseLeave={() => playing && setShowControls(false)}
    >
      <video
        ref={videoRef}
        className="video-player__media"
        poster={poster}
        playsInline
        onClick={togglePlay}
      />

      {buffering && !error && <div className="video-player__spinner" aria-hidden />}
      {error && <div className="video-player__error">{error}</div>}

      <div className="video-player__chrome">
        {live && (
          <div className="video-player__live-badge">
            <span className="video-player__live-dot" />
            LIVE
          </div>
        )}

        <div className="video-player__controls">
          {!live && (
            <input
              className="video-player__seek"
              type="range"
              min={0}
              max={duration || 0}
              step={0.1}
              value={current}
              onChange={(e) => onSeek(Number(e.target.value))}
              aria-label="Seek"
              style={{ "--progress": `${progress}%` } as React.CSSProperties}
            />
          )}

          <div className="video-player__row">
            <button type="button" className="video-player__btn" onClick={togglePlay} aria-label={playing ? "Pause" : "Play"}>
              {playing ? (
                <svg viewBox="0 0 24 24" width="22" height="22" fill="currentColor"><path d="M6 5h4v14H6V5zm8 0h4v14h-4V5z"/></svg>
              ) : (
                <svg viewBox="0 0 24 24" width="22" height="22" fill="currentColor"><path d="M8 5v14l11-7L8 5z"/></svg>
              )}
            </button>

            <div className="video-player__volume">
              <button type="button" className="video-player__btn" onClick={toggleMute} aria-label={muted ? "Unmute" : "Mute"}>
                {muted || volume === 0 ? (
                  <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor"><path d="M16.5 12c0-1.77-1.02-3.29-2.5-4.03v2.21l2.45 2.45c.03-.2.05-.41.05-.63zm2.5 0c0 .94-.2 1.82-.54 2.64l1.51 1.51C20.63 14.91 21 13.5 21 12c0-4.28-2.99-7.86-7-8.77v2.06c2.89.86 5 3.54 5 6.71zM4.27 3 3 4.27 7.73 9H3v6h4l5 5v-6.73l4.25 4.25c-.67.52-1.42.93-2.25 1.18v2.06c1.38-.31 2.63-.95 3.69-1.81L19.73 21 21 19.73l-9-9L4.27 3zM12 4 9.91 6.09 12 8.18V4z"/></svg>
                ) : (
                  <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor"><path d="M3 9v6h4l5 5V4L7 9H3zm13.5 3a4.5 4.5 0 0 0-2.5-4v8a4.5 4.5 0 0 0 2.5-4zM14 3.23v2.06c2.89.86 5 3.54 5 6.71s-2.11 5.85-5 6.71v2.06c4.01-.91 7-4.49 7-8.77s-2.99-7.86-7-8.77z"/></svg>
                )}
              </button>
              <input
                type="range"
                min={0}
                max={1}
                step={0.05}
                value={muted ? 0 : volume}
                onChange={(e) => onVolume(Number(e.target.value))}
                aria-label="Volume"
              />
            </div>

            <span className="video-player__time">
              {live ? "LIVE" : `${formatTime(current)} / ${formatTime(duration)}`}
            </span>

            <div className="video-player__spacer" />

            {title && <span className="video-player__title-inline">{title}</span>}

            <button type="button" className="video-player__btn" onClick={() => void toggleFullscreen()} aria-label="Fullscreen">
              {fullscreen ? (
                <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor"><path d="M5 16h3v3h2v-5H5v2zm3-8H5v2h5V5H8v3zm6 11h2v-3h3v-2h-5v5zm2-11V5h-2v5h5V8h-3z"/></svg>
              ) : (
                <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor"><path d="M7 14H5v5h5v-2H7v-3zm-2-4h2V7h3V5H5v5zm12 7h-3v2h5v-5h-2v3zM14 5v2h3v3h2V5h-5z"/></svg>
              )}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

export default VideoPlayer;
