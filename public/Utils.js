import Constants from './Constants.js';

export default {

  teamStyle: function(teamIdx) {
    return [
      `background: ${Constants.TeamColors[teamIdx]}`,
      `color: ${this.fg(Constants.TeamColors[teamIdx])}`,
    ].join('; ');
  },

  /**
   * Returns the best foreground color that corresponds
   * to the input background color (e.g. '#ffffff').
   */
  fg: function(bg) {
    const { r, g, b } = this.colorToRGB(bg);
    if ((r * 0.299 + g * 0.587 + b * 0.114) > 186) {
      return '#000000';
    }

    return '#ffffff';
  },

  colorToRGB: function(color) {
    if (color[0] === '#') {
      color = color.substring(1);
    }

    return {
      r: parseInt(color.substring(0, 2), 16),
      g: parseInt(color.substring(2, 4), 16),
      b: parseInt(color.substring(4), 16),
    };
  },

  rgbToColor: function(rgb) {
    const { r, g, b } = rgb;
    return '#'
      + this.normalize(r).toString(16).padStart(2, '0')
      + this.normalize(g).toString(16).padStart(2, '0')
      + this.normalize(b).toString(16).padStart(2, '0');
  },

  normalize: function(hex) {
    return Math.min(255, Math.max(0, Math.floor(hex)));
  },

  getRandomNumberInRange(min, max) {
    return Math.random() * (max - min) + min;
  },

};
