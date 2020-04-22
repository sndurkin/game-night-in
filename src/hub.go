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

type PlayerInfo struct {
	client      *Client
	name        string
	room        *GameRoom
	isRoomOwner bool
}

type GameRoom struct {
	roomCode string
	gameType string
	players  []*PlayerInfo
	teams    [][]*PlayerInfo
}

// PlayerClient is used to connect clients with player information.
type PlayerClient struct {
	client *Client
	player Player
}

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
		var createRoomReq CreateRoomRequest
		if err := json.Unmarshal(body, &createRoomReq); err != nil {
			//log.Fatal(err)
			panic(err)
		}
		h.createRoom(clientMessage, createRoomReq)
	case "join-room":
		var joinRoomReq JoinRoomRequest
		if err := json.Unmarshal(body, &joinRoomReq); err != nil {
			//log.Fatal(err)
			panic(err)
		}
		h.joinRoom(clientMessage, joinRoomReq)
	default:
		log.Fatalf("could not handle incoming action %s", incomingMessage.Action)
	}
}

func (h *Hub) createRoom(clientMessage *ClientMessage, createRoomReq CreateRoomRequest) {
	log.Printf("Create room request: %s\n", createRoomReq.Name)

	var outgoingMessage OutgoingMessage

	room := &GameRoom{
		gameType: createRoomReq.GameType,
		roomCode: strconv.Itoa(1000 + rand.Intn(9999-1000)),
		players:  make([]*PlayerInfo, 0),
		teams:    make([][]*PlayerInfo, 2),
	}
	h.rooms[room.roomCode] = room

	h.playerClients[clientMessage.client].name = createRoomReq.Name
	h.playerClients[clientMessage.client].room = room
	h.playerClients[clientMessage.client].isRoomOwner = true

	room.players = append(room.players, h.playerClients[clientMessage.client])

	room.teams[0] = make([]*PlayerInfo, 0)
	room.teams[0] = append(room.teams[0], h.playerClients[clientMessage.client])
	room.teams[1] = make([]*PlayerInfo, 0)

	log.Printf("%+v\n", h.rooms)

	outgoingMessage.Event = "created-room"
	outgoingMessage.Body = CreatedRoomEvent{
		RoomCode: room.roomCode,
		Team:     0,
	}

	h.sendOutgoingMessages(clientMessage, &outgoingMessage, nil, nil)
}

func (h *Hub) joinRoom(clientMessage *ClientMessage, joinRoomReq JoinRoomRequest) {
	log.Printf("Join room request: %s\n", joinRoomReq.Name)

	var outgoingMessage OutgoingMessage

	log.Printf("%+v\n", h.rooms)

	// TODO: add support for re-joining the game
	log.Printf("remoteAddr: %s\n", clientMessage.client.conn.UnderlyingConn().RemoteAddr())

	room, ok := h.rooms[joinRoomReq.RoomCode]
	if !ok {
		outgoingMessage.Event = "error"
		outgoingMessage.Error = "This room code does not exist."
		h.sendOutgoingMessages(clientMessage, &outgoingMessage, nil, nil)
		return
	}

	log.Printf("%+v\n", room)
	for _, playerInfo := range room.players {
		if strings.EqualFold(playerInfo.name, joinRoomReq.Name) {
			// TODO: check if same IP address so we can replace the playerClient with the one
			// that probably has lost connection?

			outgoingMessage.Event = "error"
			outgoingMessage.Error = "A player with that name is already in the room."
			h.sendOutgoingMessages(clientMessage, &outgoingMessage, nil, nil)
			return
		}
	}

	// Update the PlayerInfo instance with the room and chosen name.
	h.playerClients[clientMessage.client].name = joinRoomReq.Name
	h.playerClients[clientMessage.client].room = room

	room.players = append(room.players, h.playerClients[clientMessage.client])
	room.teams[0] = append(room.teams[0], h.playerClients[clientMessage.client])

	outgoingMessage.Event = "updated-room"
	outgoingMessage.Body = UpdatedRoomEvent{
		Teams: h.convertTeamsPlayerInfosToTeams(room.teams),
	}

	// var otherOutgoingMessage OutgoingMessage
	// otherOutgoingMessage.Event = "updated-room"
	// otherOutgoingMessage.Body = UpdatedRoomEvent{
	// 	Teams: h.convertTeamsPlayerInfosToTeams(room.teams),
	// }

	h.sendOutgoingMessages(clientMessage, &outgoingMessage, &outgoingMessage, room)
}

func (h *Hub) sendOutgoingMessages(clientMessage *ClientMessage, outgoingMessage *OutgoingMessage,
	otherOutgoingMessage *OutgoingMessage, room *GameRoom) {
	var err error
	output, err := json.Marshal(*outgoingMessage)
	if err != nil {
		panic(err) // TODO: handle properly
	}

	var otherOutput []byte
	if otherOutgoingMessage != nil {
		otherOutput, err = json.Marshal(otherOutgoingMessage)
		if err != nil {
			panic(err) // TODO: handle properly
		}
	}

	for client := range h.playerClients {
		if client == clientMessage.client {
			select {
			case client.send <- output:
			default:
				close(client.send)
				delete(h.playerClients, client)
			}
		} else if otherOutgoingMessage != nil && (room == nil || h.playerClients[client].room == room) {
			select {
			case client.send <- otherOutput:
			default:
				close(client.send)
				delete(h.playerClients, client)
			}
		}
	}
}

func (h *Hub) convertPlayerInfosToPlayers(playerInfos []*PlayerInfo) []Player {
	players := make([]Player, 0, len(playerInfos))
	for _, playerInfo := range playerInfos {
		players = append(players, Player{
			Name:    playerInfo.name,
			IsOwner: playerInfo.isRoomOwner,
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
