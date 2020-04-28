// Use of this source code is governed by a BSD-style
// Copyright 2013 The Gorilla ebSocket Authors. All rights reserved.
// license that can be found in the LICENSE file.

package main

import (
	"net"
	"encoding/json"
	"log"
	"math/rand"
	"strconv"
	"time"

	"./api"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
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
	client                 *Client
	name                   string
	room                   *GameRoom
	isRoomOwner            bool
	words                  []string
	numCardsGuessedByRound []int
}

// Game holds all the data about the Fishbowl game.
type Game struct {
	state                 string
	turnJustStarted       bool
	cardsInRound          []string
	currentServerTime     int64
	timerLength           int
	lastCardGuessed       string
	numCardsGuessedInTurn int
	teamScoresByRound     [][]int
	currentRound          int   // 0, 1, 2
	currentPlayers        []int // [ team0PlayerIdx, team1PlayerIdx ]
	currentlyPlayingTeam  int   // 0, 1, ...
}

// GameRoom holds the data about a game room.
type GameRoom struct {
	roomCode string
	gameType string
	teams    [][]*Player
	game     *Game
}

var (
	validStateTransitions = map[string][]string{
		"waiting-room": {
			"turn-start",
		},
		"turn-start": {
			"turn-active",
		},
		"turn-active": {
			"turn-start",
			"game-over",
		},
	}
)

func newHub() *Hub {
	return &Hub{
		playerClients: make(map[*Client]*Player),
		rooms:         make(map[string]*GameRoom),
		message:       make(chan *ClientMessage),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.playerClients[client] = &Player{
				client: client,
			}
		case client := <-h.unregister:
			if _, ok := h.playerClients[client]; ok {
				delete(h.playerClients, client)
				close(client.send)
			}
		case clientMessage := <-h.message:
			h.handleIncomingMessage(clientMessage)
		}
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

	actionType, ok := api.ActionLookup[incomingMessage.Action]
	if !ok {
		log.Fatalf("invalid action: %s\n", incomingMessage.Action)
		return
	}

	switch actionType {
	case api.ActionCreateRoom:
		var req api.CreateRoomRequest
		if err := json.Unmarshal(body, &req); err != nil {
			log.Fatal(err)
			return
		}
		h.createRoom(clientMessage, req)
	case api.ActionJoinRoom:
		var req api.JoinRoomRequest
		if err := json.Unmarshal(body, &req); err != nil {
			log.Fatal(err)
			return
		}
		h.joinRoom(clientMessage, req)
	case api.ActionAddTeam:
		var req api.AddTeamRequest
		if err := json.Unmarshal(body, &req); err != nil {
			log.Fatal(err)
			return
		}
		h.addTeam(clientMessage, req)
	case api.ActionRemoveTeam:
		// TODO
	case api.ActionMovePlayer:
		var req api.MovePlayerRequest
		if err := json.Unmarshal(body, &req); err != nil {
			log.Fatal(err)
			return
		}
		h.movePlayer(clientMessage, req)
	case api.ActionSubmitWords:
		var req api.SubmitWordsRequest
		if err := json.Unmarshal(body, &req); err != nil {
			log.Fatal(err)
			return
		}
		h.submitWords(clientMessage, req)
	case api.ActionStartGame:
		var req api.StartGameRequest
		if err := json.Unmarshal(body, &req); err != nil {
			log.Fatal(err)
			return
		}
		h.startGame(clientMessage, req)
	case api.ActionStartTurn:
		var req api.StartTurnRequest
		if err := json.Unmarshal(body, &req); err != nil {
			log.Fatal(err)
			return
		}
		h.startTurn(clientMessage, req)
	case api.ActionChangeCard:
		var req api.ChangeCardRequest
		if err := json.Unmarshal(body, &req); err != nil {
			log.Fatal(err)
			return
		}
		h.changeCard(clientMessage, req)
	default:
		log.Fatalf("could not handle incoming action %s", incomingMessage.Action)
	}
}

func (h *Hub) createRoom(
	clientMessage *ClientMessage,
	req api.CreateRoomRequest,
) {
	log.Printf("Create room request: %s\n", req.Name)

	room := &GameRoom{
		gameType: req.GameType,
		roomCode: h.generateUniqueRoomCode(),
		teams:    make([][]*Player, 2),
		game: &Game{
			state:             "waiting-room",
			teamScoresByRound: make([][]int, 3),
		},
	}
	h.rooms[room.roomCode] = room

	player := h.playerClients[clientMessage.client]
	player.name = req.Name
	player.room = room
	player.isRoomOwner = true
	player.words = nil

	room.teams[0] = make([]*Player, 0)
	room.teams[0] = append(room.teams[0], player)
	room.teams[1] = make([]*Player, 0)

	log.Printf("%+v\n", h.rooms)

	var msg api.OutgoingMessage
	msg.Event = "created-room"
	msg.Body = api.CreatedRoomEvent{
		RoomCode: room.roomCode,
		Team:     0,
	}

	h.sendOutgoingMessages(clientMessage.client, &msg, nil, nil)
}

func (h *Hub) generateUniqueRoomCode() string {
	for ;; {
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

func (h *Hub) joinRoom(clientMessage *ClientMessage, req api.JoinRoomRequest) {
	log.Printf("Join room request: %s for %s\n", req.Name, req.RoomCode)

	// TODO: add support for re-joining the game
	log.Printf("remoteAddr: %s\n", clientMessage.client.conn.UnderlyingConn().RemoteAddr())

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
			h.sendErrorMessage(clientMessage,
				"A player with that name is already in the room.")
			return
		}

		matchedPlayer.client.conn.Close()
		matchedPlayer.client = clientMessage.client
		h.playerClients[clientMessage.client] = matchedPlayer
		h.sendUpdatedRoomMessages(clientMessage, room)
		return
	}

	// Update the Player instance with the room and chosen name.
	player := h.playerClients[clientMessage.client]
	player.name = req.Name
	player.room = room
	player.isRoomOwner = false
	player.words = nil

	room.teams[0] = append(room.teams[0], player)

	h.sendUpdatedRoomMessages(clientMessage, room)
}

func (h *Hub) submitWords(
	clientMessage *ClientMessage,
	req api.SubmitWordsRequest,
) {
	log.Printf("Submit words request: %s\n", req.Words)

	playerClient := h.playerClients[clientMessage.client]
	room := playerClient.room
	if room == nil {
		h.sendErrorMessage(clientMessage, "You are not in a game.")
		return
	}

	playerClient.words = req.Words
	h.sendUpdatedRoomMessages(clientMessage, room)
}

func (h *Hub) movePlayer(
	clientMessage *ClientMessage,
	req api.MovePlayerRequest,
) {
	log.Printf("Move player request: %s (%d -> %d)\n", req.PlayerName,
		req.FromTeam, req.ToTeam)

	var msg api.OutgoingMessage

	playerClient := h.playerClients[clientMessage.client]
	room := playerClient.room
	if room == nil {
		h.sendErrorMessage(clientMessage, "You are not in a game.")
		return
	}

	if req.FromTeam >= len(room.teams) || req.ToTeam >= len(room.teams) {
		msg.Event = "error"
		msg.Error = "The team indexes are invalid."
		h.sendOutgoingMessages(clientMessage.client, &msg, nil, nil)
		return
	}

	player := h.removePlayerFromTeam(room, req.FromTeam, req.PlayerName)
	room.teams[req.ToTeam] = append(room.teams[req.ToTeam], player)

	h.sendUpdatedRoomMessages(clientMessage, room)
}

func (h *Hub) addTeam(
	clientMessage *ClientMessage,
	req api.AddTeamRequest,
) {
	log.Printf("Add team request\n")

	playerClient := h.playerClients[clientMessage.client]
	room := playerClient.room
	if room == nil {
		h.sendErrorMessage(clientMessage, "You are not in a game.")
		return
	}

	room.teams = append(room.teams, []*Player{})
	h.sendUpdatedRoomMessages(clientMessage, room)
}

func (h *Hub) startGame(
	clientMessage *ClientMessage,
	req api.StartGameRequest,
) {
	log.Printf("Start game request\n")

	playerClient := h.playerClients[clientMessage.client]
	room := playerClient.room
	if room == nil {
		h.sendErrorMessage(clientMessage, "You are not in a game.")
		return
	}

	if !playerClient.isRoomOwner {
		h.sendErrorMessage(clientMessage, "You are not the game owner.")
		return
	}

	game := room.game
	if !h.validateStateTransition(game.state, "turn-start") {
		h.sendErrorMessage(clientMessage, "You cannot perform that action at this time.")
		return
	}
	game.turnJustStarted = true
	game.state = "turn-start"

	h.reshuffleCardsForRound(room)

	// Init team scores
	game.teamScoresByRound[0] = make([]int, len(room.teams))
	game.teamScoresByRound[1] = make([]int, len(room.teams))
	game.teamScoresByRound[2] = make([]int, len(room.teams))

	game.currentRound = 0
	game.currentPlayers = make([]int, len(room.teams))
	for i, players := range room.teams {
		game.currentPlayers[i] = getRandomNumberInRange(0, len(players)-1)
	}
	game.currentlyPlayingTeam = getRandomNumberInRange(0, len(room.teams)-1)

	h.sendUpdatedGameMessages(room)
}

func (h *Hub) reshuffleCardsForRound(room *GameRoom) {
	game := room.game

	game.cardsInRound = []string{}
	for _, teamPlayers := range room.teams {
		for _, player := range teamPlayers {
			game.cardsInRound = append(game.cardsInRound, player.words...)
		}
	}

	// TODO: rand.Seed(time.Now().UnixNano())
	arr := game.cardsInRound
	rand.Shuffle(len(game.cardsInRound), func(i, j int) {
		arr[i], arr[j] = arr[j], arr[i]
	})
}

func (h *Hub) startTurn(
	clientMessage *ClientMessage,
	req api.StartTurnRequest,
) {
	log.Printf("Start turn request\n")

	playerClient := h.playerClients[clientMessage.client]
	room := playerClient.room
	if room == nil {
		h.sendErrorMessage(clientMessage, "You are not in a game.")
		return
	}

	currentPlayer := h.getCurrentPlayer(room)
	if currentPlayer.name != playerClient.name {
		h.sendErrorMessage(clientMessage, "You are not the current player.")
		return
	}

	game := room.game

	if !h.validateStateTransition(game.state, "turn-active") {
		h.sendErrorMessage(clientMessage, "You cannot perform that action at this time.")
		return
	}
	game.turnJustStarted = true
	game.state = "turn-active"

	game.numCardsGuessedInTurn = 0
	game.lastCardGuessed = ""
	game.timerLength = 11
	game.currentServerTime = time.Now().UnixNano() / 1000000
	timer := time.NewTimer(time.Second * time.Duration(game.timerLength))

	// Wait for timer to finish in an asynchronous goroutine
	go func() {
		// Block until timer finishes. When it is done, it sends a message
		// on the channel timer.C. No other code in
		// this goroutine is executed until that happens.
		<-timer.C

		log.Printf("Game state when timer ended: %s\n", game.state)
		if !h.validateStateTransition(game.state, "turn-start") {
			if game.state == "turn-start" || game.state == "game-over" {
				// Round or game finished before the player's turn timer expired,
				// so do nothing.
				return
			}

			log.Fatalf("Game was not in correct state when turn timer expired: %s", game.state)
			return
		}

		game.turnJustStarted = false
		game.state = "turn-start"
		h.moveToNextPlayerAndTeam(room)

		log.Printf("Sending updated game message after timer expired\n")
		h.sendUpdatedGameMessages(room)
	}()

	h.sendUpdatedGameMessages(room)
}

func (h *Hub) changeCard(
	clientMessage *ClientMessage,
	req api.ChangeCardRequest,
) {
	log.Printf("Change card request: %s\n", req.ChangeType)

	playerClient := h.playerClients[clientMessage.client]
	room := playerClient.room
	if room == nil {
		h.sendErrorMessage(clientMessage, "You are not in a game.")
		return
	}

	currentPlayer := h.getCurrentPlayer(room)
	if currentPlayer.name != playerClient.name {
		h.sendErrorMessage(clientMessage, "You are not the current player.")
		return
	}

	game := room.game
	if game.state != "turn-active" {
		// Ignore, the turn is probably over.
		return
	}

	game.turnJustStarted = false
	if req.ChangeType == "correct" {
		// Increment score for current team and the current turn.
		game.teamScoresByRound[game.currentRound][game.currentlyPlayingTeam]++
		game.numCardsGuessedInTurn++

		game.lastCardGuessed = game.cardsInRound[0]
		game.cardsInRound = game.cardsInRound[1:]

		if len(game.cardsInRound) == 0 {
			game.currentRound++
			if game.currentRound < 3 {
				// Round over, moving to next round
				game.state = "turn-start"

				h.reshuffleCardsForRound(room)
				h.moveToNextPlayerAndTeam(room)

				// Each round should start with a different team.
				//
				// TODO: The teams should be re-ordered based on score.
				game.currentlyPlayingTeam = getRandomNumberInRange(0,
					len(room.teams)-1)
			} else {
				// TODO: Game over
			}
		}
	} else {
		// Skip this card, push it to the end
		game.cardsInRound = append(game.cardsInRound[1:], game.cardsInRound[0])
	}

	h.sendUpdatedGameMessages(room)
}

func (h *Hub) moveToNextPlayerAndTeam(room *GameRoom) {
	game := room.game
	t := game.currentlyPlayingTeam
	game.currentPlayers[t] = (game.currentPlayers[t] + 1) % len(room.teams[t])
	game.currentlyPlayingTeam = (t + 1) % len(game.currentPlayers)
}

func (h *Hub) getCurrentPlayer(room *GameRoom) *Player {
	game := room.game
	players := room.teams[game.currentlyPlayingTeam]
	return players[game.currentPlayers[game.currentlyPlayingTeam]]
}

func (h *Hub) validateStateTransition(fromState, toState string) bool {
	valid, ok := validStateTransitions[fromState]
	if !ok {
		return false
	}

	if !stringInSlice(valid, toState) {
		return false
	}

	return true
}

func (h *Hub) sendUpdatedRoomMessages(
	clientMessage *ClientMessage,
	room *GameRoom,
) {
	var msg api.OutgoingMessage
	msg.Event = "updated-room"
	msg.Body = api.UpdatedRoomEvent{
		Teams: h.convertTeamsToApiTeams(room.teams),
	}

	h.sendOutgoingMessages(clientMessage.client, &msg, &msg,
		room)
}

func (h *Hub) sendErrorMessage(clientMessage *ClientMessage, err string) {
	var msg api.OutgoingMessage
	msg.Event = "error"
	msg.Error = err
	h.sendOutgoingMessages(clientMessage.client, &msg, nil, nil)
}

func (h *Hub) sendOutgoingMessages(
	primaryClient *Client,
	primaryMsg *api.OutgoingMessage,
	secondaryMsg *api.OutgoingMessage,
	room *GameRoom) {
	var err error
	primaryOutput, err := json.Marshal(*primaryMsg)
	if err != nil {
		log.Fatal(err) // TODO: handle properly
	}

	var secondaryOutput []byte
	if secondaryMsg != nil {
		secondaryOutput, err = json.Marshal(secondaryMsg)
		if err != nil {
			log.Fatal(err) // TODO: handle properly
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

func (h *Hub) removePlayerFromTeam(
	room *GameRoom,
	fromTeam int,
	playerName string,
) *Player {
	players := room.teams[fromTeam]
	for idx, player := range players {
		if player.name == playerName {
			player := players[idx]
			room.teams[fromTeam] = append(
				players[:idx],
				players[idx+1:]...,
			)
			return player
		}
	}

	return nil
}

func (h *Hub) convertPlayersToApiPlayers(players []*Player) []api.Player {
	apiPlayers := make([]api.Player, 0, len(players))
	for _, player := range players {
		apiPlayers = append(apiPlayers, api.Player{
			Name:           player.name,
			IsRoomOwner:    player.isRoomOwner,
			WordsSubmitted: len(player.words) > 0,
		})
	}
	return apiPlayers
}

func (h *Hub) convertTeamsToApiTeams(teams [][]*Player) [][]api.Player {
	apiTeams := make([][]api.Player, 0, len(teams))
	for _, players := range teams {
		apiTeams = append(apiTeams, h.convertPlayersToApiPlayers(players))
	}
	return apiTeams
}

func getRandomNumberInRange(min, max int) int {
	if min == max {
		return min
	}

	return min + rand.Intn(max-min)
}
