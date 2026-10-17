// Island loader: the server renders every page; elements marked data-island="…"
// get upgraded to interactive components. Each island's code is a separate
// chunk, downloaded only on pages that use it (the map only once it's near the
// viewport), so pages stay light.

const loaders = {
  finder: () => import('./finder.jsx'),
  map: () => import('./map.js'),
  booking: () => import('./booking.jsx'),
  elevation: () => import('./elevation.jsx'),
};

// Island data comes from a <script type="application/json" class="island-data">
// inside the element, or the page-wide #trek-data block.
function readData(el) {
  const s = el.querySelector(':scope > script.island-data') || document.getElementById('trek-data');
  try {
    return s ? JSON.parse(s.textContent) : null;
  } catch {
    return null;
  }
}

function start(el) {
  const load = loaders[el.dataset.island];
  if (!load) return;
  load()
    .then((m) => m.mount(el, readData(el)))
    .catch((err) => console.error('island failed:', el.dataset.island, err)); // server markup stays as fallback
}

const lazy = 'IntersectionObserver' in window
  ? new IntersectionObserver((entries) => {
      for (const e of entries) {
        if (e.isIntersecting) {
          lazy.unobserve(e.target);
          start(e.target);
        }
      }
    }, { rootMargin: '300px' })
  : null;

for (const el of document.querySelectorAll('[data-island]')) {
  if (el.dataset.island === 'map' && lazy) lazy.observe(el);
  else start(el);
}
