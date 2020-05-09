import { html, Component, render } from 'https://unpkg.com/htm/preact/standalone.module.js';

import ScreenWrapper from './ScreenWrapper.js';
import Constants from './Constants.js';


export default class JoinGameScreen extends Component {

  constructor(...args) {
    super(...args);

    this.state = {
      name: localStorage.getItem(Constants.LocalStorage.PLAYER_NAME) || '',
      roomCode: localStorage.getItem(Constants.LocalStorage.ROOM_CODE) || '',
      error: '',
    };

    this.onNameChange = this.onNameChange.bind(this);
    this.onRoomCodeChange = this.onRoomCodeChange.bind(this);
    this.joinGame = this.joinGame.bind(this);
  }

  render() {
    const { name, roomCode, error } = this.state;

    return html`
      <${ScreenWrapper}
        onBack=${() => this.props.transitionToScreen(Constants.Screens.HOME)}
        ...${this.props}
      >
        <div class="screen">
          ${error && html`
            <span class="label error">${error}</span>
          `}
          <form onSubmit=${this.joinGame}>
            <label>
              Room code
              <input
                name="room-code"
                autocomplete="room-code"
                type="text"
                maxlength="4"
                value=${roomCode}
                placeholder="Enter the room invite code"
                onInput=${this.onRoomCodeChange} />
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
            <button type="submit" class="lone">Join</button>
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

    const sharedProps = {
      name: this.state.name,
      isRoomOwner: false,
      roomCode: this.state.roomCode,
    };

    // Set isRoomOwner in case the current client is the room owner
    // and rejoining a game.
    if (data.body.teams) {
      for (let players of data.body.teams) {
        const player = players.find(p => p.name === this.state.name);
        if (player) {
          sharedProps.isRoomOwner = player.isRoomOwner;
          break;
        }
      }
    }

    switch (data.event) {
      case Constants.Events.UPDATED_ROOM:
        this.props.updateStoreData({
          ...sharedProps,
          gameType: data.body.gameType,
          settings: data.body.settings,
          teams: data.body.teams,
        });
        this.props.transitionToScreen(Constants.Screens.FISHBOWL_ROOM);
        break;
      case Constants.Events.UPDATED_GAME:
        this.props.updateStoreData({
          ...sharedProps,
          teams: data.body.teams,
          game: data.body,
        });
        this.props.transitionToScreen(Constants.Screens.FISHBOWL_GAME);
        break;
    }
  }

  onNameChange(e) {
    this.setState({ name: e.target.value });
  }

  onRoomCodeChange(e) {
    this.setState({ roomCode: e.target.value });
  }

  joinGame(e) {
    e.preventDefault();

    const { conn } = this.props;
    const { roomCode, name } = this.state;

    if (roomCode.length === 0) {
      this.setState({ error: 'Please enter a room code.' });
      return;
    }

    const trimmedName = name.trim();
    if (trimmedName.length === 0) {
      this.setState({ error: 'Please enter a name to join.' });
      return;
    }

    localStorage.setItem(Constants.LocalStorage.PLAYER_NAME, trimmedName);
    localStorage.setItem(Constants.LocalStorage.ROOM_CODE, roomCode);

    conn.send(JSON.stringify({
      action: Constants.Actions.JOIN_GAME,
      body: {
        roomCode: roomCode,
        name: trimmedName,
      },
    }));
  }
}
