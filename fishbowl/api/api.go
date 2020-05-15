package api

type ActionT int
type RoundT int

const (
	// Fishbowl game actions
	ActionInvalid = iota
	ActionAddTeam
	ActionRemoveTeam
	ActionMovePlayer
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

func Init() {
	for actionType, action := range Action {
		ActionLookup[action] = actionType
	}

	for roundType, round := range Round {
		RoundLookup[round] = roundType
	}
}
