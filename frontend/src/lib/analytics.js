// Google Analytics 4. The tag itself is loaded by index.html, and only on the live
// site, so local and test runs never add visitors. The app is a single page, so each
// page change is reported here as a page view.

// Addresses are cleaned before they are sent: payment links carry a secret token
// and booking pages a booking number, neither of which may reach Google.
export function analyticsPath(pathname) {
  if (/^\/pay\//.test(pathname)) return '/pay';
  if (/^\/my\/bookings\//.test(pathname)) return '/my/bookings';
  return pathname;
}

export function trackPageView(pathname) {
  if (typeof window.gtag !== 'function') return;
  const path = analyticsPath(pathname);
  window.gtag('event', 'page_view', { page_path: path, page_location: window.location.origin + path });
}
