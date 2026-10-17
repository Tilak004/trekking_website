import L from 'leaflet';
import { inr, photoURL } from './shared.jsx';

// Western Ghats map: terrain tiles, the Ghats belt through Karnataka, Bangalore
// (where weekend treks start) and a pin for every trek.

// A rough outline of the Western Ghats belt across Karnataka (approximate, for orientation).
const GHATS = [
  [15.75, 74.05], [15.1, 74.42], [14.45, 74.72], [13.8, 75.05], [13.2, 75.45], [12.6, 75.75],
  [12.05, 76.05], [11.65, 76.35], [11.45, 76.05], [11.85, 75.7], [12.4, 75.42], [13.0, 75.12],
  [13.6, 74.78], [14.25, 74.42], [14.95, 74.12], [15.65, 73.82],
];
const BANGALORE = [12.97, 77.59];

const esc = (s) => String(s).replace(/[&<>"']/g, (c) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' })[c]);

function ensureCss() {
  if (document.getElementById('leaflet-css')) return;
  const link = document.createElement('link');
  link.id = 'leaflet-css';
  link.rel = 'stylesheet';
  link.href = '/static/js/leaflet.css';
  document.head.appendChild(link);
}

function popup(t) {
  const img = t.photo ? `<span class="pop-img" style="background-image:url('${esc(photoURL(t.photo))}')"></span>` : '';
  const when = t.next
    ? `<span class="pop-open">Next: ${esc(t.next)}${t.price ? ` · ₹${inr(t.price)}` : ''}</span>`
    : '<span class="pop-soon">Opening soon</span>';
  return `<a class="map-pop" href="/treks/${esc(t.slug)}">${img}<span class="pop-body">
    <b>${esc(t.name)}</b><small>${esc(t.difficulty)}${t.distanceKm ? ` · ${t.distanceKm} km` : ''}</small>${when}
    <em>View trek →</em></span></a>`;
}

export function mount(el, treks) {
  if (!Array.isArray(treks)) return;
  ensureCss();
  el.textContent = '';
  const single = el.dataset.single === '1';
  const map = L.map(el, { scrollWheelZoom: false, zoomSnap: 0.25 });
  // Esri basemaps (no API key needed): dark grey on the dark theme, street map otherwise.
  const dark = (document.documentElement.dataset.theme || '').split(' ').includes('night');
  const esri = (layer) => `https://server.arcgisonline.com/ArcGIS/rest/services/${layer}/MapServer/tile/{z}/{y}/{x}`;
  const attribution = 'Tiles © <a href="https://www.esri.com">Esri</a> · HERE, Garmin, © OpenStreetMap contributors';
  if (dark) {
    L.tileLayer(esri('Canvas/World_Dark_Gray_Base'), { maxZoom: 16, attribution }).addTo(map);
    L.tileLayer(esri('Canvas/World_Dark_Gray_Reference'), { maxZoom: 16 }).addTo(map);
  } else {
    L.tileLayer(esri('World_Street_Map'), { maxZoom: 18, attribution }).addTo(map);
  }
  const accent = getComputedStyle(document.documentElement).getPropertyValue('--sun').trim() || '#2f7a72';

  const points = [];
  if (!single) {
    L.polygon(GHATS, {
      color: accent, weight: 1, opacity: 0.6, fillColor: accent, fillOpacity: dark ? 0.12 : 0.14, interactive: false, smoothFactor: 2,
    }).addTo(map);
    L.marker([13.62, 75.0], {
      interactive: false,
      icon: L.divIcon({ className: 'ghats-label', html: 'Western Ghats', iconSize: null }),
    }).addTo(map);
    L.marker(BANGALORE, {
      icon: L.divIcon({ className: 'city-pin', html: '<i></i>Bangalore · start', iconSize: null, iconAnchor: [6, 6] }),
      keyboard: false,
    }).addTo(map);
    points.push(BANGALORE);
  }

  for (const t of treks) {
    if (!t.lat || !t.lng) continue;
    const icon = L.divIcon({
      className: 'trek-pin' + (t.next ? ' open' : ''),
      html: `<i></i><span>${esc(t.name)}</span>`,
      iconSize: null,
      iconAnchor: [9, 24],
      popupAnchor: [0, -22],
    });
    L.marker([t.lat, t.lng], { icon, title: t.name }).addTo(map).bindPopup(popup(t), { closeButton: false, minWidth: 220 });
    points.push([t.lat, t.lng]);
  }

  if (single && points.length) map.setView(points[0], 11);
  else if (points.length) {
    // One step wider than a tight fit, so the treks sit in their wider South India context.
    const bounds = L.latLngBounds(points).pad(0.12);
    map.setView(bounds.getCenter(), map.getBoundsZoom(bounds) - 1);
  }
  else map.setView([13.3, 75.8], 7);

  // Enable wheel zoom only after the user clicks the map, so page scrolling isn't hijacked.
  map.once('focus click', () => map.scrollWheelZoom.enable());
}
