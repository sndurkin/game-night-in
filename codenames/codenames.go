package codenames

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"math/rand"
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
	cardIndicesGuessedInTurn []int
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

	content, err := ioutil.ReadFile("./codenames-words.txt")
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
	case codenames_api.ActionChangeCard:
		var req codenames_api.ChangeCardRequest
		if err := json.Unmarshal(body, &req); err != nil {
			log.Fatal(err)
		}
		g.changeCard(player, req)
	default:
		log.Fatalf("could not handle incoming action %s", incomingMessage.Action)
	}
}

func (g *Game) addTeam(
	player *models.Player,
	req codenames_api.AddTeamRequest,
) {
	log.Printf("Add team request\n")

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

	g.teams = append(g.teams, []*models.Player{})
	g.sendUpdatedGameMessages(nil)
}

func (g *Game) movePlayer(
	player *models.Player,
	req codenames_api.MovePlayerRequest,
) {
	log.Printf("Move player request: %s from %d to %d (%s)\n",
		req.PlayerName, req.FromTeam, req.ToTeam,
		req.ToTeamSpymasterRole ? "spymaster" : "guesser")

	var msg api.OutgoingMessage

	g.mutex.Lock()
	defer g.mutex.Unlock()

	room, err := g.performRoomChecks(player, true, false, false)
	if err != nil {
		g.sendErrorMessage(&models.ErrorMessageRequest{
			Player: player,
			Error:  err.Error(),
		})
		return
	}

	if req.FromTeam >= len(g.teams) || req.ToTeam >= len(g.teams) {
		msg.Event = "error"
		msg.Error = "The team indexes are invalid."
		g.sendOutgoingMessages(&models.OutgoingMessageRequest{
			PrimaryClient: player.Client,
			PrimaryMsg:    &msg,
		})
		return
	}

	fromTeam := g.teams[req.FromTeam]
	var playerToMove *models.Player
	if fromTeam.spymaster != nil && fromTeam.spymaster.Name == req.PlayerName {
		playerToMove = fromTeam.spymaster
		fromTeam.spymaster = nil
	}
	else if fromTeam.guesser != nil && fromTeam.guesser.Name == req.PlayerName {
		playerToMove = fromTeam.guesser
		fromTeam.guesser = nil
	}

	toTeam := g.teams[req.ToTeam]
	if req.ToTeamSpymasterRole {
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

	//g.cardIndicesGuessedInTurn = []int{}
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
		if g.assassinCardIdx == cardGuessIdx {
			if g.currentlyPlayingTeam == 0 {
				g.winningTeam = new(1)
			} else {
				g.winningTeam = new(0)
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
	g.teams[0] = append(g.teams[0], player)

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

	// Update the Player instance with the room and chosen name.
	player.Name = req.Name
	player.Room = g.room
	player.IsRoomOwner = false

	g.teams[0] = append(g.teams[0], player)

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
	g.initGameScores()
	g.lastCardGuessed = ""
	g.winningTeam = nil

	for _, teamPlayers := range g.teams {
		for _, player := range teamPlayers {
			playersSettings[player.Name].words = []string{}
		}
	}

	log.Println("Sending out updated game messages for rematch")
	g.sendUpdatedGameMessages(nil)
}

// This function must be called with the mutex held.
func (g *Game) removePlayerFromTeam(
	room *models.GameRoom,
	fromTeam int,
	playerName string,
) *models.Player {
	players := g.teams[fromTeam]
	for idx, player := range players {
		if player.Name == playerName {
			player := players[idx]
			g.teams[fromTeam] = append(
				players[:idx],
				players[idx+1:]...,
			)
			return player
		}
	}

	return nil
}

// This function must be called with the mutex held.
func (g *Game) reshuffleCardsForRound() {
	g.cardsInRound = []string{}
	for _, teamPlayers := range g.teams {
		for _, player := range teamPlayers {
			g.cardsInRound = append(g.cardsInRound,
				playersSettings[player.Name].words...)
		}
	}
	g.totalNumCards = len(g.cardsInRound)

	arr := g.cardsInRound
	rand.Shuffle(len(g.cardsInRound), func(i, j int) {
		arr[i], arr[j] = arr[j], arr[i]
	})
}

// This function must be called with the mutex held.
func (g *Game) moveToNextPlayerAndTeam() {
	t := g.currentlyPlayingTeam
	g.currentPlayers[t] = (g.currentPlayers[t] + 1) % len(g.teams[t])
	g.currentlyPlayingTeam = (t + 1) % len(g.currentPlayers)
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
		LastCardGuessed:       g.lastCardGuessed,
		CurrentServerTime:     currentServerTime,
		TimerLength:           timerLength,
		TotalNumCards:         g.totalNumCards,
		WinningTeam:           g.winningTeam,
		NumCardsLeftInRound:   len(g.cardsInRound),
		NumCardsGuessedInTurn: g.numCardsGuessedInTurn,
		TeamScoresByRound:     g.teamScoresByRound,
		CurrentRound:          g.currentRound,
		CurrentPlayers:        g.currentPlayers,
		CurrentlyPlayingTeam:  g.currentlyPlayingTeam,
	}

	if justJoinedClient != nil {
		log.Printf("models.Player %s just rejoined, sending updated-game event\n", currentPlayer.Name)
		if currentPlayer.Client == justJoinedClient {
			updatedGameEvent := msgToCurrentPlayer.Body.(codenames_api.UpdatedGameEvent)
			updatedGameEvent.Teams = convertTeamsToAPITeams(g.teams)
			msgToCurrentPlayer.Body = updatedGameEvent
			g.sendOutgoingMessages(&models.OutgoingMessageRequest{
				PrimaryClient: justJoinedClient,
				PrimaryMsg:    &msgToCurrentPlayer,
				Room:          room,
			})
		} else {
			updatedGameEvent := msgToOtherPlayers.Body.(codenames_api.UpdatedGameEvent)
			updatedGameEvent.Teams = convertTeamsToAPITeams(g.teams)
			msgToOtherPlayers.Body = updatedGameEvent
			g.sendOutgoingMessages(&models.OutgoingMessageRequest{
				PrimaryClient: justJoinedClient,
				PrimaryMsg:    &msgToOtherPlayers,
				Room:          room,
			})
		}
	} else {
		g.sendOutgoingMessages(&models.OutgoingMessageRequest{
			PrimaryClient: currentPlayer.Client,
			PrimaryMsg:    &msgToCurrentPlayer,
			SecondaryMsg:  &msgToOtherPlayers,
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
