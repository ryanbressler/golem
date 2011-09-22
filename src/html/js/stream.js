noderesponses=[];
jobresponses=[];

function fillindata(jobhash,len){
	var jobdata = [];
			
	for (jobid in jobhash) {
		minx=jobhash[jobid][0].x;
		maxx=jobhash[jobid].length-1;
		for(var i = 0; i<len; i++) {
			if ( i < minx) {
				jobhash[jobid].unshift({x:i,y:0});
			}
			if (i > maxx){
				jobhash[jobid].push({x:i,y:0});
			}
		}
		jobdata.push(jobhash[jobid]);
	
	}
	return jobdata;
}

function pollnodesandjobs(){
	 Ext.Ajax.request({
		method: "GET",
		url: "/jobs/",
		success: function(o) {
			var json = Ext.util.JSON.decode(o.responseText);
			jobresponses.push(json)
			jobhash = {}
			for(var i = 0; i<jobresponses.length; i++) {
				for(var j=0;j<jobresponses[i].NumberOfItems;j++){
					job = jobresponses[i].Items[j];
					if (jobhash.hasOwnProperty(job.JobId)==false) {
						jobhash[job.JobId]=[]
					}
					jobhash[job.JobId].push({x:i,y:(job.Progress.Total-job.Progress.Finished-job.Progress.Errored)})
					
				}
				
			}
			
			jobdata=fillindata(jobhash,jobresponses.length)
			d3.select("#jobs").html("")
			
			drawchart("#jobs",d3.layout.stack().offset("silhouette")(jobdata),jobresponses.length);
			
		}});
	Ext.Ajax.request({
		method: "GET",
		url: "/nodes/",
		success: function(o) {
			var json = Ext.util.JSON.decode(o.responseText);
			noderesponses.push(json)
			nodehash = {}
			for(var i = 0; i<noderesponses.length; i++) {
				for(var j=0;j<noderesponses[i].NumberOfItems;j++){
					node = noderesponses[i].Items[j];
					if (nodehash.hasOwnProperty(node.NodeId)==false) {
						nodehash[node.NodeId]=[]
					}
					nodehash[node.NodeId].push({x:i,y:node.RunningJobs})
					
				}
			}
			var nodedata = [];
			for (nodeid in nodehash) {
				nodedata.push(nodehash[nodeid]);
			
			}
			d3.select("#nodes").html("")
			drawchart("#nodes",d3.layout.stack().offset("silhouette")(nodedata),noderesponses.length);
			
		}});

}

function drawchart(vis,data,mx) {
	var color = d3.interpolateRgb("black", "green"),
	 w = 960,
		h = 200,
		my = d3.max(data, function(d) {
		  return d3.max(d, function(d) {
			return d.y0 + d.y;
		  });
		});
		
	var area = d3.svg.area()
		.x(function(d) { return d.x * w / mx; })
		.y0(function(d) { return h - d.y0 * h / my; })
		.y1(function(d) { return h - (d.y + d.y0) * h / my; });

var i = 0;
var len = data.length;
var vis = d3.select(vis)
	.append("svg:svg")
		.attr("width", w)
		.attr("height", h);
	
	vis.selectAll("path")
		.data(data)
	  .enter().append("svg:path")
		.style("fill", function() { return color(i++/len); })
		.attr("d", area);
}



function transition() {
	 var timen = document.getElementById("numberOfSecondsSince").value;
     var url = "/cluster/?numberOfSecondsSince="+timen
	 Ext.Ajax.request({
		method: "GET",
		url: url,
		success: function(o) {
			var json = Ext.util.JSON.decode(o.responseText);
			var jobdata = [];
			var workerdata = [];
			var nodedata = [];
			var available = [];
			var mintime=json.Items[0].SnapshotAt
			var maxtime=json.Items[json.NumberOfItems-1].SnapshotAt-mintime
			for (i=0;i<json.NumberOfItems;i++) {
				jobdata.push({x:json.Items[i].SnapshotAt-mintime,y:json.Items[i].JobsRunning});
				workerdata.push({x:json.Items[i].SnapshotAt-mintime,y:json.Items[i].WorkersRunning});
				nodedata.push({x:json.Items[i].SnapshotAt-mintime,y:json.Items[i].WorkersRunning+json.Items[i].WorkersAvailable});
				available.push({x:json.Items[i].SnapshotAt-mintime,y:json.Items[i].WorkersAvailable});
			
			}
			drawchart("#chart",d3.layout.stack().offset("silhouette")([jobdata]),maxtime);
			drawchart("#chart",d3.layout.stack().offset("silhouette")([workerdata]),maxtime);
			drawchart("#chart",d3.layout.stack().offset("silhouette")([nodedata]),maxtime);
			drawchart("#chart",d3.layout.stack().offset("silhouette")([workerdata,nodedata]),maxtime);
			drawchart("#chart",d3.layout.stack().offset("silhouette")([workerdata,available]),maxtime);
			
		}})
}