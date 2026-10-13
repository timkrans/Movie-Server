import React, { useCallback, useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { api, setActiveProfileId, type Profile } from "../api";
import "./ProfilePicker.css";

const COLORS = ["#e50914", "#1a73e8", "#34a853", "#f9ab00", "#a142f4", "#ff6d00", "#00acc1", "#c2185b"];

type Mode = "pick" | "create" | "edit";

const ProfilePicker: React.FC = () => {
  const navigate = useNavigate();
  const [profiles, setProfiles] = useState<Profile[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [mode, setMode] = useState<Mode>("pick");
  const [managing, setManaging] = useState(false);
  const [editId, setEditId] = useState<number | null>(null);
  const [name, setName] = useState("");
  const [color, setColor] = useState(COLORS[0]);

  const load = useCallback(async () => {
    try {
      const { data } = await api.get<Profile[]>("/profiles");
      setProfiles(data);
      setError(null);
    } catch {
      setError("Could not load profiles.");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void load();
  }, [load]);

  const select = (id: number) => {
    setActiveProfileId(id);
    navigate("/", { replace: true });
  };

  const startCreate = () => {
    setName("");
    setColor(COLORS[profiles.length % COLORS.length]);
    setEditId(null);
    setMode("create");
  };

  const startEdit = (p: Profile) => {
    setName(p.name);
    setColor(p.avatar_color || COLORS[0]);
    setEditId(p.id);
    setMode("edit");
  };

  const save = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) return;
    try {
      if (mode === "edit" && editId != null) {
        await api.put(`/profiles/${editId}`, { name: name.trim(), avatar_color: color });
      } else {
        const { data } = await api.post<Profile>("/profiles", {
          name: name.trim(),
          avatar_color: color,
        });
        setActiveProfileId(data.id);
        navigate("/", { replace: true });
        return;
      }
      setMode("pick");
      setManaging(false);
      await load();
    } catch {
      setError("Could not save profile.");
    }
  };

  const remove = async (id: number) => {
    if (!window.confirm("Delete this profile and its watch history?")) return;
    try {
      await api.delete(`/profiles/${id}`);
      await load();
    } catch {
      setError("Could not delete profile.");
    }
  };

  if (loading) {
    return <div className="profiles"><p className="profiles__lead">Loading…</p></div>;
  }

  if (mode === "create" || mode === "edit") {
    return (
      <div className="profiles">
        <h1 className="profiles__title">{mode === "edit" ? "Edit profile" : "Add profile"}</h1>
        <form className="profiles__form" onSubmit={save}>
          <label htmlFor="profile-name">Name</label>
          <input
            id="profile-name"
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            required
            autoFocus
          />
          <label>Color</label>
          <div className="profiles__colors">
            {COLORS.map((c) => (
              <button
                key={c}
                type="button"
                className={`profiles__color ${color === c ? "is-selected" : ""}`}
                style={{ background: c }}
                onClick={() => setColor(c)}
                aria-label={c}
              />
            ))}
          </div>
          <div className="profiles__form-actions">
            <button type="button" onClick={() => setMode("pick")}>
              Cancel
            </button>
            <button type="submit">Save</button>
          </div>
        </form>
        {error && <p className="profiles__error">{error}</p>}
      </div>
    );
  }

  return (
    <div className="profiles">
      <h1 className="profiles__title">Who&apos;s watching?</h1>
      {profiles.length === 0 && (
        <p className="profiles__lead">Create a profile to get started.</p>
      )}
      <div className="profiles__grid">
        {profiles.map((p) => (
          <div key={p.id} className="profile-card" style={{ display: "flex" }}>
            <button type="button" className="profile-card" onClick={() => !managing && select(p.id)}>
              <div
                className="profile-card__avatar"
                style={{ background: p.avatar_color || COLORS[0] }}
              >
                {p.name.slice(0, 1).toUpperCase()}
              </div>
              <span className="profile-card__name">{p.name}</span>
            </button>
            {managing && (
              <div className="profile-card__manage">
                <button type="button" onClick={() => startEdit(p)}>Edit</button>
                <button type="button" onClick={() => void remove(p.id)}>Delete</button>
              </div>
            )}
          </div>
        ))}
        <button type="button" className="profile-card profile-card--add" onClick={startCreate}>
          <div className="profile-card__avatar">+</div>
          <span className="profile-card__name">Add</span>
        </button>
      </div>
      {profiles.length > 0 && (
        <button
          type="button"
          className="profiles__actions"
          onClick={() => setManaging((m) => !m)}
        >
          {managing ? "Done" : "Manage profiles"}
        </button>
      )}
      {error && <p className="profiles__error">{error}</p>}
    </div>
  );
};

export default ProfilePicker;
