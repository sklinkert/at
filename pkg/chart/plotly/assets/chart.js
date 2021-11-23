function unpack(rows, key) {
  return rows.map(function (row) {
    return row[key];
  });
}

function direction2Str(direction) {
  if (direction == 0) {
    return "Long"
  }
  return "Short"
}

document.addEventListener("DOMContentLoaded", function () {
  const candleByDate = {};
  candles.forEach((candle) => {
    candleByDate[candle.Start] = candle;
  });
  const candleStickData = {
    x: unpack(candles, "Start"),
    close: unpack(candles, "Close"),
    open: unpack(candles, "Open"),
    low: unpack(candles, "Low"),
    high: unpack(candles, "High"),
    type: "candlestick",
    xaxis: "x",
    yaxis: "y",
  };

  var layout = {
    margin: {
      r: 10,
      t: 25,
      b: 40,
      l: 60,
    },
    showlegend: false,
    xaxis: {
      autorange: true,
    },
    yaxis: {
      autorange: true,
      type: "linear",
    },
    annotations: orders.map((order) => {
      const annotation = {
        x: order.CandleStart,
        y: candleByDate[order.CandleStart].Low,
        xref: "x",
        yref: "y",
        text: "B",
        hovertext: `Size: ${order.Size.toPrecision(
          4
        )}<br>Note: ${order.Note}
        <br>Direction: `+direction2Str(order.Direction),
        showarrow: true,
        arrowcolor: "green",
        valign: "bottom",
        borderpad: 4,
        arrowhead: 2,
        ax: 0,
        ay: 15,
        font: {
          size: 12,
          color: "green",
        },
      };

      if (order.Size < 0) {
        annotation.font.color = "red";
        annotation.arrowcolor = "red";
        annotation.text = "S";
        annotation.y = candleByDate[order.CandleStart].High;
        annotation.ay = -15;
      }

      return annotation;
    }),
  };

  Plotly.newPlot("graph", [candleStickData], layout);
});
