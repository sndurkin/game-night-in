package fishbowl

import (
	"github.com/sndurkin/game-night-in/models"
	fishbowl_api "github.com/sndurkin/game-night-in/fishbowl/api"
)

func convertPlayersToAPIPlayers(
	players []*models.Player,
	settings *gameSettings,
) []fishbowl_api.Player {
	apiPlayers := make([]fishbowl_api.Player, 0, len(players))
	for _, player := range players {
		apiPlayers = append(apiPlayers, fishbowl_api.Player{
			Name:           player.Name,
			IsRoomOwner:    player.IsRoomOwner,
			WordsSubmitted: len(playersSettings[player.Name].words) >= settings.numWordsRequired,
		})
	}
	return apiPlayers
}

func convertTeamsToAPITeams(
	teams [][]*models.Player,
	settings *gameSettings,
) [][]fishbowl_api.Player {
	apiTeams := make([][]fishbowl_api.Player, 0, len(teams))
	for _, players := range teams {
		apiTeams = append(apiTeams,
			convertPlayersToAPIPlayers(players, settings))
	}
	return apiTeams
}

func convertSettingsToAPISettings(
	settings *gameSettings,
) fishbowl_api.GameSettings {
	apiRounds := make([]string, 0, len(settings.rounds))
	for _, round := range settings.rounds {
		apiRounds = append(apiRounds, fishbowl_api.Round[round])
	}

	return fishbowl_api.GameSettings{
		Rounds:           apiRounds,
		TimerLength:      settings.timerLength,
		NumWordsRequired: settings.numWordsRequired,
		MaxSkipsPerTurn:  settings.maxSkipsPerTurn,
	}
}

func convertAPISettingsToSettings(
	apiSettings fishbowl_api.GameSettings,
) *gameSettings {
	rounds := make([]fishbowl_api.RoundT, 0, len(apiSettings.Rounds))
	for _, apiRound := range apiSettings.Rounds {
		rounds = append(rounds, fishbowl_api.RoundLookup[apiRound])
	}

	return &gameSettings{
		rounds:           rounds,
		timerLength:      apiSettings.TimerLength,
		numWordsRequired: apiSettings.NumWordsRequired,
		maxSkipsPerTurn:  apiSettings.MaxSkipsPerTurn,
	}
}
