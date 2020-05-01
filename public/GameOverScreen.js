import { html, Component } from 'https://unpkg.com/htm/preact/standalone.module.js';

import Utils from './Utils.js';
import Constants from './Constants.js';


export default class GameOverScreen extends Component {

  constructor(...args) {
    super(...args);

    this.rematch = this.rematch.bind(this);
    this.startOver = this.startOver.bind(this);
  }

  render() {
    const { teams, game } = this.props;

    return html`
      ${this.confetti}
      <div class="game-over">
        <div class="team-wins">Team ${game.winningTeam + 1} wins!</div>
        ${this.isCurrentPlayer ? html`
          <button class="lone" onClick=${this.rematch}>Rematch</button>
          <div class="center-horiz">or</div>
          <button class="lone" onClick=${this.startOver}>Start over</button>
        ` : null}
      </div>
    `;
  }

  handleMessage(data, e) {
    this.props.transitionToScreen(Constants.Screens.ROOM);
    this.props.updateStoreData({
      teams: data.body.teams,
    });
  }

  rematch() {
    const { conn } = this.props;
    conn.send(JSON.stringify({
      action: Constants.Actions.REMATCH,
      body: {},
    }));
  }

  startOver() {
    window.location.reload();
  }

  get confetti() {
    const { game } = this.props;
    const baseColor = Constants.TeamColors[game.winningTeam];

    const confettiStyles = [];

    console.log('Rendering confetti: ' + baseColor + ' for winning team: ' +
      game.winningTeam);
    console.log(game);
    const { r, g, b } = Utils.colorToRGB(baseColor);

    // Lighter colors
    for (let i = 1; i <= 3; i++) {
      const style = {
        background: Utils.rgbToColor({
          r: r / ((10 + i) / 10),
          g: g / ((10 + i) / 10),
          b: b / ((10 + i) / 10),
        }),
        width: Utils.getRandomNumberInRange(8, 16),
        height: Utils.getRandomNumberInRange(8, 16),
      };
      confettiStyles.push(style);
    }

    confettiStyles.push({
      background: baseColor,
      width: Utils.getRandomNumberInRange(8, 16),
      height: Utils.getRandomNumberInRange(8, 16),
    });

    // Darker colors
    for (let i = 1; i <= 6; i++) {
      const style = {
        background: Utils.rgbToColor({
          r: r * ((10 + i) / 10),
          g: g * ((10 + i) / 10),
          b: b * ((10 + i) / 10),
        }),
        width: Utils.getRandomNumberInRange(8, 16),
        height: Utils.getRandomNumberInRange(8, 16),
      };
      confettiStyles.push(style);
    }

    const sortedConfettiStyles = [];
    while (confettiStyles.length > 0) {
      const idx = Math.floor(Math.random() * confettiStyles.length);
      sortedConfettiStyles.push(confettiStyles[idx]);
      confettiStyles.splice(idx, 1);
    }

    return html`
      <div class="confetti-ct">
        ${sortedConfettiStyles.map(s => html`
          <div class="confetti" style="background: ${s.background}; width: ${s.width}px; height: ${s.height}px" />
        `)}
      </div>
    `;
  }

  // TODO: merge these with GameScreen.js
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

}
