// Google Analytics 4. The tag itself is loaded by index.html, and only on the live
// site, so local and test runs never add visitors. The app is a single page, so each
// page change is reported here as a page view.
//
// What is sent is kept anonymous on purpose: never a name, phone number, booking
// number, payment token or link address (those can hold a phone number or a private
// group invite). See deploy/ANALYTICS.md for the list of events.

// Addresses are cleaned before they are sent: payment links carry a secret token
// and booking pages a booking number, neither of which may reach Google.
export function analyticsPath(pathname) {
  if (/^\/pay\//.test(pathname)) return '/pay';
  if (/^\/my\/bookings\//.test(pathname)) return '/my/bookings';
  return pathname;
}

export function track(event, params) {
  if (typeof window.gtag === 'function') window.gtag('event', event, params || {});
}

export function trackPageView(pathname) {
  const path = analyticsPath(pathname);
  track('page_view', { page_path: path, page_location: window.location.origin + path });
}

// Runs a callback once per browser for a key ("this booking was seen as confirmed").
export function once(key, fn) {
  try {
    if (localStorage.getItem(key)) return;
    localStorage.setItem(key, '1');
  } catch {
    /* storage blocked: better to count twice than never */
  }
  fn();
}

// Which kind of link is it? Only this kind is reported, never the address.
function linkKind(a) {
  const href = a.getAttribute('href') || '';
  if (/^tel:/i.test(href)) return 'call_click';
  if (/^mailto:/i.test(href)) return 'email_click';
  let u;
  try {
    u = new URL(href, window.location.href);
  } catch {
    return '';
  }
  const host = u.hostname.replace(/^www\./, '');
  if (host === 'chat.whatsapp.com') return 'join_whatsapp_group';
  if (host === 'whatsapp.com' && u.pathname.startsWith('/channel')) return 'follow_whatsapp_channel';
  if (host === 'wa.me' || host === 'api.whatsapp.com') return 'whatsapp_click';
  if (host === 'instagram.com') return 'instagram_click';
  return '';
}

// One listener for the whole page, so every WhatsApp, call, e-mail and Instagram
// link is counted without touching each one.
let clicksOn = false;
export function trackLinkClicks() {
  if (clicksOn) return;
  clicksOn = true;
  document.addEventListener('click', (e) => {
    const a = e.target instanceof Element ? e.target.closest('a[href]') : null;
    const kind = a && linkKind(a);
    if (!kind) return;
    const path = analyticsPath(window.location.pathname);
    const trek = /^\/treks\/([^/]+)/.exec(path);
    track(kind, { link_location: path, ...(trek ? { item_id: trek[1] } : {}) });
  });
}
