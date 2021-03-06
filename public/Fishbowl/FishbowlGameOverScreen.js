import { html, Component } from 'https://unpkg.com/htm/preact/standalone.module.js';

import Utils from '../Utils.js';
import Constants from '../Constants.js';

import FishbowlConstants from './FishbowlConstants.js';


export default class FishbowlGameOverScreen extends Component {

  constructor(...args) {
    super(...args);

    this.rematch = this.rematch.bind(this);
    this.startOver = this.startOver.bind(this);
  }

  render() {
    const { isRoomOwner, teams, game } = this.props;

    const teamScores = teams.map((_, idx) => ({
      idx: idx,
      score: 0,
    }));
    for (let roundScores of game.teamScoresByRound) {
      for (let i = 0; i < teamScores.length; i++) {
        const teamScore = teamScores[i];
        teamScore.score += roundScores[i];
      }
    }
    teamScores.sort((a, b) => b.score - a.score);

    return html`
      ${this.confetti}
      <div class="game-over">
        <div class="team-wins">Team ${game.winningTeam + 1} wins!</div>
        ${isRoomOwner ? html`
          <button class="lone" onClick=${this.rematch}>Rematch</button>
          <div class="center-horiz">or</div>
          <button class="lone" onClick=${this.startOver}>Start over</button>
        ` : null}
        <h3 class="scores-table-title">Scores</h3>
        <div class="scores-table-ct">
          <table class="primary scores-table">
            <thead>
              <tr>
                <th rowspan="2" style="vertical-align: bottom">Team</th>
                <th rowspan="2"></th>
                <th colspan=${game.teamScoresByRound.length}>Rounds</th>
              </tr>
              <tr>
                ${game.teamScoresByRound.map((_, idx) => html`
                  <th>${idx + 1}</th>
                `)}
              </tr>
            </thead>
            <tbody>
              ${teamScores.map(teamScore => html`
                <tr>
                  <td style=${Utils.teamStyle(teamScore.idx)}>
                    ${teamScore.idx + 1}
                  </td>
                  <td><b>${teamScore.score}</b></td>
                  ${game.teamScoresByRound.map(roundScores => html`
                    <td>${roundScores[teamScore.idx]}</td>
                  `)}
                </tr>
              `)}
            </tbody>
          </table>
        </div>
      </div>
    `;
  }

  handleMessage(data, e) {
    switch (data.event) {
      case Constants.Events.UPDATED_ROOM:
        this.props.transitionToScreen(FishbowlConstants.Screens.ROOM);
        this.props.updateStoreData({
          teams: data.body.teams,
        });
        break;
    }
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

    // console.log('Rendering confetti: ' + baseColor + ' for winning team: ' +
    //   game.winningTeam);
    // console.log(game);
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
          <div class="confetti" style=${this.stylesToString(s)} />
        `)}
      </div>
    `;
  }

  stylesToString(styles) {
    return Object.entries(styles).map(r => {
      let rule = `${r[0]}: ${r[1]}`;
      if (r[0] === 'height' || r[0] === 'width') {
        rule = rule + 'px';
      }
      return rule;
    }).join('; ');
  }

}
