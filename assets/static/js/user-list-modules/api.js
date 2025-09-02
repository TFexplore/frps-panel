(function (exports, $) {
    'use strict';

    var i18n = {};
    var ui = null; // 将在主文件中注入

    var apiType = {Remove: 1, Enable: 2, Disable: 3};

    function add(data, index) {
        var loading = layui.layer.load();
        $.ajax({
            url: '/add', type: 'post', contentType: 'application/json', data: JSON.stringify(data),
            success: function (result) {
                if (result.success) {
                    ui.reloadTable();
                    layui.layer.close(index);
                    layui.layer.msg(i18n['OperateSuccess'], function (i) { layui.layer.close(i); });
                } else {
                    ui.errorMsg(result);
                }
            },
            complete: function () {
                layui.layer.close(loading);
            }
        });
    }

    function update(before, after) {
        before.ports.forEach(function (port, index) {
            if (/^\d+$/.test(String(port))) before.ports[index] = parseInt(String(port));
        });
        after.ports.forEach(function (port, index) {
            if (/^\d+$/.test(String(port)) && typeof port === "string") after.ports[index] = parseInt(String(port));
        });
        var loading = layui.layer.load();
        $.ajax({
            url: '/update', type: 'post', contentType: 'application/json',
            data: JSON.stringify({before: before, after: after}),
            success: function (result) {
                if (result.success) {
                    layui.layer.msg(i18n['OperateSuccess']);
                } else {
                    ui.errorMsg(result);
                }
            },
            complete: function () {
                layui.layer.close(loading);
            }
        });
    }

    function operate(type, data, index) {
        var url, extendMessage = '';
        if (type === apiType.Remove) {
            url = "/remove";
            extendMessage = ', ' + i18n['RemoveUser'] + i18n['TakeTimeMakeEffective'];
        } else if (type === apiType.Disable) {
            url = "/disable";
            extendMessage = ', ' + i18n['RemoveUser'] + i18n['TakeTimeMakeEffective'];
        } else if (type === apiType.Enable) {
            url = "/enable";
        } else {
            layui.layer.msg(i18n['OperateError']);
            return;
        }
        var loading = layui.layer.load();
        $.post({
            url: url, type: 'post', contentType: 'application/json', data: JSON.stringify({users: data}),
            success: function (result) {
                if (result.success) {
                    ui.reloadTable();
                    layui.layer.close(index);
                    layui.layer.msg(i18n['OperateSuccess'] + extendMessage, function (i) { layui.layer.close(i); });
                } else {
                    ui.errorMsg(result);
                }
            },
            complete: function () {
                layui.layer.close(loading);
            }
        });
    }

    function getMaxPort(serverName) {
        return new Promise((resolve, reject) => {
            $.ajax({
                url: '/get_max_port', // 假设后端有一个这样的接口
                type: 'get',
                data: {server: serverName},
                success: function (result) {
                    if (result.success) {
                        resolve(result.maxPort);
                    } else {
                        ui.errorMsg(result);
                        reject(result);
                    }
                },
                error: function (xhr, status, error) {
                    layui.layer.msg(i18n['NetworkError'] + ': ' + error);
                    reject(error);
                }
            });
        });
    }

    exports.init = function (lang, uiModule) {
        i18n = lang;
        ui = uiModule;
    };

    exports.type = apiType;
    exports.add = add;
    exports.update = update;
    exports.operate = operate;
    exports.getMaxPort = getMaxPort;

    function getAllMaxPorts() {
        return new Promise((resolve, reject) => {
            $.ajax({
                url: '/get_all_max_ports', // 假设后端有一个这样的接口
                type: 'get',
                success: function (result) {
                    if (result.success) {
                        resolve(result.maxPortsMap);
                    } else {
                        ui.errorMsg(result);
                        reject(result);
                    }
                },
                error: function (xhr, status, error) {
                    layui.layer.msg(i18n['NetworkError'] + ': ' + error);
                    reject(error);
                }
            });
        });
    }
    exports.getAllMaxPorts = getAllMaxPorts;

})(window.UserListAPI = window.UserListAPI || {}, layui.$);
