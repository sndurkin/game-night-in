import { html, Component, render } from 'https://unpkg.com/htm/preact/standalone.module.js';

import Constants from './Constants.js';


export default class RoomScreen extends Component {

  constructor(...args) {
    super(...args);

    this.state = {

      error: '',
    };
  }

  render() {
    const { roomCode, teams } = this.props;
    // const { gameType, name, error } = this.state;

    return html`
      <div class="screen">
        Room code: ${roomCode}
        ${(teams || []).map((team, idx) => html`
          <div class="team">
            <div class="team-title">Team ${idx + 1}</div>
            <ul class="team-list">
              ${(team || []).map(player => html`
                <li>${player.name}${player.wordsSet ? ' (words set)' : ''}</li>
              `)}
            </ul>
          </div>
        `)}
      </div>
    `;
  }

  handleMessage(data, e) {
    if (data.error) {
      this.setState({ error: data.error });
    }

    this.props.updateStoreData({ teams: data.body.teams });
  }

}
