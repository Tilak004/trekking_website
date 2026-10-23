import { Link } from 'react-router-dom';
import { Avatar, PageState, Reviews } from '../../components/common.jsx';
import { usePage, useTitle } from '../../lib/usePage.js';
import { fmtDate, lines, paras, rating } from '../../lib/util.js';
import PageShell from '../PageShell.jsx';
import NotFound from './NotFound.jsx';

const FOUNDER_LEADS = [
  { name: 'Prahladh', photo: 'prahladh.jpg', alt: 'Prahladh in the Himalayas', role: 'Ironman finisher',
    text: 'Knows pacing inside out. Sets a rhythm the whole group can keep, so everyone reaches the top together.' },
  { name: 'Lepakshi', photo: 'lepakshi.jpg', alt: "Lepakshi with a finisher's medal at the BMF Ultra", role: '24-hour adventure racer',
    text: 'Calm when the trail gets long or the weather turns, and great at keeping spirits up on the climb.' },
  { name: 'Tilak', photo: 'tilak.jpg', alt: 'Tilak jumping in the snow on a Himalayan trek', role: 'Western Ghats regular',
    text: 'Has walked these trails many times: where it gets slippery, where to rest and where the best view is.' },
];

export function Leads() {
  const page = usePage();
  const d = page.data;
  useTitle('Trek Leads');
  if (!d) return <PageShell><PageState page={page} /></PageShell>;
  return (
    <PageShell>
      <header className="page-head head-row">
        <div>
          <h1>Trek leads</h1>
          <p className="muted">Right now, every trek is led by one of us three founders. Here's who'll be walking beside you.</p>
        </div>
        <Link className="btn" to="/lead-with-us">Become a trek lead</Link>
      </header>
      {/* Until we hire trek leads, the three founders lead every trek. */}
      <div className="founder-grid">
        {FOUNDER_LEADS.map((f) => (
          <figure className="founder" key={f.name}>
            <img src={`/photos/founders/${f.photo}`} alt={f.alt} loading="lazy" />
            <figcaption><b>{f.name}</b><em>{f.role}</em><span>{f.text}</span></figcaption>
          </figure>
        ))}
      </div>

      <section className="section why">
        <div className="section-head">
          <h2>Why trek with us</h2>
        </div>
        <div className="why-grid">
          <div className="why-item"><h3>People who know the trail</h3><p>Our leads have walked these hills many times over. They set a pace everyone can enjoy and make sure nobody gets left behind.</p></div>
          <div className="why-item"><h3>Small groups, real conversations</h3><p>We cap every batch, so it feels like a trip with friends, not a crowd on a hill.</p></div>
          <div className="why-item"><h3>Ask first, pay later</h3><p>Request a seat for free. We'll call you, answer every question, and only then share payment details.</p></div>
        </div>
      </section>

      <section className="cta-band">
        <h2>Ready when you are!</h2>
        <p>No account needed. Pick a date, send a request and we'll take it from there.</p>
        <Link className="btn big" to="/#treks">See upcoming treks</Link>
      </section>
    </PageShell>
  );
}

export function LeadProfile() {
  const page = usePage();
  const d = page.data;
  useTitle(d && d.Lead.Name);
  if (page.notFound) return <NotFound />;
  if (!d) return <PageShell><PageState page={page} /></PageShell>;
  const l = d.Lead;
  const achievements = lines(l.Achievements);
  const certs = lines(l.Certifications);
  return (
    <PageShell>
      <p className="crumb"><Link to="/leads">← All trek leads</Link></p>
      <header className="page-head profile-head">
        <Avatar person={l} big />
        <div>
          <h1>{l.Name}{l.Verified && <> <span className="verified" title="ID and experience checked by our team">✓ Verified</span></>}</h1>
          <p className="muted">{l.Headline}{l.City ? ` · ${l.City}` : ''}</p>
        </div>
        <div className="stats">
          <div className="stat"><strong>{l.TreksLed}</strong><span>Treks led</span></div>
          <div className="stat"><strong>{l.TrekkersGuided}</strong><span>Trekkers guided</span></div>
          <div className="stat"><strong>{l.Ratings ? rating(l.AvgRating) : '–'}</strong><span>Avg rating{l.Ratings ? ` (${l.Ratings})` : ''}</span></div>
        </div>
      </header>

      <div className="two-col">
        <div>
          {l.Bio && <><h3>About</h3><div className="prose">{paras(l.Bio).map((p, i) => <p key={i}>{p}</p>)}</div></>}
          {achievements.length > 0 && (
            <>
              <h3>Achievements</h3>
              <ul className="tight-list">{achievements.map((x, i) => <li key={i}>{x}</li>)}</ul>
            </>
          )}
          {certs.length > 0 && (
            <>
              <h3>Certifications</h3>
              <ul className="tight-list">{certs.map((x, i) => <li key={i}>{x}</li>)}</ul>
            </>
          )}
          <h3>Reviews from trekkers</h3>
          <Reviews reviews={d.Reviews} />
        </div>
        <aside className="panel">
          <h3>Upcoming treks</h3>
          {d.Upcoming && d.Upcoming.length
            ? d.Upcoming.map((u, i) => (
              <p key={i}><Link to={`/treks/${u.TrekSlug}`}>{u.TrekName}</Link><br /><span className="muted small">{fmtDate(u.StartDate)}</span></p>
            ))
            : <p className="muted">No upcoming treks scheduled.</p>}
        </aside>
      </div>
    </PageShell>
  );
}
