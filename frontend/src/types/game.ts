export interface Player {
  id: string;
  name: string;
  isAlive: boolean;
  isHost: boolean;
  role?: string;
  votedFor?: string;
}

export interface GameState {
  id: string;
  players: { [key: string]: Player };
  phase: Phase;
  round: number;
  phaseEndTime: string;
  minPlayers: number;
  maxPlayers: number;
  mafiaCount: number;
  timeRemaining: number;
}

export type Phase = 'waiting' | 'night' | 'discuss' | 'vote' | 'gameover';

export interface LocationState {
  playerName: string;
  isHost: boolean;
} 