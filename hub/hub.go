package hub

import (
	"errors"
	"encoding/json"
	"log"
	"math/rand"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/sndurkin/game-night-in/api"
	"github.com/sndurkin/game-night-in/fishbowl"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	sync.RWMutex

	// Map of connected client to Player
	playerClients map[*Client]*Player

	// Map of room code to GameRoom
	rooms map[string]*GameRoom

	// Inbound messages from the clients.
	message chan *ClientMessage

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

// Player holds all the data about a player.
type Player struct {
	client      *Client
	name        string
	room        *GameRoom
	isRoomOwner bool

	settings    *PlayerSettings
}

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

// NewHub creates a new Hub instance which manages all incoming
// websocket messages.
func NewHub() *Hub {
	return &Hub{
		playerClients: make(map[*Client]*Player),
		rooms:         make(map[string]*GameRoom),
		message:       make(chan *ClientMessage),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
	}
}

func (h *Hub) newGame(gameType string) *Game {
	switch (gameType) {
	case "fishbowl":
		return fishbowl.NewGame()
	}

	return nil
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.Lock()
			h.playerClients[client] = &Player{
				client: client,
			}
			h.Unlock()
		case client := <-h.unregister:
			h.Lock()
			if _, ok := h.playerClients[client]; ok {
				delete(h.playerClients, client)
				close(client.send)
			}
			h.Unlock()
		case clientMessage := <-h.message:
			h.handleIncomingMessage(clientMessage)
		}
	}
}

func (h *Hub) runRoomCleanup() {
	for now := range time.Tick(time.Hour) {
		h.Lock()

		roomCodes := []string{}
		for roomCode, room := range h.rooms {
			expiryTime := room.lastInteractionTime.Add(
				time.Minute * time.Duration(30))
			if now.After(expiryTime) {
				roomCodes = append(roomCodes, roomCode)
			}
		}

		if len(roomCodes) > 0 {
			log.Printf("Cleaning up %d old rooms\n", len(roomCodes))

			for _, roomCode := range roomCodes {
				delete(h.rooms, roomCode)
			}
		}

		h.Unlock()
	}
}

func (h *Hub) handleIncomingMessage(clientMessage *ClientMessage) {
	var body json.RawMessage
	incomingMessage := api.IncomingMessage{
		Body: &body,
	}
	err := json.Unmarshal(clientMessage.message, &incomingMessage)
	if err != nil {
		log.Fatal(err)
		return
	}

	h.Lock()
	player := h.playerClients[clientMessage.client]

	actionType, ok := api.ActionLookup[incomingMessage.Action]
	if !ok {
		if player.room == nil {
			log.Fatalf("invalid action: %s\n", incomingMessage.Action)
			h.Unlock()
			return
		}

		h.Unlock()
		player.room.game.HandleIncomingMessage(player, incomingMessage)
		return
	}

	h.Unlock()

	switch actionType {
	case api.ActionCreateGame:
		var req api.CreateGameRequest
		if err := json.Unmarshal(body, &req); err != nil {
			log.Fatal(err)
			return
		}
		h.createGame(player, req)
	case api.ActionJoinGame:
		var req api.JoinGameRequest
		if err := json.Unmarshal(body, &req); err != nil {
			log.Fatal(err)
			return
		}
		h.joinGame(player, req)
	case api.ActionStartGame:
		var req api.StartGameRequest
		if err := json.Unmarshal(body, &req); err != nil {
			log.Fatal(err)
			return
		}
		h.startGame(player, req)
	case api.ActionRematch:
		var req api.RematchRequest
		if err := json.Unmarshal(body, &req); err != nil {
			log.Fatal(err)
			return
		}
		h.rematch(player, req)
	default:
		log.Fatalf("could not handle incoming action %s", incomingMessage.Action)
	}
}

func (h *Hub) performRoomChecks(
	player *Player,
	playerMustBeRoomOwner bool,
	playerMustBeCurrentPlayer bool,
) (*GameRoom, error) {
	room := player.room
	if room == nil {
		return nil, errors.New("you are not in a game")
	}

	if _, ok := h.rooms[room.roomCode]; !ok {
		return nil, errors.New("this game no longer exists")
	}

	room.lastInteractionTime = time.Now()

	if playerMustBeRoomOwner && !player.isRoomOwner {
		return nil, errors.New("you are not the game owner")
	}
	/*
	if playerMustBeCurrentPlayer {
		currentPlayer := h.getCurrentPlayer(room)
		if currentPlayer.name != player.name {
			return nil, errors.New("you are not the current player")
		}
	}
	*/
	return room, nil
}

func (h *Hub) createGame(
	player *Player,
	req api.CreateGameRequest,
) {
	log.Printf("Create game request: %s\n", req.Name)

	room := &GameRoom{
		gameType: req.GameType,
		roomCode: h.generateUniqueRoomCode(),
		lastInteractionTime: time.Now(),
		game: h.newGame(req.GameType),
	}

	h.Lock()
	defer h.Unlock()

	h.rooms[room.roomCode] = room
	log.Printf("%+v\n", h.rooms)

	player.name = req.Name
	player.room = room
	player.isRoomOwner = true

	room.game.AddPlayer(player)
}

func (h *Hub) generateUniqueRoomCode() string {
	h.RLock()
	defer h.RUnlock()

	for {
		newRoomCode := strconv.Itoa(getRandomNumberInRange(1000, 9999))
		foundDuplicate := false
		for roomCode := range h.rooms {
			if newRoomCode == roomCode {
				foundDuplicate = true
				break
			}
		}

		if !foundDuplicate {
			return newRoomCode
		}
	}
}

func (h *Hub) joinGame(clientMessage *ClientMessage, req api.JoinGameRequest) {
	log.Printf("Join game request: %s for %s\n", req.Name, req.RoomCode)

	h.Lock()
	defer h.Unlock()

	room, ok := h.rooms[req.RoomCode]
	if !ok {
		h.sendErrorMessage(clientMessage, "This room code does not exist.")
		return
	}

	matchedPlayer := h.getPlayerInRoom(room, req.Name)
	if matchedPlayer != nil {
		playerAddr := matchedPlayer.client.conn.RemoteAddr().(*net.TCPAddr)
		reqAddr := clientMessage.client.conn.RemoteAddr().(*net.TCPAddr)

		if reqAddr.IP.String() != playerAddr.IP.String() {
			log.Printf("Client with IP %s rejoining with same name as client with IP %s\n",
				reqAddr.IP.String(), playerAddr.IP.String())
			/*
				h.sendErrorMessage(clientMessage,
					"A player with that name is already in the room.")
				return
			*/
		}

		if matchedPlayer.client != clientMessage.client {
			matchedPlayer.client.conn.Close()
			delete(h.playerClients, matchedPlayer.client)
			matchedPlayer.client = clientMessage.client
		}
		h.playerClients[clientMessage.client] = matchedPlayer
		h.sendUpdatedGameMessages(room, clientMessage.client)
		return
	}

	game := room.game
	if game.state != "waiting-room" {
		h.sendErrorMessage(clientMessage,
			"You cannot join a game that has already started.")
		return
	}

	// Update the Player instance with the room and chosen name.
	player := h.playerClients[clientMessage.client]
	player.name = req.Name
	player.room = room
	player.isRoomOwner = false
	player.words = []string{}

	room.teams[0] = append(room.teams[0], player)

	h.sendUpdatedGameMessages(room, nil)
}

func (h *Hub) startGame(
	clientMessage *ClientMessage,
	req api.StartGameRequest,
) {
	log.Printf("Start game request\n")

	h.Lock()
	defer h.Unlock()

	player := h.playerClients[clientMessage.client]
	room, err := h.performRoomChecks(playerClient, true, false)
	if err != nil {
		h.sendErrorMessage(clientMessage, err.Error())
		return
	}

	room.game.Start()
}

func (h *Hub) rematch(
	clientMessage *ClientMessage,
	req api.RematchRequest,
) {
	log.Printf("Rematch request\n")

	h.Lock()
	defer h.Unlock()

	player := h.playerClients[clientMessage.client]
	room, err := h.performRoomChecks(playerClient, true, false)
	if err != nil {
		h.sendErrorMessage(clientMessage, err.Error())
		return
	}

	game.Rematch(playerClient)
}

// This function must be called with the mutex held.
func (h *Hub) sendErrorMessage(clientMessage *ClientMessage, err string) {
	var msg api.OutgoingMessage
	msg.Event = "error"
	msg.Error = err
	h.sendOutgoingMessages(clientMessage.client, &msg, nil, nil)
}

// This function must be called with the mutex held.
func (h *Hub) sendOutgoingMessages(
	primaryClient *Client,
	primaryMsg *api.OutgoingMessage,
	secondaryMsg *api.OutgoingMessage,
	room *GameRoom) {
	var err error
	primaryOutput, err := json.Marshal(*primaryMsg)
	if err != nil {
		log.Fatal(err)
		return
	}

	var secondaryOutput []byte
	if secondaryMsg != nil {
		secondaryOutput, err = json.Marshal(secondaryMsg)
		if err != nil {
			log.Fatal(err)
			return
		}
	}

	for client := range h.playerClients {
		if client == primaryClient {
			select {
			case client.send <- primaryOutput:
			default:
				close(client.send)
				delete(h.playerClients, client)
			}
		} else if secondaryMsg != nil &&
			(room == nil || h.playerClients[client].room == room) {
			select {
			case client.send <- secondaryOutput:
			default:
				close(client.send)
				delete(h.playerClients, client)
			}
		}
	}
}

// This function must be called with the mutex held.
func (h *Hub) getPlayerInRoom(room *GameRoom, name string) *Player {
	for _, players := range room.teams {
		for _, player := range players {
			if player.name == name {
				return player
			}
		}
	}

	return nil
}
