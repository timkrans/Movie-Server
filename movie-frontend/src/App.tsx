import React from "react";
import { BrowserRouter as Router, Routes, Route, Link } from "react-router-dom";
import AddMovie from "./pages/AddMovie";
import MoviesList from "./pages/MovieList";
import StreamMovie from "./pages/StreamMovie";
import HLSStream from "./pages/HLSStream";

const App: React.FC = () => {
  return (
    <Router>
      <div>
        <h1>Movie App</h1>
        <nav>
          <ul>
            <li><Link to="/">Home</Link></li>
            <li><Link to="/add-movie">Add Movie</Link></li>
          </ul>
        </nav>
        <Routes>
          <Route path="/" element={<MoviesList />} />
          <Route path="/add-movie" element={<AddMovie />} />
          <Route path="/movies/:id/stream" element={<StreamMovie />} />
          <Route path="/movies/:id/hls" element={<HLSStream />} />
        </Routes>
      </div>
    </Router>
  );
};

export default App;