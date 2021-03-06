package api

type ActionT int
type PlayerT int

// Player holds all the relevant information about a specific player
// in a game room.
type Player struct {
	Name        string `json:"name"`
	IsRoomOwner bool   `json:"isRoomOwner,omitempty"`
}

// Team holds the information about a specific team.
type Team struct {
	Players     []*Player `json:"players"`
	CardIndices []int     `json:"cardIndices"`
}

// GameSettings holds all the relevant information about a game's
// settings.
type GameSettings struct {
	SingleGuesser bool `json:"singleGuesser"`
	UseTimer      bool `json:"useTimer"`
	TimerLength   int  `json:"timerLength"`
}

// MovePlayerRequest is used by the owner of a room to move a player
// from one team to another.
type MovePlayerRequest struct {
	PlayerName   string  `json:"playerName"`
	ToTeam       int     `json:"toTeam"`
	ToPlayerType PlayerT `json:"toPlayerType"`
}

// ChangeSettingsRequest is used by the owner of a room to change
// the game settings.
type ChangeSettingsRequest struct {
	Settings GameSettings `json:"settings"`
}

// StartTurnRequest is used by the current player to start their turn.
type StartTurnRequest struct {
	NumCards int `json:"numCards"`
}

// EndTurnRequest is used by the current team to
// guess the cards and end the turn.
type EndTurnRequest struct {
	CardGuessIndices []int `json:"cardGuessIndices"`
}

// CreatedGameEvent is an event that is sent to a player
// when they create a new game room.
type CreatedGameEvent struct {
	RoomCode string       `json:"roomCode"`
	GameType string       `json:"gameType"`
	Teams    []Team       `json:"teams"`
	Settings GameSettings `json:"settings"`
}

// UpdatedRoomEvent is an event that is sent to all players
// in a room whenever a change has been made to it (e.g. player joining,
// player switching teams, etc).
type UpdatedRoomEvent struct {
	GameType string       `json:"gameType"`
	Teams    []Team       `json:"teams"`
	Settings GameSettings `json:"settings"`
}

// UpdatedGameEvent is an event that is sent to all players
// playing a game whenever a change has been made to its state.
type UpdatedGameEvent struct {
	GameType string       `json:"gameType,omitempty"`
	Teams    []Team       `json:"teams,omitempty"`
	Settings GameSettings `json:"settings,omitempty"`

	State                string   `json:"state"`
	CurrentlyPlayingTeam int      `json:"currentlyPlayingTeam"`
	Cards                []string `json:"cards,omitempty"`
	SpymasterCardIndices []int    `json:"spymasterCardIndices,omitempty"`
	CardIndicesGuessed   []int    `json:"cardIndicesGuessed"`
	WinningTeam          *int     `json:"winningTeam,omitempty"`
}

const (
	// Codenames game actions
	ActionInvalid = iota
	ActionMovePlayer
	ActionChangeSettings
	ActionStartTurn
	ActionEndTurn
)

const (
	// Player types
	PlayerSpymaster = iota
	PlayerGuesser
)

var (
	// Action holds a map of action types to protocol string.
	Action = map[ActionT]string{
		ActionInvalid:        "invalid action",
		ActionMovePlayer:     "move-player",
		ActionChangeSettings: "change-settings",
		ActionStartTurn:      "start-turn",
		ActionEndTurn:        "end-turn",
	}

	// ActionLookup holds a reverse map of Action.
	ActionLookup = make(map[string]ActionT)
)

// Init is called on program startup.
func Init() {
	for actionType, action := range Action {
		ActionLookup[action] = actionType
	}
}
