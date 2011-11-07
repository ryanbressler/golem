ItemsGrid = Ext.extend(Ext.util.Observable, {
    constructor: function(config) {
        this.itemPointers = [];
        this.gridData = [];

        Ext.apply(this, config);

        this.selectionModel = new Ext.grid.CheckboxSelectionModel();

        ItemsGrid.superclass.constructor.call(this);

        if (!this.gridPanelConfig) {
            this.gridPanelConfig = {};
        }
        
        this.on("fetched", this.refreshGrid, this);
        this.renderGrid();
        this.fetchData();
    },

    renderGrid: function() {
        var sc = this.joinArrays([{name: 'idx'}, {name: 'Uri'}], this.storeColumns);
        var gc = this.joinArrays([
            this.selectionModel,
            { header: "URI", width: 25, dataIndex: 'Uri', sortable: false, hidden: true }
        ], this.gridColumns);
        var tbar = this.joinArrays([
            { text: 'Refresh', iconCls:'refresh', ref: "../refreshButton" }
        ], this.toolbarButtons);

        this.store = new Ext.data.GroupingStore({
            reader: new Ext.data.ArrayReader({}, sc),
            data: this.gridData,
            groupField: this.groupField,
            groupOnSort: false,
            groupDir: "DESC"
        });
        if (this.multiSortInfo) {
            this.store.multiSort(this.multiSortInfo);
        }
        var defaultConfig = {
            store: this.store,
            columns: gc,
            view: new Ext.grid.GroupingView({
                forceFit:true,
                startCollapsed: false,
                groupTextTpl: '{text} ({[values.rs.length]} {[values.rs.length > 1 ? "items" : "item"]})'
            }),
            sm: this.selectionModel,
            tbar: tbar,
            menuDisabled: true,
            stripeRows: true,
            columnLines: true,
            frame:true,
            title: "Items",
            collapsible: false,
            animCollapse: false,
            iconCls: 'icon-grid'
        };

        this.grid = new Ext.grid.GridPanel(Ext.apply(defaultConfig, this.gridPanelConfig));
        this.grid.refreshButton.on("click", this.refetch, this);
    },

    refetch: function() {
        this.itemPointers = [];
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

    getItemPointer: function(item) {
        var itemPointer = this.itemPointers[item.Uri];
        if (!itemPointer) {
            itemPointer = this.gridData.length;
            this.itemPointers[item.Uri] = itemPointer;
        }
        return itemPointer;
    },

    fetchData: function() {
        Ext.Ajax.request({
            url: this.url,
            method: "GET",
            success: function(o) {
                var json = Ext.util.JSON.decode(o.responseText);
                if (json && json.Items) {
                    Ext.each(json.Items, this.addItem, this)
                    this.fireEvent("fetched");
                }
            },
            scope: this
        });
    },

    addItem: function(item) {
        var pointer = this.getItemPointer(item);
        var itemAsArray = this.getItemAsArray(item);
        this.gridData[pointer] = this.joinArrays([pointer, item.Uri], itemAsArray);
    },

    joinArrays: function(a1, a2) {
        var a3 = [];
        var pushAllFn = function(a) {
            a3.push(a);
        }
        Ext.each(a1, pushAllFn);
        Ext.each(a2, pushAllFn);
        return a3;
    }
});

