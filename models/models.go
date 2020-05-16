package models

import (
	"time"
)


// Player holds all the data about a player.
type Player struct {
	client      *Client
	name        string
	room        *GameRoom
	isRoomOwner bool

	settings    *PlayerSettings
}

// Client is just a placeholder for the real client, which is set
// when the Player is first created.
type Client struct {}

// PlayerSettings holds game-specific data about a particular player.
type PlayerSettings struct {}

// Game holds the game-specific data and logic.
type Game struct {}

// GameRoom holds the data about a game room.
type GameRoom struct {
	roomCode            string
	gameType            string
	lastInteractionTime time.Time
	game                *Game
}
