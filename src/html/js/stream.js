function drawchart(vis,data) {
	var color = d3.interpolateRgb("#aad", "#556"),
	 w = 960,
		h = 500,
		mx = m - 1,
		my = d3.max(data, function(d) {
		  return d3.max(d, function(d) {
			return d.y0 + d.y;
		  });
		});
	
	var area = d3.svg.area()
		.x(function(d) { return d.x * w / mx; })
		.y0(function(d) { return h - d.y0 * h / my; })
		.y1(function(d) { return h - (d.y + d.y0) * h / my; });
	

	vis.append("svg:svg")
		.attr("width", w)
		.attr("height", h);
	
	vis.selectAll("path")
		.data(data)
	  .enter().append("svg:path")
		.style("fill", function() { return color(Math.random()); })
		.attr("d", area);
}



function transition() {
	 Ext.Ajax.request({
		method: "GET",
		url: "http://glados:8084/cluster",
		success: function(o) {
			alert("data")
			var json = Ext.util.JSON.decode(o.responseText);
			var jobdata = [];
			var workerdata = [];
			var nodedata = [];
			for (i=0;i<json.NumberOfItems;i++) {
				jobdata.push([json[i].SnapshotAt,json[i].JobsRunning]);
				workerdata.push([json[i].SnapshotAt,json[i].WorkersRunning]);
				nodedata.push([json[i].SnapshotAt,json[i].WorkersRunning+json[i].WorkersAvailable]);
			
			}
			drawchart(d3.select("#chart"),jobdata);
			drawchart(d3.select("#chart"),workerdata);
			drawchart(d3.select("#chart"),nodedata);
			
		}})
}