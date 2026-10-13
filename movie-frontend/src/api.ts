import axios from "axios";

export const API_BASE =
  import.meta.env.VITE_API_BASE?.replace(/\/$/, "") || "http://localhost:8080";

const PROFILE_KEY = "movie-server-profile-id";

export function getActiveProfileId(): number | null {
  const raw = localStorage.getItem(PROFILE_KEY);
  if (!raw) return null;
  const id = Number(raw);
  return Number.isFinite(id) && id > 0 ? id : null;
}

export function setActiveProfileId(id: number | null) {
  if (id == null) localStorage.removeItem(PROFILE_KEY);
  else localStorage.setItem(PROFILE_KEY, String(id));
}

export const api = axios.create({ baseURL: API_BASE });

api.interceptors.request.use((config) => {
  const profileId = getActiveProfileId();
  if (profileId != null) {
    config.headers["X-Profile-Id"] = String(profileId);
  }
  return config;
});

export function movieCoverUrl(id: number | string) {
  return `${API_BASE}/movies/${id}/cover`;
}

export function movieHlsUrl(id: number | string) {
  return `${API_BASE}/movies/${id}/hls/index.m3u8`;
}

export function liveHlsUrl(id: number | string) {
  return `${API_BASE}/live/${id}/hls/index.m3u8`;
}

export function episodeCoverUrl(id: number | string) {
  return `${API_BASE}/episodes/${id}/cover`;
}

export function episodeHlsUrl(id: number | string) {
  return `${API_BASE}/episodes/${id}/hls/index.m3u8`;
}

export function showCoverUrl(id: number | string) {
  return `${API_BASE}/shows/${id}/cover`;
}

export function collectionCoverUrl(id: number | string) {
  return `${API_BASE}/collections/${id}/cover`;
}

export type Profile = {
  id: number;
  name: string;
  avatar_color: string;
  created_at: number;
};

export type ContinueItem = {
  content_type: "movie" | "episode";
  content_id: number;
  title: string;
  subtitle?: string;
  cover_url: string;
  watch_url: string;
  position_seconds: number;
  duration_seconds: number;
  progress: number;
  updated_at: number;
};

export function absoluteApiUrl(path: string) {
  if (!path) return "";
  if (path.startsWith("http")) return path;
  return `${API_BASE}${path.startsWith("/") ? path : `/${path}`}`;
}
