var mainChart = echarts.init(document.getElementById('newChart'), 'dark');
var parseNewData = true;
var chartData = [];

mainChart.setOption({
    tooltip: {
        trigger: 'axis'
    },
    grid: {
        left: '3%',
        right: '4%',
        bottom: '3%',
        containLabel: true
    },
    toolbox: {
     trigger: 'axis',
     left: 'center',
     feature: {
       dataZoom: {
         yAxisIndex: 'none'
       }
     },
    },
    title: {
      subtext: 'live'
    },
    xAxis: {
        type: 'category',
        boundaryGap: false,
        data: []
    },
    yAxis: [
	{
	  splitLine: {
        show: true,
      },
	   name: 'R',
       type: 'value',
       scale: true
    },
	{
     splitLine: {
        show: true,
     },
	   name: 'Freq',
       type: 'value',
      scale: true
    },
	{
     splitLine: {
        show: true,
     },
	   name: 'Angle',
       type: 'value',
      scale: true
    }
	],
    dataZoom: [{
        type: 'inside',
        throttle: 50
    }],
    series: [
        {
            name:'R',
            type:'line',
            yAxisIndex: 0,
            data: []
        },
        {
            name:'Freq',
            type:'line',
            yAxisIndex: 1,
            data: []
        },
        {
            name:'Angle',
            type:'line',
            yAxisIndex: 2,
            data: []
        }
    ]
});

// Detect if space key has been pressed and toggle parseNewData
document.addEventListener('keydown', function(event) {
  if (event.code == 'Space') {
    parseNewData = !parseNewData;
    console.log("parseNewData: " + parseNewData);
  }
});

function listenWsEvents(ws) {
  ws.onmessage = function (chunk) {
    if (!parseNewData) {
      return
    }

   console.log("DATA: %s", chunk.data)
   // if chunk.data is empty, return
    if (chunk.data == "") {
      return
    }

    var parsedChunk = JSON.parse(chunk.data);

    // Clear chartData
    chartData = [];

    for (let i = 0; i < parsedChunk.length; i++) {
      chartData.push(parsedChunk[i]);
    }


    mainChart.setOption({
        xAxis: {
            data: chartData.map(function (item) {
              return item[0];
            })
        },
        series: [
          {
            name: 'last',
            data: chartData.map(function (item) {
              return item[1];
            })
          }
        ]
    });
  }

  ws.onclose = function(e) {
    console.log(e);
    console.log("Reconnecting...");
    setTimeout(function(){initWs()}, 5000)
    return
  };
}

function initWs() {
  console.log("Connecting...")
  // TODO: fill in ip using template
  let ws = new WebSocket("ws://" + location.host + "/ws");
  ws.onopen = function () {
    console.log('Connected')
    listenWsEvents(ws)
  }
}


initWs()
