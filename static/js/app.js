/**
 * NCOE Case Management - Minimal JavaScript
 * Bootstrap 5.3 dark/light theme toggle with HTMX compatibility
 */

(function() {
    'use strict';

    console.log('[NCOE] app.js loaded');

    // Check if Bootstrap is available
    if (typeof bootstrap === 'undefined') {
        console.error('[NCOE] Bootstrap JS not loaded!');
    } else {
        console.log('[NCOE] Bootstrap JS available:', Object.keys(bootstrap));
    }

    // Check if HTMX is available
    if (typeof htmx === 'undefined') {
        console.error('[NCOE] HTMX not loaded!');
    } else {
        console.log('[NCOE] HTMX available, version:', htmx.version);
    }

    // Theme management using Bootstrap 5.3 color modes
    const ThemeManager = {
        STORAGE_KEY: 'theme',

        getStoredTheme: function() {
            return localStorage.getItem(this.STORAGE_KEY);
        },

        setStoredTheme: function(theme) {
            localStorage.setItem(this.STORAGE_KEY, theme);
        },

        getPreferredTheme: function() {
            const stored = this.getStoredTheme();
            if (stored) {
                return stored;
            }
            return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
        },

        setTheme: function(theme) {
            console.log('[NCOE] Setting theme to:', theme);
            document.documentElement.setAttribute('data-bs-theme', theme);
            this.updateToggleIcon(theme);
        },

        updateToggleIcon: function(theme) {
            const toggles = document.querySelectorAll('#theme-toggle, [data-theme-toggle]');
            console.log('[NCOE] Found theme toggles:', toggles.length);
            toggles.forEach(function(toggle) {
                const icon = toggle.querySelector('i');
                if (!icon) return;

                if (theme === 'dark') {
                    icon.classList.remove('bi-moon-fill');
                    icon.classList.add('bi-sun-fill');
                } else {
                    icon.classList.remove('bi-sun-fill');
                    icon.classList.add('bi-moon-fill');
                }
            });
        },

        toggle: function() {
            const currentTheme = document.documentElement.getAttribute('data-bs-theme') || 'light';
            const newTheme = currentTheme === 'dark' ? 'light' : 'dark';
            console.log('[NCOE] Toggling theme from', currentTheme, 'to', newTheme);
            this.setStoredTheme(newTheme);
            this.setTheme(newTheme);
        },

        init: function() {
            console.log('[NCOE] ThemeManager.init()');
            this.setTheme(this.getPreferredTheme());

            window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', function(e) {
                if (!ThemeManager.getStoredTheme()) {
                    ThemeManager.setTheme(e.matches ? 'dark' : 'light');
                }
            });
        }
    };

    // Use event delegation for theme toggle (works with HTMX)
    document.addEventListener('click', function(event) {
        const toggle = event.target.closest('#theme-toggle, [data-theme-toggle]');
        if (toggle) {
            console.log('[NCOE] Theme toggle clicked');
            event.preventDefault();
            ThemeManager.toggle();
        }
    });

    // Auto-dismiss alerts after 5 seconds
    function initAlertDismiss() {
        const alerts = document.querySelectorAll('.alert-dismissible');
        alerts.forEach(function(alert) {
            setTimeout(function() {
                const closeBtn = alert.querySelector('.btn-close');
                if (closeBtn) closeBtn.click();
            }, 5000);
        });
    }

    // Form validation helpers
    function initFormValidation() {
        const forms = document.querySelectorAll('form[data-validate]');
        forms.forEach(function(form) {
            form.addEventListener('submit', function(event) {
                if (!form.checkValidity()) {
                    event.preventDefault();
                    event.stopPropagation();
                }
                form.classList.add('was-validated');
            });
        });
    }

    // ==========================================================================
    // Navigation Highlighting (for HTMX SPA-style navigation)
    // ==========================================================================
    function setActiveNavFromPath() {
        var path = window.location.pathname;

        // Map pathname to nav key
        var key = null;
        if (path.startsWith('/staff/dashboard')) key = 'dashboard';
        else if (path.startsWith('/staff/cases')) key = 'cases';
        else if (path.startsWith('/staff/deadlines')) key = 'deadlines';
        else if (path.startsWith('/staff/acknowledgments')) key = 'acknowledgments';
        else if (path.startsWith('/staff/reports')) key = 'reports';
        else if (path.startsWith('/staff/search')) key = 'search';
        else if (path.startsWith('/staff/users')) key = 'users';
        else if (path.startsWith('/staff/settings')) key = 'settings';

        if (!key) return;

        // Remove active styles from all sidebar links
        document.querySelectorAll('#sidebarMenu [data-navkey]').forEach(function(a) {
            a.classList.remove('active', 'bg-primary', 'text-white');
            a.classList.add('text-white-50');
        });

        // Remove active from all desktop nav links
        document.querySelectorAll('.navbar-nav [data-navkey]').forEach(function(a) {
            a.classList.remove('active');
        });

        // Apply active styles to matching sidebar link
        var activeSidebar = document.querySelector('#sidebarMenu [data-navkey="' + key + '"]');
        if (activeSidebar) {
            activeSidebar.classList.add('active', 'bg-primary', 'text-white');
            activeSidebar.classList.remove('text-white-50');
        }

        // Apply active to matching desktop navbar link
        var activeNavbar = document.querySelector('.navbar-nav [data-navkey="' + key + '"]');
        if (activeNavbar) {
            activeNavbar.classList.add('active');
        }
    }

    // Initialize on DOM ready
    document.addEventListener('DOMContentLoaded', function() {
        console.log('[NCOE] DOMContentLoaded');
        ThemeManager.init();
        initAlertDismiss();
        initFormValidation();
        setActiveNavFromPath();

        // Debug: Check for offcanvas elements
        const sidebarMenu = document.getElementById('sidebarMenu');
        const casePanel = document.getElementById('casePanel');
        console.log('[NCOE] sidebarMenu element:', sidebarMenu ? 'FOUND' : 'NOT FOUND');
        console.log('[NCOE] casePanel element:', casePanel ? 'FOUND' : 'NOT FOUND');

        // Debug: Check for hamburger button
        const hamburger = document.querySelector('[data-bs-toggle="offcanvas"][data-bs-target="#sidebarMenu"]');
        console.log('[NCOE] Hamburger button:', hamburger ? 'FOUND' : 'NOT FOUND');

        // Debug: Test Bootstrap Offcanvas
        if (sidebarMenu && typeof bootstrap !== 'undefined' && bootstrap.Offcanvas) {
            console.log('[NCOE] Bootstrap Offcanvas class available');
        }
    });

    // ==========================================================================
    // HTMX Event Handlers
    // ==========================================================================

    document.addEventListener('htmx:beforeRequest', function(evt) {
        console.log('[NCOE] htmx:beforeRequest', evt.detail.pathInfo.requestPath);
    });

    document.addEventListener('htmx:afterRequest', function(evt) {
        console.log('[NCOE] htmx:afterRequest', evt.detail.pathInfo.requestPath, 'status:', evt.detail.xhr.status);
    });

    document.addEventListener('htmx:afterSwap', function(evt) {
        console.log('[NCOE] htmx:afterSwap, target:', evt.detail.target.id);

        // Auto-show case panel when content is loaded
        if (evt.detail.target.id === 'casePanelBody') {
            console.log('[NCOE] Case panel body updated, showing offcanvas...');
            var panelEl = document.getElementById('casePanel');
            if (panelEl) {
                if (typeof bootstrap !== 'undefined' && bootstrap.Offcanvas) {
                    var offcanvas = new bootstrap.Offcanvas(panelEl);
                    offcanvas.show();
                    console.log('[NCOE] Offcanvas.show() called');
                } else {
                    console.error('[NCOE] Bootstrap Offcanvas not available!');
                }
            } else {
                console.error('[NCOE] casePanel element not found!');
            }
        }
    });

    document.addEventListener('htmx:afterSettle', function() {
        console.log('[NCOE] htmx:afterSettle');
        ThemeManager.updateToggleIcon(ThemeManager.getPreferredTheme());
        initAlertDismiss();
        setActiveNavFromPath();
    });

    // Handle browser back/forward navigation
    window.addEventListener('popstate', setActiveNavFromPath);

    document.addEventListener('htmx:load', function() {
        console.log('[NCOE] htmx:load');
        ThemeManager.updateToggleIcon(ThemeManager.getPreferredTheme());
    });

    document.addEventListener('htmx:responseError', function(evt) {
        console.error('[NCOE] htmx:responseError', evt.detail.xhr.status, evt.detail.pathInfo.requestPath);
    });

    // Refresh table when case is updated
    document.body.addEventListener('caseUpdated', function() {
        console.log('[NCOE] caseUpdated event received');
        htmx.trigger('#cases-table', 'refresh');
    });

})();
