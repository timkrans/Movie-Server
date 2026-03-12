import React, { useState } from "react";
import axios from "axios";

const AddMovie: React.FC = () => {
  const [title, setTitle] = useState<string>("");
  const [videoFile, setVideoFile] = useState<File | null>(null);
  const [coverImage, setCoverImage] = useState<File | null>(null);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!videoFile || !coverImage) {
      alert("Please select both video and cover image.");
      return;
    }

    const formData = new FormData();
    formData.append("title", title);
    formData.append("video", videoFile);
    formData.append("cover_image", coverImage);

    setLoading(true);
    setError(null);

    try {
      const response = await axios.post("http://localhost:8080/movies", formData, {
        headers: {
          "Content-Type": "multipart/form-data",
        },
      });
      console.log("Movie added:", response.data);
      alert("Movie added successfully!");
      // Reset form inputs after submission
      setTitle("");
      setVideoFile(null);
      setCoverImage(null);
    } catch (error) {
      console.error("Error adding movie:", error);
      setError("Failed to add movie. Please try again.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div>
      <h1>Add a New Movie</h1>
      <form onSubmit={handleSubmit}>
        <label>
          Title:
          <input
            type="text"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            required
          />
        </label>
        <br />
        <label>
          Video File:
          <input
            type="file"
            onChange={(e) => setVideoFile(e.target.files ? e.target.files[0] : null)}
            required
          />
        </label>
        <br />
        <label>
          Cover Image:
          <input
            type="file"
            onChange={(e) => setCoverImage(e.target.files ? e.target.files[0] : null)}
            required
          />
        </label>
        <br />
        <button type="submit" disabled={loading}>
          {loading ? "Adding Movie..." : "Add Movie"}
        </button>
      </form>

      {error && <p style={{ color: "red" }}>{error}</p>}
    </div>
  );
};

export default AddMovie;