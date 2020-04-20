// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"log"
	"encoding/json"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// List of players
	playerClientsList []*PlayerClient

	// Inbound messages from the clients.
	message chan *ClientMessage

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

// PlayerClient is used to connect clients with player information.
type PlayerClient struct {
	client *Client
	player Player
}

func newHub() *Hub {
	return &Hub{
		message:  make(chan *ClientMessage),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case clientMessage := <-h.message:
			h.processIncomingMessage(clientMessage)
			/*
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
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
		panic(err)   // TODO: handle properly
	}

	var outgoingMessage OutgoingMessage
	var otherOutgoingMessage OutgoingMessage
	switch incomingMessage.Action {
	case "join":
		var joinReq JoinRequest
		if err := json.Unmarshal(body, &joinReq); err != nil {
			//log.Fatal(err)
			panic(err)
		}
		log.Printf("Join request: %s\n", joinReq.Name)

		var playerClient PlayerClient{
			player: Player{
				Name: joinReq.Name,
			},
			client: clientMessage.client,
		}
		if len(h.playerClientsList) == 0 {
			playerClient.player.IsOwner = true
		}
		h.playerClientsList = append(h.playerClientsList, &playerClient)

		players := make([]Player, len(h.playerClientsList))
		for i, v := range input {
			players = append(players, h.playerClientsList.player)
		}

		outgoingMessage.Event = "player-joined"
		outgoingMessage.Body = PlayerJoinedEvent{
			Players: players,
		}

		otherOutgoingMessage.Event = "other-player-joined"
		otherOutgoingMessage.Body = OtherPlayerJoinedEvent{
			Name: joinReq.Name,
		}
	}

	output, err := json.Marshal(outgoingMessage)
	if err != nil {
		panic(err)   // TODO: handle properly
	}

	otherOutput, err := json.Marshal(otherOutgoingMessage)
	if err != nil {
		panic(err)   // TODO: handle properly
	}

	for client := range h.clients {
		if client == clientMessage.client {
			select {
			case client.send <- output:
			default:
				close(client.send)
				delete(h.clients, client)
			}
		} else {
			select {
			case client.send <- otherOutput:
			default:
				close(client.send)
				delete(h.clients, client)
			}
		}
	}
}
