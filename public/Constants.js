export default {
  Screens: {
    HOME: 'home',
    CREATE_GAME: 'create-game',
    JOIN_GAME: 'join-game',
  },
  Actions: {
    START_GAME: 'start-game',
    CREATE_GAME: 'create-game',
    JOIN_GAME: 'join-game',
    REMATCH: 'rematch',
  },
  Events: {
    CREATED_GAME: 'created-game',
    UPDATED_ROOM: 'updated-room',
    UPDATED_GAME: 'updated-game',
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
