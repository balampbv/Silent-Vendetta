import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import './Lobby.css';

const Lobby: React.FC = () => {
  const [playerName, setPlayerName] = useState('');
  const [gameId, setGameId] = useState('');
  const [showJoinGame, setShowJoinGame] = useState(false);
  const [createdGameId, setCreatedGameId] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);
  const navigate = useNavigate();

  // Typing animation effect
  const [titleText, setTitleText] = useState('');
  const fullTitle = 'Silent Vendetta';
  
  useEffect(() => {
    let currentIndex = 0;
    const typingInterval = setInterval(() => {
      if (currentIndex <= fullTitle.length) {
        setTitleText(fullTitle.slice(0, currentIndex));
        currentIndex++;
      } else {
        clearInterval(typingInterval);
      }
    }, 150);

    return () => clearInterval(typingInterval);
  }, []);

  const createGame = async () => {
    setError(null);
    if (!playerName.trim()) {
      setError('Please enter your name');
      return;
    }
    
    try {
      const response = await fetch('http://localhost:3001/api/games', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ playerName }),
      });
      
      const data = await response.json();
      if (!response.ok) {
        throw new Error(data.error || 'Failed to create game');
      }
      
      setCreatedGameId(data.gameId);
      setTimeout(() => {
        navigate(`/game/${data.gameId}`, {
          state: { playerName, isHost: true }
        });
      }, 2000); // Give time to see the game ID
    } catch (error) {
      console.error('Error creating game:', error);
      setError(error instanceof Error ? error.message : 'Failed to create game');
    }
  };

  const joinGame = async () => {
    setError(null);
    setSuccessMessage(null);
    if (!playerName.trim() || !gameId.trim()) {
      setError('Please enter your name and game ID');
      return;
    }

    try {
      const response = await fetch(`http://localhost:3001/api/games/${gameId}/join`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ playerName }),
      });

      const data = await response.json();
      if (!response.ok) {
        throw new Error(data.error || 'Failed to join game');
      }

      // Show success message briefly before navigating
      setError(null);
      setCreatedGameId(null);
      setSuccessMessage(data.message || 'Successfully joined game!');

      // Navigate after a short delay
      setTimeout(() => {
        navigate(`/game/${gameId}`, {
          state: { playerName, isHost: false }
        });
      }, 1000);
    } catch (error) {
      console.error('Error joining game:', error);
      setError(error instanceof Error ? error.message : 'Failed to join game');
    }
  };

  return (
    <div className="lobby">
      <div className="background-overlay"></div>
      <div className="lobby-container">
        <div className="title-section">
          <h1 className="game-title"><span className="typing-text">{titleText}</span><span className="cursor">|</span></h1>
          <p className="game-subtitle">Trust no one. Survive the night.</p>
        </div>
        
        {error && <div className="error-message">{error}</div>}
        {successMessage && <div className="success-message">{successMessage}</div>}
        {createdGameId && (
          <div className="success-message">
            Game created! Your game code is: <span className="game-code">{createdGameId}</span>
            <br />
            Share this code with other players to join.
          </div>
        )}
        
        <div className="form-container">
          <div className="input-group">
            <label>Your Alias</label>
            <input
              type="text"
              placeholder="Enter your name"
              value={playerName}
              onChange={(e) => setPlayerName(e.target.value)}
              className="themed-input"
            />
          </div>

          {!showJoinGame ? (
            <>
              <div className="button-group">
                <button 
                  className="primary-button" 
                  onClick={createGame}
                >
                  Create New Game
                </button>
                <button 
                  className="secondary-button"
                  onClick={() => setShowJoinGame(true)}
                >
                  Join Existing Game
                </button>
              </div>
            </>
          ) : (
            <>
              <div className="input-group">
                <label>Game Code</label>
                <input
                  type="text"
                  placeholder="Enter game ID"
                  value={gameId}
                  onChange={(e) => setGameId(e.target.value)}
                  className="themed-input"
                />
              </div>
              <div className="button-group">
                <button 
                  className="primary-button" 
                  onClick={joinGame}
                >
                  Join Game
                </button>
                <button 
                  className="secondary-button"
                  onClick={() => setShowJoinGame(false)}
                >
                  Back
                </button>
              </div>
            </>
          )}
        </div>

        <div className="game-rules">
          <h3>Game Rules</h3>
          <ul>
            <li>🎭 Some players are secretly Mafia members</li>
            <li>🌙 During night phase, Mafia eliminates one player</li>
            <li>💭 During day phase, discuss and find the Mafia</li>
            <li>⚖️ Vote to eliminate suspected Mafia members</li>
            <li>🎯 Special roles: Detective can investigate, Medic can protect</li>
          </ul>
        </div>
      </div>
    </div>
  );
};

export default Lobby; 