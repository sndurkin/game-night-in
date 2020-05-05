package api

type ActionT int
type EventT int
type RoundT int

// Player holds all the relevant information about a specific player
// in a game room.
type Player struct {
	Name           string `json:"name"`
	IsRoomOwner    bool   `json:"isRoomOwner,omitempty"`
	WordsSubmitted bool   `json:"wordsSubmitted,omitempty"`
}

// GameSettings holds all the relevant information about a game's
// settings.
type GameSettings struct {
	Rounds      []string `json:"rounds"`
	TimerLength int      `json:"timerLength"`
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

// AddTeamRequest is used by the owner of a room to add a new
// team to the game.
type AddTeamRequest struct{}

// RemoveTeamRequest is used by the owner of a room to remove a
// team from the game.
type RemoveTeamRequest struct{}

// ChangeSettingsRequest is used by the owner of a room to change
// the game settings.
type ChangeSettingsRequest struct {
	Settings GameSettings `json:"settings"`
}

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

// RematchRequest is used by the owner of a room to restart
// everything.
type RematchRequest struct{}

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
	Teams    [][]Player   `json:"teams"`
	Settings GameSettings `json:"settings"`
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
	WinningTeam           *int       `json:"winningTeam,omitempty"`
	CurrentRound          int        `json:"currentRound"`
	CurrentPlayers        []int      `json:"currentPlayers"`
	CurrentlyPlayingTeam  int        `json:"currentlyPlayingTeam"`
	Teams                 [][]Player `json:"teams,omitempty"`
}

const (
	// General game actions
	ActionInvalid ActionT = iota
	ActionCreateGame
	ActionJoinGame
	ActionAddTeam
	ActionRemoveTeam
	ActionMovePlayer
	ActionChangeSettings
	ActionStartGame
	ActionStartTurn
	ActionRematch

	// Fishbowl game actions
	ActionSubmitWords
	ActionChangeCard
)

const (
	// Event types
	EventInvalid EventT = iota
	EventCreatedGame
	EventUpdatedRoom
	EventUpdatedGame
)

const (
	// Round types
	RoundInvalid RoundT = iota
	RoundDescribe
	RoundSingleWord
	RoundCharades
)

var (
	// Action holds a map of action types to protocol string.
	Action = map[ActionT]string{
		ActionInvalid:        "invalid action",
		ActionCreateGame:     "create-game",
		ActionJoinGame:       "join-game",
		ActionAddTeam:        "add-team",
		ActionRemoveTeam:     "remove-team",
		ActionMovePlayer:     "move-player",
		ActionChangeSettings: "change-settings",
		ActionStartGame:      "start-game",
		ActionStartTurn:      "start-turn",
		ActionRematch:        "rematch",
		ActionSubmitWords:    "submit-words",
		ActionChangeCard:     "change-card",
	}

	// ActionLookup holds a reverse map of Action.
	ActionLookup = make(map[string]ActionT)

	// Event holds a map of event types to protocol string.
	Event = map[EventT]string{
		EventInvalid:     "invalid event",
		EventCreatedGame: "created-game",
		EventUpdatedRoom: "updated-room",
		EventUpdatedGame: "updated-game",
	}

	// Round holds a map of round types to protocol string.
	Round = map[RoundT]string{
		RoundInvalid:    "invalid round",
		RoundDescribe:   "describe",
		RoundSingleWord: "single",
		RoundCharades:   "charades",
	}

	// RoundLookup holds a reverse map of Round.
	RoundLookup = make(map[string]RoundT)
)

func Init() {
	for actionType, action := range Action {
		ActionLookup[action] = actionType
	}

	for roundType, round := range Round {
		RoundLookup[round] = roundType
	}
}
