import React, { useCallback, useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { api } from "../api";
import "./Upload.css";

type Mode = "movie" | "episode" | "show" | "collection";

interface Collection {
  id: number;
  title: string;
}

interface Show {
  id: number;
  title: string;
}

const Upload: React.FC = () => {
  const navigate = useNavigate();
  const [mode, setMode] = useState<Mode>("movie");
  const [title, setTitle] = useState("");
  const [videoFile, setVideoFile] = useState<File | null>(null);
  const [coverImage, setCoverImage] = useState<File | null>(null);
  const [collections, setCollections] = useState<Collection[]>([]);
  const [shows, setShows] = useState<Show[]>([]);
  const [collectionId, setCollectionId] = useState<string>("");
  const [collectionOrder, setCollectionOrder] = useState("1");
  const [showId, setShowId] = useState<string>("");
  const [seasonNumber, setSeasonNumber] = useState("1");
  const [episodeNumber, setEpisodeNumber] = useState("1");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const loadMeta = useCallback(async () => {
    try {
      const [colRes, showRes] = await Promise.all([
        api.get<Collection[]>("/collections"),
        api.get<Show[]>("/shows"),
      ]);
      setCollections(colRes.data);
      setShows(showRes.data);
    } catch {
      /* ignore */
    }
  }, []);

  useEffect(() => {
    void loadMeta();
  }, [loadMeta]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);

    try {
      if (mode === "show") {
        const formData = new FormData();
        formData.append("title", title);
        if (coverImage) formData.append("cover_image", coverImage);
        const { data } = await api.post<{ id: number }>("/shows", formData);
        navigate(`/shows/${data.id}`);
        return;
      }

      if (mode === "collection") {
        const formData = new FormData();
        formData.append("title", title);
        if (coverImage) formData.append("cover_image", coverImage);
        const { data } = await api.post<{ id: number }>("/collections", formData);
        navigate(`/collections/${data.id}`);
        return;
      }

      if (!videoFile) {
        setError("Please select a video file.");
        setLoading(false);
        return;
      }

      if (mode === "movie") {
        const formData = new FormData();
        formData.append("title", title);
        formData.append("video", videoFile);
        if (coverImage) formData.append("cover_image", coverImage);
        if (collectionId) {
          formData.append("collection_id", collectionId);
          formData.append("collection_order", collectionOrder || "1");
        }
        const { data } = await api.post<{ id: number }>("/movies", formData);
        if (collectionId) {
          await api.post(`/collections/${collectionId}/movies`, {
            movie_id: data.id,
            order: Number(collectionOrder) || 1,
          }).catch(() => undefined);
        }
        navigate(`/watch/${data.id}`);
        return;
      }

      if (mode === "episode") {
        if (!showId) {
          setError("Select a show (create one first if needed).");
          setLoading(false);
          return;
        }
        const formData = new FormData();
        formData.append("title", title);
        formData.append("season_number", seasonNumber);
        formData.append("episode_number", episodeNumber);
        formData.append("video", videoFile);
        if (coverImage) formData.append("cover_image", coverImage);
        const { data } = await api.post<{ id: number }>(
          `/shows/${showId}/episodes`,
          formData,
        );
        navigate(`/watch/episode/${data.id}`);
      }
    } catch {
      setError("Failed to upload. Check that the backend is running.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="upload">
      <h1 className="upload__title">Upload</h1>
      <p className="upload__lead">
        Add movies, TV episodes, shows, or collections to your library.
      </p>

      <div className="upload__modes" role="tablist">
        {(
          [
            ["movie", "Movie"],
            ["episode", "Episode"],
            ["show", "New show"],
            ["collection", "New collection"],
          ] as const
        ).map(([value, label]) => (
          <button
            key={value}
            type="button"
            role="tab"
            className={mode === value ? "is-active" : ""}
            onClick={() => setMode(value)}
          >
            {label}
          </button>
        ))}
      </div>

      <form onSubmit={handleSubmit} className="upload__form">
        <label htmlFor="title">
          {mode === "episode" ? "Episode title" : "Title"}
        </label>
        <input
          id="title"
          type="text"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          required
        />

        {mode === "movie" && (
          <>
            <label htmlFor="collection">Collection (optional)</label>
            <select
              id="collection"
              value={collectionId}
              onChange={(e) => setCollectionId(e.target.value)}
            >
              <option value="">None — standalone movie</option>
              {collections.map((c) => (
                <option key={c.id} value={c.id}>
                  {c.title}
                </option>
              ))}
            </select>
            {collectionId && (
              <>
                <label htmlFor="order">Order in collection</label>
                <input
                  id="order"
                  type="number"
                  min={1}
                  value={collectionOrder}
                  onChange={(e) => setCollectionOrder(e.target.value)}
                />
              </>
            )}
          </>
        )}

        {mode === "episode" && (
          <>
            <label htmlFor="show">Show</label>
            <select
              id="show"
              value={showId}
              onChange={(e) => setShowId(e.target.value)}
              required
            >
              <option value="">Select a show…</option>
              {shows.map((s) => (
                <option key={s.id} value={s.id}>
                  {s.title}
                </option>
              ))}
            </select>
            <label htmlFor="season">Season number</label>
            <input
              id="season"
              type="number"
              min={1}
              value={seasonNumber}
              onChange={(e) => setSeasonNumber(e.target.value)}
              required
            />
            <label htmlFor="epnum">Episode number</label>
            <input
              id="epnum"
              type="number"
              min={1}
              value={episodeNumber}
              onChange={(e) => setEpisodeNumber(e.target.value)}
              required
            />
          </>
        )}

        {(mode === "movie" || mode === "episode") && (
          <>
            <label htmlFor="videoFile">Video (.mp4, .mov, or .zip)</label>
            <input
              id="videoFile"
              type="file"
              accept=".mp4,.mov,.zip,video/*"
              onChange={(e) => setVideoFile(e.target.files?.[0] ?? null)}
              required
            />
          </>
        )}

        <label htmlFor="coverImage">
          Cover image {mode === "movie" || mode === "episode" ? "(optional)" : "(optional)"}
        </label>
        <input
          id="coverImage"
          type="file"
          accept="image/*"
          onChange={(e) => setCoverImage(e.target.files?.[0] ?? null)}
        />

        <button type="submit" disabled={loading}>
          {loading ? "Uploading…" : mode === "show" || mode === "collection" ? "Create" : "Upload"}
        </button>
      </form>

      {error && <p className="upload__error">{error}</p>}
    </div>
  );
};

export default Upload;
