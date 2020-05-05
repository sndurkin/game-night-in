import { html, Component, render } from 'https://unpkg.com/htm/preact/standalone.module.js';

import ScreenWrapper from './ScreenWrapper.js';
import Utils from './Utils.js';
import Constants from './Constants.js';

const WORDS = [
  'apple', 'jump', 'brick', 'red', 'simple', 'potatoes', 'sack', 'lump', 'fowl', 'biscuit', 'cheese',
  'pout', 'leaf', 'tree', 'cow', 'phone', 'call', 'table', 'sawing', 'drilling', 'work', 'yellow',
  'turns', 'wet', 'lucky', 'temperate', 'climate', 'cattle', 'string', 'bell', 'cut', 'scissors',
  'time', 'long', 'afterlife', 'west', 'lol', 'points', 'game', 'score', 'computer', 'keyboard',
  'typist', 'astronomy', 'astrology', 'stars', 'castle', 'bill', 'dry', 'toast', 'less', 'more',
  'capital', 'whine', 'wine', 'plate', 'card', 'word', 'letter', 'like', 'love', 'worship', 'war',
  'famine', 'dust', 'bowl', 'iron', 'oreo', 'touchless', 'run', 'brevity', 'high', 'low', 'mumps',
  'orange', 'orangutan', 'organ', 'piano', 'violin', 'keytar', 'tar', 'key', 'lock', 'hole', 'munch',
  'rain', 'clouds', 'sky', 'thistle', 'vine', 'grow', 'plant', 'tend', 'garden', 'butterfly',
];
function getWords(num) {
  const words = [];
  for (let i = 0; i < num; i++) {
    const idx = Math.floor(Math.random() * WORDS.length);
    words.push(WORDS[idx]);
    WORDS.splice(idx, 1);
  }
  return words;
}

export default class RoomScreen extends Component {

  constructor(...args) {
    super(...args);

    this.state = {
      error: '',

      changedSettings: null,

      wordBeingEntered: '',
      words: document.location.protocol === 'https:' ? [] : getWords(5),

      showMovePlayerModal: false,
      playerToMove: null,
      teamIdxToMoveFrom: null,
    };

    this.onWordChange = this.onWordChange.bind(this);
    this.addWord = this.addWord.bind(this);
    this.submitWords = this.submitWords.bind(this);
    this.openChangeSettings = this.openChangeSettings.bind(this);
    this.closeChangeSettings = this.closeChangeSettings.bind(this);
    this.renderRoundTableRow = this.renderRoundTableRow.bind(this);
    this.selectRound = this.selectRound.bind(this);
    this.removeRound = this.removeRound.bind(this);
    this.saveSettings = this.saveSettings.bind(this);

    this.addTeam = this.addTeam.bind(this);
    this.movePlayer = this.movePlayer.bind(this);
    this.startGame = this.startGame.bind(this);
  }

  componentDidMount() {
    this.lastWordsLength = this.state.words.length;
  }

  componentDidUpdate() {
    if (this.lastWordsLength < this.state.words.length && this.inputRef) {
      this.inputRef.focus();
    }

    this.lastWordsLength = this.state.words.length;
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
          ${!this.player.wordsSubmitted ? this.renderSubmitWords() : this.renderRoom()}
        </div>
      <//>
    `;
  }

  renderChangeSettingsDialog() {
    const { error } = this.state;
    const { rounds } = this.settings;

    const header = html`
      <button class="close-settings pseudo">✖</button>
    `;

    return html`
      <${ScreenWrapper} header=${header}>
        <div class="screen">
          ${error && html`
            <span class="label error">${error}</span>
          `}
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
            ${Object.entries(Constants.Fishbowl.RoundTypes).map(e => html`
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
      ${this.renderSettingsSummary()}
      <div class="teams">
        ${teams.map((team, idx) => html`
          <div class="team">
            <div class="team-title" style=${Utils.teamStyle(idx)}>
              Team ${idx + 1}
            </div>
            <div class="team-table">
              ${(team || []).length === 0 ? html`
                <div class="empty-list">No players yet!</div>
              ` : (team || []).map(player => html`
                <div class="team-row">
                  <div class="player-ready">
                    ${player.wordsSubmitted ? '✔' : ''}
                  </div>
                  <div class="player-name">
                    ${player.name}
                  </div>
                  ${isRoomOwner ? html`
                    <a onClick=${() => this.showMovePlayerModal(idx, player)}>
                      Move
                    </a>
                  ` : player.isRoomOwner ? html`
                    <div>Owner</div>
                  ` : ''}
                </div>
              `)}
            </table>
          </div>
        `)}
      </div>
      ${isRoomOwner ? html`
        <div class="button-bar">
          <button
            onClick=${this.addTeam}
            disabled=${teams.length >= Constants.Fishbowl.MAX_TEAMS}
          >
            Add team
          </button>
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

  renderSettingsSummary() {
    const { isRoomOwner } = this.props;
    const { rounds } = this.settings;

    return html`
      <div class="settings">
        <div class="settings-header">
          <div class="settings-title">Settings</div>
          ${isRoomOwner ? html`
            <a onClick=${this.openChangeSettings}>Change</a>
          ` : null}
        </div>
        <div>
          <span>${rounds.length} round${rounds.length !== 1 ? 's' : ''}</span>
        </div>
      </div >
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
        this.props.transitionToScreen(Constants.Screens.GAME);
        break;
    }
  }

  selectRound(idx, e) {
    const newSettings = JSON.parse(JSON.stringify(this.settings));
    newSettings.rounds[idx] = e.target.value;

    this.setState({
      changedSettings: newSettings,
    });
  }

  removeRound(idx) {
    const newSettings = JSON.parse(JSON.stringify(this.settings));
    newSettings.rounds.splice(idx, 1);

    this.setState({
      changedSettings: newSettings,
    });
  }

  saveSettings() {
    const settings = this.state.changedSettings;
    this.setState({
      changedSettings: null,
    });

    const { conn } = this.props;
    conn.send(JSON.stringify({
      action: Constants.Actions.CHANGE_SETTINGS,
      body: {
        settings: settings,
      },
    }));
  }

  get settings() {
    return this.state.changedSettings || this.props.settings;
  }

  onWordChange(e) {
    this.setState({ wordBeingEntered: e.target.value });
  }

  addWord() {
    const { words, wordBeingEntered } = this.state;

    if (!wordBeingEntered) {
      return;
    }
    if (words.includes(wordBeingEntered)) {
      this.setState({
        error: 'That word is already in your list',
      });
      return;
    }

    this.setState({
      error: '',
      words: words.concat(wordBeingEntered),
      wordBeingEntered: '',
    });
  }

  deleteWord(wordIdx) {
    const { words } = this.state;

    const newWords = words.slice();
    newWords.splice(wordIdx, 1);
    this.setState({
      words: newWords,
    });
  }

  submitWords() {
    const { conn } = this.props;
    conn.send(JSON.stringify({
      action: Constants.Actions.SUBMIT_WORDS,
      body: {
        words: this.state.words,
      },
    }));
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

  addTeam() {
    const { conn } = this.props;
    conn.send(JSON.stringify({
      action: Constants.Actions.ADD_TEAM,
      body: {},
    }));
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
      action: Constants.Actions.MOVE_PLAYER,
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
    if (this.canStartGame || this.playersWithoutSubmittedWords.length) {
      let waitingStr;
      if (this.canStartGame) {
        waitingStr = `Waiting for ${this.roomOwner.name} to start the game`;
      } else if (this.playersWithoutSubmittedWords.length > 1) {
        waitingStr = `Waiting for multiple players to submit their words`;
      } else {
        waitingStr = `Waiting for ${this.playersWithoutSubmittedWords[0].name} to submit their words`;
      }

      return html`
        <div class="room-toolbar">
          ${waitingStr}
        </div>
      `;
    }

    return null;
  }

  get playersWithoutSubmittedWords() {
    const { teams } = this.props;
    let arr = [];
    for (let players of teams) {
      arr = arr.concat(players.filter(p => !p.wordsSubmitted));
    }
    return arr;
  }

  // To start the game, each team needs at least the minimum number of players
  // and everyone needs to have their words submitted.
  get canStartGame() {
    const { teams } = this.props;
    return teams.every(players => {
      return players.length >= Constants.Fishbowl.MIN_PLAYERS_PER_TEAM
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
