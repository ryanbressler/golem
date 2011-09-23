noderesponses=[];
jobresponses=[];

function fillindata(jobhash,len){
	var jobdata = [];
			
	for (jobid in jobhash) {
		nonzero=false;
		for(var i = 0; i<jobhash[jobid].length; i++) {
			if (jobhash[jobid][i].y!=0){
				nonzero=true;
			}
		}
		if (nonzero) {
			minx=jobhash[jobid][0].x;
			maxx=jobhash[jobid][jobhash[jobid].length-1].x;
			
			for(var i = 1; i<len; i++) {
				if ( i <= minx) {
					jobhash[jobid].unshift({x:minx-i,y:0});
				}
				if (i > maxx){
					jobhash[jobid].push({x:i,y:0});
				}
			}
			jobdata.push(jobhash[jobid]);
		}
	
	}
	return jobdata;
}

function pollnodesandjobs(){
	var timen = 2*parseInt(document.getElementById("realtimesecs").value);
	 Ext.Ajax.request({
		method: "GET",
		url: "/jobs/",
		success: function(o) {
			var json = Ext.util.JSON.decode(o.responseText);
			jobresponses.push(json);
			jobhash = {};
			var starti=0;
			if (timen < jobresponses.length) {
				starti=jobresponses.length-timen;
			}
			for(var i = starti; i<jobresponses.length; i++) {
				for(var j=0;j<jobresponses[i].NumberOfItems;j++){
					var job = jobresponses[i].Items[j];
					if (jobhash.hasOwnProperty(job.JobId)==false) {
						jobhash[job.JobId]=[]
					}
					jobhash[job.JobId].push({x:i-starti,y:(job.Progress.Total-job.Progress.Finished-job.Progress.Errored)})
					
				}
				
			}
			
			jobdata=fillindata(jobhash,timen)
			d3.select("#jobs").html("")
			
			drawchart("#jobs",d3.layout.stack().offset("silhouette")(jobdata),timen,"red", "black");
			
		}});
	Ext.Ajax.request({
		method: "GET",
		url: "/nodes/",
		success: function(o) {
			var json = Ext.util.JSON.decode(o.responseText);
			noderesponses.push(json)
			nodehash = {}
			var starti=0;
			if (timen < noderesponses.length) {
				starti=noderesponses.length-timen;
			}
			for(var i = starti; i<noderesponses.length; i++) {
				for(var j=0;j<noderesponses[i].NumberOfItems;j++){
					node = noderesponses[i].Items[j];
					if (nodehash.hasOwnProperty(node.NodeId)==false) {
						nodehash[node.NodeId]=[]
					}
					nodehash[node.NodeId].push({x:i-starti,y:node.RunningJobs})
					
				}
			}
			var nodedata = fillindata(nodehash,timen);
			d3.select("#nodes").html("")
			drawchart("#nodes",d3.layout.stack().offset("silhouette")(nodedata),timen,"dimgrey", "deepskyblue");
			
		}});

}

function drawchart(vis,data,mx,color1,color2) {
	var color = d3.interpolateRgb(color1,color2),
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
			drawchart("#chart",d3.layout.stack().offset("silhouette")([jobdata]),maxtime,"dimgrey", "deepskyblue");
			drawchart("#chart",d3.layout.stack().offset("silhouette")([workerdata]),maxtime,"dimgrey", "deepskyblue");
			drawchart("#chart",d3.layout.stack().offset("silhouette")([nodedata]),maxtime,"dimgrey", "deepskyblue");
			drawchart("#chart",d3.layout.stack().offset("silhouette")([workerdata,nodedata]),maxtime,"dimgrey", "deepskyblue");
			drawchart("#chart",d3.layout.stack().offset("silhouette")([workerdata,available]),maxtime,"dimgrey", "deepskyblue");
			
		}})
}