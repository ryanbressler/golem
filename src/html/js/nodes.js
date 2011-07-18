NodesGrid = Ext.extend(ItemsGrid, {
    constructor: function(config) {
        Ext.apply(this, config);

        this.gridTargetUrl = "/nodes/";

        this.toolbarButtons = [
            { text: 'Restart All', iconCls:'restart', ref: "../restartButton" },
            { text: 'Suspend New Jobs', iconCls:'stop', ref: "../suspendButton" }
        ];
        this.gridColumns = [
            { header: "Node Id", width: 25, dataIndex: 'NodeId', sortable: false },
            { header: "Hostname", width: 10, sortable: true, dataIndex: 'Hostname' },
            { header: "Max Jobs", width: 10, sortable: true, dataIndex: 'MaxJobs' },
            { header: "Running", width: 10, sortable: true, dataIndex: 'Running' }
        ];
        this.storeColumns = [
            {name: 'NodeId'},
            {name: 'Hostname'},
            {name: 'MaxJobs', type: 'int'},
            {name: 'Running' }
        ];
        NodesGrid.superclass.constructor.call(this);
    },

    getItemAsArray: function(node) {
        return [
            node.NodeId,
            node.Hostname,
            node.MaxJobs,
            node.Running
        ];
    }

});
