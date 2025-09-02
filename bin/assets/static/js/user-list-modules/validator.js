(function (exports, $) {
    'use strict';

    function verifyUser(username) {
        var valid = true;
        if (username.trim() === '' || !/^\w+$/.test(username)) {
            valid = false;
        }
        return {valid: valid, trim: username.trim()};
    }

    function verifyToken(token) {
        var valid = true;
        if (token.trim() === '' || !/^[\w!@#$%^&*()]+$/.test(token)) {
            valid = false;
        }
        return {valid: valid, trim: token.trim()};
    }

    function verifyComment(comment) {
        var valid = true;
        if (comment.trim() !== '' && /[\n\t\r]/.test(comment)) {
            valid = false;
        }
        return {valid: valid, trim: comment.trim().replace(/[\n\t\r]/g, '')};
    }

    function verifyPorts(ports) {
        var valid = true;
        if (ports.trim() !== '') {
            try {
                ports.split(",").forEach(function (port) {
                    if (/^\s*\d{1,5}\s*$/.test(port)) {
                        if (parseInt(port) < 1 || parseInt(port) > 65535) {
                            valid = false;
                        }
                    } else if (/^\s*\d{1,5}\s*-\s*\d{1,5}\s*$/.test(port)) {
                        var portRange = port.split('-');
                        if (parseInt(portRange[0]) < 1 || parseInt(portRange[0]) > 65535) {
                            valid = false;
                        } else if (parseInt(portRange[1]) < 1 || parseInt(portRange[1]) > 65535) {
                            valid = false;
                        } else if (parseInt(portRange[0]) > parseInt(portRange[1])) {
                            valid = false;
                        }
                    } else {
                        valid = false;
                    }
                    if (valid === false) {
                        throw 'break';
                    }
                });
            } catch (e) {
            }
        }
        return {valid: valid, trim: ports.replace(/\s/g, '')};
    }

    function verifyDomains(domains) {
        var valid = true;
        if (domains.trim() !== '') {
            try {
                domains.split(',').forEach(function (domain) {
                    if (!/^([a-zA-Z0-9]+(-[a-zA-Z0-9]+)*\.)+[a-zA-Z]{2,}$/.test(domain.trim())) {
                        valid = false;
                        throw 'break';
                    }
                });
            } catch (e) {
            }
        }
        return {valid: valid, trim: domains.replace(/\s/g, '')};
    }

    function verifySubdomains(subdomains) {
        var valid = true;
        if (subdomains.trim() !== '') {
            try {
                subdomains.split(',').forEach(function (subdomain) {
                    if (!/^[a-zA-z0-9][a-zA-z0-9-]{0,19}$/.test(subdomain.trim())) {
                        valid = false;
                        throw 'break';
                    }
                });
            } catch (e) {
            }
        }
        return {valid: valid, trim: subdomains.replace(/\s/g, '')};
    }

    function verifyExpireDate(expireDate) {
        var valid = true;
        if (expireDate.trim() !== '' && !/^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$/.test(expireDate.trim())) {
            valid = false;
        }
        return {valid: valid, trim: expireDate.trim()};
    }

    exports.createRules = function (i18n) {
        return {
            user: function (value, item) {
                var result = verifyUser(value);
                if (!result.valid) return i18n['UserFormatError'];
                if (item != null) {
                    typeof item === "function" ? item(result.trim) : $(item).val(result.trim);
                }
            },
            token: function (value, item) {
                var result = verifyToken(value);
                if (!result.valid) return i18n['TokenFormatError'];
                if (item != null) {
                    typeof item === "function" ? item(result.trim) : $(item).val(result.trim);
                }
            },
            comment: function (value, item) {
                var result = verifyComment(value);
                if (!result.valid) return i18n['CommentInvalid'];
                if (item != null) {
                    typeof item === "function" ? item(result.trim) : $(item).val(result.trim);
                }
            },
            ports: function (value, item) {
                var result = verifyPorts(value);
                if (!result.valid) return i18n['PortsInvalid'];
                if (item != null) {
                    typeof item === "function" ? item(result.trim) : $(item).val(result.trim);
                }
            },
            domains: function (value, item) {
                var result = verifyDomains(value);
                if (!result.valid) return i18n['DomainsInvalid'];
                if (item != null) {
                    typeof item === "function" ? item(result.trim) : $(item).val(result.trim);
                }
            },
            subdomains: function (value, item) {
                var result = verifySubdomains(value);
                if (!result.valid) return i18n['SubdomainsInvalid'];
                if (item != null) {
                    typeof item === "function" ? item(result.trim) : $(item).val(result.trim);
                }
            },
            expire_date: function (value, item) {
                var result = verifyExpireDate(value);
                if (!result.valid) return i18n['ExpireDateInvalid'];
                if (item != null) {
                    typeof item === "function" ? item(result.trim) : $(item).val(result.trim);
                }
            },
            server: function (value) {
                if (value === '') return i18n['PleaseSelectServer'];
            }
        };
    };

})(window.UserListValidator = window.UserListValidator || {}, layui.$);
