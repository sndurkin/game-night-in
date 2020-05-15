package api

type ActionT int
type EventT int
type RoundT int

// Player holds all the relevant information about a specific player
// in a game room.
type FishbowlPlayer struct {
	Name           string `json:"name"`
	IsRoomOwner    bool   `json:"isRoomOwner,omitempty"`
	WordsSubmitted bool   `json:"wordsSubmitted,omitempty"`
}

// GameSettings holds all the relevant information about a game's
// settings.
type FishbowlGameSettings struct {
	Rounds      []string `json:"rounds"`
	TimerLength int      `json:"timerLength"`
}

// SubmitWordsRequest is used by clients to submit words for the
// Fishbowl game.
type SubmitWordsRequest struct {
	Words []string `json:"words"`
}

// StartTurnRequest is used by the current player to start their turn.
type StartTurnRequest struct{}

// ChangeCardRequest is used by the current player to
// either mark the card correct or skip and move to
// the next card.
type ChangeCardRequest struct {
	ChangeType string `json:"changeType"`
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
	FishbowlActionInvalid ActionT = iota
	FishbowlActionStartTurn
	FishbowlActionSubmitWords
	FishbowlActionChangeCard
)

const (
	// FishbowlRound types
	FishbowlRoundInvalid RoundT = iota
	FishbowlRoundDescribe
	FishbowlRoundSingleWord
	FishbowlRoundCharades
)

var (
	// FishbowlAction holds a map of action types to protocol string.
	FishbowlAction = map[FishbowlActionT]string{
		FishbowlActionInvalid:        "invalid action",
		FishbowlActionStartTurn:      "start-turn",
		FishbowlActionSubmitWords:    "submit-words",
		FishbowlActionChangeCard:     "change-card",
	}

	// FishbowlRound holds a map of round types to protocol string.
	FishbowlRound = map[FishbowlRoundT]string{
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
