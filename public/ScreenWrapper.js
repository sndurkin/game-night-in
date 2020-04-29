import { html, Component, render } from 'https://unpkg.com/htm/preact/standalone.module.js';


export default class ScreenWrapper extends Component {

  constructor(...args) {
    super(...args);


  }

  render() {
    const { game, roomCode, onBack, timeLeft, children, style } = this.props;

    return html`
      <div class="screen-wrapper" style=${style}>
        <div class="screen-header">
          ${(game.state === 'turn-start' || game.state === 'turn-active') && timeLeft !== null ? html`
            <div class="time-left">${this.timeLeftStr}</div>
          ` : onBack ? html`
            <button onClick=${onBack}>Back</button>
          ` : null}
          ${roomCode && html`
            <div class="room-code">${roomCode}</div>
          `}
        </div>
        ${children}
      </div>
    `;
  }

  get timeLeftStr() {
    const { timeLeft } = this.props;
    return Math.floor(timeLeft / 60).toString().padStart(2, '0') + ':' + (timeLeft % 60).toString().padStart(2, '0');
  }

}
