const protocol = document.location.protocol === 'https:' ? 'wss' : 'ws';

export default class Connection {

  constructor(cfg) {
    Object.assign(this, cfg);
  }

  connect(reconnectAttemptNumber) {
    reconnectAttemptNumber = reconnectAttemptNumber || 0;

    return new Promise((resolve, reject) => {
      this.conn = new WebSocket(protocol + '://' + document.location.host + '/ws');
      this.onConnecting && this.onConnecting();

      this.conn.onopen = () => {
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
        this.onDisconnect && this.onDisconnect();

        if (reconnectAttemptNumber < 3) {
          this.connect(reconnectAttemptNumber + 1);
        }
      };
    });
  }

  send(...args) {
    if (this.conn.readyState !== WebSocket.OPEN) {
      alert(this.conn.readyState);
      this.connect().then(() => {
        this.conn.send(...args);
      });
    }

    this.conn.send(...args);
  }

  close(...args) {
    this.conn.close(...args);
  }

}
