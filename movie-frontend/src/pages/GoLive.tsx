import React, { useEffect, useState } from "react";
import axios from "axios";
import { Link, useNavigate } from "react-router-dom";
import { API_BASE } from "../api";
import "./GoLive.css";

interface LiveStream {
  id: number;
  title: string;
  stream_key: string;
  status: string;
  obs_server?: string;
  obs_stream_key?: string;
}

const GoLive: React.FC = () => {
  const navigate = useNavigate();
  const [title, setTitle] = useState("");
  const [stream, setStream] = useState<LiveStream | null>(null);
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  const refresh = async (id: number) => {
    const { data } = await axios.get<LiveStream>(`${API_BASE}/live/${id}`);
    setStream(data);
  };

  useEffect(() => {
    if (!stream || stream.status !== "live") return;
    const id = window.setInterval(() => {
      void refresh(stream.id).catch(() => undefined);
    }, 4000);
    return () => window.clearInterval(id);
  }, [stream?.id, stream?.status]);

  const createStream = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);
    try {
      const { data } = await axios.post<LiveStream>(`${API_BASE}/live`, { title });
      setStream(data);
      setMessage("Stream created. Copy the OBS settings, then Start listening.");
    } catch {
      setError("Failed to create stream.");
    } finally {
      setLoading(false);
    }
  };

  const start = async () => {
    if (!stream) return;
    setLoading(true);
    setError(null);
    try {
      const { data } = await axios.post(`${API_BASE}/live/${stream.id}/start`);
      setStream((prev) => (prev ? { ...prev, status: "live" } : prev));
      setMessage(
        `Listening on RTMP. In OBS set Server to ${data.obs_server} and Stream Key to ${data.obs_stream_key}, then Start Streaming.`,
      );
    } catch (err: unknown) {
      const msg =
        axios.isAxiosError(err) && err.response?.data?.error
          ? String(err.response.data.error)
          : "Could not start RTMP listener.";
      setError(msg);
    } finally {
      setLoading(false);
    }
  };

  const stop = async () => {
    if (!stream) return;
    setLoading(true);
    try {
      await axios.post(`${API_BASE}/live/${stream.id}/stop`);
      setStream((prev) => (prev ? { ...prev, status: "ended" } : prev));
      setMessage("Stream ended.");
    } catch {
      setError("Failed to stop stream.");
    } finally {
      setLoading(false);
    }
  };

  const copy = async (text: string) => {
    try {
      await navigator.clipboard.writeText(text);
      setMessage("Copied to clipboard.");
    } catch {
      setMessage(text);
    }
  };

  return (
    <div className="golive">
      <h1 className="golive__title">Go Live</h1>
      <p className="golive__lead">
        Create a stream key, start the RTMP listener, then push from OBS Studio.
      </p>

      {!stream && (
        <form className="golive__form" onSubmit={createStream}>
          <label htmlFor="live-title">Stream title</label>
          <input
            id="live-title"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            required
            placeholder="Friday night watch party"
          />
          <button type="submit" disabled={loading}>
            {loading ? "Creating…" : "Create stream"}
          </button>
        </form>
      )}

      {stream && (
        <div className="golive__panel">
          <div className="golive__status">
            Status: <strong className={`status-${stream.status}`}>{stream.status}</strong>
          </div>

          <div className="golive__field">
            <span>OBS Server</span>
            <code>{stream.obs_server || "rtmp://localhost:1935/live"}</code>
            <button
              type="button"
              onClick={() =>
                void copy(stream.obs_server || "rtmp://localhost:1935/live")
              }
            >
              Copy
            </button>
          </div>

          <div className="golive__field">
            <span>Stream key</span>
            <code>{stream.stream_key || stream.obs_stream_key}</code>
            <button
              type="button"
              onClick={() =>
                void copy(stream.stream_key || stream.obs_stream_key || "")
              }
            >
              Copy
            </button>
          </div>

          <div className="golive__actions">
            {stream.status !== "live" && stream.status !== "ended" && (
              <button type="button" onClick={() => void start()} disabled={loading}>
                Start listening
              </button>
            )}
            {stream.status === "live" && (
              <>
                <button type="button" className="danger" onClick={() => void stop()} disabled={loading}>
                  Stop stream
                </button>
                <button type="button" className="ghost" onClick={() => navigate(`/live/${stream.id}`)}>
                  Open viewer
                </button>
              </>
            )}
            {stream.status === "ended" && (
              <Link to="/go-live" onClick={() => setStream(null)}>
                Create another
              </Link>
            )}
          </div>
        </div>
      )}

      {message && <p className="golive__msg">{message}</p>}
      {error && <p className="golive__err">{error}</p>}
    </div>
  );
};

export default GoLive;
