import { html, Component, render } from 'https://unpkg.com/htm/preact/standalone.module.js';

import Constants from './Constants.js';


export default class RoomScreen extends Component {

  constructor(...args) {
    super(...args);

    this.state = {
      error: '',

      wordBeingEntered: '',
      words: [
        'a', 'b', 'c', 'd', 'e',
      ],

      showMovePlayerModal: false,
      playerToMove: null,
      teamIdxToMoveFrom: null,
    };

    this.onWordChange = this.onWordChange.bind(this);
    this.addWord = this.addWord.bind(this);
    this.submitWords = this.submitWords.bind(this);

    this.addTeam = this.addTeam.bind(this);
    this.movePlayer = this.movePlayer.bind(this);
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
      <div class="screen">
        ${error && html`
          <span class="label error">${error}</span>
        `}
        ${!this.player.wordsSet
          ? this.renderSubmitWords()
          : this.renderTeams()
        }
      </div>
    `;
  }

  renderSubmitWords() {
    const { words, wordBeingEntered } = this.state;

    return html`
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
      <div class="button-bar">
        <button onClick=${this.addWord}>Add</button>
        <div></div>
        <button
          disabled=${words.length < 5}
          onClick=${this.submitWords}
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
    const { isRoomOwner, teams } = this.props;
    const { showMovePlayerModal, teamIdxToMoveFrom } = this.state;

    return html`
      ${(teams || []).map((team, idx) => html`
        <div class="team">
          <div class="team-title">Team ${idx + 1}</div>
          <div class="team-table">
            ${(team || []).length === 0 ? html`
              <div class="empty-list">No players yet!</div>
            ` : (team || []).map(player => html`
              <div class="team-row">
                <div class="player-ready">
                  ${player.wordsSet ? '✔' : ''}
                </div>
                <div class="player-name">
                  ${player.name}
                </div>
                ${isRoomOwner && html`
                  <a onClick=${() => this.showMovePlayerModal(idx, player)}>
                    Move
                  </a>
                `}
              </div>
            `)}
          </table>
        </div>
      `)}
      ${isRoomOwner && html`
        <div><a onClick=${this.addTeam}>Add team</a></div>
      `}
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
            ${(teams || []).map((_, idx) => html`
              <span
                class="button stack"
                disabled=${teamIdxToMoveFrom === idx}
                onClick=${() => {
                  teamIdxToMoveFrom !== idx && this.movePlayer(idx)
                }}
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

    console.log(data.body.teams);
    this.props.updateStoreData({ teams: data.body.teams });
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
        error: 'That word is already in your list'
      });
      return;
    }

    this.setState({
      error: '',
      words: words.concat(wordBeingEntered),
      wordBeingEntered: ''
    });
  }

  deleteWord(wordIdx) {
    const { words } = this.state;

    const newWords = words.slice()
    newWords.splice(wordIdx, 1);
    this.setState({
      words: newWords
    });
  }

  submitWords() {
    const { conn } = this.props;
    conn.send(JSON.stringify({
      action: Constants.Actions.SUBMIT_WORDS,
      body: {
        words: this.state.words
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
