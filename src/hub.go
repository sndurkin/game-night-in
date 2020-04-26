// Use of this source code is governed by a BSD-style
// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Map of connected client to PlayerInfo
	playerClients map[*Client]*PlayerInfo

	// Map of room code to GameRoom
	rooms map[string]*GameRoom

	// Inbound messages from the clients.
	message chan *ClientMessage

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

// PlayerInfo holds all the data about a player.
type PlayerInfo struct {
	client      *Client
	name        string
	room        *GameRoom
	isRoomOwner bool
	words       []string
	numCardsGuessedByRound []int
}

type Game struct {
	state                 string
	cardsInRound          []string
	currentServerTime     int64
	timerLength           int
	currentCardsIdx       int
	numCardsGuessedInTurn int
	lastCardGuessed string
	currentRound          int   // 0, 1, 2
	currentPlayers        []int // [ team0PlayerIdx, team1PlayerIdx ]
	currentlyPlayingTeam  int   // 0, 1, ...
}

// GameRoom holds all the data about a particular game.
type GameRoom struct {
	roomCode string
	gameType string
	teams    [][]*PlayerInfo
	game     *Game
}

// PlayerClient is used to connect clients with player information.
type PlayerClient struct {
	client *Client
	player Player
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
		playerClients: make(map[*Client]*PlayerInfo),
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
			h.playerClients[client] = &PlayerInfo{
				client: client,
			}
		case client := <-h.unregister:
			if _, ok := h.playerClients[client]; ok {
				delete(h.playerClients, client)
				close(client.send)
			}
		case clientMessage := <-h.message:
			h.processIncomingMessage(clientMessage)
		}
	}
}

func (h *Hub) processIncomingMessage(clientMessage *ClientMessage) {
	var body json.RawMessage
	incomingMessage := IncomingMessage{
		Body: &body,
	}
	err := json.Unmarshal(clientMessage.message, &incomingMessage)
	if err != nil {
		panic(err) // TODO: handle properly
	}

	switch incomingMessage.Action {
	case "create-room":
		var req CreateRoomRequest
		if err := json.Unmarshal(body, &req); err != nil {
			//log.Fatal(err)
			panic(err)
		}
		h.createRoom(clientMessage, req)
	case "join-room":
		var req JoinRoomRequest
		if err := json.Unmarshal(body, &req); err != nil {
			//log.Fatal(err)
			panic(err)
		}
		h.joinRoom(clientMessage, req)
	case "submit-words":
		var req SubmitWordsRequest
		if err := json.Unmarshal(body, &req); err != nil {
			//log.Fatal(err)
			panic(err)
		}
		h.submitWords(clientMessage, req)
	case "move-player":
		var req MovePlayerRequest
		if err := json.Unmarshal(body, &req); err != nil {
			//log.Fatal(err)
			panic(err)
		}
		h.movePlayer(clientMessage, req)
	case "add-team":
		var req AddTeamRequest
		if err := json.Unmarshal(body, &req); err != nil {
			//log.Fatal(err)
			panic(err)
		}
		h.addTeam(clientMessage, req)
	case "start-game":
		var req StartGameRequest
		if err := json.Unmarshal(body, &req); err != nil {
			//log.Fatal(err)
			panic(err)
		}
		h.startGame(clientMessage, req)
	case "start-turn":
		var req StartTurnRequest
		if err := json.Unmarshal(body, &req); err != nil {
			//log.Fatal(err)
			panic(err)
		}
		h.startTurn(clientMessage, req)
	case "change-card":
		var req ChangeCardRequest
		if err := json.Unmarshal(body, &req); err != nil {
			//log.Fatal(err)
			panic(err)
		}
		h.changeCard(clientMessage, req)
	default:
		log.Fatalf("could not handle incoming action %s", incomingMessage.Action)
	}
}

func (h *Hub) createRoom(clientMessage *ClientMessage, req CreateRoomRequest) {
	log.Printf("Create room request: %s\n", req.Name)

	room := &GameRoom{
		gameType: req.GameType,
		roomCode: strconv.Itoa(getRandomNumberInRange(1000, 9999)),
		teams:    make([][]*PlayerInfo, 2),
		game: &Game{
			state: "waiting-room",
		},
	}
	h.rooms[room.roomCode] = room

	h.playerClients[clientMessage.client].name = req.Name
	h.playerClients[clientMessage.client].room = room
	h.playerClients[clientMessage.client].isRoomOwner = true
	h.playerClients[clientMessage.client].words = nil

	room.teams[0] = make([]*PlayerInfo, 0)
	room.teams[0] = append(room.teams[0], h.playerClients[clientMessage.client])
	room.teams[1] = make([]*PlayerInfo, 0)

	log.Printf("%+v\n", h.rooms)

	var msg OutgoingMessage
	msg.Event = "created-room"
	msg.Body = CreatedRoomEvent{
		RoomCode: room.roomCode,
		Team:     0,
	}

	h.sendOutgoingMessages(clientMessage.client, &msg, nil, nil)
}

func (h *Hub) joinRoom(clientMessage *ClientMessage, req JoinRoomRequest) {
	log.Printf("Join room request: %s\n", req.Name)
	log.Printf("%+v\n", h.rooms)

	// TODO: add support for re-joining the game
	log.Printf("remoteAddr: %s\n", clientMessage.client.conn.UnderlyingConn().RemoteAddr())

	room, ok := h.rooms[req.RoomCode]
	if !ok {
		h.sendErrorMessage(clientMessage, "This room code does not exist.")
		return
	}

	log.Printf("%+v\n", room)
	for _, teamPlayerInfos := range room.teams {
		for _, playerInfo := range teamPlayerInfos {
			if strings.EqualFold(playerInfo.name, req.Name) {
				// TODO: check if same IP address so we can replace the playerClient with the one
				// that probably has lost connection?

				h.sendErrorMessage(clientMessage,
					"A player with that name is already in the room.")
				return
			}
		}
	}

	// Update the PlayerInfo instance with the room and chosen name.
	h.playerClients[clientMessage.client].name = req.Name
	h.playerClients[clientMessage.client].room = room
	h.playerClients[clientMessage.client].isRoomOwner = false
	h.playerClients[clientMessage.client].words = nil

	room.teams[0] = append(room.teams[0], h.playerClients[clientMessage.client])

	h.sendUpdatedRoomMessages(clientMessage, room)
}

func (h *Hub) submitWords(clientMessage *ClientMessage, req SubmitWordsRequest) {
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

func (h *Hub) movePlayer(clientMessage *ClientMessage, req MovePlayerRequest) {
	log.Printf("Move player request: %s (%d -> %d)\n", req.PlayerName,
		req.FromTeam, req.ToTeam)

	var msg OutgoingMessage

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

func (h *Hub) addTeam(clientMessage *ClientMessage, req AddTeamRequest) {
	log.Printf("Add team request\n")

	playerClient := h.playerClients[clientMessage.client]
	room := playerClient.room
	if room == nil {
		h.sendErrorMessage(clientMessage, "You are not in a game.")
		return
	}

	room.teams = append(room.teams, []*PlayerInfo{})
	h.sendUpdatedRoomMessages(clientMessage, room)
}

func (h *Hub) startGame(clientMessage *ClientMessage, req StartGameRequest) {
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
	game.state = "turn-start"

	h.reshuffleCardsForTurn(room)

	game.currentRound = 0
	game.currentPlayers = make([]int, len(room.teams))
	for i, players := range room.teams {
		game.currentPlayers[i] = getRandomNumberInRange(0, len(players)-1)
	}
	game.currentlyPlayingTeam = getRandomNumberInRange(0, len(room.teams)-1)

	h.sendUpdatedGameMessages(clientMessage, room)
}

func (h *Hub) reshuffleCardsForTurn(room *GameRoom) []string {
	game.currentCardsIdx = 0

	game.cardsInTurn = []string{}
	for _, teamPlayerInfos := range room.teams {
		for _, playerInfo := range teamPlayerInfos {
			game.cardsInTurn = append(game.cardsInTurn, playerInfo.words...)
		}
	}

	// TODO: rand.Seed(time.Now().UnixNano())
	arr := game.cardsInTurn
	rand.Shuffle(len(game.cardsInTurn), func(i, j int) {
		arr[i], arr[j] = arr[j], arr[i]
	})
}

func (h *Hub) startTurn(clientMessage *ClientMessage, req StartTurnRequest) {
	log.Printf("Start turn request\n")

	playerClient := h.playerClients[clientMessage.client]
	room := playerClient.room
	if room == nil {
		h.sendErrorMessage(clientMessage, "You are not in a game.")
		return
	}

	currentPlayer := h.getCurrentPlayer(room)
	if strings.Compare(currentPlayer.name, playerClient.name) != 0 {
		h.sendErrorMessage(clientMessage, "You are not the current player.")
		return
	}

	game := room.game

	if !h.validateStateTransition(game.state, "turn-active") {
		h.sendErrorMessage(clientMessage, "You cannot perform that action at this time.")
		return
	}
	game.state = "turn-active"

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

		if !h.validateStateTransition(game.state, "turn-start") {
			log.Fatalf("Game was not in correct state when turn timer expired: %s", game.state)
			return
		}

		game.state = "turn-start"
		h.moveToNextPlayerAndTeam(room)

		h.sendUpdatedGameMessages(clientMessage, room)
	}()

	h.sendUpdatedGameMessages(clientMessage, room)
}

func (h *Hub) changeCard(clientMessage *ClientMessage, req ChangeCardRequest) {
	log.Printf("Change card request: %s\n", req.ChangeType)

	playerClient := h.playerClients[clientMessage.client]
	room := playerClient.room
	if room == nil {
		h.sendErrorMessage(clientMessage, "You are not in a game.")
		return
	}

	currentPlayer := h.getCurrentPlayer(room)
	if strings.Compare(currentPlayer.name, playerClient.name) != 0 {
		h.sendErrorMessage(clientMessage, "You are not the current player.")
		return
	}

	game := room.game
	if !strings.EqualFold(game.state, "turn-active") {
		// Ignore, the turn is probably over.
		return
	}

	if strings.EqualFold(req.ChangeType, "correct") {
		game.numCardsGuessedInTurn++
		game.lastCardGuessed = game.cardsInTurn[game.currentCardsIdx]
		game.cardsInTurn = append(game.cardsInTurn[0:game.currentCardsIdx],
			game.cardsInTurn[game.currentCardsIdx+1:]...)

		if len(game.cardsInTurn) == 0 {
			game.currentRound++
			if game.currentRound < 3 {
				game.state = "turn-start"
				h.reshuffleCardsForTurn(room)
				h.moveToNextPlayerAndTeam(room)
				game.currentlyPlayingTeam = getRandomNumberInRange(0,
					len(room.teams)-1)
			}
			else {
				// TODO: Game over
			}
		}
	}
	else {
		// Skip this card
		game.currentCardsIdx = (game.currentCardsIdx + 1) %
			len(game.cardsInTurn)
	}

	h.sendUpdatedGameMessages(clientMessage, room)
}

func (h *Hub) moveToNextPlayerAndTeam(room *GameRoom) {
	game := room.game
	t := game.currentlyPlayingTeam
	game.currentPlayers[t] = (game.currentPlayers[t] + 1) % len(room.teams[t])
	game.currentlyPlayingTeam = (t + 1) % len(game.currentPlayers)
}

func (h *Hub) getCurrentPlayer(room *GameRoom) *PlayerInfo {
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

func stringInSlice(arr []string, val string) bool {
	for _, s := range arr {
		if strings.EqualFold(s, val) {
			return true
		}
	}

	return false
}

func (h *Hub) sendUpdatedRoomMessages(
	clientMessage *ClientMessage,
	room *GameRoom,
) {
	var msg OutgoingMessage
	msg.Event = "updated-room"
	msg.Body = UpdatedRoomEvent{
		Teams: h.convertTeamsPlayerInfosToTeams(room.teams),
	}

	h.sendOutgoingMessages(clientMessage.client, &msg, &msg,
		room)
}

func (h *Hub) sendUpdatedGameMessages(
	clientMessage *ClientMessage,
	room *GameRoom,
) {
	game := room.game

	var lastCardGuessed string
	if game.currentCardsIdx > 0 {
		lastCardGuessed = game.cardsInTurn[game.currentCardsIdx]
	}

	var cards []string
	var currentServerTime int64
	var timerLength int
	if strings.EqualFold(game.state, "turn-active") {
		cards = game.cards[game.currentCardsIdx:]

		currentServerTime = game.currentServerTime
		timerLength = game.timerLength
	}

	var msgToCurrentPlayer OutgoingMessage
	msgToCurrentPlayer.Event = "updated-game"
	msgToCurrentPlayer.Body = UpdatedGameEvent{
		State:                 game.state,
		LastCardGuessed:       lastCardGuessed,
		CurrentServerTime:     currentServerTime,
		TimerLength:           timerLength,
		Cards:                 cards,
		NumCardsLeftInRound:   len(game.cards),
		NumCardsGuessedInTurn: game.numCardsGuessedInTurn,
		CurrentRound:          game.currentRound,
		CurrentPlayers:        game.currentPlayers,
		CurrentlyPlayingTeam:  game.currentlyPlayingTeam,
	}

	var msgToOtherPlayers OutgoingMessage
	msgToOtherPlayers.Event = "updated-game"
	msgToOtherPlayers.Body = UpdatedGameEvent{
		State:                 game.state,
		LastCardGuessed:       lastCardGuessed,
		CurrentServerTime:     currentServerTime,
		TimerLength:           timerLength,
		NumCardsLeftInRound:   len(game.cards) - game.currentCardsIdx,
		NumCardsGuessedInTurn: 0,
		CurrentRound:          game.currentRound,
		CurrentPlayers:        game.currentPlayers,
		CurrentlyPlayingTeam:  game.currentlyPlayingTeam,
	}

	h.sendOutgoingMessages(clientMessage.client, &msgToCurrentPlayer, &msgToOtherPlayers,
		room)
}

func (h *Hub) sendErrorMessage(clientMessage *ClientMessage, err string) {
	var msg OutgoingMessage
	msg.Event = "error"
	msg.Error = err
	h.sendOutgoingMessages(clientMessage.client, &msg, nil, nil)
}

func (h *Hub) sendOutgoingMessages(
	primaryClient *Client,
	primaryMsg *OutgoingMessage,
	secondaryMsg *OutgoingMessage,
	room *GameRoom) {
	var err error
	primaryOutput, err := json.Marshal(*primaryMsg)
	if err != nil {
		panic(err) // TODO: handle properly
	}

	var secondaryOutput []byte
	if secondaryMsg != nil {
		secondaryOutput, err = json.Marshal(secondaryMsg)
		if err != nil {
			panic(err) // TODO: handle properly
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
		} else if secondaryMsg != nil && (room == nil || h.playerClients[client].room == room) {
			select {
			case client.send <- secondaryOutput:
			default:
				close(client.send)
				delete(h.playerClients, client)
			}
		}
	}
}

func (h *Hub) removePlayerFromTeam(
	room *GameRoom,
	fromTeam int,
	playerName string,
) *PlayerInfo {
	players := room.teams[fromTeam]
	for idx, player := range players {
		if strings.Compare(player.name, playerName) == 0 {
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

func (h *Hub) convertPlayerInfosToPlayers(playerInfos []*PlayerInfo) []Player {
	players := make([]Player, 0, len(playerInfos))
	for _, playerInfo := range playerInfos {
		players = append(players, Player{
			Name:        playerInfo.name,
			IsRoomOwner: playerInfo.isRoomOwner,
			WordsSubmitted:    len(playerInfo.words) > 0,
		})
	}
	return players
}

func (h *Hub) convertTeamsPlayerInfosToTeams(teamsPlayerInfos [][]*PlayerInfo) [][]Player {
	teams := make([][]Player, 0, len(teamsPlayerInfos))
	for _, teamPlayerInfos := range teamsPlayerInfos {
		teams = append(teams, h.convertPlayerInfosToPlayers(teamPlayerInfos))
	}
	return teams
}

func getRandomNumberInRange(min, max int) int {
	if min == max {
		return min
	}

	return min + rand.Intn(max-min)
}
