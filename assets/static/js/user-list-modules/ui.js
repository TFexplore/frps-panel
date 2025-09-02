(function (exports, $) {
    'use strict';

    var i18n = {};
    var api = null; // 将在主文件中注入
    var validatorRules = null; // 将在主文件中注入

    function reloadTable() {
        var searchData = layui.form.val('searchForm');
        layui.table.reloadData('tokenTable', {where: searchData}, true);
    }

    function errorMsg(result) {
        var codeMap = {
            1: 'ParamError', 2: 'UserExist', 3: 'UserNotExist', 4: 'ParamError',
            5: 'UserFormatError', 6: 'TokenFormatError', 7: 'CommentInvalid',
            8: 'PortsInvalid', 9: 'DomainsInvalid', 10: 'SubdomainsInvalid'
        };
        var reason = i18n[codeMap[result.code]] || i18n['OtherError'];
        layui.layer.msg(i18n['OperateFailed'] + ',' + reason);
    }

    function updateTableField(obj, field, trim) {
        var newData = {};
        newData[field] = trim;
        obj.update(newData);
    }

    function initServerFilter() {
        var serverSelect = $('select[name="server"]');
        var servers = [];
        layui.table.cache.tokenTable.forEach(function (item) {
            if (item.server && !servers.includes(item.server)) {
                servers.push(item.server);
            }
        });
        servers.sort();

        serverSelect.find('option:not(:first)').remove();
        servers.forEach(function (server) {
            serverSelect.append('<option value="' + server + '">' + server + '</option>');
        });
        layui.form.render('select', 'searchForm');

        layui.form.on('select(serverFilter)', function () {
            reloadTable();
        });
    }

    function addPopup() {
        layui.layer.open({
            type: 1,
            title: i18n['NewUser'],
            area: ['500px'],
            content: layui.laytpl(document.getElementById('addUserTemplate').innerHTML).render(),
            success: function () {
                layui.laydate.render({elem: '#expireDate', type: 'datetime', format: 'yyyy-MM-dd HH:mm:ss'});

                var dashboardListDropdown = document.getElementById('dashboardListDropdown');
                var serverSelect = $('#addUserForm select[name="server"]');
                $(dashboardListDropdown).find('dd a').each(function () {
                    var serverName = $(this).text();
                    serverSelect.append('<option value="' + serverName + '">' + serverName + '</option>');
                });
                layui.form.render('select', 'addUserForm');
            },
            btn: [i18n['Confirm'], i18n['Cancel']],
            btn1: function (index) {
                if (layui.form.validate('#addUserForm')) {
                    var formData = layui.form.val('addUserForm');
                    if (formData.ports != null) {
                        formData.ports = formData.ports.split(',').map(function (p) {
                            return /^\d+$/.test(String(p)) ? parseInt(String(p), 10) : p;
                        });
                    }
                    if (formData.domains != null) formData.domains = formData.domains.split(',');
                    if (formData.subdomains != null) formData.subdomains = formData.subdomains.split(',');
                    api.add(formData, index);
                }
            },
            btn2: function (index) {
                layui.layer.close(index);
            }
        });
    }

    var defaultConfigTemplate = `serverAddr ={ServerIP}
serverPort = {ServerPort}
user = {User}
metadatas.token = {token}

auth.method = "token"
auth.token = "token123456"

[[proxies]]
type = "tcp"
name="每个名称都要不一样，尽量复杂一点别重复"
localIP = "127.0.0.1"
localPort = 10000
remotePort = {Port}
transport.useEncryption = true
transport.useCompression = true`;

    function editConfigTemplatePopup() {
        layui.layer.open({
            type: 1,
            title: i18n['ConfigTemplate'],
            area: ['800px', '600px'],
            content: `<form class="layui-form" style="padding: 20px;">
                <div class="layui-form-item layui-form-text">
                    <label class="layui-form-label">${ i18n['ConfigTemplate'] }</label>
                    <div class="layui-input-block">
                        <textarea name="configTemplate" id="configTemplateEditor" placeholder="${ i18n['PleaseInputConfigTemplate'] }"
                                  autocomplete="off" class="layui-textarea" style="height: 400px;">${ defaultConfigTemplate }</textarea>
                    </div>
                </div>
            </form>`,
            btn: [i18n['Confirm'], i18n['Cancel']],
            btn1: function (index) {
                defaultConfigTemplate = $('#configTemplateEditor').val();
                layui.layer.close(index);
                layui.layer.msg(i18n['OperateSuccess']);
            },
            btn2: function (index) {
                layui.layer.close(index);
            }
        });
    }

    function exportConfig(data) {
        if (data.length === 0) {
            layui.layer.msg(i18n['PleaseCheckAtLeastOneUser']);
            return;
        }

        var serverIP = window.location.hostname; // 获取当前服务器IP
        var serverPort = window.location.port; // 获取当前服务器端口
        var allConfigContents = [];

        data.forEach(function (user) {
            var configContent = defaultConfigTemplate;
            configContent = configContent.replace(/{ServerIP}/g, serverIP);
            configContent = configContent.replace(/{ServerPort}/g, serverPort);
            configContent = configContent.replace(/{User}/g, user.user);
            configContent = configContent.replace(/{token}/g, user.token);
            // 假设每个用户只有一个端口，或者需要用户选择端口
            // 这里简化处理，如果用户有多个端口，只取第一个
            var port = user.ports && user.ports.length > 0 ? user.ports[0] : '未知端口';
            configContent = configContent.replace(/{Port}/g, port);
            allConfigContents.push(`### frpc_${user.user}.ini ###\n${configContent}\n`);
        });

        var finalContent = allConfigContents.join('\n');

        layui.layer.open({
            type: 1,
            title: i18n['ExportConfig'],
            area: ['800px', '600px'],
            content: `<form class="layui-form" style="padding: 20px;">
                <div class="layui-form-item layui-form-text">
                    <label class="layui-form-label">${ i18n['ConfigTemplate'] }</label>
                    <div class="layui-input-block">
                        <textarea id="exportedConfigContent" class="layui-textarea" style="height: 400px; white-space: pre;">${ finalContent }</textarea>
                    </div>
                </div>
            </form>`,
            btn: [i18n['Confirm'], i18n['Cancel']],
            btn1: function (index) {
                var textarea = document.getElementById('exportedConfigContent');
                textarea.select();
                document.execCommand('copy');
                layui.layer.close(index);
                layui.layer.msg(i18n['OperateSuccess'] + ', 配置已复制到剪贴板');
            },
            btn2: function (index) {
                layui.layer.close(index);
            }
        });
    }

    exports.init = function (lang, apiModule, validatorModuleRules) {
        i18n = lang;
        api = apiModule;
        validatorRules = validatorModuleRules;
    };

    exports.reloadTable = reloadTable;
    exports.errorMsg = errorMsg;
    exports.updateTableField = updateTableField;
    exports.initServerFilter = initServerFilter;
    function confirmPopup(messageKey, data, type) {
        var confirmMsg = i18n[messageKey] || i18n['ConfirmOperation'];
        var confirmTitle = i18n['ConfirmTitle'] || 'Confirm';
        var confirmBtn1 = i18n['Confirm'] || 'Yes';
        var confirmBtn2 = i18n['Cancel'] || 'No';

        layui.layer.confirm(confirmMsg, {
            title: confirmTitle,
            btn: [confirmBtn1, confirmBtn2]
        }, function (index) {
            layui.layer.close(index);
            if (type === api.type.Remove) {
                api.operate(api.type.Remove, data);
            } else if (type === api.type.Disable) {
                api.disable(data);
            } else if (type === api.type.Enable) {
                api.enable(data);
            }
        });
    }

    exports.addPopup = addPopup;
    exports.confirmPopup = confirmPopup;
    exports.editConfigTemplatePopup = editConfigTemplatePopup;
    exports.exportConfig = exportConfig;

})(window.UserListUI = window.UserListUI || {}, layui.$);
