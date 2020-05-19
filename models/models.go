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

// PlayerSettings holds game-specific data about a particular player.
type PlayerSettings struct {}

// Game holds the game-specific data and logic.
type Game interface {
	HandleIncomingMessage(
		player *Player,
		incomingMessage api.IncomingMessage,
		body json.RawMessage,
		sendOutgoingMessages OutgoingMessageRequestFn,
	)
	Start(
		player *Player,
		sendOutgoingMessages OutgoingMessageRequestFn,
	)
	AddPlayer(
		player *Player,
		sendOutgoingMessages OutgoingMessageRequestFn,
	)
	Join(
		player *Player,
		newPlayerJoined bool,
		req api.JoinGameRequest,
		sendOutgoingMessages OutgoingMessageRequestFn,
	)
	Rematch(
		player *Player,
		sendOutgoingMessages OutgoingMessageRequestFn,
	)
}

type OutgoingMessageRequestFn func(*OutgoingMessageRequest)

// GameRoom holds the data about a game room.
type GameRoom struct {
	RoomCode            string
	GameType            string
	LastInteractionTime time.Time
	Game                Game
	Players             []*Player
}

// OutgoingMessageRequest is used by game-specific handlers to
// construct outgoing messages to
type OutgoingMessageRequest struct {
	PrimaryClient interface{}
	PrimaryMsg    *api.OutgoingMessage
	SecondaryMsg  *api.OutgoingMessage
	Room          *GameRoom
}
