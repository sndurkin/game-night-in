import { html, Component, render } from 'https://unpkg.com/htm/preact/standalone.module.js';

import ScreenWrapper from '../ScreenWrapper.js';
import Utils from '../Utils.js';
import Constants from '../Constants.js';

import CodenamesConstants from './CodenamesConstants.js';


const playersInTeam = [{
  label: 'Spymaster',
  playerType: CodenamesConstants.PlayerType.SPYMASTER,
}, {
  label: 'Guesser',
  playerType: CodenamesConstants.PlayerType.GUESSER,
}];

export default class CodenamesRoomScreen extends Component {

  constructor(...args) {
    super(...args);

    this.state = {
      error: '',

      changedSettings: null,

      showMovePlayerModal: false,
      playerToMove: null,
      teamIdxToMoveFrom: null,
      playerTypeToMoveFrom: null,
    };

    this.openChangeSettings = this.openChangeSettings.bind(this);
    this.closeChangeSettings = this.closeChangeSettings.bind(this);
    this.saveSettings = this.saveSettings.bind(this);

    this.movePlayer = this.movePlayer.bind(this);
    this.startGame = this.startGame.bind(this);
  }

  componentDidMount() {

  }

  componentDidUpdate() {

  }

  render() {
    const { name, teams } = this.props;
    const { changedSettings, error } = this.state;

    if (changedSettings) {
      return this.renderChangeSettingsDialog();
    }

    return html`
      <${ScreenWrapper}
        onBack=${() => this.props.transitionToScreen(Constants.Screens.HOME)}
        ...${this.props}
      >
        <div class="screen">
          ${error && html`
            <span class="label error">${error}</span>
          `}
          ${this.renderRoom()}
        </div>
      <//>
    `;
  }

  renderChangeSettingsDialog() {
    const { error } = this.state;
    const { rounds, timerLength } = this.settings;

    const header = html`
      <button
        class="close-settings pseudo"
        onClick=${this.closeChangeSettings}
      >
        ✖
      </button>
    `;

    return html`
      <${ScreenWrapper} header=${header}>
        <div class="screen">
          ${error && html`
            <span class="label error">${error}</span>
          `}
          <h3>Timer length</h3>
          <div style="display: flex; align-items: center">
            <input
              type="number"
              style="width: 4em"
              value=${timerLength}
              onChange=${this.onTimerChange}
            />
            <div style="margin-left: 1em">seconds</div>
          </div>
          <button class="lone" onClick=${this.saveSettings}>Save</button>
        </div>
      <//>
    `;
  }

  renderRoundTableRow(round, idx) {
    return html`
      <tr>
        <td>${idx + 1}</td>
        <td width="99%">
          <select onChange=${e => this.selectRound(idx, e)}>
            ${Object.entries(CodenamesConstants.RoundTypes).map(e => html`
              <option value=${e[0]} selected=${e[0] === round}>
                ${e[1].title}
              </option>
            `)}
          </select>
        </td>
        <td>
          <button class="pseudo" onClick=${() => this.removeRound(idx)}>
            ✖
          </button>
        </td>
      </tr>
    `;
  }

  renderRoom() {
    const { name, isRoomOwner } = this.props;
    const {
      showMovePlayerModal
    } = this.state;

    return html`
      ${this.renderTeams()}
      ${isRoomOwner ? html`
        <div class="button-bar">
          <div></div>
          <button
            disabled=${!this.canStartGame}
            onClick=${this.startGame}>
            Start game
          </button>
        </div>
      ` : this.waitingMessage}
      <div class="modal">
        <input
          id="move-player-modal"
          type="checkbox"
          checked=${showMovePlayerModal ? 'checked' : ''} />
        <label for="move-player-modal" class="overlay"></label>
        <article>
          <header>
            <h3>Select a team</h3>
            <label for="move-player-modal" class="close">✖</label>
          </header>
          <section class="content">
            ${this.renderTeamsToSelect()}
          </section>
          <footer></footer>
        </article>
      </div>
    `;
  }

  renderTeams() {
    const { name, isRoomOwner } = this.props;
    const teams = this.props.teams || [];

    const teamsContent = [];
    for (let teamIdx = 0; teamIdx < teams.length; teamIdx++) {
      const team = teams[teamIdx];
      const playersContent = [];
      for (let playerIdx = 0; playerIdx < playersInTeam.length; playerIdx++) {
        const player = playersInTeam[playerIdx];
        let playerName;
        let boldClass = '';
        if (team.players[player.playerType]) {
          playerName = team.players[player.playerType].name;
          if (team.players[player.playerType].name === name) {
            boldClass = ' bold';
          }
        } else {
          playerName = html`
            <span class="codenames-player-none">(none)</span>
          `;
        }

        let playerOptions = null;
        if (team.players[player.playerType]) {
          if (isRoomOwner) {
            const movePlayer = () => {
              this.showMovePlayerModal(
                team.players[player.playerType],
                teamIdx,
                player.playerType);
            };

            playerOptions = html`
              <a role="link" class="codenames-player-meta" onClick=${movePlayer}>
                Move
              </a>
            `;
          } else if (team.players[player.playerType].isRoomOwner) {
            playerOptions = html`
              <div class="codenames-player-meta">Owner</div>
            `;
          }
        }

        playersContent.push(html`
          <div class="codenames-team-row">
            <div class="codenames-player">
              <div class="codenames-player-type">
                ${player.label}
              </div>
              <div class=${'codenames-player-name' + boldClass}>
                ${playerName}
              </div>
            </div>
            ${playerOptions}
          </div>
        `);
      }

      teamsContent.push(html`
        <div class="codenames-team">
          <div class="codenames-team-header" style=${Utils.teamStyle(teamIdx)}>
            <div class="codenames-team-title">Team ${teamIdx + 1}</div>
          </div>
          <div class="codenames-team-table">
            ${playersContent}
          </div>
        </div>
      `);
    }

    return html`
      <div class="codenames-teams">
        ${teamsContent}
      </div>
    `;
  }

  renderTeamsToSelect() {
    const { teamIdxToMoveFrom, playerTypeToMoveFrom } = this.state;
    const teams = this.props.teams || [];
    const content = [];

    for (let teamIdx = 0; teamIdx < teams.length; teamIdx++) {
      for (let playerIdx = 0; playerIdx < playersInTeam.length; playerIdx++) {
        const player = playersInTeam[playerIdx];

        const disabled = teamIdxToMoveFrom === teamIdx
          && playerTypeToMoveFrom === player.playerType;

        const onClick = () => {
          if (teamIdxToMoveFrom !== teamIdx || playerTypeToMoveFrom !== player.playerType) {
            this.movePlayer({ teamIdxToMoveTo: teamIdx, playerTypeToMoveTo: player.playerType });
          }
        };

        content.push(html`
          <span class="button stack" disabled=${ disabled} onClick=${onClick}>
            Team ${ teamIdx + 1} • ${player.label}
          </span>
        `);
      }
    }

    return content;
  }

  handleMessage(data, e) {
    if (data.error) {
      this.setState({ error: data.error });
    }

    switch (data.event) {
      case Constants.Events.UPDATED_ROOM:
        this.props.updateStoreData({
          teams: data.body.teams,
          settings: data.body.settings,
        });
        break;
      case Constants.Events.UPDATED_GAME:
        this.props.updateStoreData({ game: data.body });
        this.props.transitionToScreen(CodenamesConstants.Screens.GAME);
        break;
    }
  }

  onTimerChange(e) {
    const newSettings = JSON.parse(JSON.stringify(this.settings));
    try {
      newSettings.timerLength = parseInt(e.target.value, 10);

      this.setState({
        changedSettings: newSettings,
      });
    }
    catch (e) {
      // Ignore
    }
  }

  saveSettings() {
    const settings = this.state.changedSettings;
    this.setState({
      changedSettings: null,
    });

    const { conn } = this.props;
    conn.send(JSON.stringify({
      action: CodenamesConstants.Actions.CHANGE_SETTINGS,
      body: {
        settings: settings,
      },
    }));
  }

  get settings() {
    return this.state.changedSettings || this.props.settings;
  }

  openChangeSettings() {
    this.setState({
      changedSettings: JSON.parse(JSON.stringify(this.props.settings)),
    });
  }

  closeChangeSettings() {
    this.setState({
      changedSettings: null,
    });
  }

  showMovePlayerModal(player, teamIdxToMoveFrom, playerTypeToMoveFrom) {
    const { teams } = this.props;

    this.setState({
      showMovePlayerModal: true,
      playerToMove: player,
      teamIdxToMoveFrom: teamIdxToMoveFrom,
      playerTypeToMoveFrom: playerTypeToMoveFrom,
    });
  }

  movePlayer({ name, teamIdxToMoveTo, playerTypeToMoveTo }) {
    const { conn } = this.props;
    conn.send(JSON.stringify({
      action: CodenamesConstants.Actions.MOVE_PLAYER,
      body: {
        playerName: typeof name !== 'undefined' ? name : this.state.playerToMove.name,
        toTeam: teamIdxToMoveTo,
        toPlayerType: playerTypeToMoveTo
      },
    }));

    this.setState({
      showMovePlayerModal: false,
      playerToMove: null,
      teamIdxToMoveFrom: null,
      playerTypeToMoveTo: null,
    });
  }

  startGame() {
    const { conn } = this.props;
    conn.send(JSON.stringify({
      action: Constants.Actions.START_GAME,
      body: {},
    }));
  }

  get waitingMessage() {
    if (this.canStartGame) {
      const waitingStr = `Waiting for ${this.roomOwner.name} to start the game`;
      return html`
        <div class="room-toolbar">
          ${ waitingStr}
        </div>
        `;
    }

    return null;
  }

  // To start the game, each team needs at least the minimum number of players
  // and everyone needs to have their words submitted.
  get canStartGame() {
    const { teams } = this.props;
    return teams.every(players => {
      return players.length >= CodenamesConstants.Game.MIN_PLAYERS_PER_TEAM
        && players.every(p => p.wordsSubmitted);
    });
  }

  get roomOwner() {
    const { teams } = this.props;
    for (let players of teams) {
      const player = players.find(p => p.isRoomOwner);
      if (player) {
        return player;
      }
    }

    return null;
  }

  get player() {
    const { name, teams } = this.props;
    for (let players of teams) {
      const player = players.find(p => p.name === name);
      if (player) {
        return player;
      }
    }

    return null;
  }

}
