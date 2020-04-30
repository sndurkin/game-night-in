import { html, Component } from 'https://unpkg.com/htm/preact/standalone.module.js';

import Utils from './Utils.js';
import Constants from './Constants.js';


export default class GameOverScreen extends Component {

  constructor(...args) {
    super(...args);


  }

  render() {
    const { teams, game } = this.props;

    return html`
      <div class="screen">
        ${this.confetti}
        <div>Winner: Team ${game.winningTeam + 1}</div>
        <ul>
          ${teams[game.winningTeam].map(player => html`
            <li>${player.name}</li>
          `)}
        </ul>
      </div>
    `;
  }

  handleMessage(data, e) {
    if (data.error) {
      this.setState({ error: data.error });
    }

    switch (data.event) {
      case Constants.Events.UPDATED_GAME:
        const game = data.body;
        if (game.state === Constants.States.GAME_OVER) {
          this.props.transitionToScreen(Constants.Screens.GAME_OVER);
        } else {
          this.props.updateStoreData({ game: game });
        }
        break;
    }
  }

  get confetti() {
    const { game } = this.props;
    const baseColor = Constants.TeamColors[game.winningTeam];

    const confettiStyles = [];

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

}
