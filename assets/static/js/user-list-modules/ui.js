(function (exports, $) {
    'use strict';

    var i18n = {};
    var api = null; // 将在主文件中注入
    var validatorRules = null; // 将在主文件中注入
    var dashboardsData = []; // 用于存储 dashboards 数据
    var defaultConfigTemplate = ''; // 存储配置模板
    
    // 加载配置模板
    function loadConfigTemplate() {
        $.ajax({
            url: '../static/config_template.json',
            type: 'GET',
            dataType: 'json',
            async: false, // 同步请求，确保在使用前加载完成
            success: function(data) {
                if (data && data.template) {
                    defaultConfigTemplate = data.template;
                } else {
                    console.error('配置模板格式错误');
                    // 设置一个默认值，以防加载失败
                    defaultConfigTemplate = '#frpc 配置模板加载失败';
                }
            },
            error: function(xhr, status, error) {
                console.error('加载配置模板失败:', error);
                // 设置一个默认值，以防加载失败
                defaultConfigTemplate = '#frpc 配置模板加载失败';
            }
        });
    }

    // 生成随机字符串的辅助函数
    function generateRandomString(length, characters) {
        var result = '';
        for (var i = 0; i < length; i++) {
            result += characters.charAt(Math.floor(Math.random() * characters.length));
        }
        return result;
    }

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
        // 从 dashboardsData 中获取所有服务器名称
        dashboardsData.forEach(function (dashboard) {
            if (dashboard.name && !servers.includes(dashboard.name)) {
                servers.push(dashboard.name);
            }
        });
        servers.sort();

        // 保留当前选中的值
        var currentSelectedServer = serverSelect.val();

        serverSelect.find('option:not(:first)').remove();
        servers.forEach(function (server) {
            serverSelect.append('<option value="' + server + '">' + server + '</option>');
        });

        // 恢复之前选中的值
        if (currentSelectedServer) {
            serverSelect.val(currentSelectedServer);
        }
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
            success: function (layero, index) { // 捕获 layui.layer.open 的 index
                layui.laydate.render({
                    elem: '#expireDate',
                    type: 'datetime',
                    format: 'yyyy-MM-dd HH:mm:ss',
                    done: function(value, date, endDate){
                        // 在日期选择完成后，自动关闭日期选择器

                    }
                });

                // 生成 6 位随机 user (只包含大小写字母)
                var randomUser = generateRandomString(6, 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz');
                // 生成随机 token (包含大小写字母和数字)
                var randomToken = generateRandomString(32, 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789');

                // 填充随机生成的 user 和 token
                $('#addUserForm input[name="user"]').val(randomUser);
                $('#addUserForm input[name="token"]').val(randomToken);

                var dashboardListDropdown = document.getElementById('dashboardListDropdown');
                var serverSelect = $('#addUserForm select[name="server"]');
                $(dashboardListDropdown).find('dd a').each(function () {
                    var serverName = $(this).text();
                    serverSelect.append('<option value="' + serverName + '">' + serverName + '</option>');
                });
                // 确保至少有一个服务器被选中，如果当前没有选中任何项
                if (!serverSelect.val() && serverSelect.find('option').length > 0) {
                    serverSelect.val(serverSelect.find('option:first').val());
                }
                layui.form.render('select', 'addUserForm');

                var serverMaxPortsMap = {}; // 用于存储所有服务器的最大端口号

                // 在弹窗打开时一次性获取所有服务器的最大端口号
                api.getAllMaxPorts().then(function (maxPortsMap) {
                    serverMaxPortsMap = maxPortsMap;
                    // 再次确保 serverSelect 有选中值，以防异步加载导致时序问题
                    if (!serverSelect.val() && serverSelect.find('option').length > 0) {
                        serverSelect.val(serverSelect.find('option:first').val());
                    }
                    generatePorts(); // 获取数据后立即生成一次端口
                }).catch(function (error) {
                    console.error('获取所有服务器最大端口失败:', error);
                    layui.layer.msg(i18n['OperateFailed'] + ', ' + i18n['NetworkError']);
                });

                // 端口生成逻辑
                var portsCountInput = $('#addUserForm input[name="portsCount"]');
                // 移除 portsInput 在 success 回调中的声明，改为在 generatePorts 内部获取

                function generatePorts() {
                    var addUserForm = layui.$('#addUserForm');
                    var portsInput = addUserForm.find('textarea[name="ports"]'); // 每次调用时重新获取 portsInput
                    var serverName = serverSelect.val();
                    var portsCount = parseInt(portsCountInput.val(), 10);

                    // 检查 serverName 是否在 serverMaxPortsMap 中存在，并且 portsCount 大于 0
                    if (serverMaxPortsMap.hasOwnProperty(serverName) && portsCount > 0) {
                        var maxPort = serverMaxPortsMap[serverName];
                        var startPort = maxPort + 1;
                        var endPort=startPort+portsCount
                        portsInput.val(""+startPort+"-"+endPort);
                    }
                }

                layui.form.on('select(server)', function (data) {
                    generatePorts();
                });

                portsCountInput.on('input', function () {
                    generatePorts();
                });

            },
            btn: [i18n['Confirm'], i18n['Cancel']],
            btn1: function (index) {
                if (layui.form.validate('#addUserForm')) {
                    var formData = layui.form.val('addUserForm');
                    // 端口处理逻辑调整：如果 portsCount 有值，则使用生成的端口，否则使用用户手动输入的端口
                    if (formData.portsCount && parseInt(formData.portsCount, 10) > 0) {
                        // 端口已经通过 generatePorts 函数填充到 formData.ports 中，格式为 "start-end"
                        // 直接将这个字符串作为一个元素放入数组
                        formData.ports = [formData.ports];
                    } else if (formData.ports != null) {
                        // 用户手动输入，可能是逗号分隔的单个端口或端口范围
                        var rawPorts = formData.ports.split(',');
                        formData.ports = rawPorts.map(function (p) {
                            p = p.trim();
                            // 判断是否为纯数字
                            if (/^\d+$/.test(p)) {
                                return parseInt(p, 10);
                            }
                            // 判断是否为 "start-end" 格式
                            if (/^\d+-\d+$/.test(p)) {
                                return p; // 保留为字符串
                            }
                            return p; // 其他情况也保留为字符串
                        });
                    } else {
                        formData.ports = []; // 如果没有端口输入，则为空数组
                    }
                    if (formData.domains != null) formData.domains = formData.domains.split(',');
                    if (formData.subdomains != null) formData.subdomains = formData.subdomains.split(',');
                    // 移除 portsCount 字段，因为它不是后端需要的
                    delete formData.portsCount;
                    api.add(formData, index);
                }
            },
            btn2: function (index) {
                layui.layer.close(index);
            }
        });
    }


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
                var newTemplate = $('#configTemplateEditor').val();
                // 保存到内存中
                defaultConfigTemplate = newTemplate;
                
                // 调用API保存到文件
                api.saveConfigTemplate(newTemplate)
                    .then(function() {
                        layui.layer.close(index);
                    })
                    .catch(function(error) {
                        console.error('保存配置模板失败:', error);
                    });
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

        var serverIP = window.location.hostname; // 默认值
        var serverPort = window.location.port; // 默认值
        var allConfigContents = [];

        data.forEach(function (user) {
            // 根据 user.server 匹配 dashboardsData 中的服务器信息
            var matchedDashboard = dashboardsData.find(function(dashboard) {
                return dashboard.name === user.server;
            });

            if (matchedDashboard) {
                serverIP = matchedDashboard.dashboard_addr || serverIP;
                serverPort = matchedDashboard.dashboard_port || serverPort;
            }

            var configContent = defaultConfigTemplate;
            configContent = configContent.replace(/{ServerIP}/g, serverIP);
            configContent = configContent.replace(/{ServerPort}/g, serverPort);
            configContent = configContent.replace(/{User}/g, user.user);
            configContent = configContent.replace(/{token}/g, user.token);
            // 生成 8 位随机 ProxyName (包含大小写字母和数字)
            var randomProxyName = generateRandomString(8, 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789');
            configContent = configContent.replace(/{ProxyName}/g, randomProxyName);
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

    exports.init = function (lang, apiModule, validatorModuleRules, dashboards) {
        i18n = lang;
        api = apiModule;
        validatorRules = validatorModuleRules;
        dashboardsData = dashboards; // 存储 dashboards 数据
        loadConfigTemplate(); // 加载配置模板
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
                api.operate(api.type.Disable, data);
            } else if (type === api.type.Enable) {
                api.operate(api.type.Enable, data);
            }
        });
    }

    exports.addPopup = addPopup;
    exports.confirmPopup = confirmPopup;
    exports.editConfigTemplatePopup = editConfigTemplatePopup;
    exports.exportConfig = exportConfig;

})(window.UserListUI = window.UserListUI || {}, layui.$);
