import { html, Component, render } from 'https://unpkg.com/htm/preact/standalone.module.js';


export default class JoinRoomScreen extends Component {

  constructor(...args) {
    super(...args);

    this.state = {
      name: '',
      roomCode: '',
      error: '',
    };

    this.onNameChange = this.onNameChange.bind(this);
    this.onRoomCodeChange = this.onRoomCodeChange.bind(this);
    this.joinRoom = this.joinRoom.bind(this);
  }

  render(props, state) {
    const { name, roomCode } = state;

    return html`
      <div class="screen">
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

  onNameChange(e) {
    this.setState({ name: e.target.value });
  }

  onRoomCodeChange(e) {
    this.setState({ roomCode: e.target.value });
  }

  joinRoom() {
    if (this.name.length === 0) {
      this.error = 'Please enter a name to join.';
      return;
    }

    conn.send(JSON.stringify({
      action: Constants.Actions.JOIN_ROOM,
      body: {
        roomCode: '3487',
        name: this.name
      }
    }));
  }
}
