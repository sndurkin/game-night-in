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

      wordBeingEntered: '',
      words: document.location.protocol === 'https:' ? [] : getWords(5),

      showMovePlayerModal: false,
      playerToMove: null,
      teamIdxToMoveFrom: null,
    };

    this.onWordChange = this.onWordChange.bind(this);
    this.addWord = this.addWord.bind(this);
    this.submitWords = this.submitWords.bind(this);

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
    const { error } = this.state;

    return html`
      <${ScreenWrapper}
        onBack=${() => this.props.transitionToScreen(Constants.Screens.HOME)}
        ...${this.props}
      >
        <div class="screen">
          ${error && html`
            <span class="label error">${error}</span>
          `}
          ${!this.player.wordsSubmitted ? this.renderSubmitWords() : this.renderTeams()}
        </div>
      <//>
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

  renderTeams() {
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
                onClick=${() => { teamIdxToMoveFrom !== idx && this.movePlayer(idx); }}
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
        this.props.updateStoreData({ teams: data.body.teams });
        break;
      case Constants.Events.UPDATED_GAME:
        this.props.updateStoreData({ game: data.body });
        this.props.transitionToScreen(Constants.Screens.GAME);
        break;
    }
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

  addTeam() {
    const { conn } = this.props;
    conn.send(JSON.stringify({
      action: Constants.Actions.ADD_TEAM,
      body: {},
    }));
  }

  showMovePlayerModal(teamIdx, player) {
    this.setState({
      showMovePlayerModal: true,
      playerToMove: player,
      teamIdxToMoveFrom: teamIdx,
    });
  }

  movePlayer(teamIdxToMoveTo) {
    const { conn } = this.props;
    conn.send(JSON.stringify({
      action: Constants.Actions.MOVE_PLAYER,
      body: {
        playerName: this.state.playerToMove.name,
        fromTeam: this.state.teamIdxToMoveFrom,
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
