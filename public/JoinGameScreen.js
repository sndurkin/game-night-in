import { html, Component, render } from 'https://unpkg.com/htm/preact/standalone.module.js';

import ScreenWrapper from './ScreenWrapper.js';
import Constants from './Constants.js';

import FishbowlUtils from './Fishbowl/FishbowlUtils.js';
import FishbowlConstants from './Fishbowl/FishbowlConstants.js';
import CodenamesConstants from './Codenames/CodenamesConstants.js';


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
                maxlength="40"
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

    switch (data.body.gameType) {
      case 'fishbowl':
        // Set isRoomOwner in case the current client is the room owner
        // and rejoining a game.
        if (data.body.teams) {
          const player = FishbowlUtils.getPlayerByName(data.body.teams,
            this.state.name);
          Object.assign(sharedProps, player);
        }
        break;
      case 'codenames':
        // Set isRoomOwner in case the current client is the room owner
        // and rejoining a game.
        if (data.body.teams) {
          for (let team of data.body.teams) {
            for (let player of team.players) {
              if (player && player.name === this.state.name) {
                Object.assign(sharedProps, player);
              }
            }
          }
        }
        break;
    }

    switch (data.event) {
      case Constants.Events.UPDATED_ROOM:
        this.props.updateStoreData({
          ...sharedProps,
          gameType: data.body.gameType,
          settings: data.body.settings,
          teams: data.body.teams,
        });

        switch (data.body.gameType) {
          case 'fishbowl':
            this.props.transitionToScreen(FishbowlConstants.Screens.ROOM);
            break;
          case 'codenames':
            this.props.transitionToScreen(CodenamesConstants.Screens.ROOM);
            break;
        }
        break;
      case Constants.Events.UPDATED_GAME:
        this.props.updateStoreData({
          ...sharedProps,
          gameType: data.body.gameType,
          teams: data.body.teams,
          settings: data.body.settings,
          game: data.body,
        });

        switch (data.body.gameType) {
          case 'fishbowl':
            this.props.transitionToScreen(FishbowlConstants.Screens.GAME);
            break;
          case 'codenames':
            this.props.transitionToScreen(CodenamesConstants.Screens.GAME);
            break;
        }
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
    window.top.SessionStorage[Constants.LocalStorage.PLAYER_NAME] = trimmedName;
    window.top.SessionStorage[Constants.LocalStorage.ROOM_CODE] = roomCode;

    conn.send(JSON.stringify({
      action: Constants.Actions.JOIN_GAME,
      body: {
        roomCode: roomCode,
        name: trimmedName,
      },
    }));
  }
}
