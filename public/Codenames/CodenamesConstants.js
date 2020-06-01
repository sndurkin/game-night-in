export default {
  Screens: {
    ROOM: 'codenames-room',
    GAME: 'codenames-game',
    GAME_OVER: 'codenames-game-over',
  },
  Actions: {
    MOVE_PLAYER: 'move-player',
    CHANGE_SETTINGS: 'change-settings',

    START_TURN: 'start-turn',
    END_TURN: 'end-turn',
  },
  States: {
    WAITING_ROOM: 'waiting-room',
    TURN_START: 'turn-start',
    TURN_ACTIVE: 'turn-active',
    GAME_OVER: 'game-over',
  },
  Game: {

  },
  PlayerType: {
    INVALID: 0,
    SPYMASTER: 1,
    GUESSER: 2,
  }
};
