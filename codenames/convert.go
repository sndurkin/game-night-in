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

func convertTeamsToAPITeams(teams []*team) []codenames_api.Team {
	apiTeams := make([]codenames_api.Team, 0, len(teams))
	for _, team := range teams {
		apiTeams = append(apiTeams, convertTeamToAPITeam(team))
	}
	return apiTeams
}

func convertTeamToAPITeam(team *team) codenames_api.Team {
	apiPlayers := make([]*codenames_api.Player, 0, len(team.players))
	for _, player := range team.players {
		apiPlayers = append(apiPlayers, convertPlayerToAPIPlayer(player))
	}
	return codenames_api.Team{
		Players:     apiPlayers,
		CardIndices: team.cardIndices,
	}
}

func convertPlayerToAPIPlayer(player *models.Player) *codenames_api.Player {
	if player == nil {
		return nil
	}

	return &codenames_api.Player{
		Name: player.Name,
		IsRoomOwner: player.IsRoomOwner,
	}
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
