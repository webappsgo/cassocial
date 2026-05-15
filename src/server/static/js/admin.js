/**
 * Cassocial v1.0.0 - Admin JavaScript
 * License: MIT
 */

(function () {
  'use strict';

  // ==========================================================================
  // CONFIRM DIALOGS FOR DESTRUCTIVE ACTIONS
  // ==========================================================================

  function initConfirmDialogs() {
    document.addEventListener('click', function (e) {
      var btn = e.target.closest('[data-confirm]');
      if (!btn) return;

      var message = btn.getAttribute('data-confirm') || 'Are you sure?';
      if (!window.confirm(message)) {
        e.preventDefault();
        e.stopPropagation();
      }
    });

    document.addEventListener('submit', function (e) {
      var form = e.target;
      var confirmMsg = form.getAttribute('data-confirm');
      if (!confirmMsg) return;

      if (!window.confirm(confirmMsg)) {
        e.preventDefault();
      }
    });
  }

  // ==========================================================================
  // ALERT DISMISS
  // ==========================================================================

  function initAlertDismiss() {
    document.addEventListener('click', function (e) {
      var btn = e.target.closest('.alert-close');
      if (!btn) return;

      var alert = btn.closest('.alert');
      if (alert) {
        alert.style.display = 'none';
      }
    });
  }

  // ==========================================================================
  // TABLE SORTING
  // ==========================================================================

  function initTableSorting() {
    var tables = document.querySelectorAll('.admin-table');

    tables.forEach(function (table) {
      var headers = table.querySelectorAll('th[data-sort]');

      headers.forEach(function (th) {
        th.setAttribute('tabindex', '0');
        th.setAttribute('role', 'button');
        th.setAttribute('aria-sort', 'none');

        function activateSort() {
          var col = th.getAttribute('data-sort');
          var currentDir = th.getAttribute('data-sort-dir') || 'none';
          var newDir = currentDir === 'asc' ? 'desc' : 'asc';

          headers.forEach(function (h) {
            h.removeAttribute('data-sort-active');
            h.removeAttribute('data-sort-dir');
            h.setAttribute('aria-sort', 'none');
          });

          th.setAttribute('data-sort-active', '');
          th.setAttribute('data-sort-dir', newDir);
          th.setAttribute('aria-sort', newDir === 'asc' ? 'ascending' : 'descending');

          sortTable(table, col, newDir);
        }

        th.addEventListener('click', activateSort);
        th.addEventListener('keydown', function (e) {
          if (e.key === 'Enter' || e.key === ' ') {
            e.preventDefault();
            activateSort();
          }
        });
      });
    });
  }

  function sortTable(table, col, direction) {
    var tbody = table.querySelector('tbody');
    if (!tbody) return;

    var rows = Array.from(tbody.querySelectorAll('tr'));
    var headers = Array.from(table.querySelectorAll('th'));
    var colIndex = headers.findIndex(function (th) {
      return th.getAttribute('data-sort') === col;
    });

    if (colIndex < 0) return;

    rows.sort(function (a, b) {
      var aCell = a.cells[colIndex];
      var bCell = b.cells[colIndex];
      if (!aCell || !bCell) return 0;

      var aText = (aCell.textContent || '').trim().toLowerCase();
      var bText = (bCell.textContent || '').trim().toLowerCase();

      var aNum = parseFloat(aText);
      var bNum = parseFloat(bText);

      var cmp;
      if (!isNaN(aNum) && !isNaN(bNum)) {
        cmp = aNum - bNum;
      } else {
        cmp = aText < bText ? -1 : aText > bText ? 1 : 0;
      }

      return direction === 'asc' ? cmp : -cmp;
    });

    rows.forEach(function (row) {
      tbody.appendChild(row);
    });
  }

  // ==========================================================================
  // AJAX SETTINGS FORM SUBMISSION
  // ==========================================================================

  function initAjaxForms() {
    var forms = document.querySelectorAll('form[data-ajax]');

    forms.forEach(function (form) {
      form.addEventListener('submit', function (e) {
        e.preventDefault();

        var submitBtn = form.querySelector('[type="submit"]');
        var originalText = submitBtn ? submitBtn.textContent : '';
        if (submitBtn) {
          submitBtn.disabled = true;
          submitBtn.textContent = 'Saving…';
        }

        clearFormStatus(form);

        var data = new FormData(form);
        var body = new URLSearchParams();
        data.forEach(function (val, key) {
          body.append(key, val);
        });

        fetch(form.action, {
          method: form.method || 'POST',
          headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
          body: body.toString(),
          credentials: 'same-origin'
        })
          .then(function (res) {
            return res.json().then(function (json) {
              return { ok: res.ok, data: json };
            });
          })
          .then(function (result) {
            if (result.ok) {
              showFormStatus(form, 'success', result.data.message || 'Settings saved.');
            } else {
              showFormStatus(form, 'error', result.data.error || 'An error occurred.');
            }
          })
          .catch(function () {
            showFormStatus(form, 'error', 'Network error. Please try again.');
          })
          .finally(function () {
            if (submitBtn) {
              submitBtn.disabled = false;
              submitBtn.textContent = originalText;
            }
          });
      });
    });
  }

  function showFormStatus(form, type, message) {
    var existing = form.querySelector('.form-ajax-status');
    if (existing) existing.remove();

    var div = document.createElement('div');
    div.className = 'alert ' + (type === 'success' ? 'alert-success' : 'alert-danger') + ' form-ajax-status';
    div.setAttribute('role', 'alert');
    div.textContent = message;

    form.insertAdjacentElement('afterbegin', div);
    div.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
  }

  function clearFormStatus(form) {
    var existing = form.querySelector('.form-ajax-status');
    if (existing) existing.remove();
  }

  // ==========================================================================
  // INIT
  // ==========================================================================

  function init() {
    initConfirmDialogs();
    initAlertDismiss();
    initTableSorting();
    initAjaxForms();
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();
