import React, { useCallback, useEffect, useMemo, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { api, showCoverUrl, episodeCoverUrl } from "../api";
import "./Detail.css";

interface Episode {
  id: number;
  episode_number: number;
  title: string;
  status: string;
  cover_image_file_path: string;
}

interface Season {
  id: number;
  season_number: number;
  title: string;
  episodes: Episode[];
}

interface Show {
  id: number;
  title: string;
  cover_image_file_path: string;
  seasons: Season[];
}

const ShowDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const [show, setShow] = useState<Show | null>(null);
  const [seasonIdx, setSeasonIdx] = useState(0);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  const load = useCallback(async () => {
    if (!id) return;
    try {
      const { data } = await api.get<Show>(`/shows/${id}`);
      setShow(data);
      setError(null);
    } catch {
      setError("Failed to load show.");
    } finally {
      setLoading(false);
    }
  }, [id]);

  useEffect(() => {
    void load();
  }, [load]);

  const seasons = useMemo(() => show?.seasons ?? [], [show]);
  const activeSeason = seasons[seasonIdx] ?? seasons[0];

  if (loading) return <div className="detail-state">Loading…</div>;
  if (error || !show) {
    return (
      <div className="detail-state">
        <p>{error || "Not found"}</p>
        <Link to="/">Back home</Link>
      </div>
    );
  }

  return (
    <div className="detail">
      <Link to="/" className="detail-back">← Home</Link>
      <div className="detail-hero">
        <div className="detail-hero__art">
          {show.cover_image_file_path ? (
            <img src={showCoverUrl(show.id)} alt="" />
          ) : (
            <div className="detail-hero__placeholder" />
          )}
        </div>
        <div className="detail-hero__copy">
          <h1 className="detail-title">{show.title}</h1>
          <p className="detail-lead">
            {seasons.length} season{seasons.length === 1 ? "" : "s"}
          </p>
        </div>
      </div>

      {seasons.length === 0 ? (
        <p className="detail-empty">
          No episodes yet. <Link to="/upload">Upload an episode</Link>
        </p>
      ) : (
        <>
          <div className="season-tabs">
            {seasons.map((s, i) => (
              <button
                key={s.id}
                type="button"
                className={activeSeason?.id === s.id ? "is-active" : ""}
                onClick={() => setSeasonIdx(i)}
              >
                {s.title || `Season ${s.season_number}`}
              </button>
            ))}
          </div>

          {activeSeason && (
            <ul className="episode-list">
              {(activeSeason.episodes || []).map((ep) => (
                <li key={ep.id}>
                  <Link to={`/watch/episode/${ep.id}`} className="episode-row">
                    <div className="episode-row__thumb">
                      {ep.cover_image_file_path || show.cover_image_file_path ? (
                        <img
                          src={ep.cover_image_file_path ? episodeCoverUrl(ep.id) : showCoverUrl(show.id)}
                          alt=""
                        />
                      ) : (
                        <div className="episode-row__placeholder" />
                      )}
                    </div>
                    <div className="episode-row__meta">
                      <span className="episode-row__num">
                        E{ep.episode_number}
                      </span>
                      <span className="episode-row__title">{ep.title}</span>
                      {ep.status === "processing" && (
                        <span className="episode-row__badge">Processing</span>
                      )}
                      {ep.status === "failed" && (
                        <span className="episode-row__badge episode-row__badge--fail">Failed</span>
                      )}
                    </div>
                  </Link>
                </li>
              ))}
              {(activeSeason.episodes || []).length === 0 && (
                <li className="detail-empty">No episodes in this season.</li>
              )}
            </ul>
          )}
        </>
      )}
    </div>
  );
};

export default ShowDetail;
