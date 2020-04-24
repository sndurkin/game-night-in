export default {
  Screens: {
    HOME: 'home',
    CREATE_ROOM: 'create-room',
    JOIN_ROOM: 'join-room',
    ROOM: 'room',
    GAME: 'game',
  },
  Actions: {
    CREATE_ROOM: 'create-room',
    JOIN_ROOM: 'join-room',
    SUBMIT_WORDS: 'submit-words',
    ADD_TEAM: 'add-team',
    MOVE_PLAYER: 'move-player',

    START_GAME: 'start-game',
    START_TURN: 'start-turn',
  },
  Events: {
    CREATED_ROOM: 'created-room',
    JOINED_ROOM: 'joined-room',
    PLAYER_JOINED_ROOM: 'player-joined-room',
    UPDATED_ROOM: 'updated-room',
    UPDATED_GAME: 'updated-game',
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
