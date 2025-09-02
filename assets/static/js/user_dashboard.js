layui.use(['element', 'table', 'jquery', 'form'], function () {
    var element = layui.element;
    var table = layui.table;
    var $ = layui.jquery;
    var form = layui.form; // 引入 form 模块
    var dashboardsData = [];

    $.getJSON('/api/user/dashboards', function(res) {
        if (res.data) {
            dashboardsData = res.data;
        }
    });

    var active = {
        tabAdd: function (url, id, name) {
            //新增一个Tab项
            element.tabAdd('tab', {
                title: name,
                content: '<iframe data-frameid="' + id + '" scrolling="auto" frameborder="0" src="' + url + '" style="width:100%;height:99%;"></iframe>',
                id: id //实际使用一般是规定好的id，这里以时间戳为例子
            })
            element.tabChange('tab', id);
        },
        tabDelete: function (id) {
            element.tabDelete("tab", id);//删除
        },
        tabChange: function (id) {
            element.tabChange("tab", id);//切换
        }
    };

    $('.layui-nav-item dd a').on('click', function () {
        var dataid = $(this);
        var url = dataid.attr("data-url");
        var id = dataid.attr("data-id");
        var name = dataid.attr("data-title");

        if ($(".layui-tab-title li[lay-id='" + id + "']").length > 0) {
            active.tabChange(id);
        } else {
            active.tabAdd(url, id, name);
        }
        element.init();
    });

    // 用户信息表格
    table.render({
        elem: '#userInfoTable',
        url: '/api/user/info',
        toolbar: '#userToolbar',
        defaultToolbar: [],
        cols: [[
            {field: 'user', title: '用户'},
            {field: 'token', title: '令牌'},
            {field: 'comment', title: '备注'},
            {field: 'server', title: '服务器名'},
            {field: 'ports', title: '允许端口', templet: function (d) {
                if (d.ports && d.ports.length > 0) {
                    return d.ports.join(', ');
                }
                return '无';
            }},
            {field: 'domains', title: '允许域名', templet: function (d) {
                if (d.domains && d.domains.length > 0) {
                    return d.domains.join(', ');
                }
                return '无';
            }},
            {field: 'subdomains', title: '允许子域名', templet: function (d) {
                if (d.subdomains && d.subdomains.length > 0) {
                    return d.subdomains.join(', ');
                }
                return '无';
            }},
            {field: 'create_date', title: '创建日期'},
            {field: 'expire_date', title: '到期日期', templet: function (d) {
                if (d.expire_date) {
                    return d.expire_date;
                }
                return '未设置';
            }}
        ]],
        page: false,
        done: function (res, curr, count) {
            // 如果没有数据，显示空数据提示
            if (count === 0) {
                this.elem.next('.layui-table-view').find('.layui-table-none').html('<div class="layui-none">暂无数据</div>');
            }
        }
    });

    // 代理列表表格
    table.render({
        elem: '#userProxiesTable',
        url: '/api/user/proxies?proxyType=http', // 初始化时使用默认的http协议
        toolbar: '#proxyToolbar',
        defaultToolbar: [],
        cols: [[
            {field: 'Name', title: '名称'},
            {field: 'RemotePort', title: '远程端口'},
            {field: 'Status', title: '状态', templet: '#proxyStatusTpl'},
            {field: 'Connections', title: '连接数'},
            {field: 'Traffic', title: '流入/流出流量', templet: '#proxyTrafficTpl'},
            {field: 'ClientVersion', title: '客户端版本'},
            {field: 'LastStart', title: '上次启动'},
            {field: 'LastClose', title: '上次关闭'}
        ]],
        page: true,
        limit: 10,
        limits: [10, 20, 30, 40, 50],
        done: function (res, curr, count) {
            // 如果没有数据，显示空数据提示
            if (count === 0) {
                this.elem.next('.layui-table-view').find('.layui-table-none').html('<div class="layui-none">暂无数据</div>');
            }
            // 在表格渲染完成后渲染表单，确保下拉菜单正确显示
            form.render('select', 'proxyTypeForm');
        }
    });

    // 监听协议类型选择
    form.on('select(proxyTypeSelect)', function(data){
        table.reload('userProxiesTable', {
            url: '/api/user/proxies?proxyType=' + data.value
        });
        
        // 使用Layui的form.val方法来正确设置值
        form.val('proxyTypeForm', {
            'proxyType': data.value
        });
        
        // 重新渲染表单以更新显示文本
        form.render('select', 'proxyTypeForm');
    });

    // 监听Tab切换
    element.on('tab(tab)', function (data) {
        var layId = $(this).attr('lay-id');
        if (layId === '1') { // 我的信息
            table.reload('userInfoTable', {
                url: '/api/user/info'
            });
        } else if (layId === '2') { // 我的代理
            // 保留当前选择的协议类型
            var currentProxyType = $('#proxyTypeSelect').val() || 'http';
            table.reload('userProxiesTable', {
                url: '/api/user/proxies?proxyType=' + currentProxyType
            });
        }
    });

    // 监听工具条事件
    table.on('toolbar(userInfoTable)', function (obj) {
        if (obj.event === 'refresh') {
            table.reload('userInfoTable', {
                url: '/api/user/info'
            });
        } else if (obj.event === 'export') {
            var data = table.cache.userInfoTable;
            if (data.length === 0) {
                layui.layer.msg('请先加载用户信息');
                return;
            }
            exportConfig(data);
        }
    });

    table.on('toolbar(userProxiesTable)', function (obj) {
        if (obj.event === 'refresh') {
            table.reload('userProxiesTable', {
                url: '/api/user/proxies?proxyType=' + $('#proxyTypeSelect').val()
            });
        }
    });

    function generateRandomString(length, characters) {
        var result = '';
        for (var i = 0; i < length; i++) {
            result += characters.charAt(Math.floor(Math.random() * characters.length));
        }
        return result;
    }

    function exportConfig(data) {
        $.getJSON('../static/config_template.json', function (config) {
            var defaultConfigTemplate = config.template;
            var allConfigContents = [];

            data.forEach(function (user) {
                var serverIP = window.location.hostname; // 默认值
                var serverPort = window.location.port || (window.location.protocol === 'https:' ? 443 : 80); // 默认值

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
                var randomProxyName = generateRandomString(8, 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789');
                configContent = configContent.replace(/{ProxyName}/g, randomProxyName);
                var port = user.ports && user.ports.length > 0 ? user.ports[0] : '未知端口';
                configContent = configContent.replace(/{Port}/g, port);
                allConfigContents.push(`### frpc_${user.user}.ini ###\n${configContent}\n`);
            });

            var finalContent = allConfigContents.join('\n');

            layui.layer.open({
                type: 1,
                title: '导出配置',
                area: ['800px', '600px'],
                content: `<form class="layui-form" style="padding: 20px;">
                    <div class="layui-form-item layui-form-text">
                        <label class="layui-form-label">配置模板</label>
                        <div class="layui-input-block">
                            <textarea id="exportedConfigContent" class="layui-textarea" style="height: 400px; white-space: pre;">${finalContent}</textarea>
                        </div>
                    </div>
                </form>`,
                btn: ['确认', '取消'],
                btn1: function (index) {
                    var textarea = document.getElementById('exportedConfigContent');
                    textarea.select();
                    document.execCommand('copy');
                    layui.layer.close(index);
                    layui.layer.msg('操作成功, 配置已复制到剪贴板');
                },
                btn2: function (index) {
                    layui.layer.close(index);
                }
            });
        });
    }
});
