import { html, Component, render } from 'https://unpkg.com/htm/preact/standalone.module.js';

import HomeScreen from './HomeScreen.js';
import CreateGameScreen from './CreateGameScreen.js';
import JoinGameScreen from './JoinGameScreen.js';
import Constants from './Constants.js';

import FishbowlRoomScreen from './Fishbowl/FishbowlRoomScreen.js';
import FishbowlGameScreen from './Fishbowl/FishbowlGameScreen.js';
import FishbowlGameOverScreen from './Fishbowl/FishbowlGameOverScreen.js';
import FishbowlConstants from './Fishbowl/FishbowlConstants.js';

import CodenamesRoomScreen from './Codenames/CodenamesRoomScreen.js';
import CodenamesConstants from './Codenames/CodenamesConstants.js';


class App extends Component {

  constructor(...args) {
    super(...args);

    this.state = {
      state: null,
      conn: null,

      screen: Constants.Screens.HOME,
      game: {},
    };

    this.updateStoreData = this.updateStoreData.bind(this);
    this.transitionToScreen = this.transitionToScreen.bind(this);
  }

  componentDidMount() {
    this.connect();
  }

  getActiveScreen() {
    switch (this.state.screen) {
      case Constants.Screens.HOME:
        return this.homeScreen;
      case Constants.Screens.CREATE_GAME:
        return this.createGameScreen;
      case Constants.Screens.JOIN_GAME:
        return this.joinGameScreen;
      case FishbowlConstants.Screens.ROOM:
        return this.fishbowlRoomScreen;
      case FishbowlConstants.Screens.GAME:
        return this.fishbowlGameScreen;
      case FishbowlConstants.Screens.GAME_OVER:
        return this.fishbowlGameOverScreen;
      case CodenamesConstants.Screens.ROOM:
        return this.codenamesRoomScreen;
    }

    throw new Error('Screen not supported: ' + this.state.screen);
  }

  render() {
    const { conn, state, screen, ...storeData } = this.state;

    const sharedProps = {
      conn: conn,
      ...storeData,
      updateStoreData: this.updateStoreData,
      transitionToScreen: this.transitionToScreen,
    };

    if (state === 'disconnected') {
      return html`
        <div>
          <h3>Oh no! We lost the connection to the server.</h3>
          <p>Refresh and then try to rejoin with the same room code:</p>
          <button onClick=${this.refresh}>Refresh</button>
        </div>
      `;
    }

    return html`
      <div class="app">
        ${screen === Constants.Screens.HOME && html`
          <${HomeScreen} ref=${r => this.homeScreen = r} ...${sharedProps} />
        `}
        ${screen === Constants.Screens.CREATE_GAME && html`
          <${CreateGameScreen} ref=${r => this.createGameScreen = r} ...${sharedProps} />
        `}
        ${screen === Constants.Screens.JOIN_GAME && html`
          <${JoinGameScreen} ref=${r => this.joinGameScreen = r} ...${sharedProps} />
        `}
        ${screen === FishbowlConstants.Screens.ROOM && html`
          <${FishbowlRoomScreen} ref=${r => this.fishbowlRoomScreen = r} ...${sharedProps} />
        `}
        ${screen === FishbowlConstants.Screens.GAME && html`
          <${FishbowlGameScreen} ref=${r => this.fishbowlGameScreen = r} ...${sharedProps} />
        `}
        ${screen === FishbowlConstants.Screens.GAME_OVER && html`
          <${FishbowlGameOverScreen} ref=${r => this.fishbowlGameOverScreen = r} ...${sharedProps} />
        `}
        ${screen === CodenamesConstants.Screens.ROOM && html`
          <${CodenamesRoomScreen} ref=${r => this.codenamesRoomScreen = r} ...${sharedProps} />
        `}
      </div>
      ${state === 'connecting' ? html`
        <div class="connecting" />
      ` : null}
    `;
  }

  updateStoreData(newStoreData) {
    this.setState(newStoreData);
  }

  transitionToScreen(screen) {
    this.setState({ screen: screen });
  }

  getStyle(screen) {
    return `display: ${this.state.screen === screen ? 'block' : 'none'}`;
  }

  connect(reconnectAttemptNumber) {
    reconnectAttemptNumber = reconnectAttemptNumber || 0;

    const protocol = document.location.protocol === 'https:' ? 'wss' : 'ws';
    const conn = new WebSocket(protocol + '://' + document.location.host + '/ws');

    this.setState({
      conn: conn,
      state: 'connecting'
    });

    conn.onopen = () => {
      reconnectAttemptNumber = 0;
      this.setState({
        state: 'connected'
      });
    };

    conn.onmessage = (e) => {
      const data = JSON.parse(e.data);
      this.getActiveScreen().handleMessage(data, e);
    };

    conn.onclose = (e) => {
      this.setState({
        conn: null,
        state: 'disconnected'
      });

      if (reconnectAttemptNumber < 3) {
        this.connect(reconnectAttemptNumber + 1);
      }
    };
  }

  refresh() {
    window.location.reload();
  }
}

window.onload = function () {
  if (!window['WebSocket']) {
    document.open();
    document.write('<b>Your browser does not support WebSockets.</b>');
    document.close();
    return;
  }

  render(html`<${App} />`, document.body);
};
