JobsGrid = Ext.extend(ItemsGrid, {
    constructor: function(config) {
        Ext.apply(this, config);

        var dateRenderer = Ext.util.Format.dateRenderer('m/d/Y H:i:s');

        this.groupField = "State";
        this.toolbarButtons = [
            { text: 'Stop Selected', iconCls:'stop', disabled: true, ref: "../stopButton" }
        ];
        this.gridColumns = [
            { header: "Job ID", width: 25, dataIndex: 'JobId', sortable: false, hidden:true },
            { header: "Label", width: 25, dataIndex: 'Label', sortable: false },
            { header: "Owner", width: 25, dataIndex: 'Owner', sortable: false, hidden:true },
            { header: "Type", width: 25, dataIndex: 'Type', sortable: false, hidden:true },
            { header: "Created", width: 15, dataIndex: 'FirstCreated', sortable: true, renderer: dateRenderer  },
            { header: "Modified", width: 15, dataIndex: 'LastModified', sortable: true, renderer: dateRenderer },
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
            {name: 'FirstCreated', type: "date"},
            {name: 'LastModified', type: "date"},
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

    getFormattedDate: function(v) {
       if (v) {
           var dt = v.replace("  ", " ");
           return Date.parseExact(dt, "ddd MMM d HH:mm:ss PDT yyyy"); 
       }
       return new Date();
    },

    getItemAsArray: function(job) {
        return [
            job.JobId,
            job.Label,
            job.Owner,
            job.Type,
            this.getFormattedDate(job.FirstCreated),
            this.getFormattedDate(job.LastModified),
            job.Progress.Total,
            job.Progress.Finished,
            job.Progress.Errored,
            job.State,
            job.Status
        ];
    }

});
