import { html, Component, render } from 'https://unpkg.com/htm/preact/standalone.module.js';

import ScreenWrapper from './ScreenWrapper.js';
import HomeScreen from './HomeScreen.js';
import CreateRoomScreen from './CreateRoomScreen.js';
import JoinRoomScreen from './JoinRoomScreen.js';
import RoomScreen from './RoomScreen.js';
import Constants from './Constants.js';


let conn;
class App extends Component {

  constructor(...args) {
    super(...args);

    this.state = {
      page: Constants.Pages.HOME,
    };

    this.updateStoreData = this.updateStoreData.bind(this);
    this.transitionToPage = this.transitionToPage.bind(this);
  }

  componentDidMount() {
    const { conn } = this.props;

    conn.onmessage = (e) => {
      const data = JSON.parse(e.data);

      // TODO: Screen transition handling based on messages

      this.getActiveScreen().handleMessage(data, e);
    };
  }

  getActiveScreen() {
    switch (this.state.page) {
      case Constants.Pages.HOME:
        return this.homeScreen;
      case Constants.Pages.CREATE_ROOM:
        return this.createRoomScreen;
      case Constants.Pages.JOIN_ROOM:
        return this.joinRoomScreen;
      case Constants.Pages.ROOM:
        return this.roomScreen;
    }

    return null;
  }

  render() {
    const { conn } = this.props;
    const { page, ...storeData } = this.state;

    const sharedProps = {
      conn: conn,
      ...storeData,
      updateStoreData: this.updateStoreData,
      transitionToPage: this.transitionToPage,
    };

    return html`
      <div class="app">
        ${page === Constants.Pages.HOME && html`
          <${ScreenWrapper} ...${sharedProps}>
            <${HomeScreen} ref=${r => this.homeScreen = r} ...${sharedProps} />
          <//>
        `}
        ${page === Constants.Pages.CREATE_ROOM && html`
          <${ScreenWrapper}
            onBack=${() => this.transitionToPage(Constants.Pages.HOME)}
            ...${sharedProps}
          >
            <${CreateRoomScreen} ref=${r => this.createRoomScreen = r} ...${sharedProps} />
          <//>
        `}
        ${page === Constants.Pages.JOIN_ROOM && html`
          <${ScreenWrapper}
            onBack=${() => this.transitionToPage(Constants.Pages.HOME)}
            ...${sharedProps}
          >
            <${JoinRoomScreen} ref=${r => this.joinRoomScreen = r} ...${sharedProps} />
          <//>
        `}
        ${page === Constants.Pages.ROOM && html`
          <${ScreenWrapper}
            onBack=${() => this.transitionToPage(Constants.Pages.HOME)}
            ...${sharedProps}
          >
            <${RoomScreen} ref=${r => this.roomScreen = r} ...${sharedProps} />
          <//>
        `}
      </div>
    `;
  }

  updateStoreData(newStoreData) {
    this.setState(newStoreData);
  }

  transitionToPage(page) {
    this.setState({ page: page });
  }

  getStyle(page) {
    return `display: ${this.state.page === page ? 'block' : 'none'}`;
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
