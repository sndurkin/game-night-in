package fishbowl

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/sndurkin/game-night-in/models"
	"github.com/sndurkin/game-night-in/util"
	api "github.com/sndurkin/game-night-in/api"
	fishbowl_api "github.com/sndurkin/game-night-in/fishbowl/api"
)

// FishbowlPlayerSettings holds the game-specific data about a particular player.
type FishbowlPlayerSettings struct {
	words       []string
}

// FishbowlGame holds the game-specific data and logic.
type FishbowlGame struct {
	mutex *sync.RWMutex
	room *models.GameRoom
	settings *FishbowlGameSettings

	state                 string
	turnJustStarted       bool
	cardsInRound          []string
	currentServerTime     int64
	timer                 *time.Timer
	timerLength           int
	lastCardGuessed       string
	totalNumCards         int
	numCardsGuessedInTurn int
	teams                 [][]*models.Player
	teamScoresByRound     [][]int
	winningTeam           *int
	currentRound          int   // 0, 1, 2
	currentPlayers        []int // [ team0PlayerIdx, team1PlayerIdx ]
	currentlyPlayingTeam  int   // 0, 1, ...
}

// FishbowlGameSettings holds all the data about the
// game settings.
type FishbowlGameSettings struct {
	rounds      []fishbowl_api.RoundT
	timerLength int
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

func NewGame(gameRoom *models.GameRoom, mutex *sync.RWMutex) *FishbowlGame {
	g := &FishbowlGame{
		mutex: mutex,
		settings: &FishbowlGameSettings{
			rounds: []fishbowl_api.RoundT{
				fishbowl_api.RoundDescribe,
				fishbowl_api.RoundSingleWord,
				fishbowl_api.RoundCharades,
			},
			timerLength: 45,
		},
		room:  gameRoom,
		state: "waiting-room",
		teams: make([][]*models.Player, 2),
	}

	g.teams[0] = make([]*models.Player, 0)
	g.teams[1] = make([]*models.Player, 0)

	return g
}

func (g *FishbowlGame) HandleIncomingMessage(
	player *models.Player,
	incomingMessage api.IncomingMessage,
	body json.RawMessage,
) {
	actionType, ok := api.ActionLookup[incomingMessage.Action]
	if !ok {
		log.Fatalf("invalid action: %s\n", incomingMessage.Action)
		return
	}

	switch actionType {
	case fishbowl_api.ActionAddTeam:
		var req fishbowl_api.AddTeamRequest
		if err := json.Unmarshal(body, &req); err != nil {
			log.Fatal(err)
			return
		}
		g.addTeam(player, req)
	case fishbowl_api.ActionRemoveTeam:
		// TODO
	case fishbowl_api.ActionMovePlayer:
		var req fishbowl_api.MovePlayerRequest
		if err := json.Unmarshal(body, &req); err != nil {
			log.Fatal(err)
			return
		}
		g.movePlayer(player, req)
	case fishbowl_api.ActionChangeSettings:
		var req fishbowl_api.ChangeSettingsRequest
		if err := json.Unmarshal(body, &req); err != nil {
			log.Fatal(err)
			return
		}
		g.changeSettings(player, req)
	case fishbowl_api.ActionStartTurn:
		var req fishbowl_api.StartTurnRequest
		if err := json.Unmarshal(body, &req); err != nil {
			log.Fatal(err)
			return
		}
		g.startTurn(player, req)
	case fishbowl_api.ActionSubmitWords:
		var req fishbowl_api.SubmitWordsRequest
		if err := json.Unmarshal(body, &req); err != nil {
			log.Fatal(err)
			return
		}
		g.submitWords(player, req)
	case fishbowl_api.ActionChangeCard:
		var req fishbowl_api.ChangeCardRequest
		if err := json.Unmarshal(body, &req); err != nil {
			log.Fatal(err)
			return
		}
		g.changeCard(player, req)
	default:
		log.Fatalf("could not handle incoming action %s", incomingMessage.Action)
	}
}

func (g *FishbowlGame) addTeam(
	player *models.Player,
	req fishbowl_api.AddTeamRequest,
) error {
	log.Printf("Add team request\n")

	g.mutex.Lock()
	defer g.mutex.Unlock()

	room, err := g.performRoomChecks(player, true, false)
	if err != nil {
		return err
	}

	g.teams = append(g.teams, []*models.Player{})
	g.h.sendUpdatedGameMessages(room, nil)
}

func (g *FishbowlGame) movePlayer(
	player *models.Player,
	req fishbowl_api.MovePlayerRequest,
) {
	log.Printf("Move player request: %s (%d -> %d)\n", req.PlayerName,
		req.FromTeam, req.ToTeam)

	var msg api.OutgoingMessage

	g.mutex.Lock()
	defer g.mutex.Unlock()

	room, err := g.performRoomChecks(player, true, false)
	if err != nil {
		g.h.sendErrorMessage(player.client, err.Error())
		return
	}

	if req.FromTeam >= len(g.teams) || req.ToTeam >= len(g.teams) {
		msg.Event = "error"
		msg.Error = "The team indexes are invalid."
		g.h.sendOutgoingMessages(player.client, &msg, nil, nil)
		return
	}

	playerToMove := g.removePlayerFromTeam(room, req.FromTeam, req.PlayerName)
	g.teams[req.ToTeam] = append(g.teams[req.ToTeam], playerToMove)

	g.h.sendUpdatedGameMessages(room, nil)
}

func (g *FishbowlGame) changeSettings(
	player *models.Player,
	req fishbowl_api.ChangeSettingsRequest,
) {
	log.Printf("Change settings request\n")

	g.mutex.Lock()
	defer g.mutex.Unlock()

	room, err := g.performRoomChecks(player, true, false)
	if err != nil {
		g.h.sendErrorMessage(player.client, err.Error())
		return
	}

	g.settings = convertAPISettingsToSettings(req.Settings)
	g.h.sendUpdatedGameMessages(room, nil)
}

func (g *FishbowlGame) startTurn(
	player *models.Player,
	req fishbowl_api.StartTurnRequest,
) {
	log.Printf("Start turn request\n")

	g.mutex.Lock()
	defer g.mutex.Unlock()

	room, err := g.performRoomChecks(player.client, false, true)
	if err != nil {
		g.h.sendErrorMessage(player.client, err.Error())
		return
	}

	if !g.validateStateTransition(g.state, "turn-active") {
		g.h.sendErrorMessage(clientMessage, "You cannot perform that action at this time.")
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
		g.moveToNextPlayerAndTeam(room)

		log.Printf("Sending updated game message after timer expired\n")
		g.h.sendUpdatedGameMessages(room, nil)
	}()

	g.h.sendUpdatedGameMessages(room, nil)
}

func (g *FishbowlGame) submitWords(
	player *models.Player,
	req fishbowl_api.SubmitWordsRequest,
) {
	log.Printf("Submit words request: %s\n", req.Words)

	g.mutex.Lock()
	defer g.mutex.Unlock()

	room, err := g.performRoomChecks(player, false, false)
	if err != nil {
		g.h.sendErrorMessage(clientMessage, err.Error())
		return
	}

	player.settings.words = req.Words
	g.h.sendUpdatedGameMessages(room, nil)
}

func (g *FishbowlGame) changeCard(
	player *models.Player,
	req fishbowl_api.ChangeCardRequest,
) {
	log.Printf("Change card request: %s\n", req.ChangeType)

	g.mutex.Lock()
	defer g.mutex.Unlock()

	room, err := g.performRoomChecks(player, false, true)
	if err != nil {
		g.h.sendErrorMessage(clientMessage, err.Error())
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

				g.reshuffleCardsForRound(room)
				g.moveToNextPlayerAndTeam(room)

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

	g.h.sendUpdatedGameMessages(room, nil)
}

// This function must be called with the mutex held.
func (g *FishbowlGame) removePlayerFromTeam(
	room *GameRoom,
	fromTeam int,
	playerName string,
) *models.Player {
	players := g.teams[fromTeam]
	for idx, player := range players {
		if player.name == playerName {
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

func (g *FishbowlGame) initGameScores() {
	g.teamScoresByRound = make([][]int, len(g.settings.rounds))
	for idx := range g.settings.rounds {
		g.teamScoresByRound[idx] = make([]int, len(g.teams))
	}
}

// This function must be called with the mutex held.
func (g *FishbowlGame) reshuffleCardsForRound() {
	g.cardsInRound = []string{}
	for _, teamPlayers := range g.teams {
		for _, player := range teamPlayers {
			g.cardsInRound = append(g.cardsInRound, player.settings.words...)
		}
	}
	g.totalNumCards = len(g.cardsInRound)

	arr := g.cardsInRound
	rand.Shuffle(len(g.cardsInRound), func(i, j int) {
		arr[i], arr[j] = arr[j], arr[i]
	})
}

// This function must be called with the mutex held.
func (g *FishbowlGame) moveToNextPlayerAndTeam(room *GameRoom) {
	t := g.currentlyPlayingTeam
	g.currentPlayers[t] = (g.currentPlayers[t] + 1) % len(g.teams[t])
	g.currentlyPlayingTeam = (t + 1) % len(g.currentPlayers)
}

// This function must be called with the mutex held.
func (g *FishbowlGame) getCurrentPlayer(room *GameRoom) *Player {
	players := g.teams[g.currentlyPlayingTeam]
	return players[g.currentPlayers[g.currentlyPlayingTeam]]
}

// AddPlayer adds a player to the current game.
//
// This function must be called with the mutex held.
func (g *FishbowlGame) AddPlayer(player *models.Player) {
	player.settings = &FishbowlPlayerSettings{
		words: []string{},
	}

	g.teams[0] = append(g.teams[0], player)

	var msg api.OutgoingMessage
	msg.Event = api.Event[api.EventCreatedGame]
	msg.Body = api.CreatedGameEvent{
		GameType: room.gameType,
		RoomCode: room.roomCode,
		//Team:     0,
	}

	h.sendOutgoingMessages(clientMessage.client, &msg, nil, nil)
}

// Start starts the game.
//
// This function must be called with the mutex held.
func (g *FishbowlGame) Start(player *models.Player) {
	if !g.validateStateTransition(g.state, "turn-start") {
		h.sendErrorMessage(player.client,
			"You cannot perform that action at this time.")
		return
	}
	g.turnJustStarted = true
	g.state = "turn-start"

	g.reshuffleCardsForRound(room)
	g.initGameScores(room)

	g.currentRound = 0
	g.currentPlayers = make([]int, len(g.teams))
	for i, players := range g.teams {
		g.currentPlayers[i] = util.GetRandomNumberInRange(0, len(players)-1)
	}
	g.currentlyPlayingTeam = util.GetRandomNumberInRange(0, len(g.teams)-1)

	h.sendUpdatedGameMessages(room, nil)
}

// Rematch starts a new game with the same players and settings.
//
// This function must be called with the mutex held.
func (g *FishbowlGame) Rematch(player *Player) {
	if !g.validateStateTransition(g.state, "waiting-room") {
		h.sendErrorMessage(player.client,
			"You cannot perform that action at this time.")
		return
	}

	g.state = "waiting-room"
	g.initGameScores()
	g.lastCardGuessed = ""
	g.winningTeam = nil

	for _, teamPlayers := range g.teams {
		for _, player := range teamPlayers {
			player.settings.words = []string{}
		}
	}

	log.Println("Sending out updated game messages for rematch")
	h.sendUpdatedGameMessages(room, nil)
}

// This function must be called with the mutex held.
func (g *FishbowlGame) sendUpdatedGameMessages(
	room *models.GameRoom,
	justJoinedClient *models.Client,
) {
	if g.state == "waiting-room" {
		var msg api.OutgoingMessage
		msg.Event = api.Event[api.EventUpdatedRoom]
		msg.Body = api.UpdatedRoomEvent{
			GameType: room.gameType,
			Teams:    convertTeamsToAPITeams(g.teams),
			Settings: convertSettingsToAPISettings(g.settings),
		}

		log.Printf("Sending out updated room messages\n")

		if justJoinedClient != nil {
			log.Printf("models.Player just rejoined, sending updated-room event\n")
			g.h.sendOutgoingMessages(justJoinedClient, &msg, nil, room)
		} else {
			g.h.sendOutgoingMessages(nil, &msg, &msg, room)
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
	msgToCurrentPlayer.Body = api.UpdatedGameEvent{
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
	msgToOtherPlayers.Body = api.UpdatedGameEvent{
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
		log.Printf("models.Player %s just rejoined, sending updated-game event\n", currentPlayer.name)
		if currentPlayer.client == justJoinedClient {
			updatedGameEvent := msgToCurrentPlayer.Body.(api.UpdatedGameEvent)
			updatedGameEvent.Teams = convertTeamsToAPITeams(g.teams)
			msgToCurrentPlayer.Body = updatedGameEvent
			g.h.sendOutgoingMessages(justJoinedClient, &msgToCurrentPlayer,
				nil, room)
		} else {
			updatedGameEvent := msgToOtherPlayers.Body.(api.UpdatedGameEvent)
			updatedGameEvent.Teams = convertTeamsToAPITeams(g.teams)
			msgToOtherPlayers.Body = updatedGameEvent
			g.h.sendOutgoingMessages(justJoinedClient, &msgToOtherPlayers,
				nil, room)
		}
	} else {
		g.h.sendOutgoingMessages(currentPlayer.client, &msgToCurrentPlayer,
			&msgToOtherPlayers, room)
	}
}

// This function must be called with the mutex held.
func (g *FishbowlGame) validateStateTransition(fromState, toState string) bool {
	valid, ok := validStateTransitions[fromState]
	if !ok {
		return false
	}

	if !util.StringInSlice(valid, toState) {
		return false
	}

	return true
}

func (g *FishbowlGame) performRoomChecks(
	player *models.Player,
	playerMustBeRoomOwner bool,
	playerMustBeCurrentPlayer bool,
) (*models.GameRoom, error) {
	room := player.room
	if room == nil {
		return nil, errors.New("you are not in a game")
	}

	/* TODO
	if _, ok := h.rooms[room.roomCode]; !ok {
		return nil, errors.New("this game no longer exists")
	}
	*/

	room.lastInteractionTime = time.Now()

	if playerMustBeRoomOwner && !player.isRoomOwner {
		return nil, errors.New("you are not the game owner")
	}

	if playerMustBeCurrentPlayer {
		currentPlayer := g.getCurrentPlayer(room)
		if currentPlayer.name != player.name {
			return nil, errors.New("you are not the current player")
		}
	}

	return room, nil
}
