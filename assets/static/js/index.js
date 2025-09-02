var httpPort, httpsPort, pageOptions;
(function ($) {
    $(function () {
        function init() {
            var langLoading = layui.layer.load()
            $.getJSON('/lang.json').done(function (lang) {
                pageOptions = {
                    limitTemplet: function (item) {
                        return item + lang['PerPage'];
                    },
                    skipText: [lang['Goto'], '', lang['Confirm']],
                    countText: [lang['Total'], lang['Items']]
                };

                $.ajaxSetup({
                    error: function (xhr,) {
                        if (xhr.status === 401) {
                            layui.layer.msg(lang['TokenInvalid'], function () {
                                window.location.reload();
                            });
                        }
                    },
                });

                layui.element.on('nav(leftNav)', function (elem) {
                    var id = elem.attr('id');
                    var title = elem.text();
                    if (id === 'serverInfo') {
                        loadServerInfo(lang, title.trim());
                    } else if (id === 'userList') {
                        loadUserList(lang, title.trim(), dashboardsData); // 传递 dashboardsData
                    } else if (elem.closest('.layui-nav-item').attr('id') === 'proxyList') {
                        if (id != null && id.trim() !== '') {
                            var suffix = elem.closest('.layui-nav-item').children('a').text().trim();
                            loadProxyInfo(lang, title + " " + suffix, id);
                        }
                    }
                });

                var currentIndex = -1;
                layui.element.on('nav(dashboardList)', function (elem) {
                    var newIndex = parseInt(elem.data('index'));
                    if (newIndex !== currentIndex) {
                        switchDashboard(newIndex, lang);
                    }
                });

                function updateCurrentIndex(index) {
                    currentIndex = index;
                }

                loadDashboards(lang, updateCurrentIndex); // Load dashboards after language is loaded
            }).always(function () {
                layui.layer.close(langLoading);
            });
        }

        var dashboardsData = []; // 用于存储 dashboards 数据

        function loadDashboards(lang, updateCurrentIndexCallback) {
            $.getJSON('/dashboards').done(function (res) {
                if (res.code === 0) {
                    dashboardsData = res.data; // 存储 dashboards 数据
                    var currentIndex = res.current_index;
                    var dropdown = $('#dashboardListDropdown');
                    dropdown.empty(); // Clear existing items

                    if (dashboardsData && dashboardsData.length > 0) {
                        $('#currentDashboardName').text(dashboardsData[currentIndex].name);
                        $.each(dashboardsData, function (index, dashboard) {
                            var activeClass = (index === currentIndex) ? 'layui-this' : '';
                            dropdown.append('<dd class="' + activeClass + '"><a href="javascript:;" data-index="' + index + '">' + dashboard.name + '</a></dd>');
                        });
                        updateCurrentIndexCallback(currentIndex);
                        layui.element.render('nav', 'dashboardList');
                    } else {
                        $('#currentDashboardName').text(lang['NotSet']);
                    }
                } else {
                    layui.layer.msg(lang['OperateFailed'] + ': ' + res.msg);
                }
            }).fail(function () {
                layui.layer.msg(lang['OperateFailed'] + ': ' + lang['OtherError']);
            }).always(function () {
                // 在 dashboards 加载完成后，再触发左侧导航的点击事件
                $('#leftNav .layui-this > a').click();
            });
        }

        function switchDashboard(index, lang) {
            var loading = layui.layer.load();
            $.ajax({
                url: '/switch_dashboard',
                type: 'POST',
                contentType: 'application/json',
                data: JSON.stringify({ index: index }),
                success: function (res) {
                    if (res.success) {
                        layui.layer.msg(res.message, { icon: 1, time: 1000 }, function () {
                            window.location.reload(); // Reload page to reflect new dashboard data
                        });
                    } else {
                        layui.layer.msg(res.message, { icon: 2 });
                    }
                },
                error: function () {
                    layui.layer.msg(lang['OperateFailed'] + ': ' + lang['OtherError'], { icon: 2 });
                },
                complete: function () {
                    layui.layer.close(loading);
                }
            });
        }

        function logout() {
            $.get("/logout", function (result) {
                window.location.reload();
            });
        }

        $(document).on('click.logout', '#logout', function () {
            logout();
        });

        init(); // 确保 init 函数只被调用一次
    });
})(layui.$);
