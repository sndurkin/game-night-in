import { html, Component, render } from 'https://unpkg.com/htm/preact/standalone.module.js';

import ScreenWrapper from './ScreenWrapper.js';
import HomeScreen from './HomeScreen.js';
import CreateRoomScreen from './CreateRoomScreen.js';
import JoinRoomScreen from './JoinRoomScreen.js';
import Constants from './Constants.js';


let conn;
class App extends Component {

  constructor(...args) {
    super(...args);

    this.state = {
      page: Constants.Pages.HOME,
    };

    this.transitionToPage = this.transitionToPage.bind(this);
  }

  render() {
    const sharedProps = {
      transitionToPage: this.transitionToPage
    };

    return html`
      <div class="app">
        <${ScreenWrapper} style=${this.getStyle(Constants.Pages.HOME)}>
          <${HomeScreen}  ...${sharedProps} />
        <//>
        <${ScreenWrapper}
          onBack=${() => this.transitionToPage(Constants.Pages.HOME)}
          style=${this.getStyle(Constants.Pages.CREATE_ROOM)}
        >
          <${CreateRoomScreen}  ...${sharedProps} />
        <//>
        <${ScreenWrapper}
          onBack=${() => this.transitionToPage(Constants.Pages.HOME)}
          style=${this.getStyle(Constants.Pages.JOIN_ROOM)}
        >
          <${JoinRoomScreen}  ...${sharedProps} />
        <//>
      </div>
    `;
  }

  transitionToPage(page) {
    this.setState({ page: page });
  }

  getStyle(page) {
    return `display: ${this.state.page === page ? 'block' : 'none'}`;
  }
}
/*
const Header = ({ name }) => html`<h1>${name} List</h1>`

const Footer = props => html`<footer ...${props} />`
*/
render(html`<${App} />`, document.body);


window.onload = function () {
  if (!window["WebSocket"]) {
    document.open();
    document.write('<b>Your browser does not support WebSockets.</b>');
    document.close();
    return;
  }

  conn = new WebSocket("ws://" + document.location.host + "/ws");
  conn.onclose = function (evt) {
    /*
    document.open();
    document.write('<b>Connection closed.</b>');
    document.close();
    */
  };
  conn.onmessage = function (evt) {
    const data = JSON.parse(evt.data);
    switch (data.event) {
      case 'player-joined':
        for (let player of data.body.players) {
          addPlayer(player.name, player.isOwner);
        }
        break;
      case 'other-player-joined':
        addPlayer(data.body.name);
        break;
    }
  };
};
