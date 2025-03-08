# Silent Vendetta 🎭  

A real-time multiplayer social deduction game where deception is your greatest weapon, and trust is a fragile illusion.  

> Note: 🤖 An experiment in game development using agentic AI.

## Project Structure

```
silent-vendetta/
├── frontend/          # React frontend application
│   ├── src/
│   │   ├── components/
│   │   ├── styles/
│   │   └── types/
│   └── package.json
├── cmd/              # Go backend application entry points
│   └── server/
├── pkg/              # Go backend packages
│   ├── game/
│   ├── websocket/
│   └── models/
└── go.mod           # Go module definition
```

## Game Overview 🎮  

In Silent Vendetta, players are divided into two main factions:  
- **Mafia**: Work secretly to eliminate villagers  
- **Villagers**: Must identify and eliminate the mafia  
- **Special Roles**: Detective (can investigate players) and Medic (can protect players)  

### Game Phases 🔄  

1. **Night Phase**: Mafia members choose a target to eliminate  
2. **Discussion Phase**: Players discuss and share information  
3. **Voting Phase**: Players vote to eliminate a suspected mafia member  

### Win Conditions 🏆  

- **Mafia Win**: When they outnumber the villagers  
- **Villagers Win**: When all mafia members are eliminated  

## Prerequisites

- Go 1.21 or higher
- Node.js 18 or higher
- npm 9 or higher

## Development Setup

### Backend Setup

1. Install Go dependencies:
   ```bash
   go mod download
   ```

2. Run the backend server:
   ```bash
   go run cmd/server/main.go
   ```

### Frontend Setup

1. Navigate to the frontend directory:
   ```bash
   cd frontend
   ```

2. Install dependencies:
   ```bash
   npm install
   ```

3. Start the development server:
   ```bash
   npm start
   ```

The frontend will be available at http://localhost:3000 and will proxy API requests to the backend at http://localhost:3001.

## Game Features ✨  

- Real-time multiplayer using WebSockets  
- Role-based gameplay mechanics  
- In-game chat system  
- Dynamic phase transitions  
- Voting system  
- Special role abilities  

## Tech Stack 🔧  

- **Frontend**: React with TypeScript  
- **Backend**: Go (Fiber framework)  
- **Real-time Communication**: WebSockets  
- **State Management**: Custom game state manager  

## Contributing 🤝  

1. Fork the repository  
2. Create your feature branch (`git checkout -b feature/amazing-feature`)  
3. Commit your changes (`git commit -m 'Add some amazing feature'`)  
4. Push to the branch (`git push origin feature/amazing-feature`)  
5. Open a Pull Request  

## License 📝  

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.  

---

## Features 🔥  
✅ Real-time multiplayer with WebSockets  
✅ Secret roles & hidden alliances  
✅ Strategic voting system  
✅ Interactive UI with engaging gameplay  

---

