import React, { useState, useEffect } from "react";
import { useParams } from "react-router-dom";

const HLSStream: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const [loading, setLoading] = useState<boolean>(true);
  const [error] = useState<string | null>(null);

  useEffect(() => {
    if (id) {
      setLoading(false); // No need to fetch data again since the `id` is already available
    }
  }, [id]);

  if (loading) {
    return <div>Loading HLS stream...</div>;
  }

  if (error) {
    return <div style={{ color: "red" }}>{error}</div>;
  }

  return (
    <div style={{ textAlign: "center", margin: "20px" }}>
      <h1>HLS Stream for Movie {id}</h1>
      <video
        controls
        style={{
          width: "100%",
          height: "auto", // Keeps the aspect ratio of the video
          maxHeight: "100vh", // Prevents the video from becoming larger than the viewport
        }}
        poster={`http://localhost:8080/movies/${id}/cover`} // Optional: Display a cover image until the video starts playing
        >
        <source
          src={`http://localhost:8080/movies/${id}/hls`} // Make sure the backend generates the correct stream URL
          type="application/vnd.apple.mpegurl"
        />
        <p>Your browser does not support HLS streaming.</p>
      </video>
    </div>
  );
};

export default HLSStream;