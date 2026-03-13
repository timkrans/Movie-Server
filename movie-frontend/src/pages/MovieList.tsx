import React, { useEffect, useState } from "react";
import axios from "axios";
import { Link } from "react-router-dom";
import "../MovieList.css"; // Import the CSS for styling

// Define a type for the movie data
interface Movie {
  id: number;
  title: string;
  cover_image_file_path: string;
  hls_path: string;
}

const MoviesList: React.FC = () => {
  const [movies, setMovies] = useState<Movie[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    // Fetch movies from the backend
    axios
      .get<Movie[]>("http://localhost:8080/movies")
      .then((response) => {
        setMovies(response.data);
      })
      .catch((error) => {
        console.error("Error fetching movies:", error);
        setError("Failed to load movies. Please try again later.");
      })
      .finally(() => {
        setLoading(false);
      });
  }, []);

  if (loading) {
    return <div className="loading">Loading movies...</div>;
  }

  if (error) {
    return <div className="error">{error}</div>;
  }

  return (
    <div className="movie-list">
      <h1>Movie List</h1>
      <div className="movies-grid">
        {movies.map((movie) => (
          <div key={movie.id} className="movie-card">
            <Link to={`/movies/${movie.id}/hls`} className="movie-link">
              <img
                className="movie-thumbnail"
                src={`http://localhost:5000/${movie.cover_image_file_path}`}
                alt={movie.title}
              />
              <div className="movie-title">{movie.title}</div>
            </Link>
          </div>
        ))}
      </div>
    </div>
  );
};

export default MoviesList;