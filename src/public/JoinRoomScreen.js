import { html, Component, render } from 'https://unpkg.com/htm/preact/standalone.module.js';

import Constants from './Constants.js';


export default class JoinRoomScreen extends Component {

  constructor(...args) {
    super(...args);

    this.state = {
      name: Math.random().toString(36).substring(2, 15) + Math.random().toString(36).substring(2, 15),
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

    this.props.transitionToScreen(Constants.Screens.ROOM);
    this.props.updateStoreData({
      name: this.state.name,
      isRoomOwner: false,
      roomCode: this.state.roomCode,
      teams: data.body.teams,
    });
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
