function drawchart(vis,data,mx) {
	var color = d3.interpolateRgb("#aad", "#556"),
	 w = 960,
		h = 500,
		my = d3.max(data, function(d) {
		  return d3.max(d, function(d) {
			return d.y0 + d.y;
		  });
		});
		
	var area = d3.svg.area()
		.x(function(d) { return d.x * w / mx; })
		.y0(function(d) { return h - d.y0 * h / my; })
		.y1(function(d) { return h - (d.y + d.y0) * h / my; });
	
var vis = d3.select("#chart")
	.append("svg:svg")
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
		url: "/html/cluster",
		success: function(o) {
			var json = Ext.util.JSON.decode(o.responseText);
			var jobdata = [];
			var workerdata = [];
			var nodedata = [];
			var mintime=json.Items[0].SnapshotAt
			var maxtime=json.Items[json.NumberOfItems-1].SnapshotAt-mintime
			for (i=0;i<json.NumberOfItems;i++) {
				jobdata.push({x:json.Items[i].SnapshotAt-mintime,y:json.Items[i].JobsRunning});
				workerdata.push({x:json.Items[i].SnapshotAt-mintime,y:json.Items[i].WorkersRunning});
				nodedata.push({x:json.Items[i].SnapshotAt-mintime,y:json.Items[i].WorkersRunning+json.Items[i].WorkersAvailable});
			
			}
			drawchart("#chart",d3.layout.stack().offset("silhouette")([jobdata]),maxtime);
			drawchart("#chart",d3.layout.stack().offset("silhouette")([workerdata]),maxtime);
			drawchart("#chart",d3.layout.stack().offset("silhouette")([nodedata]),maxtime);
			
		}})
}