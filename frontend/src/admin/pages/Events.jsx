import { Link, useParams } from 'react-router-dom';
import { PageState } from '../../components/common.jsx';
import { usePage, useTitle } from '../../lib/usePage.js';
import { useSubmit } from '../../lib/useSubmit.js';
import { fmtDate, inr } from '../../lib/util.js';
import RegTable from '../RegTable.jsx';

function PostButton({ path, fields, className, children, confirm }) {
  const { submit } = useSubmit();
  return (
    <form onSubmit={submit(path, confirm ? { confirm } : undefined)} className="inline">
      {Object.entries(fields || {}).map(([k, v]) => <input key={k} type="hidden" name={k} value={v} />)}
      <button className={className}>{children}</button>
    </form>
  );
}

// CancelEvent cancels a date from our side (too few trekkers, weather, closure):
// it comes off the website and every registration gets a full refund.
function CancelEvent({ e }) {
  const { submit, busy } = useSubmit();
  return (
    <section className="panel danger-zone">
      <strong>Cancel this event</strong>
      <p className="muted small">Use this when the trek can't run (for example too few seats filled). The date comes off the website,
        all {e.Active} registration{e.Active === 1 ? '' : 's'} are cancelled, and everyone who paid is owed a <b>full refund</b>
        (₹{inr(e.Collected)} collected so far). You then send each trekker a ready-written WhatsApp message and mark refunds as sent.</p>
      <form onSubmit={submit(`/admin/events/${e.ID}/cancel`, { confirm: `Cancel ${e.TrekName} on ${fmtDate(e.StartDate)}? ${e.Active} registration(s) will be cancelled with full refunds (₹${inr(e.Collected)} collected). This can't be undone.` })} className="cancel-event">
        <label>Reason (goes into the message to trekkers)
          <select name="reason" defaultValue="low">
            <option value="low">Not enough trekkers</option>
            <option value="weather">Bad weather</option>
            <option value="closed">Trail closed by the forest department</option>
            <option value="other">Other reasons</option>
          </select>
        </label>
        <button className="btn small act-rejected" disabled={busy}>Cancel event &amp; refund everyone</button>
      </form>
    </section>
  );
}

export function Events() {
  const page = usePage();
  useTitle('Events');
  const d = page.data;
  if (!d) return <PageState page={page} />;
  return (
    <>
      <div className="c-head"><h1>Events</h1><Link className="btn" to="/admin/events/new">+ New event</Link></div>
      <p className="muted">An event is a trek on a specific date. Drafts are only visible here until you publish them.</p>
      <nav className="tabs">
        {d.Filters.map((f) => <Link key={f.Key} to={`/admin/events?show=${f.Key}`} className={f.Key === d.Filter ? 'on' : undefined}>{f.Label}</Link>)}
      </nav>
      <div className="table-wrap">
        <table className="table">
          <tbody>
            <tr><th>Event</th><th>Lead</th><th>Seats</th><th>Money</th><th>Status</th><th /></tr>
            {d.Events && d.Events.length ? d.Events.map((e) => (
              <tr key={e.ID} className={e.Past ? 'past' : undefined}>
                <td><strong>{e.TrekName}</strong><br /><span className="small">{fmtDate(e.StartDate)} · {e.Region}</span></td>
                <td className="small">{e.LeadName || '—'}</td>
                <td className="small">
                  <div className="meter" title={`${e.Fill}% full`}><i style={{ width: `${e.Fill}%` }} /></div>
                  {e.Taken} / {e.Seats} seats{e.Requested ? <> · <span className="warn">{e.Requested} waiting</span></> : null}
                </td>
                <td className="small">₹{inr(e.Collected)} collected{e.Due ? <><br /><span className="muted">₹{inr(e.Due)} to collect</span></> : null}</td>
                <td><span className={`status ${e.Status}`}>{e.Status}</span>{e.TrekHidden ? <><br /><span className="muted small" title="The trek itself is hidden from trekkers">trek hidden</span></> : null}</td>
                <td className="actions">
                  <Link className="btn small ghost" to={`/admin/events/${e.ID}`}>Open</Link>
                  {e.IsPending ? (
                    <>
                      <PostButton path={`/admin/batches/${e.ID}/review`} fields={{ status: 'live' }} className="btn small act-approved">Approve</PostButton>
                      <PostButton path={`/admin/batches/${e.ID}/review`} fields={{ status: 'rejected' }} className="btn small act-rejected">Reject</PostButton>
                    </>
                  ) : e.IsDraft ? (
                    <PostButton path={`/admin/events/${e.ID}/status`} fields={{ status: 'live' }} className="btn small">Publish</PostButton>
                  ) : null}
                  {!e.Active && !e.Requested && (
                    <PostButton path={`/admin/events/${e.ID}/delete`} className="linklike danger small" confirm={`Delete the ${e.TrekName} event on ${fmtDate(e.StartDate)}? This can't be undone.`}>Delete</PostButton>
                  )}
                </td>
              </tr>
            )) : <tr><td colSpan={6} className="muted">Nothing here. <Link to="/admin/events/new">Create an event</Link>.</td></tr>}
          </tbody>
        </table>
      </div>
    </>
  );
}

// Create a new event, or view/edit one (with its publish controls and registrations).
export function EventPage() {
  const { id } = useParams();
  const page = usePage();
  const { submit, busy, error } = useSubmit();
  const d = page.data;
  useTitle(d && (d.Event.ID ? `${d.Event.TrekName} · ${fmtDate(d.Event.StartDate)}` : 'New event'));
  if (page.notFound) return <p className="muted">Event not found. <Link to="/admin/events">Back to events</Link></p>;
  if (!d) return <PageState page={page} />;
  const e = d.Event;
  const isNew = !e.ID;

  return (
    <>
      <p className="crumb"><Link to="/admin/events">← Events</Link></p>
      <div className="c-head">
        <h1>{isNew ? 'New event' : e.TrekName}</h1>
        {!isNew && <span className={`status ${e.Status} big`}>{e.Status}</span>}
      </div>

      {!isNew && (
        <>
          <div className="c-kpis">
            <div className="kpi"><span>Date</span><strong className="sm">{fmtDate(e.StartDate)}</strong></div>
            <div className="kpi"><span>Seats confirmed</span><strong>{e.Taken} / {e.Seats}</strong><em>{e.Left} left</em></div>
            <div className="kpi"><span>Waiting for approval</span><strong>{e.Requested}</strong></div>
            <div className="kpi"><span>Collected</span><strong>₹{inr(e.Collected)}</strong><em>{e.Due ? `₹${inr(e.Due)} still to collect · ` : ''}₹{inr(e.Price)} per person</em></div>
          </div>

          <section className={`c-publish ${e.Status}`}>
            {e.IsDraft ? (
              <>
                <div><strong>This event is a draft.</strong> Trekkers can't see or book it yet.{e.TrekHidden && <> <span className="warn">The trek itself is hidden, so also make it visible in <Link to={`/admin/treks/${e.TrekID}`}>trek settings</Link>.</span></>}</div>
                <PostButton path={`/admin/events/${e.ID}/status`} fields={{ status: 'live' }} className="btn">Publish event</PostButton>
              </>
            ) : e.IsPending ? (
              <>
                <div><strong>Proposed by a trek lead.</strong> Approve it to publish, or reject it.</div>
                <div className="row-btns">
                  <PostButton path={`/admin/batches/${e.ID}/review`} fields={{ status: 'live' }} className="btn act-approved">Approve &amp; publish</PostButton>
                  <PostButton path={`/admin/batches/${e.ID}/review`} fields={{ status: 'rejected' }} className="btn ghost act-rejected">Reject</PostButton>
                </div>
              </>
            ) : e.IsLive ? (
              <>
                <div><strong>Published.</strong> Trekkers can see and book this event.{e.TrekHidden && <> <span className="warn">The trek is hidden, so it won't show on the site until you make it visible in <Link to={`/admin/treks/${e.TrekID}`}>trek settings</Link>.</span></>}</div>
                <PostButton path={`/admin/events/${e.ID}/status`} fields={{ status: 'draft' }} className="btn ghost">Unpublish</PostButton>
              </>
            ) : e.IsCancelled ? (
              <div><strong>Cancelled.</strong> This date is off the website. Send each trekker the cancellation message below and mark refunds as sent.</div>
            ) : (
              <div><strong>Rejected.</strong> This proposed date is not visible to trekkers.</div>
            )}
          </section>
        </>
      )}

      <form key={`${e.ID}-${e.Status}-${e.StartDate}-${e.Seats}`} onSubmit={submit(isNew ? '/admin/events/new' : `/admin/events/${id}`)} className="form panel c-form">
        <h2>{isNew ? 'Details' : 'Event details'}</h2>
        {error && <div className="flash err">{error}</div>}
        <div className="row">
          <label>Trek
            {isNew ? (
              <>
                <select name="trek_id" required defaultValue={e.TrekID || ''}>
                  <option value="">Choose a trek…</option>
                  {d.Treks.map((t) => <option key={t.ID} value={t.ID}>{t.Name} · ₹{inr(t.Price)}{!t.Active ? ' (hidden)' : ''}</option>)}
                </select>
                <span className="hint">Not listed? <Link to="/admin/treks/new">Create a new trek</Link> first, with its itinerary and price, then come back.</span>
              </>
            ) : (
              <>
                <input value={e.TrekName} disabled />
                <span className="hint">The trek can't change after the event is created. <Link to={`/admin/treks/${e.TrekID}`}>Edit trek details (itinerary, price, photos…)</Link></span>
              </>
            )}
          </label>
          <label>Start date <input type="date" name="start_date" defaultValue={e.StartDate} min={isNew ? d.Today : undefined} required /></label>
          <label>Group size (seats) <input type="number" name="seats" defaultValue={e.Seats} min="1" max="500" required /></label>
          <label>Trek lead
            <select name="lead_id" defaultValue={e.LeadID || 0}>
              <option value="0">— assign later —</option>
              {d.Leads.map((l) => <option key={l.ID} value={l.ID}>{l.Name}</option>)}
            </select>
          </label>
        </div>
        <div className="row-btns">
          {isNew ? (
            <>
              <button className="btn ghost" disabled={busy}>Save as draft</button>
              <button className="btn" name="action" value="publish" disabled={busy}>Save &amp; publish</button>
            </>
          ) : (
            <>
              <button className="btn" disabled={busy}>Save changes</button>
              {e.IsDraft && <button className="btn ghost" name="action" value="publish" disabled={busy}>Save &amp; publish</button>}
            </>
          )}
        </div>
      </form>

      {!isNew && (
        <>
          <section className="c-block">
            <div className="c-head">
              <h2>Registrations</h2>
              <a href={`/admin/registrations.csv?event=${e.ID}`}>Download CSV</a>
            </div>
            <RegTable bookings={d.Bookings} back={d.Back} />
          </section>
          {(e.IsLive || e.IsDraft) && e.Active > 0 && <CancelEvent e={e} />}
          {!e.IsCancelled && <section className="panel danger-zone">
            <strong>Delete this event</strong>
            {d.Bookings && d.Bookings.length > 0 ? (
              <p className="muted small">This event has registrations, so it can't be deleted (that would lose booking and payment records). If the trek can't run, use <b>Cancel this event</b> above: it comes off the site and everyone is refunded, while the records are kept.</p>
            ) : (
              <>
                <p className="muted small">Removes this date for good. This can't be undone.</p>
                <PostButton path={`/admin/events/${e.ID}/delete`} className="btn small act-rejected" confirm={`Delete the ${e.TrekName} event on ${fmtDate(e.StartDate)}? This can't be undone.`}>Delete event</PostButton>
              </>
            )}
          </section>}
        </>
      )}
    </>
  );
}
