import { Link, useLocation, useNavigate } from 'react-router-dom';
import { PageState } from '../../components/common.jsx';
import { useSession } from '../../lib/session.jsx';
import { usePage, useTitle } from '../../lib/usePage.js';
import { fmtDate, inr } from '../../lib/util.js';
import RegTable from '../RegTable.jsx';

export default function Registrations() {
  const page = usePage();
  const navigate = useNavigate();
  const { search } = useLocation();
  const { PublicURL } = useSession();
  useTitle('Registrations');
  const d = page.data;
  if (!d) return <PageState page={page} />;

  // A filter changed: keep the others, drop the changed one's old value.
  const setFilter = (key, value) => {
    const q = new URLSearchParams(search);
    if (value && value !== '0') q.set(key, value);
    else q.delete(key);
    navigate(`/admin/registrations${q.toString() ? `?${q}` : ''}`);
  };

  return (
    <>
      <div className="c-head">
        <h1>Registrations</h1>
        <a className="btn ghost" href={`/admin/registrations.csv?${d.Query}`}>Download CSV</a>
      </div>
      <p className="muted">Everyone who has registered, with the details you need on the trail. {d.People} {d.People === 1 ? 'person' : 'people'} in this list (not counting cancelled or rejected).</p>

      <div className="c-filters">
        <label>Event
          <select value={d.Event || 0} onChange={(e) => setFilter('event', e.target.value)}>
            <option value="0">All {d.Scope === 'all' ? 'events' : 'upcoming events'}</option>
            {(d.Events || []).map((e) => <option key={e.ID} value={e.ID}>{e.TrekName} · {fmtDate(e.StartDate)} ({e.Active})</option>)}
          </select>
        </label>
        <label>Status
          <select value={d.Status || ''} onChange={(e) => setFilter('status', e.target.value)}>
            <option value="">Any</option>
            {d.Statuses.map((s) => <option key={s} value={s}>{s}</option>)}
          </select>
        </label>
        <label>Show
          <select value={d.Scope} onChange={(e) => setFilter('scope', e.target.value)}>
            <option value="current">Current &amp; upcoming</option>
            <option value="all">Including past events</option>
          </select>
        </label>
      </div>

      {d.Groups && d.Groups.length ? d.Groups.map((g) => (
        <section className="bk-trek c-block" key={g.Slug}>
          <h2 className="bk-trek-head">
            <a href={`${PublicURL}/treks/${g.Slug}`} target="_blank" rel="noopener noreferrer">{g.Name}</a>{' '}
            <span className="muted small">{g.Count} {g.Count === 1 ? 'registration' : 'registrations'}</span>
          </h2>
          {g.Dates.map((dt) => (
            <div className="bk-date-group" key={dt.Batch.ID}>
              <div className="bk-date-head">
                <strong>{fmtDate(dt.Batch.StartDate)}</strong>
                <span className="muted small">
                  {dt.Batch.Taken} / {dt.Batch.Seats} seats confirmed · Lead: {dt.Batch.LeadName || 'not assigned'}
                  {dt.Due ? <> · <b>₹{inr(dt.Due)} still to collect</b></> : null}
                </span>
                <Link className="small" to={`/admin/events/${dt.Batch.ID}`}>Open event →</Link>
              </div>
              <RegTable bookings={dt.Bookings} back={d.Back} />
            </div>
          ))}
        </section>
      )) : <p className="muted">No registrations here yet.</p>}
    </>
  );
}
