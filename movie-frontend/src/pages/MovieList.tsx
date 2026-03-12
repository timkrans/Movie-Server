import React, { useEffect, useState } from "react";
import axios from "axios";
import { Link } from "react-router-dom";

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
    return <div>Loading movies...</div>;
  }

  if (error) {
    return <div style={{ color: "red" }}>{error}</div>;
  }

  return (
    <div>
      <h1>Movie List</h1>
      <ul>
        {movies.map((movie) => (
          <li key={movie.id}>
            <h2>{movie.title}</h2>
            <img
              src={`http://localhost:5000/${movie.cover_image_file_path}`}
              alt={movie.title}
              style={{ maxWidth: "200px", height: "auto" }}
            />
            <p>
              <Link to={`/movies/${movie.id}/stream`}>Stream</Link> |{" "}
              <Link to={`/movies/${movie.id}/hls`}>HLS Stream</Link>
            </p>
          </li>
        ))}
      </ul>
    </div>
  );
};

export default MoviesList;