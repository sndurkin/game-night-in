package api

type ActionT int
type EventT int

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

// StartGameRequest is used by the owner of a room to start the game.
type StartGameRequest struct{}

// RematchRequest is used by the owner of a room to restart
// everything.
type RematchRequest struct{}

// OutgoingMessage is any outgoing websockets message.
type OutgoingMessage struct {
	Event string      `json:"event"`
	Error string      `json:"error,omitempty"`
	Body  interface{} `json:"body"`
}

// UpdatedGameEvent
type UpdatedGameEvent struct{}

const (
	// General game actions
	ActionInvalid ActionT = iota
	ActionCreateGame
	ActionJoinGame
	ActionStartGame
	ActionRematch
)

const (
	// Event types
	EventInvalid EventT = iota
	EventCreatedGame
	EventUpdatedRoom
	EventUpdatedGame
)

var (
	// Action holds a map of action types to protocol string.
	Action = map[ActionT]string{
		ActionInvalid:    "invalid action",
		ActionCreateGame: "create-game",
		ActionJoinGame:   "join-game",
		ActionStartGame:  "start-game",
		ActionRematch:    "rematch",
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
)

// Init is called on program startup.
func Init() {
	for actionType, action := range Action {
		ActionLookup[action] = actionType
	}
}
