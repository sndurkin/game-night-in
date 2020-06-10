package api

type ActionT int
type RoundT int

// Player holds all the relevant information about a specific player
// in a game room.
type Player struct {
	Name           string `json:"name"`
	IsRoomOwner    bool   `json:"isRoomOwner,omitempty"`
	WordsSubmitted bool   `json:"wordsSubmitted"`
}

// GameSettings holds all the relevant information about a game's
// settings.
type GameSettings struct {
	Rounds           []string `json:"rounds"`
	TimerLength      int      `json:"timerLength"`
	NumWordsRequired int      `json:"numWordsRequired"`
	MaxSkipsPerTurn  int      `json:"maxSkipsPerTurn"`
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
type RemoveTeamRequest struct{
	Team int `json:"team"`
}

// ChangeSettingsRequest is used by the owner of a room to change
// the game settings.
type ChangeSettingsRequest struct {
	Settings GameSettings `json:"settings"`
}

// StartTurnRequest is used by the current player to start their turn.
type StartTurnRequest struct{}

// ChangeCardRequest is used by the current player to
// either mark the card correct or skip and move to
// the next card.
type ChangeCardRequest struct {
	ChangeType string `json:"changeType"`
}

// CreatedGameEvent is an event that is sent to a player
// when they create a new game room.
type CreatedGameEvent struct {
	RoomCode string       `json:"roomCode"`
	GameType string       `json:"gameType"`
	Teams    [][]Player   `json:"teams"`
	//Settings GameSettings `json:"settings"`
}

// UpdatedRoomEvent is an event that is sent to all players
// in a room whenever a change has been made to it (e.g. player joining,
// player switching teams, etc).
type UpdatedRoomEvent struct {
	GameType string       `json:"gameType"`
	Teams    [][]Player   `json:"teams"`
	Settings GameSettings `json:"settings"`
}

// UpdatedGameEvent is an event that is sent to all players
// playing a game whenever a change has been made to its state.
type UpdatedGameEvent struct {
	GameType string       `json:"gameType"`
	Teams    [][]Player   `json:"teams,omitempty"`
	Settings GameSettings `json:"settings"`

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
}

const (
	// Fishbowl game actions
	ActionInvalid = iota
	ActionAddTeam
	ActionRemoveTeam
	ActionMovePlayer
	ActionChangeSettings
	ActionStartTurn
	ActionSubmitWords
	ActionChangeCard

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
		ActionAddTeam:        "add-team",
		ActionRemoveTeam:     "remove-team",
		ActionMovePlayer:     "move-player",
		ActionChangeSettings: "change-settings",
		ActionStartTurn:      "start-turn",
		ActionSubmitWords:    "submit-words",
		ActionChangeCard:     "change-card",
	}

	// ActionLookup holds a reverse map of Action.
	ActionLookup = make(map[string]ActionT)

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

// Init is called on program startup.
func Init() {
	for actionType, action := range Action {
		ActionLookup[action] = actionType
	}

	for roundType, round := range Round {
		RoundLookup[round] = roundType
	}
}
