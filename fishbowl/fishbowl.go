package fishbowl

import (
	"github.com/sndurkin/game-night-in/hub"
	"github.com/sndurkin/game-night-in/util"
	"github.com/sndurkin/game-night-in/fishbowl/api"
)

// FishbowlPlayerSettings holds the game-specific data about a particular player.
type FishbowlPlayerSettings struct {
	words       []string
}

// FishbowlGame holds the game-specific data and logic.
type FishbowlGame struct {
	h    *hub.Hub
	room *hub.GameRoom

	state                 string
	turnJustStarted       bool
	cardsInRound          []string
	currentServerTime     int64
	timer                 *time.Timer
	timerLength           int
	lastCardGuessed       string
	totalNumCards         int
	numCardsGuessedInTurn int
	teams                 [][]*hub.Player
	teamScoresByRound     [][]int
	winningTeam           *int
	currentRound          int   // 0, 1, 2
	currentPlayers        []int // [ team0PlayerIdx, team1PlayerIdx ]
	currentlyPlayingTeam  int   // 0, 1, ...
}

// FishbowlGameSettings holds all the data about the
// game settings.
type FishbowlGameSettings struct {
	rounds      []api.RoundT
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

func NewGame(h *hub.Hub, gameRoom *hub.GameRoom) *FishbowlGame {
	g := &FishbowlGame{
		h:     h,
		settings: &FishbowlGameSettings{
			rounds: []api.RoundT{
				api.RoundDescribe,
				api.RoundSingleWord,
				api.RoundCharades,
			},
			timerLength: 45,
		},
		room:  gameRoom,
		state: "waiting-room",
		teams: make([][]*Player, 2),
	}

	g.teams[0] = make([]*hub.Player, 0)
	g.teams[1] = make([]*hub.Player, 0)

	return g
}

func (g *FishbowlGame) HandleIncomingMessage(
	player *hub.Player,
	incomingMessage api.IncomingMessage,
) {
	actionType, ok := api.ActionLookup[incomingMessage.Action]
	if !ok {
		log.Fatalf("invalid action: %s\n", incomingMessage.Action)
		return
	}

	switch actionType {
	case api.ActionAddTeam:
		var req api.AddTeamRequest
		if err := json.Unmarshal(body, &req); err != nil {
			log.Fatal(err)
			return
		}
		g.addTeam(player, req)
	case api.ActionRemoveTeam:
		// TODO
	case api.ActionMovePlayer:
		var req api.MovePlayerRequest
		if err := json.Unmarshal(body, &req); err != nil {
			log.Fatal(err)
			return
		}
		g.movePlayer(player, req)
	case api.ActionChangeSettings:
		var req api.ChangeSettingsRequest
		if err := json.Unmarshal(body, &req); err != nil {
			log.Fatal(err)
			return
		}
		g.changeSettings(player, req)
	case api.ActionStartTurn:
		var req api.StartTurnRequest
		if err := json.Unmarshal(body, &req); err != nil {
			log.Fatal(err)
			return
		}
		g.startTurn(player, req)
	case api.ActionSubmitWords:
		var req api.SubmitWordsRequest
		if err := json.Unmarshal(body, &req); err != nil {
			log.Fatal(err)
			return
		}
		g.submitWords(player, req)
	case api.ActionChangeCard:
		var req api.ChangeCardRequest
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
	player *hub.Player,
	req api.AddTeamRequest,
) error {
	log.Printf("Add team request\n")

	g.h.Lock()
	defer g.h.Unlock()

	room, err := g.h.performRoomChecks(player, true, false)
	if err != nil {
		return err
	}

	room.teams = append(room.teams, []*hub.Player{})
	g.h.sendUpdatedGameMessages(room, nil)
}

func (g *FishbowlGame) movePlayer(
	player *hub.Player,
	req api.MovePlayerRequest,
) {
	log.Printf("Move player request: %s (%d -> %d)\n", req.PlayerName,
		req.FromTeam, req.ToTeam)

	var msg api.OutgoingMessage

	g.h.Lock()
	defer g.h.Unlock()

	room, err := g.h.performRoomChecks(player, true, false)
	if err != nil {
		g.h.sendErrorMessage(player.client, err.Error())
		return
	}

	if req.FromTeam >= len(room.teams) || req.ToTeam >= len(room.teams) {
		msg.Event = "error"
		msg.Error = "The team indexes are invalid."
		g.h.sendOutgoingMessages(player.client, &msg, nil, nil)
		return
	}

	playerToMove := g.removePlayerFromTeam(room, req.FromTeam, req.PlayerName)
	room.teams[req.ToTeam] = append(room.teams[req.ToTeam], playerToMove)

	g.h.sendUpdatedGameMessages(room, nil)
}

func (g *FishbowlGame) changeSettings(
	player *hub.Player,
	req api.ChangeSettingsRequest,
) {
	log.Printf("Change settings request\n")

	g.h.Lock()
	defer g.h.Unlock()

	room, err := g.h.performRoomChecks(player, true, false)
	if err != nil {
		g.h.sendErrorMessage(player.client, err.Error())
		return
	}

	g.settings = convertAPISettingsToSettings(req.Settings)
	g.h.sendUpdatedGameMessages(room, nil)
}

func (g *FishbowlGame) startTurn(
	player *hub.Player,
	req api.StartTurnRequest,
) {
	log.Printf("Start turn request\n")

	g.h.Lock()
	defer g.h.Unlock()

	room, err := g.h.performRoomChecks(player.client, false, true)
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

		g.h.Lock()
		defer g.h.Unlock()

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
		g.h.moveToNextPlayerAndTeam(room)

		log.Printf("Sending updated game message after timer expired\n")
		g.h.sendUpdatedGameMessages(room, nil)
	}()

	g.h.sendUpdatedGameMessages(room, nil)
}

func (g *FishbowlGame) submitWords(
	clientMessage *ClientMessage,
	req api.SubmitWordsRequest,
) {
	log.Printf("Submit words request: %s\n", req.Words)

	g.h.Lock()
	defer g.h.Unlock()

	playerClient := g.h.playerClients[clientMessage.client]
	room, err := g.h.performRoomChecks(playerClient, false, false)
	if err != nil {
		g.h.sendErrorMessage(clientMessage, err.Error())
		return
	}

	playerClient.words = req.Words
	g.h.sendUpdatedGameMessages(room, nil)
}

func (g *FishbowlGame) changeCard(
	player *hub.Player,
	req api.ChangeCardRequest,
) {
	log.Printf("Change card request: %s\n", req.ChangeType)

	g.h.Lock()
	defer g.h.Unlock()

	playerClient := g.h.playerClients[clientMessage.client]
	room, err := g.h.performRoomChecks(playerClient, false, true)
	if err != nil {
		g.h.sendErrorMessage(clientMessage, err.Error())
		return
	}

	game := room.game
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
			if g.currentRound < len(room.settings.rounds) {
				// Round over, moving to next round
				g.state = "turn-start"

				g.reshuffleCardsForRound(room)
				g.moveToNextPlayerAndTeam(room)

				// Each round should start with a different team.
				//
				// TODO: The teams should be re-ordered based on score.
				g.currentlyPlayingTeam = util.GetRandomNumberInRange(0,
					len(room.teams)-1)
			} else {
				g.state = "game-over" // TODO: update to use constant from api.go

				totalScores := make([]int, len(room.teams))
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
) *hub.Player {
	players := room.teams[fromTeam]
	for idx, player := range players {
		if player.name == playerName {
			player := players[idx]
			room.teams[fromTeam] = append(
				players[:idx],
				players[idx+1:]...,
			)
			return player
		}
	}

	return nil
}

func (g *FishbowlGame) initGameScores() {
	g.teamScoresByRound = make([][]int, len(g.room.settings.rounds))
	for idx := range g.settings.rounds {
		g.teamScoresByRound[idx] = make([]int, len(room.teams))
	}
}

// This function must be called with the mutex held.
func (g *FishbowlGame) reshuffleCardsForRound() {
	g.cardsInRound = []string{}
	for _, teamPlayers := range g.room.teams {
		for _, player := range teamPlayers {
			g.cardsInRound = append(g.cardsInRound, player.words...)
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
	game := room.game
	t := g.currentlyPlayingTeam
	g.currentPlayers[t] = (g.currentPlayers[t] + 1) % len(room.teams[t])
	g.currentlyPlayingTeam = (t + 1) % len(g.currentPlayers)
}

// This function must be called with the mutex held.
func (g *FishbowlGame) getCurrentPlayer(room *GameRoom) *Player {
	game := room.game
	players := room.teams[g.currentlyPlayingTeam]
	return players[g.currentPlayers[g.currentlyPlayingTeam]]
}

// AddPlayer adds a player to the current game.
//
// This function must be called with the mutex held.
func (g *FishbowlGame) AddPlayer(player *hub.Player) {
	player.settings = &FishbowlPlayerSettings{
		words: []string{},
	}

	room.teams[0] = append(room.teams[0], player)

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
func (g *FishbowlGame) Start(player *hub.Player) {
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
	g.currentPlayers = make([]int, len(room.teams))
	for i, players := range room.teams {
		g.currentPlayers[i] = getRandomNumberInRange(0, len(players)-1)
	}
	g.currentlyPlayingTeam = getRandomNumberInRange(0, len(room.teams)-1)

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

	for _, teamPlayers := range g.room.teams {
		for _, player := range teamPlayers {
			player.words = []string{}
		}
	}

	log.Println("Sending out updated game messages for rematch")
	h.sendUpdatedGameMessages(room, nil)
}

// This function must be called with the mutex held.
func (g *FishbowlGame) sendUpdatedGameMessages(
	room *hub.GameRoom,
	justJoinedClient *hub.Client,
) {
	game := room.game

	if g.state == "waiting-room" {
		var msg api.OutgoingMessage
		msg.Event = api.Event[api.EventUpdatedRoom]
		msg.Body = api.UpdatedRoomEvent{
			GameType: room.gameType,
			Teams:    convertTeamsToAPITeams(room.teams),
			Settings: convertSettingsToAPISettings(room.settings),
		}

		log.Printf("Sending out updated room messages\n")

		if justJoinedClient != nil {
			log.Printf("hub.Player just rejoined, sending updated-room event\n")
			g.h.sendOutgoingMessages(justJoinedClient, &msg, nil, room)
		} else {
			g.h.sendOutgoingMessages(nil, &msg, &msg, room)
		}
		return
	}

	currentPlayer := g.h.getCurrentPlayer(room)

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
		log.Printf("hub.Player %s just rejoined, sending updated-game event\n", currentPlayer.name)
		if currentPlayer.client == justJoinedClient {
			updatedGameEvent := msgToCurrentPlayer.Body.(api.UpdatedGameEvent)
			updatedGameEvent.Teams = convertTeamsToAPITeams(room.teams)
			msgToCurrentPlayer.Body = updatedGameEvent
			g.h.sendOutgoingMessages(justJoinedClient, &msgToCurrentPlayer,
				nil, room)
		} else {
			updatedGameEvent := msgToOtherPlayers.Body.(api.UpdatedGameEvent)
			updatedGameEvent.Teams = convertTeamsToAPITeams(room.teams)
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
