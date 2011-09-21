JobsGrid = Ext.extend(ItemsGrid, {
    constructor: function(config) {
        Ext.apply(this, config);

        this.groupField = "State";
        this.toolbarButtons = [
            { text: 'Stop Selected', iconCls:'stop', disabled: true, ref: "../stopButton" }
        ];
        this.gridColumns = [
            { header: "Job ID", width: 25, dataIndex: 'JobId', sortable: false, hidden:true },
            { header: "Label", width: 25, dataIndex: 'Label', sortable: false },
            { header: "Owner", width: 25, dataIndex: 'Owner', sortable: false, hidden:true },
            { header: "Type", width: 25, dataIndex: 'Type', sortable: false, hidden:true },
            { header: "Created", width: 25, dataIndex: 'FirstCreated', sortable: false },
            { header: "Modified", width: 25, dataIndex: 'LastModified', sortable: false },
            { header: "Total", width: 10, sortable: true, dataIndex: 'Total' },
            { header: "Finished", width: 10, sortable: true, dataIndex: 'Finished' },
            { header: "Errored", width: 10, sortable: true, dataIndex: 'Errored' },
            { header: "State", width: 10, sortable: true, dataIndex: 'State', hidden: true },
            { header: "Status", width: 10, sortable: true, dataIndex: 'Status', hidden: false }
        ];
        this.storeColumns = [
            {name: 'JobId'},
            {name: 'Label'},
            {name: 'Owner'},
            {name: 'Type'},
            {name: 'FirstCreated'},
            {name: 'LastModified'},
            {name: 'Total', type: 'int'},
            {name: 'Finished', type: 'int'},
            {name: 'Errored', type: 'int'},
            {name: 'State' },
            {name: 'Status' }
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
                url: row.data.Uri + "/stop",
                method: "post",
                failure: this.showMessage,
                scope: this
            });
        }, this);
        this.selectionModel.clearSelections(true);
    },

    getItemAsArray: function(job) {
        return [
            job.JobId,
            job.Label,
            job.Owner,
            job.Type,
            job.FirstCreated,
            job.LastModified,
            job.Progress.Total,
            job.Progress.Finished,
            job.Progress.Errored,
            job.State,
            job.Status
        ];
    }

});
