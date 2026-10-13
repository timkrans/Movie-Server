import React, { useCallback, useEffect, useState } from "react";
import { useParams, Link } from "react-router-dom";
import VideoPlayer from "../components/VideoPlayer";
import { api, movieCoverUrl, movieHlsUrl } from "../api";
import "./Watch.css";

interface Movie {
  id: number;
  title: string;
  status: string;
  hls_path: string;
}

const WatchMovie: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const [movie, setMovie] = useState<Movie | null>(null);
  const [initialPosition, setInitialPosition] = useState(0);
  const [ready, setReady] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  const load = useCallback(async () => {
    if (!id) return;
    try {
      const { data } = await api.get<Movie>(`/movies/${id}`);
      setMovie(data);

      try {
        const { data: prog } = await api.get<{ position_seconds: number }>(
          `/progress/movie/${data.id}`,
        );
        setInitialPosition(prog.position_seconds || 0);
      } catch {
        setInitialPosition(0);
      }

      setReady(true);
      setError(null);
    } catch {
      setError("Failed to load movie.");
    } finally {
      setLoading(false);
    }
  }, [id]);

  useEffect(() => {
    if (!id) {
      setError("Movie ID is missing.");
      setLoading(false);
      return;
    }
    setReady(false);
    void load();
  }, [id, load]);

  useEffect(() => {
    if (movie?.status !== "processing") return;
    const poll = window.setInterval(() => void load(), 3000);
    return () => window.clearInterval(poll);
  }, [movie?.status, load]);

  if (loading) {
    return <div className="watch-page watch-page--center">Loading…</div>;
  }

  if (error || !movie) {
    return (
      <div className="watch-page watch-page--center">
        <p>{error || "Not found"}</p>
        <Link to="/">Back home</Link>
      </div>
    );
  }

  const status = movie.status || (movie.hls_path ? "ready" : "processing");

  if (status === "processing") {
    return (
      <div className="watch-page watch-page--center">
        <p>Processing “{movie.title}” for streaming…</p>
        <Link to="/">Back home</Link>
      </div>
    );
  }

  if (status === "failed" || !movie.hls_path) {
    return (
      <div className="watch-page watch-page--center">
        <p>This video failed to process.</p>
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
        key={movie.id}
        src={movieHlsUrl(movie.id)}
        poster={movieCoverUrl(movie.id)}
        title={movie.title}
        contentType="movie"
        contentId={movie.id}
        initialPosition={initialPosition}
      />
      <div className="watch-meta">
        <Link to="/" className="watch-back">
          ← Home
        </Link>
        <h1 className="watch-title">{movie.title}</h1>
      </div>
    </div>
  );
};

export default WatchMovie;
