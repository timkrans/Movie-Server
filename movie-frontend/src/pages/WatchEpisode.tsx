import React, { useCallback, useEffect, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import VideoPlayer from "../components/VideoPlayer";
import { api, episodeCoverUrl, episodeHlsUrl } from "../api";
import "./Watch.css";

interface Episode {
  id: number;
  season_id: number;
  episode_number: number;
  title: string;
  status: string;
  hls_path: string;
  show_id: number;
  show_title: string;
  season_number: number;
}

const WatchEpisode: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [episode, setEpisode] = useState<Episode | null>(null);
  const [initialPosition, setInitialPosition] = useState(0);
  const [ready, setReady] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  const load = useCallback(async () => {
    if (!id) return;
    try {
      const { data: ep } = await api.get<Episode>(`/episodes/${id}`);
      setEpisode(ep);

      try {
        const { data: prog } = await api.get<{ position_seconds: number }>(
          `/progress/episode/${ep.id}`,
        );
        setInitialPosition(prog.position_seconds || 0);
      } catch {
        setInitialPosition(0);
      }

      setReady(true);
      setError(null);
    } catch {
      setError("Failed to load episode.");
    } finally {
      setLoading(false);
    }
  }, [id]);

  useEffect(() => {
    if (!id) {
      setError("Episode ID is missing.");
      setLoading(false);
      return;
    }
    setReady(false);
    void load();
  }, [id, load]);

  useEffect(() => {
    if (episode?.status !== "processing") return;
    const poll = window.setInterval(() => void load(), 3000);
    return () => window.clearInterval(poll);
  }, [episode?.status, load]);

  const playNext = useCallback(async () => {
    if (!episode) return;
    try {
      const { data: next } = await api.get<{ id: number }>(`/episodes/${episode.id}/next`);
      navigate(`/watch/episode/${next.id}`);
    } catch {
      /* no next episode */
    }
  }, [episode, navigate]);

  if (loading) {
    return <div className="watch-page watch-page--center">Loading…</div>;
  }

  if (error || !episode) {
    return (
      <div className="watch-page watch-page--center">
        <p>{error || "Not found"}</p>
        <Link to="/">Back home</Link>
      </div>
    );
  }

  const status = episode.status || (episode.hls_path ? "ready" : "processing");
  const label = episode.show_title
    ? `${episode.show_title} · S${episode.season_number}E${episode.episode_number} · ${episode.title}`
    : `E${episode.episode_number} · ${episode.title}`;

  if (status === "processing") {
    return (
      <div className="watch-page watch-page--center">
        <p>Processing “{episode.title}” for streaming…</p>
        <Link to="/">Back home</Link>
      </div>
    );
  }

  if (status === "failed" || !episode.hls_path) {
    return (
      <div className="watch-page watch-page--center">
        <p>This episode failed to process.</p>
        <Link to="/">Back home</Link>
      </div>
    );
  }

  if (!ready) {
    return <div className="watch-page watch-page--center">Loading…</div>;
  }

  return (
    <div className="watch-page">
      <VideoPlayer
        key={episode.id}
        src={episodeHlsUrl(episode.id)}
        poster={episodeCoverUrl(episode.id)}
        title={label}
        contentType="episode"
        contentId={episode.id}
        initialPosition={initialPosition}
        onEnded={() => void playNext()}
      />
      <div className="watch-meta">
        <Link
          to={episode.show_id ? `/shows/${episode.show_id}` : "/"}
          className="watch-back"
        >
          ← {episode.show_id ? "Show" : "Home"}
        </Link>
        <h1 className="watch-title">{label}</h1>
      </div>
    </div>
  );
};

export default WatchEpisode;
