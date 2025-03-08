import React, { useEffect, useState } from 'react';
import { useParams, useLocation, useNavigate } from 'react-router-dom';
import { Player, GameState, Phase, LocationState } from '../../types/game';
import './Game.css';

const Game: React.FC = () => {
  const { id: gameId } = useParams<{ id: string }>();
  const location = useLocation();
  const navigate = useNavigate();
  const state = location.state as LocationState;

  const [gameState, setGameState] = useState<GameState>({
    id: '',
    players: {},
    phase: 'waiting',
    round: 0,
    phaseEndTime: '',
    minPlayers: 4,
    maxPlayers: 10,
    mafiaCount: 2,
    timeRemaining: 0,
  });
  const [message, setMessage] = useState('');
  const [chat, setChat] = useState<string[]>([]);
  const [ws, setWs] = useState<WebSocket | null>(null);
  const [mafiaVotes, setMafiaVotes] = useState<{[key: string]: string}>({});
  const [timeRemaining, setTimeRemaining] = useState<number>(0);
  const [playerCount, setPlayerCount] = useState<number>(0);

  useEffect(() => {
    // If no player name is provided, redirect back to lobby
    if (!state?.playerName) {
      navigate('/');
      return;
    }

    const socket = new WebSocket(`ws://localhost:3001/ws/${gameId}`);

    socket.onopen = () => {
      console.log('Connected to game server');
      // Send player information when connected
      socket.send(JSON.stringify({
        type: 'join',
        data: {
          playerName: state?.playerName,
          isHost: state?.isHost
        }
      }));
    };

    socket.onmessage = (event) => {
      const data = JSON.parse(event.data);
      
      switch (data.type) {
        case 'gameState':
          console.log('Received raw game state data:', event.data);
          console.log('Parsed game state:', data.data);
          console.log('Players in received state:', data.data.players);
          Object.entries(data.data.players || {}).forEach(([id, player]: [string, any]) => {
            console.log(`Player ${id} full state:`, player);
            console.log(`Player ${id} IsAlive:`, player.isAlive);
          });
          
          setGameState(prevState => {
            const newState = {
              ...prevState,
              ...data.data,
              players: data.data.players || {}
            };
            // Set initial time remaining when receiving game state
            if (data.data.phaseEndTime) {
              const endTime = new Date(data.data.phaseEndTime).getTime();
              const now = new Date().getTime();
              setTimeRemaining(Math.max(0, Math.floor((endTime - now) / 1000)));
            }
            return newState;
          });
          break;
        case 'playerCount':
          setPlayerCount(data.data);
          break;
        case 'chat':
          setChat(prev => [...prev, data.data]);
          break;
        case 'error':
          console.error('Received error:', data.data);
          alert(data.data);
          break;
        case 'mafiaVote':
          if (data.data.voter && data.data.target) {
            setMafiaVotes(prev => ({
              ...prev,
              [data.data.voter]: data.data.target
            }));
          }
          break;
      }
    };

    socket.onclose = () => {
      console.log('Disconnected from game server');
    };

    setWs(socket);

    return () => {
      socket.close();
    };
  }, [gameId, state, navigate]);

  // Add timer effect
  useEffect(() => {
    let timer: NodeJS.Timeout | null = null;

    if (gameState.phase !== 'waiting' && gameState.phase !== 'gameover' && timeRemaining > 0) {
      timer = setInterval(() => {
        setTimeRemaining(prev => {
          const newTime = Math.max(0, prev - 1);
          
          // If time runs out, automatically advance to next phase
          if (newTime === 0 && gameState.phase !== 'waiting') {
            fetch(`http://localhost:3001/api/games/${gameId}/next-phase`, {
              method: 'POST',
            }).catch(error => {
              console.error('Error advancing phase:', error);
            });
          }
          
          return newTime;
        });
      }, 1000);
    }

    return () => {
      if (timer) {
        clearInterval(timer);
      }
    };
  }, [gameState.phase, timeRemaining, gameId]);

  const sendMessage = () => {
    if (!message.trim() || !ws) return;

    ws.send(JSON.stringify({
      type: 'chat',
      data: message,
    }));

    setMessage('');
  };

  const castVote = (playerId: string) => {
    if (!ws) return;

    ws.send(JSON.stringify({
      type: 'vote',
      data: playerId,
    }));
  };

  const copyGameId = () => {
    navigator.clipboard.writeText(gameId || '');
    alert('Game ID copied to clipboard!');
  };

  const startGame = () => {
    if (!ws) return;
    
    fetch(`http://localhost:3001/api/games/${gameId}/start`, {
      method: 'POST',
    })
    .then(response => {
      if (!response.ok) {
        return response.json().then(data => {
          throw new Error(data.error || 'Failed to start game');
        });
      }
    })
    .catch(error => {
      console.error('Error starting game:', error);
      alert(error.message || 'Failed to start game. Make sure there are enough players.');
    });
  };

  const handleMafiaAction = (targetId: string) => {
    if (!ws) return;
    ws.send(JSON.stringify({
      type: "mafiaAction",
      data: targetId,
    }));
  };

  const isCurrentPlayerMafia = () => {
    const currentPlayer = Object.values(gameState.players).find(p => 
      p.name.toLowerCase() === (state?.playerName || '').toLowerCase()
    );
    return currentPlayer?.role === 'mafia';
  };

  const getMafiaVoteCount = (targetId: string) => {
    return Object.values(mafiaVotes).filter(id => id === targetId).length;
  };

  const getPhaseInfo = () => {
    switch (gameState.phase) {
      case 'night':
        return {
          name: 'Night Phase',
          description: isCurrentPlayerMafia() ? 
            'Select your target to eliminate' : 
            'The mafia is choosing their target...',
          duration: 30,
          color: '#2c3e50',
          bgColor: '#34495e'
        };
      case 'discuss':
        return {
          name: 'Discussion Phase',
          description: 'Discuss who might be the mafia!',
          duration: 120,
          color: '#27ae60',
          bgColor: '#2ecc71'
        };
      case 'vote':
        return {
          name: 'Voting Phase',
          description: 'Vote for who you think is the mafia!',
          duration: 30,
          color: '#c0392b',
          bgColor: '#e74c3c'
        };
      default:
        return {
          name: gameState.phase,
          description: '',
          duration: 0,
          color: '#7f8c8d',
          bgColor: '#95a5a6'
        };
    }
  };

  // Calculate progress percentage
  const getProgress = () => {
    const phaseInfo = getPhaseInfo();
    return ((phaseInfo.duration - timeRemaining) / phaseInfo.duration) * 100;
  };

  // Get timer class based on remaining time
  const getTimerClass = () => {
    const percentage = (timeRemaining / getPhaseInfo().duration) * 100;
    if (percentage <= 25) return 'urgent';
    if (percentage <= 50) return 'warning';
    return '';
  };

  const advancePhase = () => {
    if (!ws) return;
    
    fetch(`http://localhost:3001/api/games/${gameId}/next-phase`, {
      method: 'POST',
    })
    .catch(error => {
      console.error('Error advancing phase:', error);
      alert('Failed to advance phase');
    });
  };

  return (
    <div className="game">
      <div className="game-header">
        <div className="game-info">
          {gameState.phase === 'waiting' && (
            <div className="waiting-info">
              <h2>Waiting for Players</h2>
              <p className="player-count">
                Players: {playerCount} / {gameState.maxPlayers}
                <span className="min-players">(Minimum {gameState.minPlayers} required)</span>
              </p>
              <div className="waiting-actions">
                <button 
                  className="return-home-button"
                  onClick={() => navigate('/')}
                >
                  ← Return to Home
                </button>
                {state?.isHost && (
                  <button 
                    className="start-game-button"
                    onClick={startGame}
                    disabled={playerCount < gameState.minPlayers}
                  >
                    Start Game
                  </button>
                )}
              </div>
            </div>
          )}
          {gameState.phase !== 'waiting' && gameState.phase !== 'gameover' && (
            <div 
              className="phase-info"
              style={{
                backgroundColor: getPhaseInfo().bgColor,
                borderColor: getPhaseInfo().color
              }}
            >
              <h2>{getPhaseInfo().name}</h2>
              <p className="phase-description">{getPhaseInfo().description}</p>
              <div className={`timer ${getTimerClass()}`}>
                <div className="progress-bar-container">
                  <div 
                    className="progress-bar" 
                    style={{
                      width: `${getProgress()}%`,
                      backgroundColor: getPhaseInfo().color
                    }}
                  />
                </div>
                <div className="timer-text">
                  Time Remaining: <span className="highlight">{timeRemaining}s</span>
                  <span className="duration">/{getPhaseInfo().duration}s</span>
                </div>
              </div>
              {state?.isHost && timeRemaining === 0 && (
                <button 
                  className="advance-phase-button"
                  onClick={advancePhase}
                >
                  Next Phase →
                </button>
              )}
            </div>
          )}
          {gameState.phase === 'gameover' && (
            <h2>Game Over!</h2>
          )}
        </div>
        <div className="player-info">
          <div className="current-player">
            Playing as: <span className="highlight">{state?.playerName}</span>
            {state?.isHost && <span className="host-badge">Host</span>}
            {gameState.phase !== 'waiting' && (
              <div className="role-info">
                Role: <span className="highlight">
                  {(() => {
                    const currentPlayer = Object.values(gameState.players).find(p => 
                      p.name.toLowerCase() === (state?.playerName || '').toLowerCase()
                    );
                    return currentPlayer?.role || 'Unknown';
                  })()}
                </span>
              </div>
            )}
          </div>
          <div className="game-id" onClick={copyGameId} title="Click to copy">
            Game ID: <span className="highlight">{gameId}</span>
            <span className="copy-icon">📋</span>
          </div>
        </div>
      </div>

      <div className="game-container">
        <div className="players-list">
          <h3>Players in Game</h3>
          {gameState.phase === 'waiting' && state?.isHost && (
            <div className="start-game-section">
              <p className="player-count">
                Players: {Object.keys(gameState.players).length} 
                <span className="min-players">(Minimum {gameState.minPlayers} required)</span>
              </p>
              <button 
                className="start-game-button"
                onClick={startGame}
                disabled={Object.keys(gameState.players).length < gameState.minPlayers}
              >
                Start Game
              </button>
            </div>
          )}
          {gameState.phase === 'night' && isCurrentPlayerMafia() && (
            <div className="mafia-instructions">
              <p>Select a target to eliminate:</p>
              <p className="mafia-votes-info">
                Mafia votes needed: {Math.ceil((Object.values(gameState.players)
                  .filter(p => p.isAlive && p.role === 'mafia').length + 1) / 2)}
              </p>
            </div>
          )}
          <div className="players-status-info">
            <span className="status-indicator alive">● Alive</span>
            <span className="status-indicator dead">● Dead</span>
          </div>
          {Object.values(gameState.players).map((player) => (
            <div
              key={player.id}
              className={`player-card ${!player.isAlive ? 'dead' : 'alive'} 
                ${gameState.phase === 'night' && isCurrentPlayerMafia() ? 'mafia-selecting' : ''}`}
              onClick={() => {
                if (gameState.phase === 'night' && isCurrentPlayerMafia()) {
                  handleMafiaAction(player.id);
                } else if (gameState.phase === 'vote') {
                  castVote(player.id);
                }
              }}
            >
              <div className="player-info-wrapper">
                <span className={`status-dot ${player.isAlive ? 'alive' : 'dead'}`} />
                <span className="player-name">{player.name}</span>
              </div>
              <div className="player-badges">
                {player.isHost && <span className="host-badge">Host</span>}
                <span className={`status-badge ${player.isAlive ? 'alive' : 'dead'}`}>
                  {player.isAlive ? 'Alive' : 'Dead'}
                </span>
                {gameState.phase === 'night' && isCurrentPlayerMafia() && getMafiaVoteCount(player.id) > 0 && (
                  <span className="mafia-votes">
                    Votes: {getMafiaVoteCount(player.id)}
                  </span>
                )}
              </div>
            </div>
          ))}
        </div>

        <div className="chat-container">
          <div className="chat-messages">
            {chat.map((msg, index) => (
              <div key={index} className="chat-message">
                {msg}
              </div>
            ))}
          </div>
          <div className="chat-input">
            <input
              type="text"
              value={message}
              onChange={(e) => setMessage(e.target.value)}
              onKeyPress={(e) => e.key === 'Enter' && sendMessage()}
              placeholder="Type your message..."
            />
            <button onClick={sendMessage}>Send</button>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Game; 