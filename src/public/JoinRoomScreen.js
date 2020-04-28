import { html, Component, render } from 'https://unpkg.com/htm/preact/standalone.module.js';

import Constants from './Constants.js';


export default class JoinRoomScreen extends Component {

  constructor(...args) {
    super(...args);

    this.state = {
      name: Math.random().toString(36).substring(2, 8),
      roomCode: '6374',
      error: '',
    };

    this.onNameChange = this.onNameChange.bind(this);
    this.onRoomCodeChange = this.onRoomCodeChange.bind(this);
    this.joinRoom = this.joinRoom.bind(this);
  }

  render() {
    const { name, roomCode, error } = this.state;

    return html`
      <div class="screen">
        ${error && html`
          <span class="label error">${error}</span>
        `}
        <label>
          Room code
          <input
            type="text"
            maxlength="4"
            value="${roomCode}"
            placeholder="Enter the room invite code"
            onInput=${this.onRoomCodeChange} />
        </label>
        <label>
          Name
          <input
            type="text"
            maxlength="20"
            value="${name}"
            placeholder="Enter your name"
            onInput=${this.onNameChange} />
        </label>
        <button class="lone" onClick=${this.joinRoom}>Join</button>
      </div>
    `;
  }

  handleMessage(data, e) {
    if (data.error) {
      this.setState({ error: data.error });
      return;
    }

    // TODO: pull room owner info from server, because it's possible
    // for a user to create a room, get disconnected and try to rejoin
    // the room.
    const sharedProps = {
      name: this.state.name,
      isRoomOwner: false,
      roomCode: this.state.roomCode,
    };

    switch (data.event) {
      case Constants.Events.UPDATED_ROOM:
        this.props.updateStoreData({
          ...sharedProps,
          teams: data.body.teams
        });
        this.props.transitionToScreen(Constants.Screens.ROOM);
        break;
      case Constants.Events.UPDATED_GAME:
        this.props.updateStoreData({
          ...sharedProps,
          game: data.body
        });
        this.props.transitionToScreen(Constants.Screens.GAME);
        break;
    }
  }

  onNameChange(e) {
    this.setState({ name: e.target.value });
  }

  onRoomCodeChange(e) {
    this.setState({ roomCode: e.target.value });
  }

  joinRoom() {
    const { conn } = this.props;
    const { roomCode, name } = this.state;

    if (roomCode.length === 0) {
      this.setState({ error: 'Please enter a room code.' });
      return;
    }

    if (name.length === 0) {
      this.setState({ error: 'Please enter a name to join.' });
      return;
    }

    conn.send(JSON.stringify({
      action: Constants.Actions.JOIN_ROOM,
      body: {
        roomCode: roomCode,
        name: name,
      },
    }));
  }
}
