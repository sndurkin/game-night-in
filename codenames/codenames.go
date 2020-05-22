package codenames

import (
	"encoding/json"
	"errors"
	"log"
	"math/rand"
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

	state                 string
	turnJustStarted       bool
	currentServerTime     int64
	timer                 *time.Timer

	cards                 []string
	assassinCard          int
	cardsGuessed          []int
	cardsGuessedInTurn    []int
	numCardsInTurn        int

	winningTeam           *int
	currentlyPlayingTeam  int   // 0, 1, ...

	previouslyUsedCards   []string
}

// gameSettings holds all the data about the
// game settings.
type gameSettings struct {
	rounds      []codenames_api.RoundT
	timerLength int
	useTimer    bool
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
)

func NewGame(
	gameRoom *models.GameRoom,
	mutex *sync.RWMutex,
	sendOutgoingMessages models.OutgoingMessageRequestFn,
	sendErrorMessage models.ErrorMessageRequestFn,
) *Game {
	g := &Game{
		mutex: mutex,
		sendOutgoingMessages: sendOutgoingMessages,
		sendErrorMessage: sendErrorMessage,
		settings: &gameSettings{
			useTimer: false,
			timerLength: 90,
		},
		room:  gameRoom,
		state: "waiting-room",
		teams: make([][]*models.Player, 2),
	}

	g.teams[0] = make([]*models.Player, 0)
	g.teams[1] = make([]*models.Player, 0)

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
	case codenames_api.ActionAddTeam:
		var req codenames_api.AddTeamRequest
		if err := json.Unmarshal(body, &req); err != nil {
			log.Fatal(err)
		}
		g.addTeam(player, req)
	case codenames_api.ActionRemoveTeam:
		// TODO
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
	case codenames_api.ActionSubmitWords:
		var req codenames_api.SubmitWordsRequest
		if err := json.Unmarshal(body, &req); err != nil {
			log.Fatal(err)
		}
		g.submitWords(player, req)
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

	_, err := g.performRoomChecks(player, true, false)
	if err != nil {
		g.sendErrorMessage(&models.ErrorMessageRequest{
			Player: player,
			Error: err.Error(),
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
	log.Printf("Move player request: %s (%d -> %d)\n", req.PlayerName,
		req.FromTeam, req.ToTeam)

	var msg api.OutgoingMessage

	g.mutex.Lock()
	defer g.mutex.Unlock()

	room, err := g.performRoomChecks(player, true, false)
	if err != nil {
		g.sendErrorMessage(&models.ErrorMessageRequest{
			Player: player,
			Error: err.Error(),
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

	playerToMove := g.removePlayerFromTeam(room, req.FromTeam, req.PlayerName)
	g.teams[req.ToTeam] = append(g.teams[req.ToTeam], playerToMove)

	g.sendUpdatedGameMessages(nil)
}

func (g *Game) changeSettings(
	player *models.Player,
	req codenames_api.ChangeSettingsRequest,
) {
	log.Printf("Change settings request\n")

	g.mutex.Lock()
	defer g.mutex.Unlock()

	_, err := g.performRoomChecks(player, true, false)
	if err != nil {
		g.sendErrorMessage(&models.ErrorMessageRequest{
			Player: player,
			Error: err.Error(),
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

	_, err := g.performRoomChecks(player, false, true)
	if err != nil {
		g.sendErrorMessage(&models.ErrorMessageRequest{
			Player: player,
			Error: err.Error(),
		})
		return
	}

	if !g.validateStateTransition(g.state, "turn-active") {
		g.sendErrorMessage(&models.ErrorMessageRequest{
			Player: player,
			Error: "You cannot perform that action at this time.",
		})
		return
	}
	g.turnJustStarted = true
	g.state = "turn-active"

	g.numCardsGuessedInTurn = 0
	g.lastCardGuessed = ""
	g.timerLength = g.settings.timerLength + 1
	g.currentServerTime = time.Now().UnixNano() / 1000000
	if g.timer != nil {
		g.timer.Stop()
	}
	g.timer = time.NewTimer(time.Second * time.Duration(g.timerLength))

	// Wait for timer to finish in an asynchronous goroutine
	go func() {
		// Block until timer finishes. When it is done, it sends a message
		// on the channel timer.C. No other code in
		// this goroutine is executed until that happens.
		<-g.timer.C

		log.Println("Timer expired, waiting on lock")

		g.mutex.Lock()
		defer g.mutex.Unlock()

		log.Println(" - lock obtained")

		g.timer = nil

		log.Printf("Game state when timer ended: %s\n", g.state)
		if !g.validateStateTransition(g.state, "turn-start") {
			if g.state == "turn-start" || g.state == "game-over" {
				// Round or game finished before the player's turn timer expired,
				// so do nothing.
				return
			}

			log.Fatalf("Game was not in correct state when turn timer expired: %s", g.state)
			return
		}

		g.turnJustStarted = false
		g.state = "turn-start"
		g.moveToNextPlayerAndTeam()

		log.Printf("Sending updated game message after timer expired\n")
		g.sendUpdatedGameMessages(nil)
	}()

	g.sendUpdatedGameMessages(nil)
}

func (g *Game) submitWords(
	player *models.Player,
	req codenames_api.SubmitWordsRequest,
) {
	log.Printf("Submit words request: %s\n", req.Words)

	g.mutex.Lock()
	defer g.mutex.Unlock()

	_, err := g.performRoomChecks(player, false, false)
	if err != nil {
		g.sendErrorMessage(&models.ErrorMessageRequest{
			Player: player,
			Error: err.Error(),
		})
		return
	}

	playersSettings[player.Name].words = req.Words
	g.sendUpdatedGameMessages(nil)
}

func (g *Game) changeCard(
	player *models.Player,
	req codenames_api.ChangeCardRequest,
) {
	log.Printf("Change card request: %s\n", req.ChangeType)

	g.mutex.Lock()
	defer g.mutex.Unlock()

	_, err := g.performRoomChecks(player, false, true)
	if err != nil {
		g.sendErrorMessage(&models.ErrorMessageRequest{
			Player: player,
			Error: err.Error(),
		})
		return
	}

	if g.state != "turn-active" {
		// Ignore, the turn is probably over.
		return
	}

	g.turnJustStarted = false
	if req.ChangeType == "correct" {
		// Increment score for current team and the current turn.
		g.teamScoresByRound[g.currentRound][g.currentlyPlayingTeam]++
		g.numCardsGuessedInTurn++

		g.lastCardGuessed = g.cardsInRound[0]
		g.cardsInRound = g.cardsInRound[1:]

		if len(g.cardsInRound) == 0 {
			if g.timer != nil {
				g.timer.Stop()
			}

			g.currentRound++
			if g.currentRound < len(g.settings.rounds) {
				// Round over, moving to next round
				g.state = "turn-start"

				g.reshuffleCardsForRound()
				g.moveToNextPlayerAndTeam()

				// Each round should start with a different team.
				//
				// TODO: The teams should be re-ordered based on score.
				g.currentlyPlayingTeam = util.GetRandomNumberInRange(0,
					len(g.teams)-1)
			} else {
				g.state = "game-over" // TODO: update to use constant from api.go

				totalScores := make([]int, len(g.teams))
				for _, scoresByTeam := range g.teamScoresByRound {
					for team, score := range scoresByTeam {
						totalScores[team] += score
					}
				}

				var teamWithMax, max int
				for team, totalScore := range totalScores {
					if totalScore > max {
						max = totalScore
						teamWithMax = team
					}
				}

				g.winningTeam = &teamWithMax
			}
		}
	} else {
		// Skip this card, push it to the end
		g.cardsInRound = append(g.cardsInRound[1:], g.cardsInRound[0])
	}

	g.sendUpdatedGameMessages(nil)
}

// AddPlayer adds a player to the current game.
//
// This function must be called with the mutex held.
func (g *Game) AddPlayer(player *models.Player) {
	playersSettings[player.Name] = &playerSettings{
		words: []string{},
	}

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
			Error: "You cannot join a game that has already started.",
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
	playersSettings[player.Name] = &playerSettings{
		words: []string{},
	}

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
			Error: "You cannot perform that action at this time.",
		})
		return
	}
	g.turnJustStarted = true
	g.state = "turn-start"

	g.reshuffleCardsForRound()
	g.initGameScores()

	g.currentRound = 0
	g.currentPlayers = make([]int, len(g.teams))
	for i, players := range g.teams {
		g.currentPlayers[i] = util.GetRandomNumberInRange(0, len(players)-1)
	}
	g.currentlyPlayingTeam = util.GetRandomNumberInRange(0, len(g.teams)-1)

	g.sendUpdatedGameMessages(nil)
}

// Rematch starts a new game with the same players and settings.
//
// This function must be called with the mutex held.
func (g *Game) Rematch(player *models.Player) {
	if !g.validateStateTransition(g.state, "waiting-room") {
		g.sendErrorMessage(&models.ErrorMessageRequest{
			Player: player,
			Error: "You cannot perform that action at this time.",
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

func (g *Game) initGameScores() {
	g.teamScoresByRound = make([][]int, len(g.settings.rounds))
	for idx := range g.settings.rounds {
		g.teamScoresByRound[idx] = make([]int, len(g.teams))
	}
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
func (g *Game) getCurrentPlayer(room *models.GameRoom) *models.Player {
	players := g.teams[g.currentlyPlayingTeam]
	return players[g.currentPlayers[g.currentlyPlayingTeam]]
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

	currentPlayer := g.getCurrentPlayer(room)

	var currentCard string
	var currentServerTime int64
	var timerLength int
	if g.state == "turn-active" {
		currentCard = g.cardsInRound[0]

		if g.turnJustStarted || justJoinedClient != nil {
			currentServerTime = g.currentServerTime
			timerLength = g.timerLength
		}
	}

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
	playerMustBeCurrentPlayer bool,
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

	if playerMustBeCurrentPlayer {
		currentPlayer := g.getCurrentPlayer(room)
		if currentPlayer.Name != player.Name {
			return nil, errors.New("you are not the current player")
		}
	}

	return room, nil
}
