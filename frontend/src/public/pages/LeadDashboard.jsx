import { Link } from 'react-router-dom';
import { PageState } from '../../components/common.jsx';
import { usePage, useTitle } from '../../lib/usePage.js';
import { useSubmit } from '../../lib/useSubmit.js';
import { firstWord, fmtDate, inr, rating } from '../../lib/util.js';
import PageShell from '../PageShell.jsx';

function Batch({ b }) {
  const { submit } = useSubmit();
  return (
    <article className="dash-batch">
      <div className="dash-batch-head">
        <div>
          <h3><Link to={`/treks/${b.TrekSlug}`}>{b.TrekName}</Link></h3>
          <span className="muted small">{fmtDate(b.StartDate)} · {b.Taken} / {b.Seats} seats confirmed</span>
        </div>
        {b.Status === 'live' ? <span className="status approved">Live</span>
          : b.Status === 'pending' ? <span className="status pending">Waiting for approval</span>
          : <span className="status rejected">Not approved</span>}
      </div>
      {b.Status === 'live' ? (
        b.Participants && b.Participants.length ? (
          <div className="table-wrap">
            <table className="table small">
              <tbody>
                <tr><th>Trekker</th><th>Group</th><th>Pickup</th><th>Medical notes</th></tr>
                {b.Participants.map((p) => (
                  <tr key={p.ID}>
                    <td><strong>{p.UserName}</strong>{p.Status === 'requested' && <><br /><span className="status pending">Not confirmed yet</span></>}</td>
                    <td>{p.People} · age {p.Age}</td>
                    <td>{p.Pickup}</td>
                    <td>
                      {p.Medical ? <span className="warn">{p.Medical}</span> : <span className="muted">None</span>}
                      {p.Note ? <><br /><span className="muted">Note: {p.Note}</span></> : null}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : <p className="muted small">No bookings yet. Share the trek page with your friends and followers!</p>
      ) : (
        <form onSubmit={submit(`/lead/dates/${b.ID}/withdraw`, { confirm: 'Withdraw this date?' })} className="inline">
          <button className="linklike danger small">Withdraw this date</button>
        </form>
      )}
    </article>
  );
}

export default function LeadDashboard() {
  const page = usePage();
  const { submit, busy, error } = useSubmit();
  const d = page.data;
  useTitle('Lead dashboard');
  if (!d) return <PageShell><PageState page={page} /></PageShell>;
  const l = d.Lead;
  return (
    <PageShell>
      {l && (
        <header className="page-head head-row">
          <div>
            <p className="eyebrow">Lead dashboard</p>
            <h1>Hi {firstWord(l.Name)}</h1>
            <p className="muted">Your treks, your groups, and new dates you'd like to lead. Trekkers' contact details stay with our team; we'll add you to each trek's WhatsApp group.</p>
          </div>
          <div className="stats">
            <div className="stat"><strong>{l.TreksLed}</strong><span>Treks led</span></div>
            <div className="stat"><strong>{l.TrekkersGuided}</strong><span>Trekkers guided</span></div>
            <div className="stat"><strong>{l.Ratings ? rating(l.AvgRating) : '–'}</strong><span>Avg rating</span></div>
          </div>
        </header>
      )}

      <div className="two-col">
        <div>
          <h2 className="dash-h">Your upcoming treks</h2>
          {d.Upcoming && d.Upcoming.length
            ? d.Upcoming.map((b) => <Batch key={b.ID} b={b} />)
            : <p className="muted">No upcoming treks yet. Propose a date →</p>}

          {d.Past && d.Past.length > 0 && (
            <>
              <h2 className="dash-h">Past treks</h2>
              <ul className="tight-list">
                {d.Past.map((b) => <li key={b.ID}>{b.TrekName} · {fmtDate(b.StartDate)}{b.Status === 'live' ? ` · ${b.Taken} trekkers` : ''}</li>)}
              </ul>
            </>
          )}
        </div>

        <aside className="panel">
          <h3>Propose a date</h3>
          <p className="muted small">Pick a trek and a date you'd like to lead. It goes live after our team approves it. The price is the trek's standard price.</p>
          <form onSubmit={submit('/lead/dates')} className="form">
            {error && <div className="flash err">{error}</div>}
            <label>Trek
              <select name="trek_id" required defaultValue="">
                <option value="">Choose a trek…</option>
                {(d.Treks || []).map((t) => <option key={t.ID} value={t.ID}>{t.Name}{t.Price ? ` · ₹${inr(t.Price)}` : ''}</option>)}
              </select>
            </label>
            <div className="row">
              <label>Date <input type="date" name="start_date" min={d.MinDate} required /></label>
              <label>Group size <input type="number" name="seats" min="2" max="40" defaultValue="15" required /></label>
            </div>
            <button className="btn" disabled={busy}>Send for approval</button>
          </form>
          <p className="muted small" style={{ marginTop: 12 }}><Link to="/lead-with-us">Edit your public profile →</Link></p>
        </aside>
      </div>
    </PageShell>
  );
}
