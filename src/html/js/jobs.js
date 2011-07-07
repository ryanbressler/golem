JobsGrid = Ext.extend(Ext.util.Observable, {
    constructor: function(config) {
        Ext.apply(this, config);

        this.jobPointers = [];
        this.gridData = [];
        if (!this.gridPanelConfig) {
            this.gridPanelConfig = {};
        }

        JobsGrid.superclass.constructor.call(this);

        this.on("fetched", this.refreshGrid, this);

        this.initSelectionModel();
        this.renderGrid();
        this.fetchData();
    },

    initSelectionModel: function() {
        if (this.selectionModel != null) {
            return;
        }

        this.selectionModel = new Ext.grid.CheckboxSelectionModel();
        this.selectionModel.on("selectionchange", this.enableButtonsOnSelectionChange, this);
    },

    enableButtonsOnSelectionChange: function(sm) {
        if (sm.getCount()) {
            this.grid.stopButton.enable();
        } else {
            this.grid.stopButton.disable();
        }
    },

    renderGrid: function() {
        this.store = new Ext.data.GroupingStore({
            reader: new Ext.data.ArrayReader({}, this.getStoreColumns()),
            data: this.gridData,
            groupField:'Running',
            groupOnSort: false,
            groupDir: "DESC"
        });

        var defaultConfig = {
            store: this.store,
            columns: this.getGridColumns(),
            view: new Ext.grid.GroupingView({
                forceFit:true,
                startCollapsed: false,
                groupTextTpl: '{text} ({[values.rs.length]} {[values.rs.length > 1 ? "items" : "item"]})'
            }),
            sm: this.selectionModel,
            tbar: [
                { text: 'Refresh', iconCls:'refresh', ref: "../refreshButton" },
                '-',
                { text: 'Stop Selected', iconCls:'stop', disabled: true, ref: "../stopButton" }
            ],
            stripeRows: true,
            columnLines: true,
            frame:true,
            title: "Jobs",
            collapsible: false,
            animCollapse: false,
            iconCls: 'icon-grid'
        };

        this.grid = new Ext.grid.GridPanel(Ext.apply(defaultConfig, this.gridPanelConfig));

        this.grid.stopButton.on("click", function() {
            this.selectionModel.each(function(row) {
                Ext.Ajax.request({
                    url: row.data.uri + "/stop",
                    method: "post",
                    failure: this.showMessage,
                    scope: this
                });
            }, this);
            this.selectionModel.clearSelections(true);
        }, this);

        this.grid.refreshButton.on("click", this.refetch, this);
    },

    refetch: function() {
        this.jobPointers = [];
        this.gridData = [];
        this.refreshGrid();
        this.fetchData();
    },

    refreshGrid: function() {
        this.store.loadData(this.gridData);
    },

    showMessage: function(o) {
        var json = Ext.util.JSON.decode(o.responseText);
        if (json && json.message) {
            Ext.MessageBox.alert('Status', json.message);
        }
    },

    getJobPointer: function(job) {
        var jobPointer = this.jobPointers[job.uri];
        if (!jobPointer) {
            jobPointer = this.gridData.length;
            this.jobPointers[job.uri] = jobPointer;
        }
        return jobPointer;
    },

    getStoreColumns: function() {
        return [
            {name: 'idx'},
            {name: 'uri'},
            {name: 'SubId'},
            {name: 'TotalJobs', type: 'int'},
            {name: 'FinishedJobs', type: 'int'},
            {name: 'ErroredJobs', type: 'int'},
            {name: 'Running' }
        ];
    },

    getGridColumns: function() {
        return [
            this.selectionModel,
            { header: "URI", width: 25, dataIndex: 'uri', sortable: false, hidden: true },
            { header: "Sub ID", width: 25, dataIndex: 'SubId', sortable: false },
            { header: "Total Jobs", width: 10, sortable: true, dataIndex: 'TotalJobs' },
            { header: "Finished", width: 10, sortable: true, dataIndex: 'FinishedJobs' },
            { header: "Errored", width: 10, sortable: true, dataIndex: 'ErroredJobs' },
            { header: "Running", width: 10, sortable: true, dataIndex: 'Running' }
        ];
    },

    fetchData: function() {
        Ext.Ajax.request({
            url: "/jobs/",
            method: "GET",
            success: function(o) {
                var json = Ext.util.JSON.decode(o.responseText);
                if (json && json.items) {
                    Ext.each(json.items, this.addJob, this)
                    this.fireEvent("fetched");
                }
            },
            scope: this
        });
    },

    addJob: function(job) {
        var jobPointer = this.getJobPointer(job);
        this.gridData[jobPointer] = [
            jobPointer,
            job.uri,
            job.SubId,
            job.TotalJobs,
            job.FinishedJobs,
            job.ErroredJobs,
            job.Running
        ];
    }

});
