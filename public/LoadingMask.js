import { html, Component } from 'https://unpkg.com/htm/preact/standalone.module.js';


export default class LoadingMask extends Component {

  render() {
    return html`
      <div class="loading-mask-ct">
        <div class="loading-mask" />
        <div class="loading-mask-message">${this.props.message || 'Loading...'}</div>
      </div>
    `;
  }

}