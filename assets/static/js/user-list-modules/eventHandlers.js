(function (exports, $) {
    'use strict';

    var ui = null; // 将在主文件中注入
    var api = null; // 将在主文件中注入
    var validatorRules = null; // 将在主文件中注入

    function onTableEdit(obj) {
        var field = obj.field, value = obj.value, oldValue = obj.oldValue;
        var before = $.extend(true, {}, obj.data), after = $.extend(true, {}, obj.data);
        var verifyMsg = false;

        if (['token', 'comment', 'ports', 'domains', 'subdomains', 'expire_date'].includes(field)) {
            verifyMsg = validatorRules[field](value, function (trim) {
                ui.updateTableField(obj, field, trim);
            });
            if (verifyMsg) {
                layui.layer.msg(verifyMsg);
                return obj.reedit();
            }
            before[field] = oldValue;
            after[field] = value;
        } else if (field === 'server') {
            before.server = oldValue;
            after.server = value;
        } else {
            return;
        }

        before.ports = before.ports.split(',');
        before.domains = before.domains.split(',');
        before.subdomains = before.subdomains.split(',');
        after.ports = after.ports.split(',');
        after.domains = after.domains.split(',');
        after.subdomains = after.subdomains.split(',');

        api.update(before, after);
    }

    function processTableData(data) {
        data.forEach(function (temp) {
            temp.ports = temp.ports.split(',').map(function (p) {
                return /^\d+$/.test(String(p)) ? parseInt(String(p), 10) : p;
            });
            temp.domains = temp.domains.split(',');
            temp.subdomains = temp.subdomains.split(',');
        });
        return data;
    }

    function onTableToolbar(obj) {
        var data = processTableData(layui.table.checkStatus(obj.config.id).data);
        switch (obj.event) {
            case 'add':
                ui.addPopup();
                break;
            case 'remove':
                ui.confirmPopup('ConfirmRemoveUser', data, api.type.Remove);
                break;
            case 'disable':
                ui.confirmPopup('ConfirmDisableUser', data, api.type.Disable);
                break;
            case 'enable':
                ui.confirmPopup('ConfirmEnableUser', data, api.type.Enable);
                break;
            case 'editConfigTemplate':
                ui.editConfigTemplatePopup();
                break;
        }
    }

    function onTableTool(obj) {
        var data = processTableData([obj.data])[0];
        switch (obj.event) {
            case 'remove':
                ui.confirmPopup('ConfirmRemoveUser', [data], api.type.Remove);
                break;
            case 'exportConfig':
                ui.exportConfig([data]);
                break;
            case 'disable':
                ui.confirmPopup('ConfirmDisableUser', [data], api.type.Disable);
                break;
            case 'enable':
                ui.confirmPopup('ConfirmEnableUser', [data], api.type.Enable);
                break;
        }
    }

    function bindTableEvents() {
        layui.table.on('edit(tokenTable)', onTableEdit);
        layui.table.on('toolbar(tokenTable)', onTableToolbar);
        layui.table.on('tool(tokenTable)', onTableTool);
    }

    function bindDocumentEvents() {
        $(document).on('click.search', '#searchBtn', function () {
            ui.reloadTable();
            return false;
        }).on('click.reset', '#resetBtn', function () {
            $('#searchForm')[0].reset();
            ui.reloadTable();
            return false;
        });
    }

    exports.init = function (uiModule, apiModule, validatorModuleRules) {
        ui = uiModule;
        api = apiModule;
        validatorRules = validatorModuleRules;
    };

    exports.bindTableEvents = bindTableEvents;
    exports.bindDocumentEvents = bindDocumentEvents;

})(window.UserListEventHandlers = window.UserListEventHandlers || {}, layui.$);
