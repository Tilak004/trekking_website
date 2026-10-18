import { Link } from 'react-router-dom';
import { Avatar, PageState, Reviews } from '../../components/common.jsx';
import { useSession } from '../../lib/session.jsx';
import { usePage, useTitle } from '../../lib/usePage.js';
import { fmtDate, lines, paras, rating } from '../../lib/util.js';
import PageShell from '../PageShell.jsx';
import NotFound from './NotFound.jsx';

export function Leads() {
  const page = usePage();
  const { User } = useSession();
  const d = page.data;
  useTitle('Trek Leads');
  if (!d) return <PageShell><PageState page={page} /></PageShell>;
  return (
    <PageShell>
      <header className="page-head head-row">
        <div>
          <h1>Trek leads</h1>
          <p className="muted">The people who take you up the mountain. Every profile builds up from real treks led and real reviews from trekkers.</p>
        </div>
        <Link className="btn" to="/lead-with-us">Become a trek lead</Link>
      </header>
      <div className="grid leads-grid">
        {d.Leads && d.Leads.length ? d.Leads.map((l) => (
          <Link className="card" to={`/leads/${l.Slug}`} key={l.Slug}>
            <Avatar person={l} />
            <h3>{l.Name}{l.Verified && <> <span className="verified sm">✓</span></>}</h3>
            <p className="muted">{l.Headline}</p>
            <span className="card-link">View profile →</span>
          </Link>
        )) : <p className="muted">No trek leads listed yet.</p>}
      </div>

      <section className="section why">
        <div className="section-head">
          <p className="eyebrow">Why trek with us</p>
          <h2>Built by people who race mountains</h2>
        </div>
        <div className="why-grid">
          <div className="why-item"><div className="why-icon">01</div><h3>Athlete-led groups</h3><p>Your lead is an endurance athlete who knows pacing, safety and how to get every member to the top.</p></div>
          <div className="why-item"><div className="why-icon">02</div><h3>Small, personal batches</h3><p>Limited seats per batch, so the trail feels like yours and nobody gets left behind.</p></div>
          <div className="why-item"><div className="why-icon">03</div><h3>Honest, simple booking</h3><p>Send a request for free. We confirm your seat personally before any payment.</p></div>
        </div>
      </section>

      <section className="cta-band">
        <h2>Your next summit is a weekend away.</h2>
        <p>Create a free account and request a seat in under two minutes.</p>
        {User ? <Link className="btn big" to="/#treks">Find a trek</Link> : <Link className="btn big" to="/signup">Sign up free</Link>}
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
