export default {
  Screens: {
    HOME: 'home',
    CREATE_GAME: 'create-game',
    JOIN_GAME: 'join-game',
    ROOM: 'room',
    GAME: 'game',
    GAME_OVER: 'game-over',
  },
  Actions: {
    CREATE_GAME: 'create-game',
    JOIN_GAME: 'join-game',
    SUBMIT_WORDS: 'submit-words',
    ADD_TEAM: 'add-team',
    MOVE_PLAYER: 'move-player',

    START_GAME: 'start-game',
    START_TURN: 'start-turn',
    CHANGE_CARD: 'change-card',

    REMATCH: 'rematch',
  },
  States: {
    WAITING_ROOM: 'waiting-room',
    TURN_START: 'turn-start',
    TURN_ACTIVE: 'turn-active',
    GAME_OVER: 'game-over',
  },
  Events: {
    CREATED_GAME: 'created-game',
    UPDATED_ROOM: 'updated-room',
    UPDATED_GAME: 'updated-game',
  },
  CardChange: {
    CORRECT: 'correct',
    SKIP: 'skip',
  },
  Fishbowl: {
    MIN_PLAYERS_PER_TEAM: 1,
    MAX_TEAMS: 10,
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
  },
  TeamColors: [
    '#cc0000',    // Red
    '#0000cc',    // Blue
    '#ffff00',    // Yellow
    '#00cc00',    // Green
    '#cc6000',    // Orange
    '#cc00cc',    // Purple
    '#cccccc',    // Gray
    '#663300',    // Brown
    '#ffffff',    // White
    '#000000',    // Black
  ],
  LocalStorage: {
    PLAYER_NAME: 'playerName',
    ROOM_CODE: 'roomCode',
  },
};
