export default {
  Screens: {
    ROOM: 'fishbowl-room',
    GAME: 'fishbowl-game',
    GAME_OVER: 'fishbowl-game-over',
  },
  Actions: {
    SUBMIT_WORDS: 'submit-words',
    ADD_TEAM: 'add-team',
    REMOVE_TEAM: 'remove-team',
    MOVE_PLAYER: 'move-player',
    KICK_PLAYER: 'kick-player',
    CHANGE_SETTINGS: 'change-settings',

    START_TURN: 'start-turn',
    CHANGE_CARD: 'change-card',
  },
  States: {
    WAITING_ROOM: 'waiting-room',
    TURN_START: 'turn-start',
    TURN_ACTIVE: 'turn-active',
    GAME_OVER: 'game-over',
  },
  CardChange: {
    CORRECT: 'correct',
    SKIP: 'skip',
  },
  Game: {
    MIN_PLAYERS_PER_TEAM: 1,
    MAX_TEAMS: 10,
  },
  RoundTypes: {
    'describe': {
      title: 'Describe',
      short: 'Describe this:',
      long: [
        'Use anything to describe the word or phrase to your team',
        'except for spelling hints, rhymes, and gestures.',
      ].join(' '),
    },
    'single': {
      title: 'Single word',
      short: 'Describe this in 1 word:',
      long: [
        'You can only use a single word to describe the word or phrase',
        'to your team.',
      ].join(' '),
    },
    'charades': {
      title: 'Charades',
      short: 'Act this out:',
      long: [
        'You can only act out clues for the word or phrase, no talking!',
      ].join(' '),
    },
  },
};
