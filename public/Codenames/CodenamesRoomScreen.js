import { html, Component, render } from 'https://unpkg.com/htm/preact/standalone.module.js';

import ScreenWrapper from '../ScreenWrapper.js';
import Utils from '../Utils.js';
import Constants from '../Constants.js';

import CodenamesConstants from './CodenamesConstants.js';


export default class CodenamesRoomScreen extends Component {

  constructor(...args) {
    super(...args);

    this.state = {
      error: '',

      changedSettings: null,

      showMovePlayerModal: false,
      playerToMove: null,
      teamIdxToMoveFrom: null,
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
          <h3>Rounds</h3>
          <table class="primary rounds-table">
            <tbody>
              ${rounds.map(this.renderRoundTableRow)}
            </tbody>
          </table>
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

  renderSubmitWords() {
    const { words, wordBeingEntered } = this.state;

    return html`
      <form>
        <label>
          Word
          <input
            ref=${r => this.inputRef = r}
            type="text"
            maxlength="20"
            value="${wordBeingEntered}"
            placeholder="Enter a word or phrase"
            onInput=${this.onWordChange} />
        </label>
      </form>
      <div class="button-bar">
        <button onClick=${this.addWord}>Add</button>
        <div></div>
        <button
          disabled=${words.length < 5}
          onClick=${this.submitWords}
          style="font-size: 0.9em"
        >
          Submit words
        </button>
      </div>
      <div class="word-list">
        <h4 class="word-list-title">Words:</h4>
        ${words.map((word, idx) => html`
          <div class="word-row">
            <div class="word">${word}</div>
            <div
              class="word-delete"
              onClick=${() => this.deleteWord(idx)}
            >
              ✖
            </div>
          </div>
        `)}
      </div>
    `;
  }

  renderRoom() {
    const { isRoomOwner } = this.props;
    const { showMovePlayerModal, teamIdxToMoveFrom } = this.state;
    const teams = this.props.teams || [];

    return html`
      <div class="teams">
        ${teams.map((team, idx) => html`
          <div class="team">
            <div class="team-title" style=${Utils.teamStyle(idx)}>
              Team ${idx + 1}
            </div>
            <div class="team-table">
              <div class="team-row">
                <div class="player-type">
                  Spymaster
                </div>
                <div class="player-name">
                  ${team.spymaster ? team.spymaster.name : 'None'}
                </div>
                ${isRoomOwner ? html`
                  <a onClick=${() => this.showMovePlayerModal(idx, player)}>
                    Move
                  </a>
                ` : player.isRoomOwner ? html`
                  <div>Owner</div>
                ` : ''}
              </div>
              <div class="team-row">
                <div class="player-type">
                  Guesser
                </div>
                <div class="player-name">
                  ${team.guesser ? team.guesser.name : 'None'}
                </div>
                ${isRoomOwner ? html`
                  <a onClick=${() => this.showMovePlayerModal(idx, player)}>
                    Move
                  </a>
                ` : player.isRoomOwner ? html`
                  <div>Owner</div>
                ` : ''}
              </div>
            </table>
          </div>
        `)
      }
      </div>
  ${
      isRoomOwner ? html`
        <div class="button-bar">
          <div></div>
          <button
            disabled=${!this.canStartGame}
            onClick=${this.startGame}>
            Start game
          </button>
        </div>
      ` : this.waitingMessage
      }
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
      ${teams.map((_, idx) => html`
              <span
                class="button stack"
                disabled=${teamIdxToMoveFrom === idx}
                onClick=${() => { teamIdxToMoveFrom !== idx && this.movePlayer({ teamIdxToMoveTo: idx }); }}
              >
                Team ${idx + 1}
              </span>
            `)}
    </section>
    <footer></footer>
  </article>
</div>
`;
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

  showMovePlayerModal(teamIdx, player) {
    const { teams } = this.props;
    if (teams.length === 2) {
      // If there are only 2 teams, don't bother showing the modal.
      this.movePlayer({
        name: player.name,
        teamIdxToMoveFrom: teamIdx,
        teamIdxToMoveTo: teamIdx === 0 ? 1 : 0,
      });
      return;
    }

    this.setState({
      showMovePlayerModal: true,
      playerToMove: player,
      teamIdxToMoveFrom: teamIdx,
    });
  }

  movePlayer({ name, teamIdxToMoveFrom, teamIdxToMoveTo }) {
    const { conn } = this.props;
    conn.send(JSON.stringify({
      action: CodenamesConstants.Actions.MOVE_PLAYER,
      body: {
        playerName: typeof name !== 'undefined' ? name : this.state.playerToMove.name,
        fromTeam: typeof teamIdxToMoveFrom !== 'undefined' ? teamIdxToMoveFrom : this.state.teamIdxToMoveFrom,
        toTeam: teamIdxToMoveTo,
      },
    }));

    this.setState({
      showMovePlayerModal: false,
      playerToMove: null,
      teamIdxToMoveFrom: null,
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
  < div class="room-toolbar" >
    ${ waitingStr}
        </div >
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
