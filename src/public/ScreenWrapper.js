import { html, Component, render } from 'https://unpkg.com/htm/preact/standalone.module.js';


export default class ScreenWrapper extends Component {

  constructor(...args) {
    super(...args);


  }

  render() {
    const { game, roomCode, onBack, children, style } = this.props;

    return html`
      <div class="screen-wrapper" style=${style}>
        <div class="screen-header">
          ${onBack ? html`
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

}
