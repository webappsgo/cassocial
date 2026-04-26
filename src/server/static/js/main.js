/**
 * Cassocial v1.0.0 - Main JavaScript
 * License: MIT
 */

(function() {
  'use strict';

  // ============================================================================
  // THEME MANAGEMENT
  // ============================================================================

  const ThemeManager = {
    STORAGE_KEY: 'cassocial-theme',

    init() {
      this.loadTheme();
      this.attachToggleListeners();
    },

    loadTheme() {
      const savedTheme = localStorage.getItem(this.STORAGE_KEY) || 'dark';
      this.setTheme(savedTheme);
    },

    setTheme(theme) {
      document.documentElement.setAttribute('data-theme', theme);
      localStorage.setItem(this.STORAGE_KEY, theme);

      // Update toggle states
      const toggles = document.querySelectorAll('.theme-toggle');
      toggles.forEach(toggle => {
        toggle.classList.toggle('active', theme === 'light');
      });
    },

    toggleTheme() {
      const currentTheme = document.documentElement.getAttribute('data-theme') || 'dark';
      const newTheme = currentTheme === 'dark' ? 'light' : 'dark';
      this.setTheme(newTheme);
    },

    attachToggleListeners() {
      const toggles = document.querySelectorAll('.theme-toggle');
      toggles.forEach(toggle => {
        toggle.addEventListener('click', () => this.toggleTheme());
      });
    }
  };

  // ============================================================================
  // MOBILE MENU
  // ============================================================================

  const MobileMenu = {
    init() {
      const toggle = document.querySelector('.mobile-menu-toggle');
      const nav = document.querySelector('nav');

      if (!toggle || !nav) return;

      toggle.addEventListener('click', () => {
        nav.classList.toggle('active');
        const expanded = nav.classList.contains('active');
        toggle.setAttribute('aria-expanded', expanded);
      });

      // Close menu when clicking outside
      document.addEventListener('click', (e) => {
        if (!toggle.contains(e.target) && !nav.contains(e.target)) {
          nav.classList.remove('active');
          toggle.setAttribute('aria-expanded', 'false');
        }
      });

      // Close menu when pressing Escape
      document.addEventListener('keydown', (e) => {
        if (e.key === 'Escape' && nav.classList.contains('active')) {
          nav.classList.remove('active');
          toggle.setAttribute('aria-expanded', 'false');
          toggle.focus();
        }
      });
    }
  };

  // ============================================================================
  // FORM VALIDATION
  // ============================================================================

  const FormValidator = {
    init() {
      const forms = document.querySelectorAll('form[data-validate]');
      forms.forEach(form => this.attachValidation(form));
    },

    attachValidation(form) {
      form.addEventListener('submit', (e) => {
        if (!this.validateForm(form)) {
          e.preventDefault();
          this.focusFirstError(form);
        }
      });

      // Real-time validation
      const inputs = form.querySelectorAll('input, textarea, select');
      inputs.forEach(input => {
        input.addEventListener('blur', () => this.validateField(input));
        input.addEventListener('input', () => {
          if (input.classList.contains('error')) {
            this.validateField(input);
          }
        });
      });
    },

    validateForm(form) {
      let isValid = true;
      const inputs = form.querySelectorAll('input, textarea, select');

      inputs.forEach(input => {
        if (!this.validateField(input)) {
          isValid = false;
        }
      });

      return isValid;
    },

    validateField(field) {
      const formGroup = field.closest('.form-group');
      if (!formGroup) return true;

      let isValid = true;
      let errorMessage = '';

      // Required validation
      if (field.hasAttribute('required') && !field.value.trim()) {
        isValid = false;
        errorMessage = 'This field is required';
      }

      // Email validation
      if (field.type === 'email' && field.value) {
        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        if (!emailRegex.test(field.value)) {
          isValid = false;
          errorMessage = 'Please enter a valid email address';
        }
      }

      // URL validation
      if (field.type === 'url' && field.value) {
        try {
          new URL(field.value);
        } catch {
          isValid = false;
          errorMessage = 'Please enter a valid URL';
        }
      }

      // Min length validation
      if (field.hasAttribute('minlength') && field.value) {
        const minLength = parseInt(field.getAttribute('minlength'));
        if (field.value.length < minLength) {
          isValid = false;
          errorMessage = `Must be at least ${minLength} characters`;
        }
      }

      // Max length validation
      if (field.hasAttribute('maxlength') && field.value) {
        const maxLength = parseInt(field.getAttribute('maxlength'));
        if (field.value.length > maxLength) {
          isValid = false;
          errorMessage = `Must be no more than ${maxLength} characters`;
        }
      }

      // Pattern validation
      if (field.hasAttribute('pattern') && field.value) {
        const pattern = new RegExp(field.getAttribute('pattern'));
        if (!pattern.test(field.value)) {
          isValid = false;
          errorMessage = field.getAttribute('data-pattern-message') || 'Invalid format';
        }
      }

      // Update UI
      this.updateFieldStatus(formGroup, field, isValid, errorMessage);
      return isValid;
    },

    updateFieldStatus(formGroup, field, isValid, errorMessage) {
      const errorElement = formGroup.querySelector('.form-error');

      if (isValid) {
        formGroup.classList.remove('error');
        field.setAttribute('aria-invalid', 'false');
        if (errorElement) {
          errorElement.textContent = '';
        }
      } else {
        formGroup.classList.add('error');
        field.setAttribute('aria-invalid', 'true');
        if (errorElement) {
          errorElement.textContent = errorMessage;
        }
      }
    },

    focusFirstError(form) {
      const firstError = form.querySelector('.form-group.error input, .form-group.error textarea, .form-group.error select');
      if (firstError) {
        firstError.focus();
      }
    }
  };

  // ============================================================================
  // FLASH MESSAGES AUTO-HIDE
  // ============================================================================

  const FlashMessages = {
    DURATION: 5000, // 5 seconds

    init() {
      const alerts = document.querySelectorAll('.alert');
      alerts.forEach(alert => {
        this.attachCloseButton(alert);
        this.autoHide(alert);
      });
    },

    attachCloseButton(alert) {
      const closeBtn = alert.querySelector('.alert-close');
      if (closeBtn) {
        closeBtn.addEventListener('click', () => {
          this.hideAlert(alert);
        });
      }
    },

    autoHide(alert) {
      if (alert.hasAttribute('data-no-auto-hide')) return;

      setTimeout(() => {
        this.hideAlert(alert);
      }, this.DURATION);
    },

    hideAlert(alert) {
      alert.style.opacity = '0';
      alert.style.transition = 'opacity 0.3s ease';

      setTimeout(() => {
        alert.remove();
      }, 300);
    }
  };

  // ============================================================================
  // DRAG & DROP LINK REORDERING
  // ============================================================================

  const DragDrop = {
    draggedElement: null,

    init() {
      const containers = document.querySelectorAll('[data-drag-container]');
      containers.forEach(container => this.attachDragEvents(container));
    },

    attachDragEvents(container) {
      const items = container.querySelectorAll('[draggable="true"]');

      items.forEach(item => {
        item.addEventListener('dragstart', (e) => this.handleDragStart(e));
        item.addEventListener('dragend', (e) => this.handleDragEnd(e));
        item.addEventListener('dragover', (e) => this.handleDragOver(e));
        item.addEventListener('drop', (e) => this.handleDrop(e));
        item.addEventListener('dragenter', (e) => this.handleDragEnter(e));
        item.addEventListener('dragleave', (e) => this.handleDragLeave(e));
      });
    },

    handleDragStart(e) {
      this.draggedElement = e.currentTarget;
      e.currentTarget.classList.add('dragging');
      e.dataTransfer.effectAllowed = 'move';
      e.dataTransfer.setData('text/html', e.currentTarget.innerHTML);
    },

    handleDragEnd(e) {
      e.currentTarget.classList.remove('dragging');

      // Remove all drag-over classes
      const items = document.querySelectorAll('.drag-over');
      items.forEach(item => item.classList.remove('drag-over'));

      this.draggedElement = null;
    },

    handleDragOver(e) {
      if (e.preventDefault) {
        e.preventDefault();
      }
      e.dataTransfer.dropEffect = 'move';
      return false;
    },

    handleDragEnter(e) {
      if (e.currentTarget !== this.draggedElement) {
        e.currentTarget.classList.add('drag-over');
      }
    },

    handleDragLeave(e) {
      e.currentTarget.classList.remove('drag-over');
    },

    handleDrop(e) {
      if (e.stopPropagation) {
        e.stopPropagation();
      }

      e.preventDefault();

      if (this.draggedElement !== e.currentTarget) {
        const container = e.currentTarget.parentNode;
        const allItems = Array.from(container.children);
        const draggedIndex = allItems.indexOf(this.draggedElement);
        const targetIndex = allItems.indexOf(e.currentTarget);

        if (draggedIndex < targetIndex) {
          container.insertBefore(this.draggedElement, e.currentTarget.nextSibling);
        } else {
          container.insertBefore(this.draggedElement, e.currentTarget);
        }

        // Trigger custom event for saving order
        this.saveOrder(container);
      }

      return false;
    },

    saveOrder(container) {
      const items = container.querySelectorAll('[draggable="true"]');
      const order = Array.from(items).map((item, index) => ({
        id: item.getAttribute('data-id'),
        position: index
      }));

      // Dispatch custom event with new order
      const event = new CustomEvent('orderChanged', {
        detail: { order },
        bubbles: true
      });
      container.dispatchEvent(event);

      // If there's a data-save-url, send AJAX request
      const saveUrl = container.getAttribute('data-save-url');
      if (saveUrl) {
        this.sendOrderUpdate(saveUrl, order);
      }
    },

    sendOrderUpdate(url, order) {
      fetch(url, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ order })
      })
      .then(response => response.json())
      .then(data => {
        if (data.success) {
          console.log('Order updated successfully');
        }
      })
      .catch(error => {
        console.error('Error updating order:', error);
      });
    }
  };

  // ============================================================================
  // MODAL MANAGEMENT
  // ============================================================================

  const ModalManager = {
    init() {
      const triggers = document.querySelectorAll('[data-modal-trigger]');
      triggers.forEach(trigger => {
        trigger.addEventListener('click', (e) => {
          e.preventDefault();
          const modalId = trigger.getAttribute('data-modal-trigger');
          this.openModal(modalId);
        });
      });

      const closeButtons = document.querySelectorAll('.modal-close, [data-modal-close]');
      closeButtons.forEach(btn => {
        btn.addEventListener('click', () => {
          this.closeAllModals();
        });
      });

      // Close on overlay click
      const overlays = document.querySelectorAll('.modal-overlay');
      overlays.forEach(overlay => {
        overlay.addEventListener('click', (e) => {
          if (e.target === overlay) {
            this.closeAllModals();
          }
        });
      });

      // Close on Escape key
      document.addEventListener('keydown', (e) => {
        if (e.key === 'Escape') {
          this.closeAllModals();
        }
      });
    },

    openModal(modalId) {
      const modal = document.getElementById(modalId);
      if (modal) {
        modal.classList.add('active');
        document.body.style.overflow = 'hidden';

        // Focus first focusable element
        const focusable = modal.querySelector('button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])');
        if (focusable) {
          focusable.focus();
        }
      }
    },

    closeAllModals() {
      const modals = document.querySelectorAll('.modal-overlay.active');
      modals.forEach(modal => {
        modal.classList.remove('active');
      });
      document.body.style.overflow = '';
    }
  };

  // ============================================================================
  // ANALYTICS CHARTS (Basic Setup)
  // ============================================================================

  const Analytics = {
    charts: {},

    init() {
      // Initialize charts if Chart.js is loaded
      if (typeof Chart !== 'undefined') {
        this.initCharts();
      }
    },

    initCharts() {
      // Views Chart
      const viewsCanvas = document.getElementById('views-chart');
      if (viewsCanvas) {
        this.charts.views = this.createLineChart(viewsCanvas, 'Profile Views');
      }

      // Clicks Chart
      const clicksCanvas = document.getElementById('clicks-chart');
      if (clicksCanvas) {
        this.charts.clicks = this.createBarChart(clicksCanvas, 'Link Clicks');
      }

      // Device Chart
      const deviceCanvas = document.getElementById('device-chart');
      if (deviceCanvas) {
        this.charts.device = this.createPieChart(deviceCanvas);
      }
    },

    createLineChart(canvas, label) {
      const ctx = canvas.getContext('2d');
      return new Chart(ctx, {
        type: 'line',
        data: {
          labels: [],
          datasets: [{
            label: label,
            data: [],
            borderColor: 'rgb(189, 147, 249)',
            backgroundColor: 'rgba(189, 147, 249, 0.1)',
            tension: 0.4
          }]
        },
        options: {
          responsive: true,
          maintainAspectRatio: false,
          plugins: {
            legend: {
              display: true,
              labels: {
                color: getComputedStyle(document.documentElement).getPropertyValue('--text-primary')
              }
            }
          },
          scales: {
            y: {
              beginAtZero: true,
              ticks: {
                color: getComputedStyle(document.documentElement).getPropertyValue('--text-secondary')
              },
              grid: {
                color: getComputedStyle(document.documentElement).getPropertyValue('--border')
              }
            },
            x: {
              ticks: {
                color: getComputedStyle(document.documentElement).getPropertyValue('--text-secondary')
              },
              grid: {
                color: getComputedStyle(document.documentElement).getPropertyValue('--border')
              }
            }
          }
        }
      });
    },

    createBarChart(canvas, label) {
      const ctx = canvas.getContext('2d');
      return new Chart(ctx, {
        type: 'bar',
        data: {
          labels: [],
          datasets: [{
            label: label,
            data: [],
            backgroundColor: 'rgba(80, 250, 123, 0.6)',
            borderColor: 'rgb(80, 250, 123)',
            borderWidth: 1
          }]
        },
        options: {
          responsive: true,
          maintainAspectRatio: false,
          plugins: {
            legend: {
              display: true,
              labels: {
                color: getComputedStyle(document.documentElement).getPropertyValue('--text-primary')
              }
            }
          },
          scales: {
            y: {
              beginAtZero: true,
              ticks: {
                color: getComputedStyle(document.documentElement).getPropertyValue('--text-secondary')
              },
              grid: {
                color: getComputedStyle(document.documentElement).getPropertyValue('--border')
              }
            },
            x: {
              ticks: {
                color: getComputedStyle(document.documentElement).getPropertyValue('--text-secondary')
              },
              grid: {
                color: getComputedStyle(document.documentElement).getPropertyValue('--border')
              }
            }
          }
        }
      });
    },

    createPieChart(canvas) {
      const ctx = canvas.getContext('2d');
      return new Chart(ctx, {
        type: 'doughnut',
        data: {
          labels: ['Mobile', 'Desktop', 'Tablet'],
          datasets: [{
            data: [],
            backgroundColor: [
              'rgba(139, 233, 253, 0.6)',
              'rgba(189, 147, 249, 0.6)',
              'rgba(80, 250, 123, 0.6)'
            ],
            borderColor: [
              'rgb(139, 233, 253)',
              'rgb(189, 147, 249)',
              'rgb(80, 250, 123)'
            ],
            borderWidth: 1
          }]
        },
        options: {
          responsive: true,
          maintainAspectRatio: false,
          plugins: {
            legend: {
              position: 'bottom',
              labels: {
                color: getComputedStyle(document.documentElement).getPropertyValue('--text-primary')
              }
            }
          }
        }
      });
    },

    updateChart(chartName, labels, data) {
      if (this.charts[chartName]) {
        this.charts[chartName].data.labels = labels;
        this.charts[chartName].data.datasets[0].data = data;
        this.charts[chartName].update();
      }
    }
  };

  // ============================================================================
  // COPY TO CLIPBOARD
  // ============================================================================

  const ClipboardManager = {
    init() {
      const copyButtons = document.querySelectorAll('[data-copy]');
      copyButtons.forEach(button => {
        button.addEventListener('click', (e) => {
          e.preventDefault();
          const text = button.getAttribute('data-copy');
          this.copyToClipboard(text, button);
        });
      });
    },

    async copyToClipboard(text, button) {
      try {
        await navigator.clipboard.writeText(text);
        this.showCopyFeedback(button, true);
      } catch (err) {
        // Fallback for older browsers
        this.fallbackCopy(text, button);
      }
    },

    fallbackCopy(text, button) {
      const textarea = document.createElement('textarea');
      textarea.value = text;
      textarea.style.position = 'fixed';
      textarea.style.opacity = '0';
      document.body.appendChild(textarea);
      textarea.select();

      try {
        const successful = document.execCommand('copy');
        this.showCopyFeedback(button, successful);
      } catch (err) {
        this.showCopyFeedback(button, false);
      }

      document.body.removeChild(textarea);
    },

    showCopyFeedback(button, success) {
      const originalText = button.textContent;
      button.textContent = success ? 'Copied!' : 'Failed';
      button.disabled = true;

      setTimeout(() => {
        button.textContent = originalText;
        button.disabled = false;
      }, 2000);
    }
  };

  // ============================================================================
  // COLOR PICKER PREVIEW
  // ============================================================================

  const ColorPicker = {
    init() {
      const colorInputs = document.querySelectorAll('input[type="color"]');
      colorInputs.forEach(input => {
        this.createPreview(input);
        input.addEventListener('input', () => this.updatePreview(input));
      });
    },

    createPreview(input) {
      const preview = document.createElement('span');
      preview.className = 'color-preview';
      preview.style.backgroundColor = input.value;
      input.parentNode.appendChild(preview);

      preview.addEventListener('click', () => input.click());
    },

    updatePreview(input) {
      const preview = input.parentNode.querySelector('.color-preview');
      if (preview) {
        preview.style.backgroundColor = input.value;
      }
    }
  };

  // ============================================================================
  // INITIALIZATION
  // ============================================================================

  function init() {
    // Initialize all modules
    ThemeManager.init();
    MobileMenu.init();
    FormValidator.init();
    FlashMessages.init();
    DragDrop.init();
    ModalManager.init();
    Analytics.init();
    ClipboardManager.init();
    ColorPicker.init();

    console.log('Cassocial v1.0.0 initialized');
  }

  // Run on DOM ready
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }

  // Export to global scope for external access
  window.Cassocial = {
    ThemeManager,
    ModalManager,
    Analytics,
    ClipboardManager
  };

})();
