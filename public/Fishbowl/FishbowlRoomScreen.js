import { html, Component, render } from 'https://unpkg.com/htm/preact/standalone.module.js';

import ScreenWrapper from '../ScreenWrapper.js';
import Utils from '../Utils.js';
import Constants from '../Constants.js';

import FishbowlUtils from './FishbowlUtils.js';
import FishbowlConstants from './FishbowlConstants.js';


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

export default class FishbowlRoomScreen extends Component {

  constructor(...args) {
    super(...args);

    this.state = {
      error: '',

      changedSettings: null,

      enteringWords: false,
      wordBeingEntered: '',
      words: document.location.protocol === 'https:' ? [] : getWords(5),

      showMovePlayerModal: false,
      playerToMove: null,
      teamIdxToMoveFrom: null,

      showKickPlayerModal: false,
      playerToKick: null,
    };

    this.onWordChange = this.onWordChange.bind(this);
    this.addWord = this.addWord.bind(this);
    this.enterWords = this.enterWords.bind(this);
    this.submitWords = this.submitWords.bind(this);
    this.openChangeSettings = this.openChangeSettings.bind(this);
    this.closeChangeSettings = this.closeChangeSettings.bind(this);
    this.renderRoundTableRow = this.renderRoundTableRow.bind(this);
    this.onNumWordsRequiredChange = this.onNumWordsRequiredChange.bind(this);
    this.onTimerChange = this.onTimerChange.bind(this);
    this.selectRound = this.selectRound.bind(this);
    this.removeRound = this.removeRound.bind(this);
    this.saveSettings = this.saveSettings.bind(this);

    this.addTeam = this.addTeam.bind(this);
    this.removeTeam = this.removeTeam.bind(this);
    this.movePlayer = this.movePlayer.bind(this);
    this.kickPlayer = this.kickPlayer.bind(this);
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
    const { changedSettings, enteringWords, error } = this.state;

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
          ${enteringWords ? this.renderEnterWords() : this.renderRoom()}
        </div>
      <//>
    `;
  }

  renderChangeSettingsDialog() {
    const { error } = this.state;
    const { numWordsRequired, timerLength, rounds } = this.settings;

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
          <h3>Words per player required</h3>
          <div style="display: flex; align-items: center">
            <input
              type="number"
              style="width: 4em"
              value=${numWordsRequired}
              onChange=${this.onNumWordsRequiredChange}
            />
            <div style="margin-left: 1em">words</div>
          </div>
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
            ${Object.entries(FishbowlConstants.RoundTypes).map(e => html`
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

  renderEnterWords() {
    const { words, wordBeingEntered } = this.state;
    const { numWordsRequired } = this.settings;

    return html`
      <form>
        <label>
          Word
          <input
            ref=${r => this.inputRef = r}
            type="text"
            maxlength="60"
            value="${wordBeingEntered}"
            placeholder="Enter a word or phrase"
            onInput=${this.onWordChange} />
        </label>
      </form>
      <div class="button-bar">
        <button onClick=${this.addWord}>Add</button>
        <div></div>
        <button
          disabled=${words.length < numWordsRequired}
          onClick=${this.submitWords}
          style="font-size: 0.9em"
        >
          Submit words
        </button>
      </div>
      <div class="word-list">
        <h4 class="word-list-title">
          Words (${words.length} out of ${numWordsRequired}):
        </h4>
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
    const {
      showMovePlayerModal, teamIdxToMoveFrom,
      showKickPlayerModal, playerToKick,
    } = this.state;
    const teams = this.props.teams || [];

    return html`
      ${this.renderSettingsSummary()}
      <div class="teams">
        ${teams.map((team, idx) => html`
          <div class="team">
            <div class="team-header" style=${Utils.teamStyle(idx)}>
              <div class="team-title">Team ${idx + 1}</div>
              ${isRoomOwner && idx >= 2 ? html`
                <button
                  class="team-remove pseudo"
                  onClick=${() => this.removeTeam(idx)}
                >
                  ✖
                </button>
              ` : null}
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
                    ${!player.isRoomOwner ? html`
                      <a
                        role="link"
                        onClick=${() => this.showKickPlayerModal(player)}
                      >
                       Kick
                      </a>
                      <span class="inline-actions-separator"> • </span>
                    ` : null}
                    <a
                      role="link"
                      onClick=${() => this.showMovePlayerModal(idx, player)}
                    >
                      Move
                    </a>
                  ` : player.isRoomOwner ? html`
                    <div>Game creator</div>
                  ` : ''}
                </div>
              `)}
            </table>
          </div>
        `)}
        ${isRoomOwner ? html`
          <button
            onClick=${this.addTeam}
            disabled=${teams.length >= FishbowlConstants.Game.MAX_TEAMS}
          >
            Add team
          </button>
        ` : null}
      </div>
      ${this.renderButtonBar()}
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
      <div class="modal">
        <input
          id="kick-player-modal"
          type="checkbox"
          checked=${showKickPlayerModal ? 'checked' : ''} />
        <label for="kick-player-modal" class="overlay"></label>
        <article>
          <header>
            <h3>Kick ${playerToKick && playerToKick.name}?</h3>
            <label for="kick-player-modal" class="close">✖</label>
          </header>
          <section class="content">
            <div class="button-bar">
              <button class="pseudo" onClick=${this.hideKickPlayerModal}>
                Cancel
              </button>
              <div />
              <button onClick=${this.kickPlayer}>
                Kick
              </button>
            </div>
          </section>
          <footer></footer>
        </article>
      </div>
    `;
  }

  renderSettingsSummary() {
    const { isRoomOwner } = this.props;
    const { numWordsRequired, timerLength, rounds } = this.settings;

    const wordsStr = `${numWordsRequired} words per player`;
    const timerStr = `${timerLength}s timer`;
    const roundsStr = `${rounds.length} round${rounds.length !== 1 ? 's' : ''}`;

    return html`
      <div class="settings">
        <div class="settings-header">
          <div class="settings-title">Settings</div>
          ${isRoomOwner ? html`
            <a role="link" onClick=${this.openChangeSettings}>Change</a>
          ` : null}
        </div>
        <div><span>${wordsStr}, ${timerStr}, ${roundsStr}</span></div>
      </div>
    `;
  }

  renderButtonBar() {
    const { isRoomOwner, wordsSubmitted } = this.props;

    return html`
      <div>
        <div>${this.waitingMessage}</div>
        <div class="button-bar">
          <button
            onClick=${this.enterWords}
            class=${wordsSubmitted ? 'secondary' : ''}
          >
            ${!wordsSubmitted ? 'Enter words' : 'Edit words'}
          </button>
          <div></div>
          ${isRoomOwner ? html`
            <button
              disabled=${!this.canStartGame}
              onClick=${this.startGame}
            >
              Start game
            </button>
          ` : null}
        </div>
      </div>
    `;
  }

  handleMessage(data, e) {
    if (data.error) {
      if (data.errorIsFatal) {
        window.location.reload();
        return;
      }

      this.setState({ error: data.error });
    }

    switch (data.event) {
      case Constants.Events.UPDATED_ROOM:
        const playerProps = {};

        if (data.body.teams) {
          const player = FishbowlUtils.getPlayerByName(data.body.teams,
            this.props.name);
          Object.assign(playerProps, player);
        }

        this.props.updateStoreData({
          ...playerProps,
          teams: data.body.teams,
          settings: data.body.settings,
        });
        break;
      case Constants.Events.UPDATED_GAME:
        this.props.updateStoreData({ game: data.body });
        this.props.transitionToScreen(FishbowlConstants.Screens.GAME);
        break;
    }
  }

  onNumWordsRequiredChange(e) {
    const newSettings = JSON.parse(JSON.stringify(this.settings));
    try {
      newSettings.numWordsRequired = parseInt(e.target.value, 10);

      this.setState({
        changedSettings: newSettings,
      });
    } catch (e) {
      // Ignore
    }
  }

  onTimerChange(e) {
    const newSettings = JSON.parse(JSON.stringify(this.settings));
    try {
      newSettings.timerLength = parseInt(e.target.value, 10);

      this.setState({
        changedSettings: newSettings,
      });
    } catch (e) {
      // Ignore
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
      action: FishbowlConstants.Actions.CHANGE_SETTINGS,
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
    const { words } = this.state;

    const wordBeingEntered = this.state.wordBeingEntered.trim();
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
      words: [wordBeingEntered].concat(words),
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

  enterWords() {
    this.setState({
      enteringWords: true,
    });
  }

  submitWords() {
    this.setState({
      enteringWords: false,
    });

    const { conn } = this.props;
    conn.send(JSON.stringify({
      action: FishbowlConstants.Actions.SUBMIT_WORDS,
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
      action: FishbowlConstants.Actions.ADD_TEAM,
      body: {},
    }));
  }

  removeTeam(idx) {
    const { conn } = this.props;
    conn.send(JSON.stringify({
      action: FishbowlConstants.Actions.REMOVE_TEAM,
      body: {
        team: idx,
      },
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

  showKickPlayerModal(player) {
    this.setState({
      showKickPlayerModal: true,
      playerToKick: player,
    });
  }

  hideKickPlayerModal() {
    this.setState({
      showKickPlayerModal: false,
      playerToKick: null,
    });
  }

  movePlayer({ name, teamIdxToMoveFrom, teamIdxToMoveTo }) {
    const { conn } = this.props;
    conn.send(JSON.stringify({
      action: FishbowlConstants.Actions.MOVE_PLAYER,
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

  kickPlayer() {
    const { conn } = this.props;
    conn.send(JSON.stringify({
      action: FishbowlConstants.Actions.KICK_PLAYER,
      body: {
        playerName: this.state.playerToKick.name,
      },
    }));

    this.setState({
      showKickPlayerModal: false,
      playerToKick: null,
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
      return players.length >= FishbowlConstants.Game.MIN_PLAYERS_PER_TEAM
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
