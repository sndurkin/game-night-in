// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"strings"
	"encoding/json"
	"log"
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
	client *Client
	name   string
	room   *GameRoom
}

type GameRoom struct {
	roomCode   string
	gameType   string
	playerList []*PlayerInfo
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
			/*
				for client := range h.playerClients {
					select {
					case client.send <- message:
					default:
						close(client.send)
						delete(h.playerClients, client)
					}
				}
			*/
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
	case "join-room":
		var joinReq JoinRequest
		if err := json.Unmarshal(body, &joinReq); err != nil {
			//log.Fatal(err)
			panic(err)
		}
		h.joinRoom(clientMessage, joinReq)
	}
}

func (h *Hub) joinRoom(clientMessage *ClientMessage, joinReq JoinRequest) {
	log.Printf("Join request: %s\n", joinReq.Name)

	for _, playerInfo := range h.playerClients {
		if strings.EqualFold(playerInfo.name, joinReq.Name) {
			// TODO: check if same IP address so we can replace the playerClient with the one
			// that probably has lost connection?

			outgoingMessage.Event = "error"
			outgoingMessage.Error = "A player with that name is already in the room."
			sendOutgoingMessages(clientMessage, &outgoingMessage, nil)
			return
		}
	}

	h.playerClients[clientMessage.client].name = joinReq.Name
	_, ok := h.rooms[joinReq.RoomCode]
	if !ok {
		outgoingMessage.Event = "error"
		outgoingMessage.Error = "This room code does not exist."
		sendOutgoingMessages(clientMessage, &outgoingMessage, nil)
		return
	}

	players := make([]Player, 5)

	outgoingMessage.Event = "joined"
	outgoingMessage.Body = JoinedEvent{
		Players: players,
	}

	otherOutgoingMessage.Event = "player-joined"
	otherOutgoingMessage.Body = PlayerJoinedEvent{
		Name: joinReq.Name,
	}

	sendOutgoingMessages(clientMessage, &outgoingMessage, &otherOutgoingMessage)
}

func (h *Hub) sendOutgoingMessages(clientMessage *ClientMessage, outgoingMessage *OutgoingMessage,
	otherOutgoingMessage *OutgoingMessage) {
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
		} else if otherOutgoingMessage != nil { {
			select {
			case client.send <- otherOutput:
			default:
				close(client.send)
				delete(h.playerClients, client)
			}
		}
	}
}
