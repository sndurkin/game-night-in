import { html, Component, render } from 'https://unpkg.com/htm/preact/standalone.module.js';

import Constants from './Constants.js';


export default class HomeScreen extends Component {

  constructor(...args) {
    super(...args);

    this.state = {
      name: '',
      error: '',
    };

    this.createRoom = this.createRoom.bind(this);
    this.joinRoom = this.joinRoom.bind(this);
  }

  render(props, state) {
    return html`
      <div class="screen">
        <button class="lone" onClick=${this.createRoom}>Create new game</button>
        <div class="center-horiz">or</div>
        <button class="lone" onClick=${this.joinRoom}>Join existing game</button>
      </div>
    `;
  }

  createRoom() {
    this.props.transitionToPage(Constants.Pages.CREATE_ROOM);
  }

  joinRoom() {
    this.props.transitionToPage(Constants.Pages.JOIN_ROOM);
  }

  onNameChange(e) {
    this.setState({ name: e.target.value });
  }

  join() {
    if (this.name.length === 0) {
      this.error = 'Please enter a name to join.';
      return;
    }

    conn.send(JSON.stringify({
      action: 'join',
      body: {
        name: this.name
      }
    }));
  }
}
