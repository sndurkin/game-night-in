package main

// Player holds all the relevant information about a specific player
// in a game room.
type Player struct {
	Name        string `json:"name"`
	IsRoomOwner bool   `json:"isRoomOwner,omitempty"`
	WordsSet    bool   `json:"wordsSet,omitempty"`
}

// IncomingMessage holds any incoming websocket message.
type IncomingMessage struct {
	Action string      `json:"action"`
	Body   interface{} `json:"body"`
}

// CreateRoomRequest is used by clients to create a new game room.
type CreateRoomRequest struct {
	GameType string `json:"gameType"`
	Name     string `json:"name"`
}

// JoinRoomRequest is used by clients to officially join a game room.
type JoinRoomRequest struct {
	RoomCode string `json:"roomCode"`
	Name     string `json:"name"`
}

// SubmitWordsRequest is used by clients to submit words for the
// Fishbowl game.
type SubmitWordsRequest struct {
	Words []string `json:"words"`
}

// MovePlayerRequest is used by the owner of a room to move a player
// from one team to another.
type MovePlayerRequest struct {
	PlayerName string `json:"playerName"`
	FromTeam   int    `json:"fromTeam"`
	ToTeam     int    `json:"toTeam"`
}

// AddTeamRequest is used by the owner of a room to create a new
// team.
type AddTeamRequest struct{}

// StartGameRequest is used by the owner of a room to start the game.
type StartGameRequest struct{}

// StartTurnRequest is used by the current player to start their turn.
type StartTurnRequest struct{}

// OutgoingMessage is any outgoing websockets message.
type OutgoingMessage struct {
	Event string      `json:"event"`
	Error string      `json:"error,omitempty"`
	Body  interface{} `json:"body"`
}

// CreatedRoomEvent is an event that is sent to a player
// when they create a new game room.
type CreatedRoomEvent struct {
	RoomCode string `json:"roomCode"`
	Team     int    `json:"team"`
}

// UpdatedRoomEvent is an event that is sent to all players
// in a room whenever a change has been made to it (e.g. player joining,
// player switching teams, etc).
type UpdatedRoomEvent struct {
	Teams [][]Player `json:"teams"`
}

// UpdatedGameEvent is an event that is sent to all players
// playing a game whenever a change has been made to its state.
type UpdatedGameEvent struct {
	State                string `json:"state"`
	CurrentServerTime    uint   `json:"currentServerTime,omitempty"`
	TimerLength          int    `json:"timerLength,omitempty"`
	CardsLeftInRound     int    `json:"cardsLeftInRound"`
	CardsGuessedInTurn   int    `json:"cardsGuessedInTurn"`
	CurrentRound         int    `json:"currentRound"`
	CurrentPlayers       []int  `json:"currentPlayers"`
	CurrentlyPlayingTeam int    `json:"currentlyPlayingTeam"`
}
