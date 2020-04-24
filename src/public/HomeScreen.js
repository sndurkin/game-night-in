import { html, Component, render } from 'https://unpkg.com/htm/preact/standalone.module.js';

import Constants from './Constants.js';


export default class HomeScreen extends Component {

  constructor(...args) {
    super(...args);

    this.openCreateRoom = this.openCreateRoom.bind(this);
    this.openJoinRoom = this.openJoinRoom.bind(this);
  }

  render() {
    return html`
      <div class="screen">
        <button class="lone" onClick=${this.openCreateRoom}>Create new game</button>
        <div class="center-horiz">or</div>
        <button class="lone" onClick=${this.openJoinRoom}>Join existing game</button>
      </div>
    `;
  }

  openCreateRoom() {
    this.props.transitionToScreen(Constants.Screens.CREATE_ROOM);
  }

  openJoinRoom() {
    this.props.transitionToScreen(Constants.Screens.JOIN_ROOM);
  }

}
