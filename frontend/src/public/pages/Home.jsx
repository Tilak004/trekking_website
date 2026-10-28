import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import Icon from '../../components/Icon.jsx';
import { PageState, Rating } from '../../components/common.jsx';
import { useTitle, usePage } from '../../lib/usePage.js';
import { photoURL, rating } from '../../lib/util.js';
import PageShell from '../PageShell.jsx';
import Finder from './Finder.jsx';
import MapView from './MapView.jsx';

// "--ar:.75;--sky:.4" -> { '--ar': '.75', '--sky': '.4' }
function cssVars(str) {
  const out = {};
  for (const part of String(str || '').split(';')) {
    const i = part.indexOf(':');
    if (i > 0) out[part.slice(0, i).trim()] = part.slice(i + 1).trim();
  }
  return out;
}

function HeroFeatures() {
  return (
    <div className="hero-features">
      <div className="wrap features">
        <div className="feature"><Icon name="mountain" /><span>Leads who know every turn</span></div>
        <div className="feature"><Icon name="users" /><span>Small groups, never crowds</span></div>
        <div className="feature"><Icon name="calendar" /><span>Pickups in Bangalore</span></div>
        <div className="feature"><Icon name="wallet" /><span>Free to request a seat</span></div>
      </div>
    </div>
  );
}

function Slides({ slides }) {
  const [on, setOn] = useState(0);
  useEffect(() => {
    if (slides.length < 2 || window.matchMedia('(prefers-reduced-motion: reduce)').matches) return undefined;
    const t = setInterval(() => setOn((i) => (i + 1) % slides.length), 6000);
    return () => clearInterval(t);
  }, [slides.length]);
  return (
    <div className="hero-slides" aria-hidden="true">
      {slides.map((s, i) => {
        // Only load the photos already shown, the current one and the next one.
        const load = s.Photo && (i <= on + 1 || (on === slides.length - 1 && i === 0));
        return <div key={s.Photo || 'base'} className={'slide' + (i === on ? ' on' : '')} style={load ? { backgroundImage: `url('${photoURL(s.Photo)}')` } : undefined} />;
      })}
    </div>
  );
}

function Hero({ d, brand }) {
  if (d.HeroLayers) {
    return (
      <>
        <section className="hero layered ref" style={cssVars(d.HeroStyle)}>
          <img className="hl-bg" src="/photos/hero-layers/background.webp" alt="Green hills of the Western Ghats" />
          <h1 className="sr-only">{brand} · Weekend treks in the Western Ghats</h1>
          <img className="hl-fg" src="/photos/hero-layers/foreground.webp" alt="" aria-hidden="true" />
          {d.HeroSlides && d.HeroSlides.length > 0 && <Slides slides={[{ Photo: '' }, ...d.HeroSlides.filter((s) => s.Photo !== d.HeroPhoto)]} />}
          <div className="hl-corner hl-left">
            <div className="hero-ctas">
              <a className="btn" href="#treks"><Icon name="compass" /> Explore treks</a>
              <Link className="btn ghost-light" to="/leads"><Icon name="users" /> Meet the leads</Link>
            </div>
          </div>
        </section>
        <HeroFeatures />
      </>
    );
  }
  return (
    <section className="hero" style={d.HeroPhoto ? { backgroundImage: `url('${photoURL(d.HeroPhoto)}')` } : undefined}>
      {d.HeroSlides && d.HeroSlides.length > 0 && <Slides slides={d.HeroSlides} />}
      <div className="wrap hero-inner">
        <p className="hero-script">Explore</p>
        <h1 className="hero-title">The Western Ghats</h1>
        <p className="hero-tagline">For the soul that seeks more.</p>
        <p className="hero-intro">Small groups, honest prices and a trek lead who knows every turn of the trail. All you bring is your shoes.</p>
        <div className="hero-ctas">
          <a className="btn big" href="#treks"><Icon name="compass" /> Explore treks</a>
          <Link className="btn big ghost-light" to="/leads"><Icon name="users" /> Meet the leads</Link>
        </div>
      </div>
      <HeroFeatures />
    </section>
  );
}

export default function Home() {
  const page = usePage();
  const d = page.data;
  useTitle('For the soul that seeks more');
  const state = <PageState page={page} />;
  if (!d) return <PageShell>{state}</PageShell>;

  return (
    <PageShell hero={<Hero d={d} brand={d.Brand} />}>
      <section id="treks" className="section">
        <div className="section-head">
          <h2>Where to next?</h2>
        </div>
        <Finder treks={d.FinderData || []} />
      </section>

      {d.Stats && d.Stats.Trekkers > 0 && (
        <section className="trust-strip">
          <div><strong>{d.Stats.TreksDone}</strong><span>Treks completed</span></div>
          <div><strong>{d.Stats.Trekkers}</strong><span>Trekkers guided</span></div>
          {d.Stats.Reviews > 0 && <div><strong>★ {rating(d.Stats.AvgRating)}</strong><span>From {d.Stats.Reviews} reviews</span></div>}
        </section>
      )}

      <section className="section">
        <div className="section-head">
          <h2>Where we wander</h2>
          <p className="muted">Shola grasslands, misty ridges and granite hills. Our kind of country, and soon yours!</p>
        </div>
        <MapView />
      </section>

      {d.PhotoReviews && d.PhotoReviews.length > 0 && (
        <section className="section">
          <div className="section-head">
            <h2>Straight from the trail</h2>
          </div>
          <div className="photo-reviews">
            {d.PhotoReviews.map((r, i) => (
              <figure key={i}>
                <img src={photoURL(r.Photo)} alt={`Photo by ${r.UserName}`} loading="lazy" />
                <figcaption>
                  <Rating n={r.TrekRating} />
                  {r.Comment ? <q>{r.Comment}</q> : null}
                  <b>{r.UserName} · {r.TrekName}</b>
                </figcaption>
              </figure>
            ))}
          </div>
        </section>
      )}
    </PageShell>
  );
}
