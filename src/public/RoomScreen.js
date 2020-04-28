import { html, Component, render } from 'https://unpkg.com/htm/preact/standalone.module.js';

import Constants from './Constants.js';


export default class RoomScreen extends Component {

  constructor(...args) {
    super(...args);

    this.state = {
      error: '',

      wordBeingEntered: '',
      words: [
        '1', '2', '3', '4', '5',
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
      <div class="screen">
        ${error && html`
          <span class="label error">${error}</span>
        `}
        ${!this.player.wordsSubmitted ? this.renderSubmitWords() : this.renderTeams()}
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

    // To start the game, each team needs at least 2 people
    // and everyone needs to have their words submitted.
    const canStartGame = teams.every(players => {
      return players.length >= Constants.Fishbowl.MIN_PLAYERS_PER_TEAM
        && players.every(p => p.wordsSubmitted);
    });

    return html`
      <div class="teams">
        ${teams.map((team, idx) => html`
          <div class="team">
            <div class="team-title">Team ${idx + 1}</div>
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
      ${isRoomOwner && html`
        <div class="button-bar">
          <button onClick=${this.addTeam}>Add team</button>
          <div></div>
          <button
            disabled=${!canStartGame}
            onClick=${this.startGame}>
            Start game
          </button>
        </div>
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
            ${teams.map((_, idx) => html`
              <span
                class="button stack"
                disabled=${teamIdxToMoveFrom === idx}
                onClick=${() => {
        teamIdxToMoveFrom !== idx && this.movePlayer(idx);
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