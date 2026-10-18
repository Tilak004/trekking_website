import { PageState } from '../../components/common.jsx';
import { usePage, useTitle } from '../../lib/usePage.js';
import { inr, rating } from '../../lib/util.js';

function HBars({ bars }) {
  if (!bars || !bars.length) return <p className="muted">No data yet.</p>;
  return (
    <div className="hbars">
      {bars.map((b) => (
        <div className="hbar" key={b.Label}>
          <span>{b.Label}</span>
          <div><i style={{ width: `${b.Pct}%` }} /></div>
          <b>{b.Value}</b>
        </div>
      ))}
    </div>
  );
}

function Columns({ bars }) {
  if (!bars || !bars.length) return <p className="muted">No data yet.</p>;
  return (
    <div className="cols">
      {bars.map((b) => (
        <div className="col" key={b.Label} title={`${b.Label}: ${b.Value}${b.Sub ? ` · ${b.Sub}` : ''}`}>
          <b>{inr(b.Value)}</b>
          <div className="col-bar"><i style={{ height: `${b.Pct}%` }} /></div>
          <span>{b.Label}</span>
          {b.Sub ? <em>{b.Sub}</em> : null}
        </div>
      ))}
    </div>
  );
}

export default function Analytics() {
  const page = usePage();
  useTitle('Analytics');
  const d = page.data;
  if (!d) return <PageState page={page} />;
  return (
    <>
      <div className="c-head"><h1>Analytics</h1></div>
      <p className="note"><strong>Collected</strong> is money you have marked as paid on a booking. <strong>Booked value</strong> is trek price × people on approved or completed bookings.</p>

      <h2 className="c-h">Money</h2>
      <div className="c-kpis">
        <div className="kpi"><span>Collected</span><strong>₹{inr(d.Collected)}</strong><em>marked paid</em></div>
        <div className="kpi"><span>Still to collect</span><strong>₹{inr(d.ToCollect)}</strong><em>confirmed, not yet paid</em></div>
        <div className="kpi"><span>Booked value</span><strong>₹{inr(d.ConfirmedValue)}</strong><em>approved + completed</em></div>
        <div className="kpi"><span>In the pipeline</span><strong>₹{inr(d.PipelineValue)}</strong><em>requests not yet approved</em></div>
      </div>
      <div className="c-kpis">
        <div className="kpi"><span>Lost</span><strong>₹{inr(d.LostValue)}</strong><em>cancelled + rejected</em></div>
        <div className="kpi"><span>Avg. booking value</span><strong>₹{inr(d.AvgValue)}</strong></div>
      </div>

      <div className="c-two">
        <section className="c-block"><h3>Collected by month (when marked paid)</h3><Columns bars={d.CollectedMonthly} /></section>
        <section className="c-block"><h3>Booked value by month of trek</h3><Columns bars={d.Monthly} /></section>
      </div>

      <h2 className="c-h">People</h2>
      <div className="c-kpis">
        <div className="kpi"><span>Trekkers confirmed</span><strong>{d.People}</strong><em>across {d.Bookings} bookings</em></div>
        <div className="kpi"><span>Avg. group per booking</span><strong>{Number(d.AvgGroup).toFixed(1)}</strong></div>
        <div className="kpi"><span>Repeat trekkers</span><strong>{d.Repeat}</strong><em>{d.RepeatPct}% of {d.Unique} who booked</em></div>
        <div className="kpi"><span>Seats filled</span><strong>{d.FillPct}%</strong><em>across all published events</em></div>
      </div>

      <div className="c-two">
        <section className="c-block">
          <h3>Bookings by status</h3>
          <div className="hbars">
            {d.Funnel.map((f) => (
              <div className="hbar" key={f.Status}>
                <span>{f.Status}</span>
                <div><i className={`s-${f.Status}`} style={{ width: `${f.Pct}%` }} /></div>
                <b>{f.Bookings}</b>
              </div>
            ))}
          </div>
        </section>
        <section className="c-block"><h3>Age of the person booking</h3><HBars bars={d.Ages} /></section>
        <section className="c-block"><h3>Where people board (people)</h3><HBars bars={d.Pickups} /></section>
        <section className="c-block"><h3>New accounts per month</h3><Columns bars={d.Signups} /></section>
      </div>

      <h2 className="c-h">By trek</h2>
      <div className="table-wrap">
        <table className="table">
          <tbody>
            <tr><th>Trek</th><th>Events</th><th>People</th><th>Seats filled</th><th>Booked value</th><th>Collected</th><th>Rating</th></tr>
            {d.Treks && d.Treks.length ? d.Treks.map((t) => (
              <tr key={t.Name}>
                <td><strong>{t.Name}</strong></td><td>{t.Events}</td><td>{t.People}</td>
                <td className="small"><div className="meter"><i style={{ width: `${t.Fill}%` }} /></div>{t.Fill}%</td>
                <td>₹{inr(t.Value)}</td><td>₹{inr(t.Collected)}</td>
                <td>{t.Ratings ? `${rating(t.Rating)} (${t.Ratings})` : '—'}</td>
              </tr>
            )) : <tr><td colSpan={7} className="muted">No treks yet.</td></tr>}
          </tbody>
        </table>
      </div>

      <h2 className="c-h">By trek lead</h2>
      <div className="table-wrap">
        <table className="table">
          <tbody>
            <tr><th>Lead</th><th>Events</th><th>People guided</th><th>Booked value</th></tr>
            {d.Leads && d.Leads.length ? d.Leads.map((l) => (
              <tr key={l.Name}><td><strong>{l.Name}</strong></td><td>{l.Events}</td><td>{l.People}</td><td>₹{inr(l.Value)}</td></tr>
            )) : <tr><td colSpan={4} className="muted">No published events with a lead yet.</td></tr>}
          </tbody>
        </table>
      </div>
    </>
  );
}
