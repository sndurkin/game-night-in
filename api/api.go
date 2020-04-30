package api

type ActionT int
type EventT int

// Player holds all the relevant information about a specific player
// in a game room.
type Player struct {
	Name           string `json:"name"`
	IsRoomOwner    bool   `json:"isRoomOwner,omitempty"`
	WordsSubmitted bool   `json:"wordsSubmitted,omitempty"`
}

// IncomingMessage holds any incoming websocket message.
type IncomingMessage struct {
	Action string      `json:"action"`
	Body   interface{} `json:"body"`
}

// CreateGameRequest is used by clients to create a new game room.
type CreateGameRequest struct {
	GameType string `json:"gameType"`
	Name     string `json:"name"`
}

// JoinGameRequest is used by clients to officially join a game room.
type JoinGameRequest struct {
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

// ChangeCardRequest is used by the current player to
// either mark the card correct or skip and move to
// the next card.
type ChangeCardRequest struct {
	ChangeType string `json:"changeType"`
}

// OutgoingMessage is any outgoing websockets message.
type OutgoingMessage struct {
	Event string      `json:"event"`
	Error string      `json:"error,omitempty"`
	Body  interface{} `json:"body"`
}

// CreatedGameEvent is an event that is sent to a player
// when they create a new game room.
type CreatedGameEvent struct {
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
	State                 string     `json:"state"`
	CurrentServerTime     int64      `json:"currentServerTime,omitempty"`
	TimerLength           int        `json:"timerLength,omitempty"`
	LastCardGuessed       string     `json:"lastCardGuessed"`
	CurrentCard           string     `json:"currentCard,omitempty"`
	TotalNumCards         int        `json:"totalNumCards"`
	NumCardsLeftInRound   int        `json:"numCardsLeftInRound"`
	NumCardsGuessedInTurn int        `json:"numCardsGuessedInTurn"`
	TeamScoresByRound     [][]int    `json:"teamScoresByRound"`
	WinningTeam           int        `json:"winningTeam"`
	CurrentRound          int        `json:"currentRound"`
	CurrentPlayers        []int      `json:"currentPlayers"`
	CurrentlyPlayingTeam  int        `json:"currentlyPlayingTeam"`
	Teams                 [][]Player `json:"teams,omitempty"`
}

const (
	// General game actions
	ActionInvalid    ActionT = 0
	ActionCreateGame ActionT = 1
	ActionJoinGame   ActionT = 2
	ActionAddTeam    ActionT = 3
	ActionRemoveTeam ActionT = 4
	ActionMovePlayer ActionT = 5
	ActionStartGame  ActionT = 6
	ActionStartTurn  ActionT = 7

	// Fishbowl game actions
	ActionSubmitWords ActionT = 30
	ActionChangeCard  ActionT = 31

	EventInvalid     EventT = 0
	EventCreatedGame EventT = 1
	EventUpdatedRoom EventT = 2
	EventUpdatedGame EventT = 3
)

var (
	// Action holds a map of action types to protocol string.
	Action = map[ActionT]string{
		ActionInvalid:     "invalid action",
		ActionCreateGame:  "create-game",
		ActionJoinGame:    "join-game",
		ActionAddTeam:     "add-team",
		ActionRemoveTeam:  "remove-team",
		ActionMovePlayer:  "move-player",
		ActionStartGame:   "start-game",
		ActionStartTurn:   "start-turn",
		ActionSubmitWords: "submit-words",
		ActionChangeCard:  "change-card",
	}

	// Event holds a map of event types to protocol string.
	Event = map[EventT]string{
		EventInvalid:     "invalid event",
		EventCreatedGame: "created-game",
		EventUpdatedRoom: "updated-room",
		EventUpdatedGame: "updated-game",
	}

	// ActionLookup holds a reverse map of Action.
	ActionLookup = map[string]ActionT{
		"invalid action": ActionInvalid,
		"create-game":    ActionCreateGame,
		"join-game":      ActionJoinGame,
		"add-team":       ActionAddTeam,
		"remove-team":    ActionRemoveTeam,
		"move-player":    ActionMovePlayer,
		"start-game":     ActionStartGame,
		"start-turn":     ActionStartTurn,
		"submit-words":   ActionSubmitWords,
		"change-card":    ActionChangeCard,
	}
)
