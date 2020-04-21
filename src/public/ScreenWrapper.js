import { html, Component, render } from 'https://unpkg.com/htm/preact/standalone.module.js';


export default class ScreenWrapper extends Component {

  constructor(...args) {
    super(...args);

    // this.onBack = this.onBack.bind(this);
  }

  render(props, state) {
    const { onBack, children, style } = props;

    return html`
      <div style=${style}>
        ${onBack && html`
          <button onClick=${onBack}>Back</button>
        `}
        ${children}
      </div>
    `;
  }

}
