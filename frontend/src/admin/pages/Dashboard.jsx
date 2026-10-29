import { Link } from 'react-router-dom';
import { PageState } from '../../components/common.jsx';
import { usePage, useTitle } from '../../lib/usePage.js';
import { fmtDate, inr } from '../../lib/util.js';
import RegTable from '../RegTable.jsx';

export function EventList({ events }) {
  return (
    <div className="table-wrap">
      <table className="table">
        <tbody>
          <tr><th>Event</th><th>Seats</th><th>Status</th><th /></tr>
          {events && events.length ? events.map((e) => (
            <tr key={e.ID}>
              <td><strong>{e.TrekName}</strong><br /><span className="small">{fmtDate(e.StartDate)}{e.LeadName ? ` · Lead: ${e.LeadName}` : ''}</span></td>
              <td className="small">{e.Taken} / {e.Seats}{e.Requested ? <><br /><span className="warn">{e.Requested} waiting</span></> : null}</td>
              <td><span className={`status ${e.Status}`}>{e.Status}</span></td>
              <td><Link to={`/admin/events/${e.ID}`}>Open →</Link></td>
            </tr>
          )) : <tr><td colSpan={4} className="muted">No upcoming events. <Link to="/admin/events/new">Create one</Link>.</td></tr>}
        </tbody>
      </table>
    </div>
  );
}

export default function Dashboard() {
  const page = usePage();
  useTitle('Dashboard');
  const d = page.data;
  if (!d) return <PageState page={page} />;
  const D = d.D;
  return (
    <>
      <div className="c-head"><h1>Dashboard</h1><Link className="btn" to="/admin/events/new">+ New event</Link></div>

      <div className="c-kpis">
        <div className="kpi"><span>Collected</span><strong>₹{inr(D.Collected)}</strong><em>marked paid · all time</em></div>
        <div className="kpi"><span>Still to collect</span><strong>₹{inr(D.ToCollect)}</strong><em>confirmed, not yet paid</em></div>
        <div className="kpi"><span>Trekkers confirmed</span><strong>{D.Trekkers}</strong><em>all time</em></div>
        <div className="kpi"><span>Seats filled</span><strong>{d.Fill}%</strong><em>{D.SeatsTaken} of {D.SeatsTotal} · {D.UpcomingLive} live events</em></div>
      </div>

      {(D.Claimed || D.Requested || D.Pending || D.Drafts || D.LeadApps) ? (
        <section className="c-todo">
          <h2>Needs your attention</h2>
          <ul>
            {D.Claimed ? <li><Link to="/admin/registrations"><b>{D.Claimed}</b> {D.Claimed === 1 ? 'payment' : 'payments'} claimed, check your UPI app and confirm</Link></li> : null}
            {D.Requested ? <li><Link to="/admin/registrations?status=requested"><b>{D.Requested}</b> registration {D.Requested === 1 ? 'request' : 'requests'} waiting for approval</Link></li> : null}
            {D.Pending ? <li><Link to="/admin/events?show=pending"><b>{D.Pending}</b> {D.Pending === 1 ? 'date' : 'dates'} proposed by trek leads</Link></li> : null}
            {D.Drafts ? <li><Link to="/admin/events?show=draft"><b>{D.Drafts}</b> draft {D.Drafts === 1 ? 'event' : 'events'} not yet published</Link></li> : null}
            {D.LeadApps ? <li><Link to="/admin/leads"><b>{D.LeadApps}</b> trek-lead {D.LeadApps === 1 ? 'application' : 'applications'} to review</Link></li> : null}
          </ul>
        </section>
      ) : null}

      <section className="c-block">
        <div className="c-head"><h2>Upcoming events</h2><Link to="/admin/events">All events →</Link></div>
        <EventList events={d.Upcoming} />
      </section>
      <section className="c-block">
        <div className="c-head"><h2>Payments to verify &amp; requests</h2><Link to="/admin/registrations">All →</Link></div>
        <RegTable bookings={d.Bookings} back={d.Back} showEvent />
      </section>
    </>
  );
}
