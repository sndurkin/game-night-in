import { html, Component } from 'https://unpkg.com/htm/preact/standalone.module.js';

import ScreenWrapper from './ScreenWrapper.js';
import Constants from './Constants.js';

import FishbowlConstants from './Fishbowl/FishbowlConstants.js';


export default class CreateGameScreen extends Component {

  constructor(...args) {
    super(...args);

    this.state = {
      gameType: 'fishbowl',
      name: localStorage.getItem(Constants.LocalStorage.PLAYER_NAME) || '',
      error: '',
    };

    this.onSelectGameType = this.onSelectGameType.bind(this);
    this.onNameChange = this.onNameChange.bind(this);
    this.createGame = this.createGame.bind(this);
  }

  render() {
    const { gameType, name, error } = this.state;

    return html`
      <${ScreenWrapper}
        onBack=${() => this.props.transitionToScreen(Constants.Screens.HOME)}
        ...${this.props}
      >
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
                value=${name}
                placeholder="Enter your name"
                onInput=${this.onNameChange} />
            </label>
            <button type="submit" class="lone">Create</button>
          </form>
        </div>
      <//>
    `;
  }

  handleMessage(data, e) {
    if (data.error) {
      this.setState({ error: data.error });
      return;
    }

    switch (data.event) {
      case Constants.Events.CREATED_GAME:
        localStorage.setItem(Constants.LocalStorage.ROOM_CODE, data.body.roomCode);

        this.props.transitionToScreen(FishbowlConstants.Screens.ROOM);
        this.props.updateStoreData({
          name: this.state.name,
          isRoomOwner: true,
          roomCode: data.body.roomCode,
          gameType: data.body.gameType,
          teams: [
            [{ name: this.state.name, isRoomOwner: true }],
          ],
        });
        break;
    }
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
    const trimmedName = name.trim();
    if (trimmedName.length === 0) {
      this.setState({ error: 'Please enter a name first.' });
      return;
    }

    localStorage.setItem(Constants.LocalStorage.PLAYER_NAME, trimmedName);

    conn.send(JSON.stringify({
      action: Constants.Actions.CREATE_GAME,
      body: {
        gameType: gameType,
        name: trimmedName,
      },
    }));
  }
}
