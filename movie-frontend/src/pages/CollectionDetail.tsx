import React, { useCallback, useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { api, collectionCoverUrl, movieCoverUrl } from "../api";
import "./Detail.css";

interface Movie {
  id: number;
  title: string;
  status: string;
  cover_image_file_path: string;
  collection_order?: number | null;
}

interface Collection {
  id: number;
  title: string;
  cover_image_file_path: string;
  movies: Movie[];
}

const CollectionDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const [collection, setCollection] = useState<Collection | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  const load = useCallback(async () => {
    if (!id) return;
    try {
      const { data } = await api.get<Collection>(`/collections/${id}`);
      setCollection(data);
      setError(null);
    } catch {
      setError("Failed to load collection.");
    } finally {
      setLoading(false);
    }
  }, [id]);

  useEffect(() => {
    void load();
  }, [load]);

  if (loading) return <div className="detail-state">Loading…</div>;
  if (error || !collection) {
    return (
      <div className="detail-state">
        <p>{error || "Not found"}</p>
        <Link to="/">Back home</Link>
      </div>
    );
  }

  const movies = collection.movies || [];

  return (
    <div className="detail">
      <Link to="/" className="detail-back">← Home</Link>
      <div className="detail-hero">
        <div className="detail-hero__art">
          {collection.cover_image_file_path ? (
            <img src={collectionCoverUrl(collection.id)} alt="" />
          ) : movies[0]?.cover_image_file_path ? (
            <img src={movieCoverUrl(movies[0].id)} alt="" />
          ) : (
            <div className="detail-hero__placeholder" />
          )}
        </div>
        <div className="detail-hero__copy">
          <h1 className="detail-title">{collection.title}</h1>
          <p className="detail-lead">
            {movies.length} film{movies.length === 1 ? "" : "s"}
          </p>
        </div>
      </div>

      {movies.length === 0 ? (
        <p className="detail-empty">
          No movies in this collection. <Link to="/upload">Upload one</Link>
        </p>
      ) : (
        <ul className="episode-list">
          {movies.map((m, i) => (
            <li key={m.id}>
              <Link to={`/watch/${m.id}`} className="episode-row">
                <div className="episode-row__thumb">
                  {m.cover_image_file_path ? (
                    <img src={movieCoverUrl(m.id)} alt="" />
                  ) : (
                    <div className="episode-row__placeholder" />
                  )}
                </div>
                <div className="episode-row__meta">
                  <span className="episode-row__num">
                    {m.collection_order ?? i + 1}
                  </span>
                  <span className="episode-row__title">{m.title}</span>
                  {m.status === "processing" && (
                    <span className="episode-row__badge">Processing</span>
                  )}
                </div>
              </Link>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
};

export default CollectionDetail;
