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

func convertSettingsToAPISettings(settings *GameSettings) api.GameSettings {
	apiRounds := make([]string, 0, len(settings.rounds))
	for _, round := range settings.rounds {
		apiRounds = append(apiRounds, api.Round[round])
	}

	return api.GameSettings{
		Rounds: apiRounds,
		TimerLength: settings.timerLength,
	}
}

func convertAPISettingsToSettings(apiSettings api.GameSettings) *GameSettings {
	rounds := make([]api.RoundT, 0, len(apiSettings.Rounds))
	for _, apiRound := range apiSettings.Rounds {
		rounds = append(rounds, api.RoundLookup[apiRound])
	}

	return &GameSettings{
		rounds: rounds,
		timerLength: apiSettings.TimerLength,
	}
}
