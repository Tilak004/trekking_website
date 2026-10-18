// Formatting helpers that match the Go template functions.

export const inr = (n) => Number(n || 0).toLocaleString('en-IN');

// "2026-11-07" -> "Sat, 7 Nov 2026"
export function fmtDate(s) {
  if (!s) return '';
  const m = /^(\d{4})-(\d{2})-(\d{2})/.exec(s);
  if (!m) return s;
  const d = new Date(Date.UTC(+m[1], +m[2] - 1, +m[3]));
  const day = d.toLocaleDateString('en-GB', { weekday: 'short', timeZone: 'UTC' });
  const mon = d.toLocaleDateString('en-GB', { month: 'short', timeZone: 'UTC' });
  return `${day}, ${d.getUTCDate()} ${mon} ${d.getUTCFullYear()}`;
}

export const photoURL = (p) => '/photos/' + String(p).split('/').map(encodeURIComponent).join('/');

export const initial = (s) => ([...String(s || '').trim()][0] || '?').toUpperCase();
export const firstWord = (s) => String(s || '').trim().split(/\s+/)[0] || '';
export const slug = (s) => String(s || '').toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-|-$/g, '');
export const rating = (f) => Number(f || 0).toFixed(1);
export const plural = (n, one, many) => `${n} ${n === 1 ? one : many}`;

export function stars(n) {
  const k = Math.max(0, Math.min(5, Math.round(n)));
  return '★'.repeat(k) + '☆'.repeat(5 - k);
}

export const lines = (s) => String(s || '').split('\n').map((l) => l.trim()).filter(Boolean);
export const paras = (s) => String(s || '').replace(/\r\n/g, '\n').split('\n\n').map((p) => p.trim()).filter(Boolean);

// "Title\nbody" chunks separated by a blank line (itinerary, FAQs).
export function blocks(s) {
  return paras(s).map((chunk) => {
    const i = chunk.indexOf('\n');
    return i < 0 ? { Title: chunk, Body: '' } : { Title: chunk.slice(0, i).trim(), Body: chunk.slice(i + 1).trim() };
  });
}

export const digitsOnly = (s) => String(s || '').replace(/\D/g, '');

// A wa.me link to a phone number (10-digit Indian numbers get +91) with text filled in.
export function whatsappTo(phone, text) {
  let d = digitsOnly(phone);
  if (d.length === 10) d = '91' + d;
  return `https://wa.me/${d}` + (text ? `?text=${encodeURIComponent(text).replace(/\+/g, '%20')}` : '');
}

export const telLink = (phone) => String(phone || '').replace(/\s+/g, '');

export const today = () => new Date().toISOString().slice(0, 10);
