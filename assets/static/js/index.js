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
                        loadUserList(lang, title.trim());
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

                $('#leftNav .layui-this > a').click();
                loadDashboards(lang, updateCurrentIndex); // Load dashboards after language is loaded
            }).always(function () {
                layui.layer.close(langLoading);
            });
        }

        function loadDashboards(lang, updateCurrentIndexCallback) {
            $.getJSON('/dashboards').done(function (res) {
                if (res.code === 0) {
                    var dashboards = res.data;
                    var currentIndex = res.current_index;
                    var dropdown = $('#dashboardListDropdown');
                    dropdown.empty(); // Clear existing items

                    if (dashboards && dashboards.length > 0) {
                        $('#currentDashboardName').text(dashboards[currentIndex].name);
                        $.each(dashboards, function (index, dashboard) {
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

        init();
    });
})(layui.$);
