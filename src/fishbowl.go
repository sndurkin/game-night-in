package main

import (
	"./api"
)


func (h *Hub) sendUpdatedGameMessages(room *GameRoom) {
	game := room.game
	currentPlayer := h.getCurrentPlayer(room)

	var currentCard string
	var currentServerTime int64
	var timerLength int
	if game.state == "turn-active" {
		currentCard = game.cardsInRound[0]

		if game.turnJustStarted {
			currentServerTime = game.currentServerTime
			timerLength = game.timerLength
		}
	}

	var msgToCurrentPlayer api.OutgoingMessage
	msgToCurrentPlayer.Event = "updated-game"
	msgToCurrentPlayer.Body = api.UpdatedGameEvent{
		State:                 game.state,
		LastCardGuessed:       game.lastCardGuessed,
		CurrentServerTime:     currentServerTime,
		TimerLength:           timerLength,
		CurrentCard:           currentCard,
		NumCardsLeftInRound:   len(game.cardsInRound),
		NumCardsGuessedInTurn: game.numCardsGuessedInTurn,
		TeamScoresByRound:     game.teamScoresByRound,
		CurrentRound:          game.currentRound,
		CurrentPlayers:        game.currentPlayers,
		CurrentlyPlayingTeam:  game.currentlyPlayingTeam,
	}

	var msgToOtherPlayers api.OutgoingMessage
	msgToOtherPlayers.Event = "updated-game"
	msgToOtherPlayers.Body = api.UpdatedGameEvent{
		State:                 game.state,
		LastCardGuessed:       game.lastCardGuessed,
		CurrentServerTime:     currentServerTime,
		TimerLength:           timerLength,
		NumCardsLeftInRound:   len(game.cardsInRound),
		NumCardsGuessedInTurn: game.numCardsGuessedInTurn,
		TeamScoresByRound:     game.teamScoresByRound,
		CurrentRound:          game.currentRound,
		CurrentPlayers:        game.currentPlayers,
		CurrentlyPlayingTeam:  game.currentlyPlayingTeam,
	}

	h.sendOutgoingMessages(currentPlayer.client, &msgToCurrentPlayer,
		&msgToOtherPlayers, room)
}
