package codenames

import (
	"github.com/sndurkin/game-night-in/models"
	codenames_api "github.com/sndurkin/game-night-in/codenames/api"
)

func convertPlayersToAPIPlayers(
	players []*models.Player,
) []codenames_api.Player {
	apiPlayers := make([]codenames_api.Player, 0, len(players))
	for _, player := range players {
		apiPlayers = append(apiPlayers, codenames_api.Player{
			Name:           player.Name,
			IsRoomOwner:    player.IsRoomOwner,
		})
	}
	return apiPlayers
}

func convertTeamsToAPITeams(
	teams [][]*models.Player,
) [][]codenames_api.Player {
	apiTeams := make([][]codenames_api.Player, 0, len(teams))
	for _, players := range teams {
		apiTeams = append(apiTeams, convertPlayersToAPIPlayers(players))
	}
	return apiTeams
}

func convertSettingsToAPISettings(
	settings *gameSettings,
) codenames_api.GameSettings {
	return codenames_api.GameSettings{

	}
}

func convertAPISettingsToSettings(
	apiSettings codenames_api.GameSettings,
) *gameSettings {
	return &gameSettings{

	}
}
