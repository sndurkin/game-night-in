import { html, Component, render } from 'https://unpkg.com/htm/preact/standalone.module.js';

import Constants from './Constants.js';


export default class GameScreen extends Component {

  constructor(...args) {
    super(...args);

    this.state = {

    };
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
          <div class="game-card"></div>
          <div class="game-info">
            <div class="cards-left">${game.cardsLeftInRound} cards left</div>
            <div class="cards-guessed">+ ${game.cardsGuessedInTurn}</div>
          </div>
          <div class="current-player">${this.currentPlayer.name}</div>
        </div>
        <div class="button-bar">
          ${game.state === 'turn-start' && html`
            <div></div>
          `}
          <button>Skip</button>
          <div></div>
          <button onClick=${this.markWordCorrect}>Correct!</button>
        </div>
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

  markWordCorrect() {
    const { conn } = this.props;
    conn.send(JSON.stringify({
      action: Constants.Actions.START_TURN,
      body: {},
    }));
  }

  get isCurrentPlayer() {
    const { name } = this.props;
    this.currentPlayer.name === name;
  }

  get currentPlayer() {
    const { teams, game } = this.props;

    const players = teams[game.currentlyPlayingTeam];
    const playerIdx = game.currentPlayers[game.currentlyPlayingTeam];
    return players[playerIdx];
  }

}
