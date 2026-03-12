import React, { useEffect, useState } from "react";
import { useParams } from "react-router-dom";

const StreamMovie: React.FC = () => {
  const { id } = useParams<{ id: string }>(); // Grab the ID directly from the URL
  const [] = useState<any>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    // Same here, the ID is already available, no need to fetch it again
    if (id) {
      setLoading(false); // Assuming the movie is available for streaming based on ID
      // Your stream logic can be handled here
    }
  }, [id]);

  if (loading) {
    return <div>Loading video...</div>;
  }

  if (error) {
    return <div style={{ color: "red" }}>{error}</div>;
  }

  return (
    <div>
      <h1>Stream Movie {id}</h1>
      <video controls style={{ width: "100%" }}>
        <source
          src={`http://localhost:8080/movies/${id}/stream`} 
          type="video/mp4"
        />
        <p>Your browser does not support the video tag.</p>
      </video>
    </div>
  );
};

export default StreamMovie;