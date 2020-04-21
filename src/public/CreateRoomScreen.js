import { html, Component, render } from 'https://unpkg.com/htm/preact/standalone.module.js';

import Constants from './Constants.js';


export default class CreateRoomScreen extends Component {

  constructor(...args) {
    super(...args);

    this.state = {
      gameType: 'fishbowl',
      name: '',
      error: '',
    };

    this.onSelectGameType = this.onSelectGameType.bind(this);
    this.onNameChange = this.onNameChange.bind(this);
    this.createRoom = this.createRoom.bind(this);
  }

  render(props, state) {
    const { gameType, name, error } = state;

    return html`
      <div class="screen">
        ${error && html`
          <span class="label error">${error}</span>
        `}
        <label class="select">
          Game type
          <select value=${gameType} onChange=${this.onSelectGameType}>
            <option value="fishbowl">Fishbowl</option>
          </select>
        </label>
        <label>
          Username
          <input
            type="text"
            maxlength="20"
            value="${name}"
            placeholder="Enter your name"
            onInput=${this.onNameChange} />
        </label>
        <button class="lone" onClick=${this.createRoom}>Create</button>
      </div>
    `;
  }

  onSelectGameType(e) {
    this.setState({ gameType: e.target.value });
  }

  onNameChange(e) {
    this.setState({ name: e.target.value });
  }

  createRoom() {
    const { gameType, name } = this.state;
    if (name.length === 0) {
      this.setState({ error: 'Please enter a name to join.' });
      return;
    }

    conn.send(JSON.stringify({
      action: Constants.Actions.CREATE_ROOM,
      body: {
        gameType: gameType,
        name: name
      }
    }));
  }
}
