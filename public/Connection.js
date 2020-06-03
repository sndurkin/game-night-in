import Constants from './Constants.js';


const protocol = document.location.protocol === 'https:' ? 'wss' : 'ws';

export default class Connection {

  constructor(cfg) {
    Object.assign(this, cfg);
  }

  connect(reconnectAttemptNumber) {
    reconnectAttemptNumber = reconnectAttemptNumber || 0;

    console.log('connect(' + reconnectAttemptNumber + ')');
    return new Promise((resolve, reject) => {
      let params = '';
      if (window.top.SessionStorage[Constants.LocalStorage.ROOM_CODE]) {
        params += '?name=' + window.top.SessionStorage[Constants.LocalStorage.PLAYER_NAME]
          + '&roomCode=' + window.top.SessionStorage[Constants.LocalStorage.ROOM_CODE];
      }
      this.conn = new WebSocket(protocol + '://' + document.location.host + '/ws' + params);
      this.onConnecting && this.onConnecting();

      this.conn.onopen = () => {
        console.log('WebSocket.onopen');
        this.onConnect && this.onConnect();
        resolve();
      };

      this.conn.onerror = () => {
        reject();
      };

      this.conn.onmessage = (e) => {
        const data = JSON.parse(e.data);
        this.onMessage && this.onMessage(data, e);
      };

      this.conn.onclose = (e) => {
        console.log('WebSocket.onclose');
        if (reconnectAttemptNumber < 3) {
          this.connect(reconnectAttemptNumber + 1);
        } else {
          this.onDisconnect && this.onDisconnect();
        }
      };
    });
  }

  send(...args) {
    console.log('WebSocket.send readyState: ' + this.conn.readyState);
    if (this.conn.readyState !== WebSocket.OPEN) {
      this.connect().then(() => {
        this.conn.send(...args);
      });
      return;
    }

    this.conn.send(...args);
  }

  close(...args) {
    this.conn.close(...args);
  }

}
