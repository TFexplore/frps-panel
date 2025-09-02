var loadUserList = (function ($) {
    'use strict';

    var i18n = {};

    // 导入模块
    var validator = window.UserListValidator;
    var ui = window.UserListUI;
    var api = window.UserListAPI;
    var eventHandlers = window.UserListEventHandlers;

    /**
     * 主函数
     */
    function loadUserList(lang, title, dashboards) {
        i18n = lang;
        $("#title").text(title);
        $('#content').html(layui.laytpl($('#userListTemplate').html()).render());

        // 初始化模块并注入依赖
        var validatorRules = validator.createRules(i18n);
        layui.form.verify(validatorRules); // 初始化layui表单验证规则

        ui.init(i18n, api, validatorRules, dashboards); // 传递 dashboards
        api.init(i18n, ui);
        eventHandlers.init(ui, api, validatorRules);

        var $section = $('#content > section');
        layui.table.render({
            elem: '#tokenTable',
            height: $section.height() - $('#searchForm').height() + 8,
            text: {none: i18n['EmptyData']},
            url: '/tokens',
            method: 'get',
            where: {},
            dataType: 'json',
            editTrigger: 'dblclick',
            page: pageOptions,
            toolbar: '#userListToolbarTemplate',
            defaultToolbar: false,
            initSort: {
                field: 'create_date',
                type: 'desc'
            },
            cols: [[
                {type: 'checkbox'},
                {field: 'user', title: i18n['User'], width: 120, sort: true},
                {field: 'token', title: i18n['Token'], width: 180, sort: true, edit: true},
                {field: 'server', title: i18n['Server'], width: 100, sort: true, edit: true},
                {field: 'create_date', title: i18n['CreateDate'], width: 150, sort: true},
                {field: 'expire_date', title: i18n['ExpireDate'], width: 150, sort: true, edit: true},
                {field: 'comment', title: i18n['Notes'], sort: true, edit: 'textarea'},
                {field: 'ports', title: i18n['AllowedPorts'], sort: true, edit: 'textarea'},
                {field: 'domains', title: i18n['AllowedDomains'], sort: true, edit: 'textarea'},
                {field: 'subdomains', title: i18n['AllowedSubdomains'], sort: true, edit: 'textarea'},
                {
                    field: 'enable', title: i18n['Status'], width: 100,
                    templet: '<span>{{d.enable? "' + i18n['Enable'] + '":"' + i18n['Disable'] + '"}}</span>',
                    sort: true
                },
                {title: i18n['Operation'], width: 180, toolbar: '#userListOperationTemplate'}
            ]],
            parseData: function (res) {
                if (res.data) {
                    res.data.forEach(function (data) {
                        data.ports = data.ports.join(',');
                        data.domains = data.domains.join(',');
                        data.subdomains = data.subdomains.join(',');
                        data.server = data.server || i18n['NotSet'];
                        data.create_date = data.create_date || i18n['NotSet'];
                        data.expire_date = data.expire_date || i18n['NotSet'];
                    });
                }
                return res;
            },
            done: function () {
                ui.initServerFilter();
            }
        });

        eventHandlers.bindTableEvents();
    }

    eventHandlers.bindDocumentEvents();

    return loadUserList;
})(layui.$);
