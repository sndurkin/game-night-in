window.onload = function () {
  let conn;

  function addPlayer(playerName, isOwner) {
    $('<li></li>').text(playerName).appendTo($('#player-list'));
  }

  $('#join').click(function (e) {
    if (!conn) {
      e.preventDefault();
      return false;
    }
    const name = $('#name').val();
    if (!name) {
      e.preventDefault();
      return false;
    }
    conn.send(JSON.stringify({
      action: 'join',
      body: {
        name: name
      }
    }));
    e.preventDefault();
    return false;
  });

  if (window["WebSocket"]) {
    conn = new WebSocket("ws://" + document.location.host + "/ws");
    conn.onclose = function (evt) {
      var item = document.createElement("div");
      item.innerHTML = "<b>Connection closed.</b>";
      //appendLog(item);
    };
    conn.onmessage = function (evt) {
      const data = JSON.parse(evt.data);
      switch (data.event) {
        case 'player-joined':
          for (let player of data.body.players) {
            addPlayer(player.name, player.isOwner);
          }
          break;
        case 'other-player-joined':
          addPlayer(data.body.name);
          break;
      }
    };
  } else {
    var item = document.createElement("div");
    item.innerHTML = "<b>Your browser does not support WebSockets.</b>";
    appendLog(item);
  }
};