export default {
  Screens: {
    HOME: 'home',
    CREATE_GAME: 'create-game',
    JOIN_GAME: 'join-game',
    ROOM: 'room',
    GAME: 'game',
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
    Rounds: [
      'Describe',
      'Single word',
      'Charades',
    ],
  },
};
