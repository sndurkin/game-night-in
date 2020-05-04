package main

import (
	"log"

	"github.com/sndurkin/gaming-remotely/api"
)

// This function must be called with the mutex held.
func (h *Hub) sendUpdatedGameMessages(room *GameRoom, justJoinedClient *Client) {
	game := room.game

	if game.state == "waiting-room" {
		var msg api.OutgoingMessage
		msg.Event = api.Event[api.EventUpdatedRoom]
		msg.Body = api.UpdatedRoomEvent{
			Teams: convertTeamsToAPITeams(room.teams),
			Settings: convertSettingsToAPISettings(room.settings),
		}

		log.Printf("Sending out updated room messages\n")

		if justJoinedClient != nil {
			log.Printf("Player just rejoined, sending updated-room event\n")
			h.sendOutgoingMessages(justJoinedClient, &msg, nil, room)
		} else {
			h.sendOutgoingMessages(nil, &msg, &msg, room)
		}
		return
	}

	currentPlayer := h.getCurrentPlayer(room)

	var currentCard string
	var currentServerTime int64
	var timerLength int
	if game.state == "turn-active" {
		currentCard = game.cardsInRound[0]

		if game.turnJustStarted || justJoinedClient != nil {
			currentServerTime = game.currentServerTime
			timerLength = game.timerLength
		}
	}

	var msgToCurrentPlayer api.OutgoingMessage
	msgToCurrentPlayer.Event = api.Event[api.EventUpdatedGame]
	msgToCurrentPlayer.Body = api.UpdatedGameEvent{
		State:                 game.state,
		LastCardGuessed:       game.lastCardGuessed,
		CurrentServerTime:     currentServerTime,
		TimerLength:           timerLength,
		CurrentCard:           currentCard,
		TotalNumCards:         game.totalNumCards,
		WinningTeam:           game.winningTeam,
		NumCardsLeftInRound:   len(game.cardsInRound),
		NumCardsGuessedInTurn: game.numCardsGuessedInTurn,
		TeamScoresByRound:     game.teamScoresByRound,
		CurrentRound:          game.currentRound,
		CurrentPlayers:        game.currentPlayers,
		CurrentlyPlayingTeam:  game.currentlyPlayingTeam,
	}

	var msgToOtherPlayers api.OutgoingMessage
	msgToOtherPlayers.Event = api.Event[api.EventUpdatedGame]
	msgToOtherPlayers.Body = api.UpdatedGameEvent{
		State:                 game.state,
		LastCardGuessed:       game.lastCardGuessed,
		CurrentServerTime:     currentServerTime,
		TimerLength:           timerLength,
		TotalNumCards:         game.totalNumCards,
		WinningTeam:           game.winningTeam,
		NumCardsLeftInRound:   len(game.cardsInRound),
		NumCardsGuessedInTurn: game.numCardsGuessedInTurn,
		TeamScoresByRound:     game.teamScoresByRound,
		CurrentRound:          game.currentRound,
		CurrentPlayers:        game.currentPlayers,
		CurrentlyPlayingTeam:  game.currentlyPlayingTeam,
	}

	if justJoinedClient != nil {
		log.Printf("Player %s just rejoined, sending updated-game event\n", currentPlayer.name)
		if currentPlayer.client == justJoinedClient {
			updatedGameEvent := msgToCurrentPlayer.Body.(api.UpdatedGameEvent)
			updatedGameEvent.Teams = convertTeamsToAPITeams(room.teams)
			msgToCurrentPlayer.Body = updatedGameEvent
			h.sendOutgoingMessages(justJoinedClient, &msgToCurrentPlayer,
				nil, room)
		} else {
			updatedGameEvent := msgToOtherPlayers.Body.(api.UpdatedGameEvent)
			updatedGameEvent.Teams = convertTeamsToAPITeams(room.teams)
			msgToOtherPlayers.Body = updatedGameEvent
			h.sendOutgoingMessages(justJoinedClient, &msgToOtherPlayers,
				nil, room)
		}
	} else {
		h.sendOutgoingMessages(currentPlayer.client, &msgToCurrentPlayer,
			&msgToOtherPlayers, room)
	}
}
