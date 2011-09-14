NodesGrid = Ext.extend(ItemsGrid, {
    constructor: function(config) {
        Ext.apply(this, config);

        this.groupField = "Running";
        this.toolbarButtons = [
            { text: 'Restart All', iconCls:'restart', ref: "../restartButton" },
            { text: 'Suspend New Jobs', iconCls:'stop', ref: "../suspendButton" }
        ];
        this.gridColumns = [
            { header: "Node Id", width: 25, dataIndex: 'NodeId', sortable: false },
            { header: "Hostname", width: 10, sortable: true, dataIndex: 'Hostname' },
            { header: "Max Jobs", width: 10, sortable: true, dataIndex: 'MaxJobs' },
            { header: "Running Jobs", width: 10, sortable: true, dataIndex: 'RunningJobs' },
            { header: "Running", width: 10, sortable: true, dataIndex: 'Running', hidden: true }
        ];
        this.storeColumns = [
            {name: 'NodeId'},
            {name: 'Hostname'},
            {name: 'MaxJobs', type: 'int'},
            {name: 'RunningJobs', type: 'int'},
            {name: 'Running' }
        ];
        NodesGrid.superclass.constructor.call(this);
    },

    getItemAsArray: function(node) {
        return [
            node.NodeId,
            node.Hostname,
            node.MaxJobs,
            node.RunningJobs,
            node.Running
        ];
    }

});
