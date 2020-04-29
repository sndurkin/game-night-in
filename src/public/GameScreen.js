import { html, Component, render } from 'https://unpkg.com/htm/preact/standalone.module.js';

import Constants from './Constants.js';


export default class GameScreen extends Component {

  constructor(...args) {
    super(...args);

    this.state = {
      timeLeft: null,
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

    console.log('component did mount/update:', game.state, this.state.timeLeft);
    if (game.state === 'turn-start' && this.state.timeLeft !== null) {
      clearInterval(this.intervalId);
      this.setState({
        timeLeft: null,
      });
    } else if (game.currentServerTime && game.timerLength && !this.state.timeLeft) {
      console.log('starting timer');

      clearInterval(this.intervalId);

      this.timerEndTime = game.currentServerTime + (game.timerLength * 1000);
      const timeLeft = Math.max(0, Math.floor((this.timerEndTime - new Date().getTime()) / 1000));
      if (timeLeft === 0) {
        // The timer has already expired, so wait for the game-updated message
        // from the server.
        return;
      }

      this.intervalId = setInterval(() => {
        console.log('timer interval');
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
    const { error, timeLeft } = this.state;

    return html`
      <div class="screen">
        ${error && html`
          <span class="label error">${error}</span>
        `}

        <div class="game-area">
          <div class="game-card">
            ${this.card}
          </div>
          ${game.state === 'turn-start' || game.state === 'turn-active' ? html`
            <div class="time-left">${timeLeft}</div>
          ` : null}
          <div class="game-info">
            <div class="cards-left-ct">
              <div class="cards-left-num">${game.numCardsLeftInRound}</div>
              <div class="cards-left-text">cards left</div>
            </div>
            <div class="cards-guessed">+ ${game.numCardsGuessedInTurn}</div>
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
    `;
  }

  handleMessage(data, e) {
    if (data.error) {
      this.setState({ error: data.error });
    }

    switch (data.event) {
      case Constants.Events.UPDATED_GAME:
        this.props.updateStoreData({ game: data.body });
        break;
    }
  }

  startTurn() {
    const { conn } = this.props;
    conn.send(JSON.stringify({
      action: Constants.Actions.START_TURN,
      body: {},
    }));
  }

  changeCard(changeType) {
    const { conn } = this.props;
    conn.send(JSON.stringify({
      action: Constants.Actions.CHANGE_CARD,
      body: {
        changeType: changeType,
      },
    }));
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

  get card() {
    const { game } = this.props;

    if (this.isCurrentPlayer) {
      return game.currentCard;
    }

    return html`
      <span style="color: #777">${game.lastCardGuessed}</span>
    `;
  }

  get buttonBar() {
    const { game } = this.props;
    switch (game.state) {
      case 'turn-start':
        return html`
          <button onClick=${this.startTurn}>Start!</button>
        `;
      case 'turn-active':
        return html`
          <button
            class="pseudo"
            onClick=${() => this.changeCard(Constants.CardChange.SKIP)}
          >
            Skip
          </button>
          <div></div>
          <button
            class="success"
            onClick=${() => this.changeCard(Constants.CardChange.CORRECT)}
          >
            Correct!
          </button>
        `;
    }

    return null;
  }

}
