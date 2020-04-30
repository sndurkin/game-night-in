import Constants from './Constants.js';

export default {

  teamStyle: function (teamIdx) {
    return [
      `background: ${Constants.TeamColors[teamIdx]}`,
      `color: ${this.fg(Constants.TeamColors[teamIdx])}`,
    ].join('; ');
  },

  /**
   * Returns the best foreground color that corresponds
   * to the input background color (e.g. '#ffffff').
   */
  fg: function (bg) {
    const r = parseInt(bg.substring(1, 3), 16);
    const g = parseInt(bg.substring(3, 5), 16);
    const b = parseInt(bg.substring(5), 16);

    if ((r * 0.299 + g * 0.587 + b * 0.114) > 186) {
      return '#000000';
    }

    return '#ffffff';
  },

};
