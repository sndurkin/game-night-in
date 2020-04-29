package main

import (
	"github.com/sndurkin/gaming-remotely/api"
)

func convertPlayersToAPIPlayers(players []*Player) []api.Player {
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

func convertTeamsToAPITeams(teams [][]*Player) [][]api.Player {
	apiTeams := make([][]api.Player, 0, len(teams))
	for _, players := range teams {
		apiTeams = append(apiTeams, convertPlayersToAPIPlayers(players))
	}
	return apiTeams
}
