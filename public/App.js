import { html, Component, render } from 'https://unpkg.com/htm/preact/standalone.module.js';

import HomeScreen from './HomeScreen.js';
import CreateGameScreen from './CreateGameScreen.js';
import JoinGameScreen from './JoinGameScreen.js';
import Constants from './Constants.js';

import FishbowlRoomScreen from './Fishbowl/FishbowlRoomScreen.js';
import FishbowlGameScreen from './Fishbowl/FishbowlGameScreen.js';
import FishbowlGameOverScreen from './Fishbowl/FishbowlGameOverScreen.js';
import FishbowlConstants from './Fishbowl/FishbowlConstants.js';


let conn;
class App extends Component {

  constructor(...args) {
    super(...args);

    this.state = {
      screen: Constants.Screens.HOME,
      game: {},
    };

    this.updateStoreData = this.updateStoreData.bind(this);
    this.transitionToScreen = this.transitionToScreen.bind(this);
  }

  componentDidMount() {
    const { conn } = this.props;

    conn.onmessage = (e) => {
      const data = JSON.parse(e.data);
      this.getActiveScreen().handleMessage(data, e);
    };
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
    }

    return null;
  }

  render() {
    const { conn } = this.props;
    const { screen, ...storeData } = this.state;

    const sharedProps = {
      conn: conn,
      ...storeData,
      updateStoreData: this.updateStoreData,
      transitionToScreen: this.transitionToScreen,
    };

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
      </div>
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
}

window.onload = function () {
  if (!window['WebSocket']) {
    document.open();
    document.write('<b>Your browser does not support WebSockets.</b>');
    document.close();
    return;
  }

  const protocol = document.location.protocol === 'https:' ? 'wss' : 'ws';
  conn = new WebSocket(protocol + '://' + document.location.host + '/ws');
  conn.onclose = function (e) {
    while (document.body.firstChild) {
      document.body.removeChild(document.body.firstChild);
    }

    render(html`
      <div>
        <h3>Oh no! We lost the connection to the server.</h3>
        <p>Refresh and then try to rejoin with the same room code:</p>
        <button onClick=${refresh}>Refresh</button>
      </div>
    `, document.body);
  };

  render(html`<${App} conn=${conn} />`, document.body);
};

function refresh() {
  window.location.reload();
}
