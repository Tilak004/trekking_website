import { render } from 'preact';
import { useMemo, useState } from 'preact/hooks';
import { TrekCard } from './shared.jsx';

// Trek finder: a few quick questions, and the trek grid re-sorts live.

const LEVELS = [
  ['any', 'Any'],
  ['easy', 'Beginner friendly'],
  ['moderate', 'Moderate'],
  ['hard', 'Challenging'],
];
const WHEN = [
  ['any', 'Any time'],
  ['open', 'Booking open'],
  ['soon', 'Opening soon'],
];
const LENGTH = [
  ['any', 'Any'],
  ['short', 'Half day'],
  ['day', 'Full day'],
  ['multi', '2 days +'],
];
const VIBES = [
  ['sunrise', 'Sunrise'],
  ['waterfalls', 'Waterfalls'],
  ['grasslands', 'Grasslands'],
  ['forest', 'Shola forest'],
  ['summit', 'Summit views'],
  ['rocky', 'Rocky climbs'],
];

function levelOf(d) {
  d = d.toLowerCase();
  return {
    easy: d.startsWith('easy'),
    moderate: d.includes('moderate'),
    hard: d.includes('difficult'),
  };
}

function lengthOf(duration) {
  const d = duration.toLowerCase();
  if (/\b([2-9]) days\b/.test(d)) return 'multi';
  if (/half|morning|sunrise|early/.test(d)) return 'short';
  return 'day';
}

function Chips({ options, value, onChange, multi }) {
  return (
    <div class="chips">
      {options.map(([key, label]) => {
        const on = multi ? value.includes(key) : value === key;
        return (
          <button
            type="button"
            class={on ? 'chip on' : 'chip'}
            aria-pressed={on}
            onClick={() => onChange(multi ? (on ? value.filter((v) => v !== key) : [...value, key]) : key)}
          >
            {label}
          </button>
        );
      })}
    </div>
  );
}

function Finder({ treks }) {
  const [level, setLevel] = useState('any');
  const [when, setWhen] = useState('any');
  const [length, setLength] = useState('any');
  const [vibes, setVibes] = useState([]);

  const results = useMemo(() => {
    const list = treks
      .filter((t) => level === 'any' || levelOf(t.difficulty)[level])
      .filter((t) => when === 'any' || (when === 'open') === !!t.next)
      .filter((t) => length === 'any' || lengthOf(t.duration) === length)
      .map((t, i) => ({ t, i, score: vibes.filter((v) => t.tags.includes(v)).length }));
    // Best vibe match first, then bookable treks, then the original order.
    list.sort((a, b) => b.score - a.score || Number(!!b.t.next) - Number(!!a.t.next) || a.i - b.i);
    return list;
  }, [treks, level, when, length, vibes]);

  const filtered = level !== 'any' || when !== 'any' || length !== 'any' || vibes.length > 0;
  const reset = () => {
    setLevel('any');
    setWhen('any');
    setLength('any');
    setVibes([]);
  };

  return (
    <div class="finder">
      <div class="finder-panel">
        <div class="finder-q"><span>How fit are you?</span><Chips options={LEVELS} value={level} onChange={setLevel} /></div>
        <div class="finder-q"><span>How long?</span><Chips options={LENGTH} value={length} onChange={setLength} /></div>
        <div class="finder-q"><span>What do you want to see?</span><Chips options={VIBES} value={vibes} onChange={setVibes} multi /></div>
        <div class="finder-q"><span>When?</span><Chips options={WHEN} value={when} onChange={setWhen} /></div>
        <div class="finder-meta">
          <strong>{results.length}</strong> {results.length === 1 ? 'trek matches' : 'treks match'}
          {filtered && <button type="button" class="linklike" onClick={reset}>Reset</button>}
        </div>
      </div>
      {results.length ? (
        <div class="grid finder-grid">
          {results.map(({ t, score }, n) => (
            <TrekCard key={t.slug} t={t} badge={vibes.length && score > 0 && n === 0 ? 'Best match' : null} />
          ))}
        </div>
      ) : (
        <div class="finder-empty">
          No trek matches all of that yet. <button type="button" class="linklike" onClick={reset}>Clear filters</button>
        </div>
      )}
    </div>
  );
}

export function mount(el, treks) {
  if (!Array.isArray(treks)) return;
  el.textContent = '';
  render(<Finder treks={treks} />, el);
}
