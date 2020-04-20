import { html, Component, render } from 'https://unpkg.com/htm/preact/standalone.module.js';

import HomeScreen from './HomeScreen.js';


let conn;
class App extends Component {  
  render() {
    return html`
      <div class="app">
        <${HomeScreen} />
      </div>
    `;
  }
}
/*
const Header = ({ name }) => html`<h1>${name} List</h1>`

const Footer = props => html`<footer ...${props} />`
*/
render(html`<${App} page="All" />`, document.body);


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