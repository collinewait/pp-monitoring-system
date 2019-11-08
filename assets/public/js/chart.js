function createChart(parentNode, data) {
  const node = $(
    '<div class="col-sm-12 panel panel-default">' +
      '  <div class="panel-heading row">' +
      '    <div class="col-sm-2 text-right">Name</div>' +
      '    <div class="col-sm-10">' +
      "      <b>" +
      data.name +
      "</b>" +
      "    </div>" +
      '    <div class="col-sm-2 text-right">Serial No</div>' +
      '  <div class="col-sm-10">' +
      "      <b>" +
      data.serialNo +
      "</b>" +
      "    </div>" +
      '    <div class="col-sm-2 text-right">SafeRange</div>' +
      '    <div class="col-sm-10">' +
      "      <b>" +
      "       <span>" +
      data.minSafeValue +
      "       </span> - " +
      "       <span>" +
      +data.maxSafeValue +
      "         </span>" +
      "        <span>" +
      data.unitType +
      "</span>" +
      "      </b>" +
      "    </div>" +
      "  </div>" +
      '  <div class="panel-body ' +
      data.name +
      '"' +
      '    style="width:100%;height:200px;" >' +
      "  </div>" +
      "</div>"
  );

  $(parentNode).append(node);

  const avg = (data.maxSafeValue + data.minSafeValue) / 2;
  const valueRange = data.maxSafeValue - data.minSafeValue;

  $("." + data.name).CanvasJSChart({
    axisY: {
      minimum: avg - valueRange * 0.6,
      maximum: avg + valueRange * 0.6
    },
    axisX: {
      labelFormatter: function(e) {
        return e.value.toTimeString().substr(0, 8);
      }
    },
    data: [
      {
        type: "line",
        dataPoints: []
      },
      {
        type: "rangeArea",
        dataPoints: [],
        color: "green",
        fillOpacity: 0.2
      }
    ]
  });

  let chart = $("." + data.name)[0];
  chart.dataset["minSafeValue"] = data.minSafeValue;
  chart.dataset["maxSafeValue"] = data.maxSafeValue;
}

function updateChart(msg) {
  const node = $("." + msg.Name);
  if (node.length == 0) return;

  const chart = node.CanvasJSChart();
  const pts = chart.options.data[0].dataPoints;
  const range = chart.options.data[1].dataPoints;
  const minSafeValue = parseFloat(node[0].dataset["minSafeValue"]);
  const maxSafeValue = parseFloat(node[0].dataset["maxSafeValue"]);

  pts.push({ x: new Date(msg.Timestamp), y: msg.Value });
  while (pts.length > 20) {
    pts.shift();
  }
  range[0] = { x: pts[0].x, y: [minSafeValue, maxSafeValue] };
  range[1] = {
    x: pts[pts.length - 1].x,
    y: [minSafeValue, maxSafeValue]
  };
  chart.render();
}
