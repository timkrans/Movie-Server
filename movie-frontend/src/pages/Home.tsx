import React, { useCallback, useEffect, useState } from "react";
import { Link } from "react-router-dom";
import {
  api,
  absoluteApiUrl,
  movieCoverUrl,
  showCoverUrl,
  collectionCoverUrl,
  type ContinueItem,
} from "../api";
import "./Home.css";

interface Movie {
  id: number;
  title: string;
  status: string;
  cover_image_file_path: string;
  collection_id?: number | null;
}

interface Show {
  id: number;
  title: string;
  cover_image_file_path: string;
}

interface Collection {
  id: number;
  title: string;
  cover_image_file_path: string;
}

interface LiveStream {
  id: number;
  title: string;
  status: string;
}

type Tab = "movies" | "shows" | "live";

const Home: React.FC = () => {
  const [tab, setTab] = useState<Tab>("movies");
  const [movies, setMovies] = useState<Movie[]>([]);
  const [shows, setShows] = useState<Show[]>([]);
  const [collections, setCollections] = useState<Collection[]>([]);
  const [continueItems, setContinueItems] = useState<ContinueItem[]>([]);
  const [lives, setLives] = useState<LiveStream[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    try {
      const [movieRes, showRes, colRes, liveRes, contRes] = await Promise.all([
        api.get<Movie[]>("/movies"),
        api.get<Show[]>("/shows"),
        api.get<Collection[]>("/collections"),
        api.get<LiveStream[]>("/live"),
        api.get<ContinueItem[]>("/progress/continue").catch(() => ({ data: [] as ContinueItem[] })),
      ]);
      setMovies(movieRes.data);
      setShows(showRes.data);
      setCollections(colRes.data);
      setLives(liveRes.data);
      setContinueItems(contRes.data);
      setError(null);
    } catch {
      setError("Could not reach the movie server.");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void load();
  }, [load]);

  if (loading) {
    return <div className="home-state">Loading…</div>;
  }

  if (error) {
    return <div className="home-state home-state--error">{error}</div>;
  }

  const standaloneMovies = movies.filter((m) => !m.collection_id);

  return (
    <div className="home">
      {continueItems.length > 0 && (
        <section className="home-section home-section--continue">
          <h2 className="home-heading">Continue Watching</h2>
          <div className="continue-row">
            {continueItems.map((item) => (
              <Link
                key={`${item.content_type}-${item.content_id}`}
                to={item.watch_url}
                className="thumb continue-thumb"
                aria-label={item.title}
              >
                <div className="thumb__media">
                  {item.cover_url ? (
                    <img src={absoluteApiUrl(item.cover_url)} alt="" loading="lazy" />
                  ) : (
                    <div className="thumb__placeholder" />
                  )}
                  <span className="thumb__play" aria-hidden>▶</span>
                  <div className="continue-bar" aria-hidden>
                    <div
                      className="continue-bar__fill"
                      style={{ width: `${Math.min(100, item.progress * 100)}%` }}
                    />
                  </div>
                </div>
                <div className="thumb__title">{item.title}</div>
                {item.subtitle && <div className="thumb__sub">{item.subtitle}</div>}
              </Link>
            ))}
          </div>
        </section>
      )}

      <div className="home-tabs" role="tablist">
        <button
          type="button"
          role="tab"
          aria-selected={tab === "movies"}
          className={tab === "movies" ? "is-active" : ""}
          onClick={() => setTab("movies")}
        >
          Movies
        </button>
        <button
          type="button"
          role="tab"
          aria-selected={tab === "shows"}
          className={tab === "shows" ? "is-active" : ""}
          onClick={() => setTab("shows")}
        >
          Shows
        </button>
        <button
          type="button"
          role="tab"
          aria-selected={tab === "live"}
          className={tab === "live" ? "is-active" : ""}
          onClick={() => setTab("live")}
        >
          Live
        </button>
      </div>

      {tab === "movies" && (
        <section className="home-section">
          {collections.length > 0 && (
            <>
              <h2 className="home-heading">Collections</h2>
              <div className="thumb-grid" style={{ marginBottom: 32 }}>
                {collections.map((col) => (
                  <Link
                    key={col.id}
                    to={`/collections/${col.id}`}
                    className="thumb"
                    aria-label={col.title}
                  >
                    <div className="thumb__media">
                      {col.cover_image_file_path ? (
                        <img src={collectionCoverUrl(col.id)} alt="" loading="lazy" />
                      ) : (
                        <div className="thumb__placeholder" />
                      )}
                    </div>
                    <div className="thumb__title">{col.title}</div>
                  </Link>
                ))}
              </div>
            </>
          )}

          <h2 className="home-heading">Movies</h2>
          {standaloneMovies.length === 0 && collections.length === 0 ? (
            <p className="home-empty">
              No movies yet. <Link to="/upload">Upload one</Link>
            </p>
          ) : standaloneMovies.length === 0 ? (
            <p className="home-empty">All movies are in collections.</p>
          ) : (
            <div className="thumb-grid">
              {standaloneMovies.map((movie) => (
                <Link
                  key={movie.id}
                  to={`/watch/${movie.id}`}
                  className="thumb"
                  aria-label={movie.title}
                >
                  <div className="thumb__media">
                    {movie.cover_image_file_path ? (
                      <img src={movieCoverUrl(movie.id)} alt="" loading="lazy" />
                    ) : (
                      <div className="thumb__placeholder" />
                    )}
                    {movie.status === "processing" && (
                      <span className="thumb__badge">Processing</span>
                    )}
                    {movie.status === "failed" && (
                      <span className="thumb__badge thumb__badge--fail">Failed</span>
                    )}
                    <span className="thumb__play" aria-hidden>▶</span>
                  </div>
                  <div className="thumb__title">{movie.title}</div>
                </Link>
              ))}
            </div>
          )}
        </section>
      )}

      {tab === "shows" && (
        <section className="home-section">
          <div className="home-section__row">
            <h2 className="home-heading">TV Shows</h2>
            <Link to="/upload" className="home-cta">
              Add
            </Link>
          </div>
          {shows.length === 0 ? (
            <p className="home-empty">
              No shows yet. <Link to="/upload">Add a show</Link>
            </p>
          ) : (
            <div className="thumb-grid">
              {shows.map((show) => (
                <Link
                  key={show.id}
                  to={`/shows/${show.id}`}
                  className="thumb"
                  aria-label={show.title}
                >
                  <div className="thumb__media">
                    {show.cover_image_file_path ? (
                      <img src={showCoverUrl(show.id)} alt="" loading="lazy" />
                    ) : (
                      <div className="thumb__placeholder" />
                    )}
                  </div>
                  <div className="thumb__title">{show.title}</div>
                </Link>
              ))}
            </div>
          )}
        </section>
      )}

      {tab === "live" && (
        <section className="home-section">
          <div className="home-section__row">
            <h2 className="home-heading">Live streams</h2>
            <Link to="/go-live" className="home-cta">
              Go Live
            </Link>
          </div>
          {(() => {
            const activeLives = lives.filter((s) => s.status === "live");
            if (activeLives.length === 0) {
              return (
                <p className="home-empty">
                  Nothing live right now. <Link to="/go-live">Go Live</Link>
                </p>
              );
            }
            return (
              <div className="thumb-grid">
                {activeLives.map((stream) => (
                  <Link
                    key={stream.id}
                    to={`/live/${stream.id}`}
                    className="thumb"
                    aria-label={stream.title}
                  >
                    <div className="thumb__media thumb__media--live">
                      <div className="thumb__placeholder thumb__placeholder--live" />
                      <span className="thumb__badge thumb__badge--live">LIVE</span>
                    </div>
                    <div className="thumb__title">{stream.title}</div>
                  </Link>
                ))}
              </div>
            );
          })()}
        </section>
      )}
    </div>
  );
};

export default Home;
