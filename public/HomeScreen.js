import { html, Component, render } from 'https://unpkg.com/htm/preact/standalone.module.js';

import ScreenWrapper from './ScreenWrapper.js';
import Constants from './Constants.js';


export default class HomeScreen extends Component {

  constructor(...args) {
    super(...args);

    this.openCreateGame = this.openCreateGame.bind(this);
    this.openJoinGame = this.openJoinGame.bind(this);
  }

  render() {
    return html`
      <div class="screen">
        <div class="logo"></div>
        <button class="lone" onClick=${this.openCreateGame}>Create new game</button>
        <div class="center-horiz">or</div>
        <button class="lone" onClick=${this.openJoinGame}>Join existing game</button>
      </div>
    `;
  }

  openCreateGame() {
    this.props.transitionToScreen(Constants.Screens.CREATE_GAME);
  }

  openJoinGame() {
    this.props.transitionToScreen(Constants.Screens.JOIN_GAME);
  }

}
