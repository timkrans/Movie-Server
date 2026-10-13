import React, { useEffect, useState } from "react";
import {
  BrowserRouter as Router,
  Routes,
  Route,
  Link,
  useLocation,
  Navigate,
  useNavigate,
} from "react-router-dom";
import Home from "./pages/Home";
import Upload from "./pages/Upload";
import GoLive from "./pages/GoLive";
import WatchMovie from "./pages/WatchMovie";
import WatchLive from "./pages/WatchLive";
import WatchEpisode from "./pages/WatchEpisode";
import ShowDetail from "./pages/ShowDetail";
import CollectionDetail from "./pages/CollectionDetail";
import ProfilePicker from "./pages/ProfilePicker";
import { api, getActiveProfileId, setActiveProfileId, type Profile } from "./api";
import { getPreferredTheme, setTheme, type Theme } from "./theme";
import "./App.css";

function ThemeToggle() {
  const [theme, setThemeState] = useState<Theme>(() => getPreferredTheme());

  const toggle = () => {
    const next: Theme = theme === "dark" ? "light" : "dark";
    setTheme(next);
    setThemeState(next);
  };

  const toLight = theme === "dark";

  return (
    <button
      type="button"
      className="theme-toggle"
      onClick={toggle}
      aria-label={toLight ? "Switch to light mode" : "Switch to dark mode"}
      title={toLight ? "Light mode" : "Dark mode"}
    >
      {toLight ? (
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden>
          <circle cx="12" cy="12" r="4" />
          <path d="M12 2v2M12 20v2M4.9 4.9l1.4 1.4M17.7 17.7l1.4 1.4M2 12h2M20 12h2M4.9 19.1l1.4-1.4M17.7 6.3l1.4-1.4" />
        </svg>
      ) : (
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden>
          <path d="M21 14.5A8.5 8.5 0 1 1 9.5 3a7 7 0 0 0 11.5 11.5z" />
        </svg>
      )}
    </button>
  );
}

function RequireProfile({ children }: { children: React.ReactNode }) {
  const location = useLocation();
  const profileId = getActiveProfileId();
  if (!profileId && location.pathname !== "/profiles") {
    return <Navigate to="/profiles" replace />;
  }
  return <>{children}</>;
}

function Shell() {
  const location = useLocation();
  const navigate = useNavigate();
  const hideChrome =
    location.pathname.startsWith("/watch/") ||
    location.pathname.startsWith("/live/") ||
    location.pathname === "/profiles";

  const [profile, setProfile] = useState<Profile | null>(null);

  useEffect(() => {
    const id = getActiveProfileId();
    if (!id) {
      setProfile(null);
      return;
    }
    void api
      .get<Profile[]>("/profiles")
      .then(({ data }) => {
        const found = data.find((p) => p.id === id) ?? null;
        if (!found) {
          setActiveProfileId(null);
          setProfile(null);
          navigate("/profiles", { replace: true });
          return;
        }
        setProfile(found);
      })
      .catch(() => undefined);
  }, [location.pathname, navigate]);

  const switchProfile = () => {
    setActiveProfileId(null);
    navigate("/profiles");
  };

  return (
    <div className={`shell ${hideChrome ? "shell--theater" : ""}`}>
      {!hideChrome && (
        <header className="topbar">
          <Link to="/" className="brand" aria-label="Movie Server home">
            <span className="brand__name">Movie Server</span>
          </Link>
          <nav className="topbar__nav">
            <Link to="/">Home</Link>
            <Link to="/go-live">Live</Link>
            <Link to="/upload">Upload</Link>
            {profile && (
              <button
                type="button"
                className="profile-chip"
                onClick={switchProfile}
                title="Switch profile"
              >
                <span
                  className="profile-chip__dot"
                  style={{ background: profile.avatar_color }}
                />
                {profile.name}
              </button>
            )}
            <ThemeToggle />
          </nav>
        </header>
      )}

      <main className="shell__main">
        <RequireProfile>
          <Routes>
            <Route path="/profiles" element={<ProfilePicker />} />
            <Route path="/" element={<Home />} />
            <Route path="/upload" element={<Upload />} />
            <Route path="/add-movie" element={<Upload />} />
            <Route path="/go-live" element={<GoLive />} />
            <Route path="/shows/:id" element={<ShowDetail />} />
            <Route path="/collections/:id" element={<CollectionDetail />} />
            <Route path="/watch/episode/:id" element={<WatchEpisode />} />
            <Route path="/watch/:id" element={<WatchMovie />} />
            <Route path="/movies/:id/hls" element={<WatchMovie />} />
            <Route path="/live/:id" element={<WatchLive />} />
          </Routes>
        </RequireProfile>
      </main>
    </div>
  );
}

const App: React.FC = () => (
  <Router>
    <Shell />
  </Router>
);

export default App;
