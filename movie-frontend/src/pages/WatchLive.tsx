import React, { useEffect, useState } from "react";
import axios from "axios";
import { useParams, Link } from "react-router-dom";
import VideoPlayer from "../components/VideoPlayer";
import { API_BASE, liveHlsUrl } from "../api";
import "./Watch.css";

interface LiveStream {
  id: number;
  title: string;
  status: string;
  hls_dir: string;
}

const WatchLive: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const [stream, setStream] = useState<LiveStream | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!id) {
      setError("Stream ID is missing.");
      setLoading(false);
      return;
    }

    let cancelled = false;
    const load = async () => {
      try {
        const { data } = await axios.get<LiveStream>(`${API_BASE}/live/${id}`);
        if (!cancelled) {
          setStream(data);
          setError(null);
        }
      } catch {
        if (!cancelled) setError("Failed to load live stream.");
      } finally {
        if (!cancelled) setLoading(false);
      }
    };

    void load();
    const poll = window.setInterval(() => void load(), 4000);
    return () => {
      cancelled = true;
      window.clearInterval(poll);
    };
  }, [id]);

  if (loading) {
    return <div className="watch-page watch-page--center">Loading…</div>;
  }

  if (error || !stream) {
    return (
      <div className="watch-page watch-page--center">
        <p>{error || "Not found"}</p>
        <Link to="/live">Back to Live</Link>
      </div>
    );
  }

  if (stream.status !== "live") {
    return (
      <div className="watch-page watch-page--center">
        <p>
          “{stream.title}” is {stream.status}. Start the stream from Go Live, then connect OBS.
        </p>
        <Link to={`/go-live`}>Go Live</Link>
      </div>
    );
  }

  return (
    <div className="watch-page">
      <VideoPlayer src={liveHlsUrl(stream.id)} title={stream.title} live />
      <div className="watch-meta">
        <Link to="/live" className="watch-back">
          ← Live
        </Link>
        <h1 className="watch-title">{stream.title}</h1>
      </div>
    </div>
  );
};

export default WatchLive;
