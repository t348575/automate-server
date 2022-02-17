// @ts-ignore

document.getElementById("set-button")!.onclick = (ev: MouseEvent) => {
  // @ts-ignore
  let socket = new WebSocket(document.getElementById("ws_url")!.value);

  let input = document.getElementById("input_data")!;

  socket.onopen = function() {
    console.log("[open] Connection established");
  };

  let counter = 0;

  document.getElementById("start-button")!.onclick = (ev: MouseEvent) => {
    console.log("Sending to server");
    socket.send(JSON.stringify({ data: `some data: ${++counter}` }));
  }

  socket.onmessage = function(event) {
    console.log(event.data);
  };
  
  socket.onclose = function(event) {
    if (event.wasClean) {
      console.log(`[close] Connection closed cleanly, code=${event.code} reason=${event.reason}`);
    } else {
      // e.g. server process killed or network down
      // event.code is usually 1006 in this case
      console.log('[close] Connection died');
    }
  };
  
  socket.onerror = function(error: Event) {
    console.log(`[error] ${error}`);
  };
}

