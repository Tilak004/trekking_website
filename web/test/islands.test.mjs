// Smoke tests for the islands in a simulated DOM: node test/islands.test.mjs
import { Window } from 'happy-dom';
import { readdirSync } from 'node:fs';
import assert from 'node:assert/strict';

const win = new Window({ url: 'http://localhost/' });
for (const k of ['window', 'document', 'HTMLElement', 'Event', 'MouseEvent', 'Node', 'getComputedStyle']) {
  globalThis[k] = k === 'window' ? win : win[k];
}

const js = readdirSync(new URL('../../static/js/', import.meta.url));
const chunk = (name) => import(new URL('../../static/js/' + js.find((f) => f.startsWith(name + '-')), import.meta.url));
const flush = () => new Promise((r) => setTimeout(r, 0));
const click = async (el) => { el.dispatchEvent(new win.MouseEvent('click', { bubbles: true })); await flush(); };
const type = async (el, v) => { el.value = v; el.dispatchEvent(new win.Event('input', { bubbles: true })); await flush(); };
const byText = (root, sel, text) => [...root.querySelectorAll(sel)].find((e) => e.textContent.trim().startsWith(text));

const treks = [
  { slug: 'a', name: 'Alpha', region: 'R', difficulty: 'Easy', duration: 'Half day (morning)', tags: ['sunrise'], next: 'Sat, 12 Dec 2026', price: 999, rating: 0, ratings: 0, lead: null, photo: '' },
  { slug: 'b', name: 'Bravo', region: 'R', difficulty: 'Moderate', duration: '1 day (overnight travel)', tags: ['waterfalls', 'forest'], next: '', price: 0, rating: 0, ratings: 0, lead: { name: 'Tilak', photo: '', verified: true }, photo: '' },
  { slug: 'c', name: 'Charlie', region: 'R', difficulty: 'Difficult', duration: '2 days / 1 night', tags: ['summit'], next: '', price: 0, rating: 0, ratings: 0, lead: null, photo: '' },
];

// ---- finder ----
{
  const el = document.createElement('div');
  document.body.appendChild(el);
  (await chunk('finder')).mount(el, treks);
  await flush();
  const names = () => [...el.querySelectorAll('.trek-card h3')].map((h) => h.textContent);
  assert.deepEqual(names(), ['Alpha', 'Bravo', 'Charlie'], 'all treks, bookable first');
  await click(byText(el, '.chip', 'Waterfalls'));
  assert.equal(names()[0], 'Bravo', 'vibe match ranks first');
  assert.ok(el.querySelector('.tag.match'), 'best-match badge');
  await click(byText(el, '.chip', 'Beginner friendly'));
  assert.deepEqual(names(), ['Alpha'], 'level filter');
  await click(byText(el, '.chip', '2 days +'));
  assert.ok(el.querySelector('.finder-empty'), 'empty state');
  await click(byText(el, '.finder-empty button', 'Clear'));
  assert.equal(names().length, 3, 'reset');
  assert.match(el.textContent, /Led by Tilak/, 'lead shown on card');
  console.log('finder ✓');
}

// ---- booking flow (guest: no account) ----
{
  const el = document.createElement('div');
  document.body.appendChild(el);
  const d = { slug: 'a', price: 999, loggedIn: false, name: '', phone: '', pickups: ['Bangalore - Majestic', 'Base'], batches: [
    { id: 7, label: 'Sat, 12 Dec 2026', year: '2026', left: 3, lead: 'Tilak' },
    { id: 8, label: 'Sat, 19 Dec 2026', left: 0, lead: '' },
  ] };
  (await chunk('booking')).mount(el, d);
  await flush();
  const form = el.querySelector('form');
  const hidden = (n) => form.querySelector(`input[name="${n}"]`).value;
  assert.equal(hidden('batch_id'), '7', 'first open batch preselected');
  assert.ok(byText(el, '.bk-date', 'Sat, 19 Dec').disabled, 'full batch disabled');
  for (let i = 0; i < 5; i++) await click(el.querySelector('[aria-label="More"]'));
  assert.equal(hidden('people'), '3', 'people capped at seats left');
  assert.match(el.textContent, /₹2,997/, 'total for 3');
  await click(byText(el, '.bk-nav button', 'Book now'));                 // -> details
  form.dispatchEvent(new window.Event('submit', { cancelable: true }));
  await flush();
  assert.match(el.querySelector('.bk-err').textContent, /WhatsApp/, 'name + phone required');
  const [nameIn, phoneIn] = el.querySelectorAll('.bk-pane input');
  await type(nameIn, 'Ravi');
  await type(phoneIn, '98765 43210');
  await click(byText(el, '.chip', 'Base'));
  assert.ok(byText(el, '.bk-nav button', 'Send booking request'), 'submit on the details step');
  assert.deepEqual(['batch_id', 'people', 'pickup', 'name', 'phone'].map(hidden),
    ['7', '3', 'Base', 'Ravi', '98765 43210'], 'posts the server form fields');
  console.log('booking ✓');
}

// ---- elevation ----
{
  const el = document.createElement('div');
  el.dataset.points = '0 800\n2 1100\n4 1000\n6 1500';
  document.body.appendChild(el);
  (await chunk('elevation')).mount(el);
  await flush();
  assert.match(el.textContent, /6 km/);
  assert.match(el.textContent, /↑ 800 m/, 'total climb = 300 + 500');
  assert.match(el.textContent, /1,500 m/);
  console.log('elevation ✓');
}

await win.happyDOM.close();
