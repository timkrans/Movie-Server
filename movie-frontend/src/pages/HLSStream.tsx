import React, { useState, useEffect, useRef } from "react";
import { useParams } from "react-router-dom";
import "../HLSStream.css"; // Import external CSS file

const HLSStream: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const videoRef = useRef<HTMLVideoElement | null>(null); // Reference to the video element
  const [isFullscreen, setIsFullscreen] = useState(false);

  useEffect(() => {
    if (!id) {
      setError("Movie ID is missing.");
      setLoading(false);
    } else {
      setLoading(false);
    }
  }, [id]);

  const toggleFullscreen = async () => {
    if (videoRef.current) {
      if (document.fullscreenElement) {
        // If video is already fullscreen, exit fullscreen
        document.exitFullscreen();
        setIsFullscreen(false);
      } else {
        // Request fullscreen for the video element
        await videoRef.current.requestFullscreen();
        setIsFullscreen(true);
      }
    }
  };

  if (loading) {
    return <div className="loading-container">Loading...</div>;
  }

  if (error) {
    return <div className="error-container">{error}</div>;
  }

  return (
    <div className="hls-container">
      <div className="video-container">
        <video
          ref={videoRef}
          className="video"
          controls
          poster={`http://localhost:8080/movies/${id}/cover`} // Optional cover image
        >
          <source
            src={`http://localhost:8080/movies/${id}/hls`} // Ensure backend provides the correct HLS stream URL
            type="application/vnd.apple.mpegurl"
          />
          <p>Your browser does not support HLS streaming.</p>
        </video>
        <div className="video-title">HLS Stream for Movie {id}</div>
      </div>

      <button onClick={toggleFullscreen} className="fullscreen-btn">
        {isFullscreen ? "Exit Fullscreen" : "Go Fullscreen"}
      </button>
    </div>
  );
};

export default HLSStream;