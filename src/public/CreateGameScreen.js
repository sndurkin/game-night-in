import { html, Component, render } from 'https://unpkg.com/htm/preact/standalone.module.js';

import Constants from './Constants.js';


export default class CreateGameScreen extends Component {

  constructor(...args) {
    super(...args);

    this.state = {
      gameType: 'fishbowl',
      name: Math.random().toString(36).substring(2, 8),
      error: '',
    };

    this.onSelectGameType = this.onSelectGameType.bind(this);
    this.onNameChange = this.onNameChange.bind(this);
    this.createGame = this.createGame.bind(this);
  }

  render() {
    const { gameType, name, error } = this.state;

    return html`
      <div class="screen">
        ${error && html`
          <span class="label error">${error}</span>
        `}
        <form onSubmit=${this.createGame}>
          <label class="select">
            Game type
            <select value=${gameType} onChange=${this.onSelectGameType}>
              <option value="fishbowl">Fishbowl</option>
            </select>
          </label>
          <label>
            Name
            <input
              name="name"
              autocomplete="given-name"
              type="text"
              maxlength="20"
              value="${name}"
              placeholder="Enter your name"
              onInput=${this.onNameChange} />
          </label>
        </form>
        <button type="submit" class="lone">Create</button>
      </div>
    `;
  }

  handleMessage(data, e) {
    if (data.error) {
      this.setState({ error: data.error });
      return;
    }

    this.props.transitionToScreen(Constants.Screens.ROOM);
    this.props.updateStoreData({
      name: this.state.name,
      isRoomOwner: true,
      roomCode: data.body.roomCode,
      teams: [
        [{ name: this.state.name, isRoomOwner: true }],
      ],
    });
  }

  onSelectGameType(e) {
    this.setState({ gameType: e.target.value });
  }

  onNameChange(e) {
    this.setState({ name: e.target.value });
  }

  createGame(e) {
    e.preventDefault();

    const { conn } = this.props;
    const { gameType, name } = this.state;
    if (name.length === 0) {
      this.setState({ error: 'Please enter a name first.' });
      return;
    }

    conn.send(JSON.stringify({
      action: Constants.Actions.CREATE_GAME,
      body: {
        gameType: gameType,
        name: name,
      },
    }));
  }
}
