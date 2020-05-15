package fishbowl

import (
	"github.com/sndurkin/game-night-in/fishbowl/api"
)

// FishbowlPlayerSettings holds the game-specific data about a particular player.
type FishbowlPlayerSettings struct {
	words       []string
}

// FishbowlGame holds the game-specific data and logic.
type FishbowlGame struct {
	state                 string
	turnJustStarted       bool
	cardsInRound          []string
	currentServerTime     int64
	timer                 *time.Timer
	timerLength           int
	lastCardGuessed       string
	totalNumCards         int
	numCardsGuessedInTurn int
	teams                 [][]*Player
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

func (g *FishbowlGame) HandleIncomingMessage(
	player *Player,
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
		g.movePlayer(clientMessage, req)
	case api.ActionStartTurn:
		var req api.StartTurnRequest
		if err := json.Unmarshal(body, &req); err != nil {
			log.Fatal(err)
			return
		}
		g.startTurn(clientMessage, req)
	case api.ActionSubmitWords:
		var req api.SubmitWordsRequest
		if err := json.Unmarshal(body, &req); err != nil {
			log.Fatal(err)
			return
		}
		g.submitWords(clientMessage, req)
	case api.ActionChangeCard:
		var req api.ChangeCardRequest
		if err := json.Unmarshal(body, &req); err != nil {
			log.Fatal(err)
			return
		}
		g.changeCard(clientMessage, req)
	default:
		log.Fatalf("could not handle incoming action %s", incomingMessage.Action)
	}
}

func (g *FishbowlGame) addTeam(
	player *Player,
	req api.AddTeamRequest,
) error {
	log.Printf("Add team request\n")

	h.Lock()
	defer h.Unlock()

	room, err := h.performRoomChecks(player, true, false)
	if err != nil {
		return err
	}

	room.teams = append(room.teams, []*Player{})
	h.sendUpdatedGameMessages(room, nil)
}

func (g *FishbowlGame) movePlayer(
	clientMessage *ClientMessage,
	req api.MovePlayerRequest,
) {
	log.Printf("Move player request: %s (%d -> %d)\n", req.PlayerName,
		req.FromTeam, req.ToTeam)

	var msg api.OutgoingMessage

	h.Lock()
	defer h.Unlock()

	playerClient := h.playerClients[clientMessage.client]
	room, err := h.performRoomChecks(playerClient, true, false)
	if err != nil {
		h.sendErrorMessage(clientMessage, err.Error())
		return
	}

	if req.FromTeam >= len(room.teams) || req.ToTeam >= len(room.teams) {
		msg.Event = "error"
		msg.Error = "The team indexes are invalid."
		h.sendOutgoingMessages(clientMessage.client, &msg, nil, nil)
		return
	}

	player := h.removePlayerFromTeam(room, req.FromTeam, req.PlayerName)
	room.teams[req.ToTeam] = append(room.teams[req.ToTeam], player)

	h.sendUpdatedGameMessages(room, nil)
}

func (g *FishbowlGame) startTurn(
	clientMessage *ClientMessage,
	req api.StartTurnRequest,
) {
	log.Printf("Start turn request\n")

	h.Lock()
	defer h.Unlock()

	playerClient := h.playerClients[clientMessage.client]
	room, err := h.performRoomChecks(playerClient, false, true)
	if err != nil {
		h.sendErrorMessage(clientMessage, err.Error())
		return
	}

	game := room.game

	if !h.validateStateTransition(game.state, "turn-active") {
		h.sendErrorMessage(clientMessage, "You cannot perform that action at this time.")
		return
	}
	game.turnJustStarted = true
	game.state = "turn-active"

	game.numCardsGuessedInTurn = 0
	game.lastCardGuessed = ""
	game.timerLength = room.settings.timerLength + 1
	game.currentServerTime = time.Now().UnixNano() / 1000000
	if game.timer != nil {
		game.timer.Stop()
	}
	game.timer = time.NewTimer(time.Second * time.Duration(game.timerLength))

	// Wait for timer to finish in an asynchronous goroutine
	go func() {
		// Block until timer finishes. When it is done, it sends a message
		// on the channel timer.C. No other code in
		// this goroutine is executed until that happens.
		<-game.timer.C

		log.Println("Timer expired, waiting on lock")

		h.Lock()
		defer h.Unlock()

		log.Println(" - lock obtained")

		game.timer = nil

		log.Printf("Game state when timer ended: %s\n", game.state)
		if !h.validateStateTransition(game.state, "turn-start") {
			if game.state == "turn-start" || game.state == "game-over" {
				// Round or game finished before the player's turn timer expired,
				// so do nothing.
				return
			}

			log.Fatalf("Game was not in correct state when turn timer expired: %s", game.state)
			return
		}

		game.turnJustStarted = false
		game.state = "turn-start"
		h.moveToNextPlayerAndTeam(room)

		log.Printf("Sending updated game message after timer expired\n")
		h.sendUpdatedGameMessages(room, nil)
	}()

	h.sendUpdatedGameMessages(room, nil)
}

func (g *FishbowlGame) submitWords(
	clientMessage *ClientMessage,
	req api.SubmitWordsRequest,
) {
	log.Printf("Submit words request: %s\n", req.Words)

	h.Lock()
	defer h.Unlock()

	playerClient := h.playerClients[clientMessage.client]
	room, err := h.performRoomChecks(playerClient, false, false)
	if err != nil {
		h.sendErrorMessage(clientMessage, err.Error())
		return
	}

	playerClient.words = req.Words
	h.sendUpdatedGameMessages(room, nil)
}

func (g *FishbowlGame) changeCard(
	clientMessage *ClientMessage,
	req api.ChangeCardRequest,
) {
	log.Printf("Change card request: %s\n", req.ChangeType)

	h.Lock()
	defer h.Unlock()

	playerClient := h.playerClients[clientMessage.client]
	room, err := h.performRoomChecks(playerClient, false, true)
	if err != nil {
		h.sendErrorMessage(clientMessage, err.Error())
		return
	}

	game := room.game
	if game.state != "turn-active" {
		// Ignore, the turn is probably over.
		return
	}

	game.turnJustStarted = false
	if req.ChangeType == "correct" {
		// Increment score for current team and the current turn.
		game.teamScoresByRound[game.currentRound][game.currentlyPlayingTeam]++
		game.numCardsGuessedInTurn++

		game.lastCardGuessed = game.cardsInRound[0]
		game.cardsInRound = game.cardsInRound[1:]

		if len(game.cardsInRound) == 0 {
			if game.timer != nil {
				game.timer.Stop()
			}

			game.currentRound++
			if game.currentRound < len(room.settings.rounds) {
				// Round over, moving to next round
				game.state = "turn-start"

				h.reshuffleCardsForRound(room)
				h.moveToNextPlayerAndTeam(room)

				// Each round should start with a different team.
				//
				// TODO: The teams should be re-ordered based on score.
				game.currentlyPlayingTeam = getRandomNumberInRange(0,
					len(room.teams)-1)
			} else {
				game.state = "game-over" // TODO: update to use constant from api.go

				totalScores := make([]int, len(room.teams))
				for _, scoresByTeam := range game.teamScoresByRound {
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

				game.winningTeam = &teamWithMax
			}
		}
	} else {
		// Skip this card, push it to the end
		game.cardsInRound = append(game.cardsInRound[1:], game.cardsInRound[0])
	}

	h.sendUpdatedGameMessages(room, nil)
}

// This function must be called with the mutex held.
func (g *FishbowlGame) removePlayerFromTeam(
	room *GameRoom,
	fromTeam int,
	playerName string,
) *Player {
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
