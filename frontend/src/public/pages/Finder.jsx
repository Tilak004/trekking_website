import { useMemo, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { inr, photoURL } from '../../lib/util.js';

// Trek finder: a few quick questions, and the trek grid re-sorts live.

const LEVELS = [['any', 'Any'], ['easy', 'Beginner friendly'], ['moderate', 'Moderate'], ['hard', 'Challenging']];
const WHEN = [['any', 'Any time'], ['open', 'Booking open'], ['soon', 'Opening soon']];
const LENGTH = [['any', 'Any'], ['short', 'Half day'], ['day', 'Full day'], ['multi', '2 days +']];
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
  return { easy: d.startsWith('easy'), moderate: d.includes('moderate'), hard: d.includes('difficult') };
}

function lengthOf(duration) {
  const d = duration.toLowerCase();
  if (/\b([2-9]) days\b/.test(d)) return 'multi';
  if (/half|morning|sunrise|early/.test(d)) return 'short';
  return 'day';
}

function Chips({ options, value, onChange, multi }) {
  return (
    <div className="chips">
      {options.map(([key, label]) => {
        const on = multi ? value.includes(key) : value === key;
        return (
          <button key={key} type="button" className={on ? 'chip on' : 'chip'} aria-pressed={on}
            onClick={() => onChange(multi ? (on ? value.filter((v) => v !== key) : [...value, key]) : key)}>
            {label}
          </button>
        );
      })}
    </div>
  );
}

const Mountains = () => (
  <svg viewBox="0 0 400 200" preserveAspectRatio="none" aria-hidden="true">
    <path d="M0 200 90 90l50 55 80-110 70 85 40-35 70 115z" fill="rgba(255,255,255,.12)" />
    <path d="M0 200 70 140l60 40 90-75 80 60 50-30 50 65z" fill="rgba(255,255,255,.18)" />
  </svg>
);

// A trek card from the finder's trimmed-down trek data.
function FinderCard({ t, badge }) {
  const open = !!t.next;
  const navigate = useNavigate();
  const book = (e) => { e.preventDefault(); e.stopPropagation(); navigate(`/treks/${t.slug}#book`); };
  return (
    <Link className="card trek-card" to={`/treks/${t.slug}`}>
      <div className="card-img" style={t.photo ? { backgroundImage: `url('${photoURL(t.photo)}')` } : undefined}>
        {!t.photo && <Mountains />}
        <span className="tag diff">{t.difficulty}</span>
        {t.trail === 'closed' ? <span className="tag open closed">Trail closed</span> : open && <span className="tag open">Booking open</span>}
        {open && t.price > 0 && <span className="card-price">₹{inr(t.price)}</span>}
        {badge && <span className="tag match">{badge}</span>}
      </div>
      <div className="card-body">
        <p className="eyebrow">{t.region}</p>
        <h3>{t.name}</h3>
        <div className="card-bottom">
          {open ? <span>Next: <strong>{t.next}</strong></span> : <span className="soon-tag">Opening soon</span>}
          {t.ratings > 0 && <span className="rating">★ {t.rating.toFixed(1)}</span>}
        </div>
        {open && t.trail !== 'closed' && (
          <span className="btn small card-book" role="button" tabIndex={0} onClick={book}
            onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ' ') book(e); }}>Book now</span>
        )}
      </div>
    </Link>
  );
}

export default function Finder({ treks }) {
  const [level, setLevel] = useState('any');
  const [when, setWhen] = useState('any');
  const [length, setLength] = useState('any');
  const [vibes, setVibes] = useState([]);
  const [open, setOpen] = useState(false); // phones: filters folded away until asked for

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

  const active = (level !== 'any') + (when !== 'any') + (length !== 'any') + (vibes.length > 0);
  return (
    <div className="finder">
      <button type="button" className={'finder-toggle' + (open ? ' open' : '')} aria-expanded={open} onClick={() => setOpen(!open)}>
        <span>Filter treks{active ? ` · ${active} on` : ''}</span>
        <span className="muted">{results.length} {results.length === 1 ? 'trek' : 'treks'} <i aria-hidden="true">{open ? '▴' : '▾'}</i></span>
      </button>
      <div className={'finder-panel' + (open ? ' open' : '')}>
        <div className="finder-q"><span>How fit are you?</span><Chips options={LEVELS} value={level} onChange={setLevel} /></div>
        <div className="finder-q"><span>How long?</span><Chips options={LENGTH} value={length} onChange={setLength} /></div>
        <div className="finder-q"><span>What do you want to see?</span><Chips options={VIBES} value={vibes} onChange={setVibes} multi /></div>
        <div className="finder-q"><span>When?</span><Chips options={WHEN} value={when} onChange={setWhen} /></div>
        <div className="finder-meta">
          <strong>{results.length}</strong> {results.length === 1 ? 'trek matches' : 'treks match'}
          {filtered && <button type="button" className="linklike" onClick={reset}>Reset</button>}
        </div>
      </div>
      {results.length ? (
        <div className="grid finder-grid">
          {results.map(({ t, score }, n) => (
            <FinderCard key={t.slug} t={t} badge={vibes.length && score > 0 && n === 0 ? 'Best match' : null} />
          ))}
        </div>
      ) : (
        <div className="finder-empty">
          No trek matches all of that yet. <button type="button" className="linklike" onClick={reset}>Clear filters</button>
        </div>
      )}
    </div>
  );
}
