import { html, Component, render } from 'https://unpkg.com/htm/preact/standalone.module.js';

import Constants from './Constants.js';


export default class GameScreen extends Component {

  constructor(...args) {
    super(...args);

    this.state = {

    };

    this.startTurn = this.startTurn.bind(this);
    this.changeCard = this.changeCard.bind(this);
  }

  render() {
    const { game } = this.props;
    const { error } = this.state;

    return html`
      <div class="screen">
        ${error && html`
          <span class="label error">${error}</span>
        `}

        <div class="game-area">
          <div class="game-card">

          </div>
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
