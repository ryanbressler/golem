noderesponses=[];
jobresponses=[];

Ext.onReady(function() {
    Ext.TaskMgr.start({ run: pollNodesAndJobs, interval: 1000 });
});

function pollNodesAndJobs(){
    var realTimeSecs = parseInt(Ext.getDom("realtimesecs").value);

    Ext.Ajax.request({
	    method: "GET",
		url: "/jobs/",
		success: function(o) {
			var json = Ext.util.JSON.decode(o.responseText);
			jobresponses.push(json);
			jobhash = {};
			var starti=0;
			if (realTimeSecs < jobresponses.length) {
				starti=jobresponses.length-realTimeSecs;
			}
			for(var i = starti; i<jobresponses.length; i++) {
				for(var j=0;j<jobresponses[i].NumberOfItems;j++){
					var job = jobresponses[i].Items[j];
					if (!jobhash.hasOwnProperty(job.JobId)) {
						jobhash[job.JobId]=[];
					}
					jobhash[job.JobId].push({x:(i-starti),y:(job.Progress.Total-job.Progress.Finished-job.Progress.Errored)});
				}
			}
			jobresponses=jobresponses.slice(starti);
			
			jobdata=fillInData(jobhash,realTimeSecs);
			d3.select("#jobs").html("");
            if (jobdata && jobdata.length) {
                drawChart("#jobs",d3.layout.stack().offset("silhouette")(jobdata),realTimeSecs,"red", "black");
            }
		}
     });

	Ext.Ajax.request({
		method: "GET",
		url: "/nodes/",
		success: function(o) {
			var json = Ext.util.JSON.decode(o.responseText);
			noderesponses.push(json);
			nodehash = {};
			var starti=0;
			if (realTimeSecs < noderesponses.length) {
				starti=noderesponses.length-realTimeSecs;
			}
			for(var i = starti; i<noderesponses.length; i++) {
				for(var j=0;j<noderesponses[i].NumberOfItems;j++){
					node = noderesponses[i].Items[j];
					if (!nodehash.hasOwnProperty(node.NodeId)) {
						nodehash[node.NodeId]=[];
					}
					nodehash[node.NodeId].push({x:(i-starti),y:node.RunningJobs});
				}
			}
			noderesponses=noderesponses.slice(starti);
			var nodedata = fillInData(nodehash,realTimeSecs);
			d3.select("#nodes").html("");
            if (nodedata && nodedata.length) {
                drawChart("#nodes",d3.layout.stack().offset("silhouette")(nodedata),realTimeSecs,"dimgrey", "deepskyblue");
            }
		}});
}

function fillInData(jobhash,realTimeSecs){
	var jobdata = [];

	for (jobid in jobhash) {
		nonzero=false;
        jobArray = jobhash[jobid];

        Ext.each(jobArray, function(jobItem) {
            if (jobItem.y!=0) {
                nonzero=true;
            }
        });

		if (nonzero) {
			minx=jobArray[0].x;
			maxx=jobArray[jobArray.length-1].x;

			for(var i = 1; i<realTimeSecs; i++) {
				if ( i <= minx) {
					jobArray.unshift({x:minx-i,y:0});
				}
				if (i > maxx){
					jobArray.push({x:i,y:0});
				}
			}
			jobdata.push(jobArray);
		}
	}
	return jobdata;
}

function transition() {
	 var timing = Ext.getDom("numberOfSecondsSince").value;
	 Ext.Ajax.request({
		method: "GET",
		url: "/cluster/?numberOfSecondsSince="+timing,
		success: function(o) {
			var json = Ext.util.JSON.decode(o.responseText);
			var jobdata = [];
			var workerdata = [];
			var nodedata = [];
			var available = [];
			var mintime=json.Items[0].SnapshotAt;
			var maxtime=json.Items[json.NumberOfItems-1].SnapshotAt-mintime;

            Ext.each(json.Items, function(jsonItem) {
                var x = jsonItem.SnapshotAt-mintime;
				jobdata.push({x:x,y:jsonItem.JobsRunning});
				workerdata.push({x:x,y:jsonItem.WorkersRunning});
				nodedata.push({x:x,y:jsonItem.WorkersRunning+jsonItem.WorkersAvailable});
				available.push({x:x,y:jsonItem.WorkersAvailable});
			});
            
			drawChart("#chart",d3.layout.stack().offset("silhouette")([jobdata]),maxtime,"dimgrey", "deepskyblue");
			drawChart("#chart",d3.layout.stack().offset("silhouette")([workerdata]),maxtime,"dimgrey", "deepskyblue");
			drawChart("#chart",d3.layout.stack().offset("silhouette")([nodedata]),maxtime,"dimgrey", "deepskyblue");
			drawChart("#chart",d3.layout.stack().offset("silhouette")([workerdata,nodedata]),maxtime,"dimgrey", "deepskyblue");
			drawChart("#chart",d3.layout.stack().offset("silhouette")([workerdata,available]),maxtime,"dimgrey", "deepskyblue");
		}
     });
}

function drawChart(vis,data,mx,color1,color2) {
    if (data && data.length) {
        var color = d3.interpolateRgb(color1,color2);
        var w = 960;
        var h = 200;
        var my = d3.max(data, function(d) {
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
        var vis = d3.select(vis).append("svg:svg").attr("width", w).attr("height", h);
        vis.selectAll("path").data(data).enter().append("svg:path")
            .style("fill", function() { return color(i++/len); }).attr("d", area);
    }
}
