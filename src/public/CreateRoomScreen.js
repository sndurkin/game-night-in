import { html, Component, render } from 'https://unpkg.com/htm/preact/standalone.module.js';

import Constants from './Constants.js';


export default class CreateRoomScreen extends Component {

  constructor(...args) {
    super(...args);

    this.state = {
      name: '',
      error: '',
    };

    this.onNameChange = this.onNameChange.bind(this);
  }

  render(props, state) {
    const { name } = state;

    return html`
      <div class="screen">
        <input type="text" maxlength="20" value="${name}" onInput=${this.onNameChange} autofocus />
        <button>Create</button>
      </div>
    `;
  }

  onNameChange(e) {
    this.setState({ name: e.target.value });
  }

  create() {
    if (this.name.length === 0) {
      this.error = 'Please enter a name to join.';
      return;
    }

    conn.send(JSON.stringify({
      action: Constants.Actions.CREATE_ROOM,
      body: {
        gameType: 'fishbowl',
        name: this.name
      }
    }));
  }
}
