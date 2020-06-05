package models

import (
	"encoding/json"
	"time"

	"github.com/sndurkin/game-night-in/api"
)


// Player holds all the data about a player.
type Player struct {
	Client      interface{}
	Name        string
	Room        *GameRoom
	IsRoomOwner bool
}

// Game holds the game-specific data and logic.
type Game interface {
	HandleIncomingMessage(
		player *Player,
		incomingMessage api.IncomingMessage,
		body json.RawMessage,
	)
	Start(player *Player)
	AddPlayer(player *Player)
	Join(
		player *Player,
		newPlayerJoined bool,
		req api.JoinGameRequest,
	)
	Kick(playerName string)
	Rematch(player *Player)
}

type ErrorMessageRequestFn func(*ErrorMessageRequest)
type OutgoingMessageRequestFn func(*OutgoingMessageRequest)

// GameRoom holds the data about a game room.
type GameRoom struct {
	RoomCode            string
	GameType            string
	LastInteractionTime time.Time
	Game                Game
	Players             []*Player
}

// ErrorMessageRequest is used by game-specific handlers to
// construct an error message to 1 client.
type ErrorMessageRequest struct {
	Player   *Player
	Fatal bool
	Error string
}

// OutgoingMessageRequest is used by game-specific handlers to
// construct outgoing messages to clients.
type OutgoingMessageRequest struct {
	PrimaryClient interface{}
	PrimaryMsg    *api.OutgoingMessage
	SecondaryMsg  *api.OutgoingMessage
	Room          *GameRoom
}
