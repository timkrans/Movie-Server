import React, { useState } from "react";
import axios from "axios";
import "../styles/AddMovie.css"; 

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
      //reset form inputs after submission
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
    <div className="add-movie-container">
      <h1 className="form-title">Add a New Movie</h1>
      <form onSubmit={handleSubmit} className="movie-form">
        <div className="form-group">
          <label htmlFor="title" className="form-label">Title:</label>
          <input
            id="title"
            type="text"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            className="form-input"
            required
          />
        </div>

        <div className="form-group">
          <label htmlFor="videoFile" className="form-label">Video File:</label>
          <input
            id="videoFile"
            type="file"
            onChange={(e) => setVideoFile(e.target.files ? e.target.files[0] : null)}
            className="form-input"
            required
          />
        </div>

        <div className="form-group">
          <label htmlFor="coverImage" className="form-label">Cover Image:</label>
          <input
            id="coverImage"
            type="file"
            onChange={(e) => setCoverImage(e.target.files ? e.target.files[0] : null)}
            className="form-input"
            required
          />
        </div>

        <button type="submit" className="submit-btn" disabled={loading}>
          {loading ? <span className="loading-text">Adding Movie...</span> : "Add Movie"}
        </button>
      </form>

      {error && <p className="error-message">{error}</p>}
    </div>
  );
};

export default AddMovie;