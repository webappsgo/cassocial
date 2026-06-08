/**
 * Cassocial v1.0.0 - Theme Init
 * License: MIT
 *
 * Loaded synchronously in <head> to set the data-theme attribute before
 * the page paints, preventing a flash of the wrong theme. Must be tiny
 * and side-effect-free beyond reading localStorage and the prefers-color-scheme
 * media query.
 */
(function() {
  'use strict';
  try {
    var saved = localStorage.getItem('cassocial-theme');
    if (saved) {
      document.documentElement.setAttribute('data-theme', saved);
    } else if (window.matchMedia && window.matchMedia('(prefers-color-scheme: light)').matches) {
      document.documentElement.setAttribute('data-theme', 'light');
    }
  } catch (_) {
    // localStorage may be unavailable in private mode; ignore and keep default.
  }
})();
