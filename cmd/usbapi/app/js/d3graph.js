//START LOCAL FUNCTINO
(function () {

//Get Screen width and height from jquery
var SCREEN_WIDTH = $(document).width();
var SCREEN_HEIGHT = $(document).height();

// define dimensions of graph
var m = [32, 32, 0, 0]; // margins top right bottom left
var yAxisWidth = 48;
var xAxisHeight = 48;
var isSmallDevice = (SCREEN_WIDTH < 860 || SCREEN_HEIGHT < 400);
if (isSmallDevice) {
  yAxisWidth = 48;
  xAxisHeight = 24;
  m = [16, 32, 0, 0];
}
var w = SCREEN_WIDTH - yAxisWidth - m[1] - m[3]; // width
var h = SCREEN_HEIGHT- xAxisHeight - m[0] - m[2]; // height
// create a simple data array that we'll plot with a line (this array represents only the Y values, X will just be the index location)
//var data = [3, 6, 2, 7, 5, 2, 0, 3, 8, 9, 2, 5, 9, 3, 6, 3, 6, 2, 7, 5, 2, 1, 3, 8, 9, 2, 5, 9, 2, 7];
$.getJSON("http://192.168.59.103:8086/db/test1/series?u=root&p=root&q=select%20*%20from%20/test.*/%20limit%201000", function (record) {

var points = record || [];
points = points && record[0].points || [];
var points_length = points.length;
var date = new Date(points[0][0]);
$('#clockdiv').html(date.getHours() + ":" + date.getMinutes() + ":" + date.getSeconds());
//var data = points;
// X scale will fit all values from data[] within pixels 0-w
//var x = d3.time.scale().domain([points[0][0], points[points_length-1][0]]).range([0, w]);
var x = d3.scale.linear().domain([0, points_length]).range([0, w]);
// Y scale will fit values from 0-10 within pixels h-0 (Note the inverted domain for the y-scale: bigger is up!)
var y = d3.scale.linear().domain([900, 1300]).range([h, 0]);
// automatically determining max range can work something like this
// var y = d3.scale.linear().domain([0, d3.max(data)]).range([h, 0]);
// create a line function that can convert data[] into x and y points
var line = d3.svg.line()
// assign the X function to plot our line as we wish
.x(function(d,i) {
// verbose logging to show what's actually being done
//console.log('Plotting X value for data point: ' + d[0]);
// return the X coordinate where we want to plot this datapoint
return x(i);
})
.y(function(d) {
// verbose logging to show what's actually being done
//console.log('Plotting Y value for data point: ' + d[2]);
// return the Y coordinate where we want to plot this datapoint
return y(d[2]);
});
//.interpolate("basis");
 
// Add an SVG element with the desired dimensions and margin.
var graph = d3.select("#mygraph").append("svg")
            .attr("width", SCREEN_WIDTH)
            .attr("height", SCREEN_HEIGHT)
            .attr("preserveAspectRatio", "xMinYMin slice")
            .append("g")
            .attr("transform", "translate(" + (yAxisWidth+m[3]) + "," + m[0] + ")");

graph.append("defs").append("clipPath")
      .attr("id", "clip")
      .append("rect")
      .attr("width", w)
      .attr("height", h);

 
// create xAxis
var xAxis = d3.svg.axis().scale(x).tickFormat(function(d) {
  var date = new Date(d); 
  if (isSmallDevice)
    return ":" + date.getSeconds();
  else
    return date.getHours() + ":" + date.getMinutes() + ":" + date.getSeconds() + "." + date.getMilliseconds();
});
// Add the x-axis.
var xAxisLine = graph.append("g")
    .attr("class", "x axis")
    .attr("transform", "translate(0," + h + ")")
    .attr("stroke", "#666")
    .call(xAxis);
     
 
// create left yAxis
var yAxis = d3.svg.axis().scale(y).ticks(10).orient("left");
// Add the y-axis to the left
var yAxisLine = graph.append("g")
    .attr("class", "y axis")
    .attr("stroke", "#666")
    .call(yAxis);

var path = graph.append("g")
    .attr("clip-path", "url(#clip)")
  .append("path")
    .datum(points)
    .attr("class", "line")
    .attr("stroke", "#F3Ef21")
    .attr("style", "stroke-width:2px;fill: transparent");
    
var transition = d3.select({}).transition()
    .duration(10)
    .ease("linear");
var count = 0;

// TICK FUNCTION
// callback function called by transition 
// see time in duration parameters
(function tick() {
  transition = transition.each(function() {
    //x.domain([now - (n - 2) * duration, now - duration]);
    //y.domain([0, d3.max(points, function(d) {return d[2];})]);

    // push the accumulated count onto the back, and reset the count
    points.push([Date.now(), 0, (100 * Math.sin(count/Math.PI/10)) + 1000 + Math.random() * 50]);
    count++;

    // redraw the line
    graph.select(".line")
        .attr("d", line)
        .attr("transform", null);

    // slide the x-axis left
    //xAxisLine.call(xAxis);

    // slide the line left
    path.transition()
        .attr("transform", "translate(-1)");

    // pop the old data point off the front
    points.shift();
  }).transition().each("start", tick);
})();
// End Tick function

});
// End GetJSON function

}()); //END LOCAL FUNCTION