(function() {
  const socket = new WebSocket("ws://localhost:3000/ws");
  let sources = [];

  socket.addEventListener("message", function(e) {
    const msg = JSON.parse(e.data);

    switch (msg.type) {
      case "source":
        if (!sources.some(source => source == msg.data.name)) {
          sources.push(msg.data.name);
          createChart($("#chartContainer"), msg.data);
        }
        break;

      case "reading":
        console.log("msg....", msg);
        updateChart(msg.data);
        break;
    }
  });

  socket.addEventListener("open", function(e) {
    socket.send(
      JSON.stringify({
        type: "discover"
      })
    );
  });
})();
