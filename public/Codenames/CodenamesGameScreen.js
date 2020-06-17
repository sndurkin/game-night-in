import { html, Component } from 'https://unpkg.com/htm/preact/standalone.module.js';

import ScreenWrapper from '../ScreenWrapper.js';
import LoadingMask from '../LoadingMask.js';
import Constants from '../Constants.js';

import CodenamesConstants from './CodenamesConstants.js';


export default class CodenamesGameScreen extends Component {

  constructor(...args) {
    super(...args);

    this.state = {
      timeLeft: null,

      currentCard: null,
      changingCard: false,

      numSkipsUsed: 0,
    };

    this.startTurn = this.startTurn.bind(this);
    this.changeCard = this.changeCard.bind(this);
  }

  componentDidMount() {
    this.handleGameStateChange();
  }

  componentDidUpdate() {
    this.handleGameStateChange();
  }

  handleGameStateChange() {
    const { game } = this.props;

    // console.log('component did mount/update:', game.state, this.state.timeLeft);
    if (game.state === CodenamesConstants.States.TURN_START && this.state.timeLeft !== null) {
      clearInterval(this.intervalId);
      this.setState({
        timeLeft: null,
      });
    } else if (game.currentServerTime && game.timerLength && !this.state.timeLeft) {
      // console.log('starting timer');

      clearInterval(this.intervalId);

      this.timerEndTime = game.currentServerTime + (game.timerLength * 1000);
      const timeLeft = Math.max(0, Math.floor((this.timerEndTime - new Date().getTime()) / 1000));
      if (timeLeft === 0) {
        // The timer has already expired, so wait for the game-updated message
        // from the server.
        return;
      }

      this.intervalId = setInterval(() => {
        // console.log('timer interval');
        const timeLeft = Math.max(0, Math.floor((this.timerEndTime - new Date().getTime()) / 1000));
        this.setState({
          timeLeft: timeLeft,
        });

        if (timeLeft === 0) {
          clearInterval(this.intervalId);
        }
      }, 500);

      this.setState({
        timeLeft: timeLeft,
      });
    }
  }

  render() {
    const { game } = this.props;
    const { timeLeft, changingCard } = this.state;

    return html`
      <${ScreenWrapper} ...${this.props} header=${this.header}>
        <div class="screen">
          <div class="game-info-bar">
            <div class="cards-guessed">+ ${game.numCardsGuessedInTurn}</div>
          </div>
          <div class="game-area">
            ${this.gameAreaContents}
            <div class="game-info-bar">
              <div class="cards-left-ct">
                <div class="cards-left-num">${game.numCardsLeftInRound}</div>
                <div class="cards-left-text">cards left</div>
              </div>
              ${this.timeLeftComponent}
            </div>
            <div class="current-player">
              ${this.turnStr}
            </div>
          </div>
          ${this.isCurrentPlayer && html`
            <div class="button-bar game-actions">
              ${this.buttonBar}
            </div>
          `}
        </div>
        ${changingCard ? html`
          <${LoadingMask} />
        ` : null}
      <//>
    `;
  }

  handleMessage(data, e) {
    if (data.error) {
      if (data.errorIsFatal) {
        window.location.reload();
        return;
      }

      this.setState({ error: data.error });
    }

    switch (data.event) {
      case Constants.Events.UPDATED_GAME:
        const game = data.body;

        if (this.state.changingCard && this.state.currentCard !== game.currentCard) {
          // Card was just changed.
          this.setState({
            changingCard: false,
          });
        }

        this.props.updateStoreData({ game: game });
        if (game.state === CodenamesConstants.States.GAME_OVER) {
          this.props.transitionToScreen(CodenamesConstants.Screens.GAME_OVER);
        }
        break;
    }
  }

  startTurn() {
    this.setState({
      numSkipsUsed: 0,
    });

    const { conn } = this.props;
    conn.send(JSON.stringify({
      action: CodenamesConstants.Actions.START_TURN,
      body: {},
    }));
  }

  changeCard(changeType) {
    const { game } = this.props;
    const { numSkipsUsed } = this.state;

    const newState = {
      currentCard: game.currentCard,
      changingCard: true,
    };
    if (changeType === CodenamesConstants.CardChange.SKIP) {
      newState.numSkipsUsed = numSkipsUsed + 1;
    }
    this.setState(newState);

    const { conn } = this.props;
    conn.send(JSON.stringify({
      action: CodenamesConstants.Actions.CHANGE_CARD,
      body: {
        changeType: changeType,
      },
    }));
  }

  get header() {
    const { teams, game, settings } = this.props;

    return teams.map((_, teamIdx) => {
      const result = [];
      if (teamIdx > 0) {
        result.push(html`<div style="width: 0.25em"></div>`);
      }

      const teamColor = Constants.TeamColors[teamIdx];

      let totalScore = 0;
      for (let i = 0; i < settings.rounds.length; i++) {
        totalScore += game.teamScoresByRound[i][teamIdx];
      }

      result.push(html`
        <div class="team-score" style="border-bottom-color: ${teamColor}">
          ${totalScore}
        </div>
      `);

      return result;
    });
  }

  get isCurrentPlayer() {
    const { name } = this.props;
    return this.currentPlayer.name === name;
  }

  get currentPlayer() {
    const { teams, game } = this.props;

    const players = teams[game.currentlyPlayingTeam];
    const playerIdx = game.currentPlayers[game.currentlyPlayingTeam];
    return players[playerIdx];
  }

  get turnStr() {
    if (this.isCurrentPlayer) {
      return "It's your turn!";
    }

    if (this.currentPlayer.name.endsWith('s')) {
      return this.currentPlayer.name + "' turn";
    }

    return this.currentPlayer.name + "'s turn";
  }

  get gameAreaContents() {
    const { settings, game } = this.props;

    if (game.state === CodenamesConstants.States.TURN_START
      && game.numCardsLeftInRound === game.totalNumCards) {
      const currentRoundType = settings.rounds[game.currentRound];
      return html`
        <div class="round-start">
          <div class="round-start-title">
            Round ${game.currentRound + 1}
          </div>
          <div class="round-start-subtitle">
            ${CodenamesConstants.RoundTypes[currentRoundType].title}
          </div>
          <div class="round-start-desc">
            ${CodenamesConstants.RoundTypes[currentRoundType].long}
          </div>
        </div>
      `;
    }

    if (game.state === CodenamesConstants.States.TURN_ACTIVE
      && this.isCurrentPlayer) {
      return html`
        <div class="game-card">${game.currentCard}</div>
      `;
    }

    return html`
      <div class="game-card">
        <span style="color: #777">${game.lastCardGuessed}</span>
      </div>
    `;
  }

  get buttonBar() {
    const { game, settings } = this.props;
    const { numSkipsUsed } = this.state;
    const skipAvailable = numSkipsUsed < settings.maxSkipsPerTurn;

    switch (game.state) {
      case 'turn-start':
        return html`
          <button onClick=${this.startTurn}>Start!</button>
        `;
      case 'turn-active':
        return html`
          <button
            class=${'pseudo' + (!skipAvailable ? ' invisible' : '')}
            onClick=${() => this.changeCard(CodenamesConstants.CardChange.SKIP)}
          >
            Skip
          </button>
          <div></div>
          <button
            class="success"
            onClick=${() => this.changeCard(CodenamesConstants.CardChange.CORRECT)}
          >
            Correct!
          </button>
        `;
    }

    return null;
  }

  get timeLeftComponent() {
    const { game } = this.props;
    const { timeLeft } = this.state;

    if (timeLeft === null) {
      return null;
    }

    switch (game.state) {
      case CodenamesConstants.States.TURN_START:
      case CodenamesConstants.States.TURN_ACTIVE:
        const minutes = Math.floor(timeLeft / 60).toString().padStart(2, '0');
        const seconds = (timeLeft % 60).toString().padStart(2, '0');

        return html`
          <div class="Codenames-time-left">${minutes}:${seconds}</div>
        `;
    }

    return null;
  }

}
