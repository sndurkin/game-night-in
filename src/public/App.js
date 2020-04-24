import { html, Component, render } from 'https://unpkg.com/htm/preact/standalone.module.js';

import ScreenWrapper from './ScreenWrapper.js';
import HomeScreen from './HomeScreen.js';
import CreateRoomScreen from './CreateRoomScreen.js';
import JoinRoomScreen from './JoinRoomScreen.js';
import RoomScreen from './RoomScreen.js';
import GameScreen from './GameScreen.js';
import Constants from './Constants.js';


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
      case Constants.Screens.CREATE_ROOM:
        return this.createRoomScreen;
      case Constants.Screens.JOIN_ROOM:
        return this.joinRoomScreen;
      case Constants.Screens.ROOM:
        return this.roomScreen;
      case Constants.Screens.GAME:
        return this.gameScreen;
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
          <${ScreenWrapper} ...${sharedProps}>
            <${HomeScreen} ref=${r => this.homeScreen = r} ...${sharedProps} />
          <//>
        `}
        ${screen === Constants.Screens.CREATE_ROOM && html`
          <${ScreenWrapper}
            onBack=${() => this.transitionToScreen(Constants.Screens.HOME)}
            ...${sharedProps}
          >
            <${CreateRoomScreen} ref=${r => this.createRoomScreen = r} ...${sharedProps} />
          <//>
        `}
        ${screen === Constants.Screens.JOIN_ROOM && html`
          <${ScreenWrapper}
            onBack=${() => this.transitionToScreen(Constants.Screens.HOME)}
            ...${sharedProps}
          >
            <${JoinRoomScreen} ref=${r => this.joinRoomScreen = r} ...${sharedProps} />
          <//>
        `}
        ${screen === Constants.Screens.ROOM && html`
          <${ScreenWrapper}
            onBack=${() => this.transitionToScreen(Constants.Screens.HOME)}
            ...${sharedProps}
          >
            <${RoomScreen} ref=${r => this.roomScreen = r} ...${sharedProps} />
          <//>
        `}
        ${screen === Constants.Screens.GAME && html`
          <${ScreenWrapper} ...${sharedProps}>
            <${GameScreen} ref=${r => this.gameScreen = r} ...${sharedProps} />
          <//>
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

window.onload = function() {
  if (!window['WebSocket']) {
    document.open();
    document.write('<b>Your browser does not support WebSockets.</b>');
    document.close();
    return;
  }

  conn = new WebSocket('ws://' + document.location.host + '/ws');
  conn.onclose = function(e) {
    document.open();
    document.write('');
    document.close();
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
