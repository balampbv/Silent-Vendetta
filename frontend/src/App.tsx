import React from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import Game from './components/game/Game';
import Lobby from './components/lobby/Lobby';
import './styles/app.css';

const App: React.FC = () => {
  return (
    <Router>
      <div className="App">
        <header className="App-header">
          <h1>Silent Vendetta</h1>
        </header>
        <main>
          <Routes>
            <Route path="/" element={<Lobby />} />
            <Route path="/game/:id" element={<Game />} />
          </Routes>
        </main>
      </div>
    </Router>
  );
};

export default App; 