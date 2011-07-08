JobsGrid = Ext.extend(ItemsGrid, {
    constructor: function(config) {
        Ext.apply(this, config);

        this.toolbarButtons = [
            { text: 'Stop Selected', iconCls:'stop', disabled: true, ref: "../stopButton" }
        ];
        this.gridColumns = [
            { header: "Sub ID", width: 25, dataIndex: 'SubId', sortable: false },
            { header: "Total Jobs", width: 10, sortable: true, dataIndex: 'TotalJobs' },
            { header: "Finished", width: 10, sortable: true, dataIndex: 'FinishedJobs' },
            { header: "Errored", width: 10, sortable: true, dataIndex: 'ErroredJobs' },
            { header: "Running", width: 10, sortable: true, dataIndex: 'Running' }
        ];
        this.storeColumns = [
            {name: 'SubId'},
            {name: 'TotalJobs', type: 'int'},
            {name: 'FinishedJobs', type: 'int'},
            {name: 'ErroredJobs', type: 'int'},
            {name: 'Running' }
        ];

        JobsGrid.superclass.constructor.call(this);

        this.selectionModel.on("selectionchange", this.enableButtonsOnSelectionChange, this);
        this.grid.stopButton.on("click", this.onStop, this);
    },

    enableButtonsOnSelectionChange: function(sm) {
        if (sm.getCount()) {
            this.grid.stopButton.enable();
        } else {
            this.grid.stopButton.disable();
        }
    },

    onStop: function() {
        this.selectionModel.each(function(row) {
            Ext.Ajax.request({
                url: row.data.uri + "/stop",
                method: "post",
                failure: this.showMessage,
                scope: this
            });
        }, this);
        this.selectionModel.clearSelections(true);
    },

    getItemAsArray: function(job) {
        return [
            job.SubId,
            job.TotalJobs,
            job.FinishedJobs,
            job.ErroredJobs,
            job.Running
        ];
    }

});
