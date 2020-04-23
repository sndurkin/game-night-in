import { html, Component, render } from 'https://unpkg.com/htm/preact/standalone.module.js';


export default class ScreenWrapper extends Component {

  constructor(...args) {
    super(...args);

    // this.onBack = this.onBack.bind(this);
  }

  render(props, state) {
    const { roomCode, onBack, children, style } = props;

    return html`
      <div class="screen-wrapper" style=${style}>
        <div class="screen-header">
          ${onBack && html`
            <button onClick=${onBack}>Back</button>
          `}
          ${roomCode && html`
            <div class="room-code">${roomCode}</div>
          `}
        </div>
        ${children}
      </div>
    `;
  }

}
