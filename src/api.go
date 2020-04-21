package main

// Player holds all the relevant information about a specific player
// in a game room.
type Player struct {
	Name    string `json:"name"`
	IsOwner bool   `json:"isOwner,omitempty"`
}

// IncomingMessage holds any incoming websocket message.
type IncomingMessage struct {
	Action string      `json:"action"`
	Body   interface{} `json:"body"`
}

// JoinRequest is used by clients to officially join a game room.
type JoinRequest struct {
	RoomCode string `json:"roomCode"`
	Name     string `json:"name"`
}

// OutgoingMessage is any outgoing websockets message.
type OutgoingMessage struct {
	Event string      `json:"event"`
	Error string      `json:"error"`
	Body  interface{} `json:"body"`
}

// JoinedEvent is an event that is sent to a player
// when they join a game room.
type JoinedEvent struct {
	Players []Player `json:"players"`
}

// PlayerJoinedEvent is an event that is sent to all other players
// when a new player joins a game room.
type PlayerJoinedEvent struct {
	Name string `json:"name"`
}
