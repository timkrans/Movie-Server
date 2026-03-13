import React, { useState, useEffect } from "react";
import { BrowserRouter as Router, Routes, Route, Link } from "react-router-dom";
import AddMovie from "./pages/AddMovie";
import MoviesList from "./pages/MovieList";
import HLSStream from "./pages/HLSStream";
import "./App.css";

const App: React.FC = () => {
  const [darkMode, setDarkMode] = useState<boolean>(false);

  useEffect(() => {
    // Check if the user prefers dark mode
    const prefersDarkMode = window.matchMedia("(prefers-color-scheme: dark)").matches;
    setDarkMode(prefersDarkMode);
  }, []);

  const toggleDarkMode = () => {
    setDarkMode(!darkMode);
  };

  useEffect(() => {
    // Apply dark mode class to body
    if (darkMode) {
      document.body.classList.add("dark-mode");
    } else {
      document.body.classList.remove("dark-mode");
    }
  }, [darkMode]);

  return (
    <Router>
      <div className="appContainer">
        <div className="appHeader">
          <h1>Movie App</h1>
          <button onClick={toggleDarkMode} className="darkModeToggle">
            {darkMode ? "Switch to Light Mode" : "Switch to Dark Mode"}
          </button>
        </div>
        <nav className="nav">
          <ul>
            <li><Link to="/">Home</Link></li>
            <li><Link to="/add-movie">Add Movie</Link></li>
          </ul>
        </nav>
        <Routes>
          <Route path="/" element={<MoviesList />} />
          <Route path="/add-movie" element={<AddMovie />} />
          <Route path="/movies/:id/hls" element={<HLSStream />} />
        </Routes>
      </div>
    </Router>
  );
};

export default App;