package codenames

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"sync"
	"time"

	api "github.com/sndurkin/game-night-in/api"
	codenames_api "github.com/sndurkin/game-night-in/codenames/api"
	"github.com/sndurkin/game-night-in/models"
	"github.com/sndurkin/game-night-in/util"
)

// Game holds the game-specific data and logic.
type Game struct {
	mutex                *sync.RWMutex
	sendOutgoingMessages models.OutgoingMessageRequestFn
	sendErrorMessage     models.ErrorMessageRequestFn
	room                 *models.GameRoom
	settings             *gameSettings

	state             string
	turnJustStarted   bool
	currentServerTime int64
	timer             *time.Timer

	cards                    []string
	assassinCardIdx          int
	cardIndicesGuessed       []int
	numCardsInTurn           int

	teams                []*team
	winningTeam          *int
	currentlyPlayingTeam int // 0 or 1

	previouslyUsedCards []string
}

type team struct {
	spymaster   *models.Player `json:"spymaster"`
	guesser     *models.Player `json:"guesser"`
	cardIndices []int          `json:"cardIndices"`
}

// gameSettings holds all the data about the
// game settings.
type gameSettings struct {

}

var (
	validStateTransitions = map[string][]string{
		"waiting-room": {
			"turn-start",
		},
		"turn-start": {
			"turn-active",
		},
		"turn-active": {
			"turn-start",
			"game-over",
		},
		"game-over": {
			"waiting-room",
		},
	}

	allCards []string
)

// Init is called on program startup.
func Init() {
	codenames_api.Init()

	content, err := ioutil.ReadFile("./codenames/codenames-words.txt")
	if err != nil {
		log.Fatal(err)
	}
	allCards = strings.Split(strings.Replace(string(content), "\r\n", "\n", -1), "\n")
}

func NewGame(
	gameRoom *models.GameRoom,
	mutex *sync.RWMutex,
	sendOutgoingMessages models.OutgoingMessageRequestFn,
	sendErrorMessage models.ErrorMessageRequestFn,
) *Game {
	g := &Game{
		mutex:                mutex,
		sendOutgoingMessages: sendOutgoingMessages,
		sendErrorMessage:     sendErrorMessage,
		settings: &gameSettings{},
		room:  gameRoom,
		state: "waiting-room",
		teams: make([]*team, 2),
		cards: make([]string, 0),
	}

	// Generate 25 unique cards for the game.
	cardIndices := make(map[int]bool, 25)
	for i := 0; i < 25; i++ {
		for {
			newCardIdx := util.GetRandomNumberInRange(0, len(allCards)-1)
			if _, exists := cardIndices[newCardIdx]; exists {
				continue
			}

			cardIndices[newCardIdx] = true
			g.cards = append(g.cards, allCards[newCardIdx])
			break
		}
	}

	// Generate the assassin card.
	cardIndices = make(map[int]bool, 12)
	g.assassinCardIdx = util.GetRandomNumberInRange(0, 24)
	cardIndices[g.assassinCardIdx] = true

	// Generate the first team's cards.
	g.teams[0] = &team{}
	for i := 0; i < 6; i++ {
		for {
			newCardIdx := util.GetRandomNumberInRange(0, 24)
			if _, exists := cardIndices[newCardIdx]; exists {
				continue
			}

			cardIndices[newCardIdx] = true
			g.teams[0].cardIndices = append(g.teams[0].cardIndices, newCardIdx)
			break
		}
	}

	// Generate the second team's cards.
	g.teams[1] = &team{}
	for i := 0; i < 5; i++ {
		for {
			newCardIdx := util.GetRandomNumberInRange(0, 24)
			if _, exists := cardIndices[newCardIdx]; exists {
				continue
			}

			cardIndices[newCardIdx] = true
			g.teams[1].cardIndices = append(g.teams[1].cardIndices, newCardIdx)
			break
		}
	}

	return g
}

func (g *Game) HandleIncomingMessage(
	player *models.Player,
	incomingMessage api.IncomingMessage,
	body json.RawMessage,
) {
	actionType, ok := codenames_api.ActionLookup[incomingMessage.Action]
	if !ok {
		log.Fatalf("invalid codenames action: %s\n", incomingMessage.Action)
	}

	switch actionType {
	case codenames_api.ActionMovePlayer:
		var req codenames_api.MovePlayerRequest
		if err := json.Unmarshal(body, &req); err != nil {
			log.Fatal(err)
		}
		g.movePlayer(player, req)
	case codenames_api.ActionChangeSettings:
		var req codenames_api.ChangeSettingsRequest
		if err := json.Unmarshal(body, &req); err != nil {
			log.Fatal(err)
		}
		g.changeSettings(player, req)
	case codenames_api.ActionStartTurn:
		var req codenames_api.StartTurnRequest
		if err := json.Unmarshal(body, &req); err != nil {
			log.Fatal(err)
		}
		g.startTurn(player, req)
	case codenames_api.ActionEndTurn:
		var req codenames_api.EndTurnRequest
		if err := json.Unmarshal(body, &req); err != nil {
			log.Fatal(err)
		}
		g.endTurn(player, req)
	default:
		log.Fatalf("could not handle incoming action %s", incomingMessage.Action)
	}
}

func (g *Game) movePlayer(
	player *models.Player,
	req codenames_api.MovePlayerRequest,
) {
	var roleName string
	if req.ToPlayerType == codenames_api.PlayerSpymaster {
		roleName = "spymaster"
	} else {
		roleName = "guesser"
	}
	log.Printf("Move player request: %s to %d (%s)\n", req.PlayerName,
		req.ToTeam, roleName)

	var msg api.OutgoingMessage

	g.mutex.Lock()
	defer g.mutex.Unlock()

	_, err := g.performRoomChecks(player, true, false, false)
	if err != nil {
		g.sendErrorMessage(&models.ErrorMessageRequest{
			Player: player,
			Error:  err.Error(),
		})
		return
	}

	if req.ToTeam < 0 || req.ToTeam >= len(g.teams) {
		msg.Event = "error"
		msg.Error = "The team indexes are invalid."
		g.sendOutgoingMessages(&models.OutgoingMessageRequest{
			PrimaryClient: player.Client,
			PrimaryMsg:    &msg,
		})
		return
	}

	toTeam := g.teams[req.ToTeam]
	var playerToSwitchWith *models.Player
	if req.ToPlayerType == codenames_api.PlayerSpymaster {
		if toTeam.spymaster != nil {
			playerToSwitchWith = toTeam.spymaster
		}
	} else {
		if toTeam.guesser != nil {
			playerToSwitchWith = toTeam.guesser
		}
	}

	var playerToMove *models.Player
	for _, team := range g.teams {
		if team.spymaster != nil && team.spymaster.Name == req.PlayerName {
			playerToMove = team.spymaster
			team.spymaster = playerToSwitchWith
			break
		} else if team.guesser != nil && team.guesser.Name == req.PlayerName {
			playerToMove = team.guesser
			team.guesser = playerToSwitchWith
			break
		}
	}

	if req.ToPlayerType == codenames_api.PlayerSpymaster {
		toTeam.spymaster = playerToMove
	} else {
		toTeam.guesser = playerToMove
	}

	g.sendUpdatedGameMessages(nil)
}

func (g *Game) changeSettings(
	player *models.Player,
	req codenames_api.ChangeSettingsRequest,
) {
	log.Printf("Change settings request\n")

	g.mutex.Lock()
	defer g.mutex.Unlock()

	_, err := g.performRoomChecks(player, true, false, false)
	if err != nil {
		g.sendErrorMessage(&models.ErrorMessageRequest{
			Player: player,
			Error:  err.Error(),
		})
		return
	}

	g.settings = convertAPISettingsToSettings(req.Settings)
	g.sendUpdatedGameMessages(nil)
}

func (g *Game) startTurn(
	player *models.Player,
	req codenames_api.StartTurnRequest,
) {
	log.Printf("Start turn request\n")

	g.mutex.Lock()
	defer g.mutex.Unlock()

	_, err := g.performRoomChecks(player, false, true, false)
	if err != nil {
		g.sendErrorMessage(&models.ErrorMessageRequest{
			Player: player,
			Error:  err.Error(),
		})
		return
	}

	if !g.validateStateTransition(g.state, "turn-active") {
		g.sendErrorMessage(&models.ErrorMessageRequest{
			Player: player,
			Error:  "You cannot perform that action at this time.",
		})
		return
	}
	g.turnJustStarted = true
	g.state = "turn-active"

	g.numCardsInTurn = req.NumCards

	g.sendUpdatedGameMessages(nil)
}

func (g *Game) endTurn(
	player *models.Player,
	req codenames_api.EndTurnRequest,
) {
	log.Printf("End turn request\n")

	g.mutex.Lock()
	defer g.mutex.Unlock()

	_, err := g.performRoomChecks(player, false, false, true)
	if err != nil {
		g.sendErrorMessage(&models.ErrorMessageRequest{
			Player: player,
			Error:  err.Error(),
		})
		return
	}

	if !g.validateStateTransition(g.state, "turn-start") {
		g.sendErrorMessage(&models.ErrorMessageRequest{
			Player: player,
			Error:  "You cannot perform that action at this time.",
		})
		return
	}
	g.state = "turn-start"

	if g.numCardsInTurn > len(req.CardGuessIndices) {
		g.sendErrorMessage(&models.ErrorMessageRequest{
			Player: player,
			Error:  "You cannot make that many guesses.",
		})
		return
	}

	for _, cardGuessIdx := range req.CardGuessIndices {
		if util.IntInSlice(g.cardIndicesGuessed, cardGuessIdx) {
			g.sendErrorMessage(&models.ErrorMessageRequest{
				Player: player,
				Error:  fmt.Sprintf("Card \"%s\" has already been guessed.",
					g.cards[cardGuessIdx]),
			})
			return
		}
	}

	for _, cardGuessIdx := range req.CardGuessIndices {
		if g.assassinCardIdx == cardGuessIdx {
			if g.currentlyPlayingTeam == 0 {
				*g.winningTeam = 1
			} else {
				*g.winningTeam = 0
			}

			g.state = "game-over"
			g.sendUpdatedGameMessages(nil)
			return
		}

		g.cardIndicesGuessed = append(g.cardIndicesGuessed, cardGuessIdx)
	}

	g.sendUpdatedGameMessages(nil)
}

// AddPlayer adds a player to the current game.
//
// This function must be called with the mutex held.
func (g *Game) AddPlayer(player *models.Player) {
	g.teams[0].spymaster = player

	var msg api.OutgoingMessage
	msg.Event = api.Event[api.EventCreatedGame]
	msg.Body = codenames_api.CreatedGameEvent{
		GameType: g.room.GameType,
		RoomCode: g.room.RoomCode,
		Team:     0,
	}

	g.sendOutgoingMessages(&models.OutgoingMessageRequest{
		PrimaryClient: player.Client,
		PrimaryMsg:    &msg,
	})
}

func (g *Game) Join(
	player *models.Player,
	newPlayerJoined bool,
	req api.JoinGameRequest,
) {
	if g.state != "waiting-room" {
		g.sendErrorMessage(&models.ErrorMessageRequest{
			Player: player,
			Error:  "You cannot join a game that has already started.",
		})
		return
	}

	if !newPlayerJoined {
		// Only the player's connection was updated, a new player has not joined.
		g.sendUpdatedGameMessages(nil)
		return
	}

	if g.teams[0].spymaster == nil {
		g.teams[0].spymaster = player
	} else if g.teams[0].guesser == nil {
		g.teams[0].guesser = player
	} else if g.teams[1].spymaster == nil {
		g.teams[1].spymaster = player
	} else if g.teams[1].guesser == nil {
		g.teams[1].guesser = player
	} else {
		g.sendErrorMessage(&models.ErrorMessageRequest{
			Player: player,
			Error:  "This game is full.",
		})
		return
	}

	// Update the Player instance with the room and chosen name.
	player.Name = req.Name
	player.Room = g.room
	player.IsRoomOwner = false

	g.sendUpdatedGameMessages(nil)
}

// Start starts the game.
//
// This function must be called with the mutex held.
func (g *Game) Start(player *models.Player) {
	if !g.validateStateTransition(g.state, "turn-start") {
		g.sendErrorMessage(&models.ErrorMessageRequest{
			Player: player,
			Error:  "You cannot perform that action at this time.",
		})
		return
	}
	g.state = "turn-start"

	g.cardIndicesGuessed = []int{}
	g.currentlyPlayingTeam = 0

	g.sendUpdatedGameMessages(nil)
}

// Kick removes a player from the game.
//
// This function must be called with the mutex held.
func (g *Game) Kick(playerName string) {
	// TODO

	g.sendUpdatedGameMessages(nil)
}

// Rematch starts a new game with the same players and settings.
//
// This function must be called with the mutex held.
func (g *Game) Rematch(player *models.Player) {
	if !g.validateStateTransition(g.state, "waiting-room") {
		g.sendErrorMessage(&models.ErrorMessageRequest{
			Player: player,
			Error:  "You cannot perform that action at this time.",
		})
		return
	}

	g.state = "waiting-room"
	g.cardIndicesGuessed = []int{}
	g.winningTeam = nil

	log.Println("Sending out updated game messages for rematch")
	g.sendUpdatedGameMessages(nil)
}

// This function must be called with the mutex held.
func (g *Game) removePlayerFromTeam(
	room *models.GameRoom,
	fromTeam int,
	playerName string,
) *models.Player {
	team := g.teams[fromTeam]

	if team.spymaster != nil && team.spymaster.Name == playerName {
		player := team.spymaster
		team.spymaster = nil
		return player
	}

	if team.guesser != nil && team.guesser.Name == playerName {
		player := team.guesser
		team.guesser = nil
		return player
	}

	return nil
}

// This function must be called with the mutex held.
func (g *Game) getCurrentSpymaster(room *models.GameRoom) *models.Player {
	team := g.teams[g.currentlyPlayingTeam]
	return team.spymaster
}

// This function must be called with the mutex held.
func (g *Game) getCurrentGuesser(room *models.GameRoom) *models.Player {
	team := g.teams[g.currentlyPlayingTeam]
	return team.guesser
}

// This function must be called with the mutex held.
func (g *Game) sendUpdatedGameMessages(justJoinedClient interface{}) {
	room := g.room
	if g.state == "waiting-room" {
		var msg api.OutgoingMessage
		msg.Event = api.Event[api.EventUpdatedRoom]
		msg.Body = codenames_api.UpdatedRoomEvent{
			GameType: room.GameType,
			Teams:    convertTeamsToAPITeams(g.teams),
			Settings: convertSettingsToAPISettings(g.settings),
		}

		log.Printf("Sending out updated room messages\n")

		if justJoinedClient != nil {
			log.Printf("models.Player just rejoined, sending updated-room event\n")
			g.sendOutgoingMessages(&models.OutgoingMessageRequest{
				PrimaryClient: justJoinedClient,
				PrimaryMsg:    &msg,
				Room:          room,
			})
		} else {
			g.sendOutgoingMessages(&models.OutgoingMessageRequest{
				PrimaryMsg:   &msg,
				SecondaryMsg: &msg,
				Room:         room,
			})
		}
		return
	}
	/*
	var msgToCurrentPlayer api.OutgoingMessage
	msgToCurrentPlayer.Event = api.Event[api.EventUpdatedGame]
	msgToCurrentPlayer.Body = codenames_api.UpdatedGameEvent{
		State:                 g.state,
		LastCardGuessed:       g.lastCardGuessed,
		CurrentServerTime:     currentServerTime,
		TimerLength:           timerLength,
		CurrentCard:           currentCard,
		TotalNumCards:         g.totalNumCards,
		WinningTeam:           g.winningTeam,
		NumCardsLeftInRound:   len(g.cardsInRound),
		NumCardsGuessedInTurn: g.numCardsGuessedInTurn,
		TeamScoresByRound:     g.teamScoresByRound,
		CurrentRound:          g.currentRound,
		CurrentPlayers:        g.currentPlayers,
		CurrentlyPlayingTeam:  g.currentlyPlayingTeam,
	}
	*/
	var msgToOtherPlayers api.OutgoingMessage
	msgToOtherPlayers.Event = api.Event[api.EventUpdatedGame]
	msgToOtherPlayers.Body = codenames_api.UpdatedGameEvent{
		State:                 g.state,
		CardIndicesGuessed:    g.cardIndicesGuessed,
		WinningTeam:           g.winningTeam,
		CurrentlyPlayingTeam:  g.currentlyPlayingTeam,
		//Teams:                 convertTeamsToAPITeams(g.teams),
	}

	if justJoinedClient != nil {
		var justJoinedPlayer *models.Player
		for _, player := range room.Players {
			if player.Client == justJoinedClient {
				justJoinedPlayer = player
				break
			}
		}

		log.Printf(
			"Player %s just rejoined, sending updated-game event\n",
			justJoinedPlayer.Name,
		)

		updatedGameEvent := msgToOtherPlayers.Body.(codenames_api.UpdatedGameEvent)
		updatedGameEvent.Teams = convertTeamsToAPITeams(g.teams)
		msgToOtherPlayers.Body = updatedGameEvent
		g.sendOutgoingMessages(&models.OutgoingMessageRequest{
			PrimaryClient: justJoinedClient,
			PrimaryMsg:    &msgToOtherPlayers,
			Room:          room,
		})
	} else {
		g.sendOutgoingMessages(&models.OutgoingMessageRequest{
			PrimaryMsg:    &msgToOtherPlayers,
			Room:          room,
		})
	}
}

// This function must be called with the mutex held.
func (g *Game) validateStateTransition(fromState, toState string) bool {
	valid, ok := validStateTransitions[fromState]
	if !ok {
		return false
	}

	if !util.StringInSlice(valid, toState) {
		return false
	}

	return true
}

func (g *Game) performRoomChecks(
	player *models.Player,
	playerMustBeRoomOwner bool,
	playerMustBeCurrentSpymaster bool,
	playerMustBeCurrentGuesser bool,
) (*models.GameRoom, error) {
	room := player.Room
	if room == nil {
		return nil, errors.New("you are not in a game")
	}

	/* TODO
	if _, ok := h.rooms[room.roomCode]; !ok {
		return nil, errors.New("this game no longer exists")
	}
	*/

	room.LastInteractionTime = time.Now()

	if playerMustBeRoomOwner && !player.IsRoomOwner {
		return nil, errors.New("you are not the game owner")
	}

	if playerMustBeCurrentSpymaster {
		currentSpymaster := g.getCurrentSpymaster(room)
		if currentSpymaster.Name != player.Name {
			return nil, errors.New("you are not the current spymaster")
		}
	}

	if playerMustBeCurrentGuesser {
		currentGuesser := g.getCurrentGuesser(room)
		if currentGuesser.Name != player.Name {
			return nil, errors.New("you are not the current guesser")
		}
	}

	return room, nil
}
