import { useEffect, useRef } from 'react';

// Western Ghats map: terrain tiles, the Ghats belt through Karnataka, Bangalore
// (where weekend treks start) and well-known landmarks. Exact trek trailheads are
// kept private. Leaflet only loads once the map is near the viewport.

// A rough outline of the Western Ghats belt across Karnataka (approximate, for orientation).
const GHATS = [
  [15.75, 74.05], [15.1, 74.42], [14.45, 74.72], [13.8, 75.05], [13.2, 75.45], [12.6, 75.75],
  [12.05, 76.05], [11.65, 76.35], [11.45, 76.05], [11.85, 75.7], [12.4, 75.42], [13.0, 75.12],
  [13.6, 74.78], [14.25, 74.42], [14.95, 74.12], [15.65, 73.82],
];
const BANGALORE = [12.97, 77.59];
const LANDMARKS = [
  ['Nandi Hills', 13.37, 77.68], ['Savandurga', 12.92, 77.29], ['Ramanagara', 12.72, 77.28],
  ['Chikmagalur', 13.32, 75.77], ['Coorg', 12.42, 75.74], ['Sakleshpur', 12.94, 75.78], ['Agumbe', 13.5, 75.09],
];

const esc = (s) => String(s).replace(/[&<>"']/g, (c) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' })[c]);

async function build(el) {
  const [{ default: L }] = await Promise.all([import('leaflet'), import('leaflet/dist/leaflet.css')]);
  const map = L.map(el, { scrollWheelZoom: false, zoomSnap: 0.25 });
  const esri = (layer) => `https://server.arcgisonline.com/ArcGIS/rest/services/${layer}/MapServer/tile/{z}/{y}/{x}`;
  const attribution = 'Tiles © <a href="https://www.esri.com">Esri</a> · HERE, Garmin, © OpenStreetMap contributors';
  L.tileLayer(esri('World_Street_Map'), { maxZoom: 18, attribution }).addTo(map);
  const accent = getComputedStyle(document.documentElement).getPropertyValue('--sun').trim() || '#2f7a72';

  const points = [BANGALORE];
  L.polygon(GHATS, { color: accent, weight: 1, opacity: 0.6, fillColor: accent, fillOpacity: 0.14, interactive: false, smoothFactor: 2 }).addTo(map);
  L.marker([13.62, 75.0], {
    interactive: false,
    icon: L.divIcon({ className: 'ghats-label', html: 'Western Ghats', iconSize: null }),
  }).addTo(map);
  L.marker(BANGALORE, {
    icon: L.divIcon({ className: 'city-pin', html: '<i></i>Bangalore · start', iconSize: null, iconAnchor: [6, 6] }),
    keyboard: false,
  }).addTo(map);
  for (const [name, lat, lng] of LANDMARKS) {
    L.marker([lat, lng], {
      keyboard: false,
      title: name,
      icon: L.divIcon({ className: 'trek-pin landmark', html: `<i></i><span>${esc(name)}</span>`, iconSize: null, iconAnchor: [9, 24] }),
    }).addTo(map);
    points.push([lat, lng]);
  }
  // One step wider than a tight fit, so the pins sit in their wider South India context.
  const bounds = L.latLngBounds(points).pad(0.12);
  map.setView(bounds.getCenter(), map.getBoundsZoom(bounds) - 1);
  // Enable wheel zoom only after the user clicks the map, so page scrolling isn't hijacked.
  map.once('focus click', () => map.scrollWheelZoom.enable());
  return map;
}

export default function MapView() {
  const ref = useRef(null);
  useEffect(() => {
    const el = ref.current;
    let map = null;
    let gone = false;
    const start = () => build(el).then((m) => {
      if (gone) m.remove();
      else map = m;
    }).catch((err) => console.error('map failed', err));
    let io = null;
    if ('IntersectionObserver' in window) {
      io = new IntersectionObserver((entries) => {
        if (entries.some((e) => e.isIntersecting)) {
          io.disconnect();
          start();
        }
      }, { rootMargin: '300px' });
      io.observe(el);
    } else start();
    return () => {
      gone = true;
      if (io) io.disconnect();
      if (map) map.remove();
    };
  }, []);
  return <div className="ghats-map" ref={ref} />;
}
