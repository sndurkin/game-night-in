package main

import (
	"errors"
	"encoding/json"
	"log"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/sndurkin/game-night-in/api"
	"github.com/sndurkin/game-night-in/models"
	"github.com/sndurkin/game-night-in/fishbowl"
	"github.com/sndurkin/game-night-in/util"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	mutex sync.RWMutex

	// Map of connected client to Player
	playerClients map[*Client]*models.Player

	// Map of room code to GameRoom
	rooms map[string]*models.GameRoom

	// Inbound messages from the clients.
	message chan *ClientMessage

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

// newHub creates a new Hub instance which manages all incoming
// websocket messages.
func newHub() *Hub {
	return &Hub{
		playerClients: make(map[*Client]*models.Player),
		rooms:         make(map[string]*models.GameRoom),
		message:       make(chan *ClientMessage),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
	}
}

func (h *Hub) newGame(gameType string, room *models.GameRoom) interface{} {
	switch (gameType) {
	case "fishbowl":
		return fishbowl.NewGame(room, &h.mutex)
	}

	return nil
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.playerClients[client] = &models.Player{
				Client: client,
			}
			h.mutex.Unlock()
		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.playerClients[client]; ok {
				delete(h.playerClients, client)
				close(client.send)
			}
			h.mutex.Unlock()
		case clientMessage := <-h.message:
			h.handleIncomingMessage(clientMessage)
		}
	}
}

func (h *Hub) runRoomCleanup() {
	for now := range time.Tick(time.Hour) {
		h.mutex.Lock()

		roomCodes := []string{}
		for roomCode, room := range h.rooms {
			expiryTime := room.LastInteractionTime.Add(
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

		h.mutex.Unlock()
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

	h.mutex.Lock()
	player := h.playerClients[clientMessage.client]

	actionType, ok := api.ActionLookup[incomingMessage.Action]
	if !ok {
		if player.Room == nil {
			log.Fatalf("invalid action: %s\n", incomingMessage.Action)
			h.mutex.Unlock()
			return
		}

		h.mutex.Unlock()
		player.Room.Game.HandleIncomingMessage(
			player,
			incomingMessage,
			body,
			h.sendOutgoingMessages,
		)
		return
	}

	h.mutex.Unlock()

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
	player *models.Player,
	playerMustBeRoomOwner bool,
	playerMustBeCurrentPlayer bool,
) (*models.GameRoom, error) {
	room := player.Room
	if room == nil {
		return nil, errors.New("you are not in a game")
	}

	if _, ok := h.rooms[room.RoomCode]; !ok {
		return nil, errors.New("this game no longer exists")
	}

	room.LastInteractionTime = time.Now()

	if playerMustBeRoomOwner && !player.IsRoomOwner {
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
	player *models.Player,
	req api.CreateGameRequest,
) {
	log.Printf("Create game request: %s\n", req.Name)

	room := &models.GameRoom{
		GameType: req.GameType,
		RoomCode: h.generateUniqueRoomCode(),
		LastInteractionTime: time.Now(),
		Players: make([]*models.Player, 0),
	}
	room.Game = h.newGame(req.GameType, room).(*models.Game)

	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.rooms[room.RoomCode] = room
	log.Printf("%+v\n", h.rooms)

	player.Name = req.Name
	player.Room = room
	player.IsRoomOwner = true
	room.Players = append(room.Players, player)

	room.Game.AddPlayer(player)
}

func (h *Hub) generateUniqueRoomCode() string {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for {
		newRoomCode := strconv.Itoa(util.GetRandomNumberInRange(1000, 9999))
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

func (h *Hub) joinGame(player *models.Player, req api.JoinGameRequest) {
	log.Printf("Join game request: %s for %s\n", req.Name, req.RoomCode)

	h.mutex.Lock()
	defer h.mutex.Unlock()

	room, ok := h.rooms[req.RoomCode]
	if !ok {
		h.sendErrorMessage(player, "This room code does not exist.")
		return
	}

	matchedPlayer, playerIdx := h.getPlayerInRoom(room, req.Name)
	if matchedPlayer != nil {
		matchedPlayerClient := matchedPlayer.Client.(*Client)
		playerClient := player.Client.(*Client)
		playerAddr := matchedPlayerClient.conn.RemoteAddr().(*net.TCPAddr)
		reqAddr := playerClient.conn.RemoteAddr().(*net.TCPAddr)

		if reqAddr.IP.String() != playerAddr.IP.String() {
			log.Printf("Client with IP %s rejoining with same name as client with IP %s\n",
				reqAddr.IP.String(), playerAddr.IP.String())
			/*
				h.sendErrorMessage(clientMessage,
					"A player with that name is already in the room.")
				return
			*/
		}

		if matchedPlayerClient != playerClient {
			matchedPlayerClient.conn.Close()
			delete(h.playerClients, matchedPlayerClient)
			matchedPlayer.Client = playerClient
		}
		h.playerClients[playerClient] = matchedPlayer
		room.Players[playerIdx] = matchedPlayer
		room.Game.Join(matchedPlayer, false, req, h.sendOutgoingMessages)
		return
	}

	room.Players = append(room.Players, player)
	room.Game.Join(player, true, req, h.sendOutgoingMessages)
}

func (h *Hub) startGame(
	player *models.Player,
	req api.StartGameRequest,
) {
	log.Printf("Start game request\n")

	h.mutex.Lock()
	defer h.mutex.Unlock()

	room, err := h.performRoomChecks(player, true, false)
	if err != nil {
		h.sendErrorMessage(player, err.Error())
		return
	}

	room.Game.Start(player, h.sendOutgoingMessages)
}

func (h *Hub) rematch(
	player *models.Player,
	req api.RematchRequest,
) {
	log.Printf("Rematch request\n")

	h.mutex.Lock()
	defer h.mutex.Unlock()

	room, err := h.performRoomChecks(player, true, false)
	if err != nil {
		h.sendErrorMessage(player, err.Error())
		return
	}

	room.Game.Rematch(player, h.sendOutgoingMessages)
}

// This function must be called with the mutex held.
func (h *Hub) sendErrorMessage(player *models.Player, err string) {
	var msg api.OutgoingMessage
	msg.Event = "error"
	msg.Error = err
	h.sendOutgoingMessages(&models.OutgoingMessageRequest{
		PrimaryClient: player.Client,
		PrimaryMsg: &msg,
	})
}

// This function must be called with the mutex held.
func (h *Hub) sendOutgoingMessages(
	req *models.OutgoingMessageRequest,
) {
	var err error
	primaryOutput, err := json.Marshal(*req.PrimaryMsg)
	if err != nil {
		log.Fatal(err)
		return
	}

	var secondaryOutput []byte
	if req.SecondaryMsg != nil {
		secondaryOutput, err = json.Marshal(req.SecondaryMsg)
		if err != nil {
			log.Fatal(err)
			return
		}
	}

	for client := range h.playerClients {
		if client == req.PrimaryClient {
			select {
			case client.send <- primaryOutput:
			default:
				close(client.send)
				delete(h.playerClients, client)
			}
		} else if req.SecondaryMsg != nil &&
			(req.Room == nil || h.playerClients[client].Room == req.Room) {
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
func (h *Hub) getPlayerInRoom(
	room *models.GameRoom,
	name string,
) (*models.Player, int) {
	for idx, player := range room.Players {
		if player.Name == name {
			return player, idx
		}
	}

	return nil, -1
}
