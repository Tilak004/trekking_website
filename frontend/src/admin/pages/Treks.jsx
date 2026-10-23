import { Link, useParams } from 'react-router-dom';
import { PageState } from '../../components/common.jsx';
import { useSession } from '../../lib/session.jsx';
import { usePage, useTitle } from '../../lib/usePage.js';
import { useSubmit } from '../../lib/useSubmit.js';
import { fmtDate, inr, rating } from '../../lib/util.js';

export function Treks() {
  const page = usePage();
  const { submit } = useSubmit();
  useTitle('Treks');
  const d = page.data;
  if (!d) return <PageState page={page} />;
  return (
    <>
      {d.Pending && d.Pending.length > 0 && (
        <section className="panel pending-dates">
          <h2>Dates waiting for approval <span className="status pending">{d.Pending.length}</span></h2>
          <div className="table-wrap">
            <table className="table">
              <tbody>
                <tr><th>Trek</th><th>Date</th><th>Group size</th><th>Proposed by</th><th /></tr>
                {d.Pending.map((p) => (
                  <tr key={p.ID}>
                    <td><strong>{p.TrekName}</strong></td>
                    <td>{fmtDate(p.StartDate)}</td>
                    <td>{p.Seats}</td>
                    <td>{p.LeadName || '–'}<br /><span className="muted small">{p.ProposedAt}</span></td>
                    <td className="actions">
                      <form onSubmit={submit(`/admin/batches/${p.ID}/review`)} className="inline"><input type="hidden" name="status" value="live" /><button className="btn small act-approved">Approve</button></form>
                      <form onSubmit={submit(`/admin/batches/${p.ID}/review`)} className="inline"><input type="hidden" name="status" value="rejected" /><button className="btn small act-rejected">Reject</button></form>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </section>
      )}
      <div className="head-row">
        <h1>Treks</h1>
        <span className="head-actions"><Link className="btn ghost" to="/admin/trek-defaults">Trek defaults</Link> <Link className="btn" to="/admin/treks/new">+ New trek</Link></span>
      </div>
      <div className="table-wrap">
        <table className="table">
          <tbody>
            <tr><th>Trek</th><th>Difficulty</th><th>Price</th><th>Next batch</th><th>Rating</th><th>Visible</th><th /></tr>
            {d.Treks && d.Treks.length ? d.Treks.map((t) => (
              <tr key={t.ID}>
                <td><strong>{t.Name}</strong><br /><span className="muted small">{t.Region}</span></td>
                <td>{t.Difficulty}</td>
                <td>₹{inr(t.Price)}</td>
                <td>{t.NextBatch ? fmtDate(t.NextBatch) : <span className="muted">none</span>}</td>
                <td>{t.Ratings ? `${rating(t.AvgRating)} (${t.Ratings})` : '–'}</td>
                <td>{t.Active ? 'Yes' : <span className="muted">Hidden</span>}</td>
                <td><Link to={`/admin/treks/${t.ID}`}>Edit &amp; batches →</Link></td>
              </tr>
            )) : <tr><td colSpan={7} className="muted">No treks yet.</td></tr>}
          </tbody>
        </table>
      </div>
    </>
  );
}

function BatchRow({ b, leads, today }) {
  const { submit } = useSubmit();
  return (
    <tr className={b.StartDate < today ? 'past' : undefined}>
      <td><Link to={`/admin/events/${b.ID}`}>{fmtDate(b.StartDate)}</Link>{b.Status !== 'live' && <> <span className={`status ${b.Status}`}>{b.Status}</span></>}</td>
      <td>{b.Taken} / {b.Seats}</td>
      <td>{b.Bookings}</td>
      <td>
        <form onSubmit={submit(`/admin/batches/${b.ID}/lead`)} className="inline-form">
          <select name="lead_id" defaultValue={b.LeadID || 0}>
            <option value="0">— no lead —</option>
            {leads.map((l) => <option key={l.ID} value={l.ID}>{l.Name}</option>)}
          </select>
          <input type="number" name="seats" defaultValue={b.Seats} min="1" className="narrow" title="Seats" />
          <button className="btn small ghost">Update</button>
        </form>
      </td>
      <td>
        {b.Bookings === 0 && (
          <form onSubmit={submit(`/admin/batches/${b.ID}/delete`, { confirm: 'Delete this batch?' })} className="inline">
            <button className="linklike danger">Delete</button>
          </form>
        )}
      </td>
    </tr>
  );
}

// UseDefault copies the common text from Trek defaults into one field of the
// trek form. It asks before replacing text that is already there.
function UseDefault({ field, defaults }) {
  const text = defaults && defaults[field];
  if (!text) return null;
  const fill = (e) => {
    const box = e.currentTarget.closest('label').querySelector('textarea');
    if (box.value.trim() && box.value.trim() !== text.trim() && !window.confirm('Replace what is in this box with the trek defaults?')) return;
    box.value = text;
  };
  return <button type="button" className="linklike small use-default" onClick={fill}>Use trek defaults</button>;
}

export function TrekDefaults() {
  const page = usePage();
  const { submit, busy, error } = useSubmit();
  useTitle('Trek defaults');
  const d = page.data;
  if (!d) return <PageState page={page} />;
  const v = d.Defaults || {};
  return (
    <>
      <p className="crumb"><Link to="/admin/treks">← All treks</Link></p>
      <h1>Trek defaults</h1>
      <p className="muted">Text that's the same for most treks. New treks start with it filled in, and on any trek you can press
        "Use trek defaults" to pull it in. Each trek keeps its own copy, so you can still change it per trek.</p>
      {error && <div className="flash err">{error}</div>}
      <form onSubmit={submit('/admin/trek-defaults')} className="form panel">
        <label>Not included (one per line) <textarea name="exclusions" rows="6" defaultValue={v.exclusions} /></label>
        <label>What to carry (one per line) <textarea name="things_to_carry" rows="10" defaultValue={v.things_to_carry} />
          <span className="hint">Optional groups: a line ending in a colon, like <code>Clothing:</code>, starts a group.</span></label>
        <label>FAQs <textarea name="faqs" rows="12" defaultValue={v.faqs} />
          <span className="hint">First line is the question, the next lines are the answer. Leave a blank line between questions.</span></label>
        <button className="btn" disabled={busy}>Save defaults</button>
      </form>
    </>
  );
}

export function TrekForm() {
  const { id } = useParams();
  const page = usePage();
  const { PublicURL } = useSession();
  const { submit, busy, error } = useSubmit();
  const batchForm = useSubmit();
  const trailForm = useSubmit();
  const d = page.data;
  useTitle(d && (d.Trek.ID ? `Edit ${d.Trek.Name}` : 'New trek'));
  if (page.notFound) return <p className="muted">Trek not found. <Link to="/admin/treks">Back to treks</Link></p>;
  if (!d) return <PageState page={page} />;
  const t = d.Trek;
  const isNew = !t.ID;

  return (
    <>
      <p className="crumb">
        <Link to="/admin/treks">← All treks</Link>
        {!isNew && <> · <a href={`${PublicURL}/treks/${t.Slug}`} target="_blank" rel="noopener noreferrer">View public page</a></>}
      </p>
      <h1>{isNew ? 'New trek' : `Edit ${t.Name}`}</h1>

      {!isNew && (
        <form key={`trail-${t.ID}-${t.TrailUpdated}`} onSubmit={trailForm.submit(`/admin/treks/${t.ID}/trail`)} className="form panel trail-form">
          <strong>Trail status</strong> <span className="muted small">Shown on the trek page and cards.{t.TrailUpdated ? ` Last updated ${fmtDate(t.TrailUpdated)}.` : ''}</span>
          {trailForm.error && <div className="flash err">{trailForm.error}</div>}
          <div className="row">
            <label>Status
              <select name="trail_status" defaultValue={t.TrailStatus || ''}>
                <option value="">Not shown</option>
                <option value="open">Open</option>
                <option value="closed">Closed</option>
              </select>
            </label>
            <label>Note (optional) <input name="trail_note" defaultValue={t.TrailNote} maxLength={160} placeholder="e.g. Closed by forest department till 15 Nov" /></label>
          </div>
          <button className="btn small" disabled={trailForm.busy}>Update status</button>
        </form>
      )}

      <form key={t.ID || 'new'} onSubmit={submit(isNew ? '/admin/treks/new' : `/admin/treks/${id}`)} className="form panel">
        {(error || d.Err) && <div className="flash err">{error || d.Err}</div>}
        <div className="row">
          <label>Name <input name="name" defaultValue={t.Name} required maxLength={120} /></label>
          <label>URL slug <input name="slug" defaultValue={t.Slug} placeholder="auto from name" maxLength={120} /></label>
        </div>
        <div className="row">
          <label>Region <input name="region" defaultValue={t.Region} placeholder="e.g. Chikkamagaluru" /></label>
          <label>Duration <input name="duration" defaultValue={t.Duration} placeholder="e.g. 2 days / 1 night" /></label>
        </div>
        <div className="row">
          <label>Distance (km) <input type="number" name="distance_km" defaultValue={t.DistanceKm} min="0" /></label>
          <label>Max altitude (m) <input type="number" name="max_altitude" defaultValue={t.MaxAltitude} min="0" /></label>
          <label>Difficulty
            <select name="difficulty" defaultValue={t.Difficulty}>
              {d.Difficulties.map((x) => <option key={x}>{x}</option>)}
            </select>
          </label>
          <label>Price per person (₹) <input type="number" name="price" defaultValue={t.Price} min="0" required /></label>
        </div>
        <label>Short summary (shown on cards) <input name="summary" defaultValue={t.Summary} maxLength={300} /></label>
        <label>Full description <textarea name="description" rows="8" defaultValue={t.Description} /></label>
        <label>Best season <input name="best_season" defaultValue={t.BestSeason} placeholder="e.g. September – February" maxLength={80} /></label>
        <label>Itinerary
          <textarea name="itinerary" rows="10" defaultValue={t.Itinerary} />
          <span className="hint">One step per paragraph. The first line is the time or day, the next lines are the details. Leave a blank line between steps.</span>
        </label>
        <div className="row">
          <label>Included (one per line) <textarea name="inclusions" rows="6" defaultValue={t.Inclusions} /></label>
          <label>Not included (one per line) <textarea name="exclusions" rows="6" defaultValue={t.Exclusions} /><UseDefault field="exclusions" defaults={d.Defaults} /></label>
        </div>
        <label>What to carry (one per line) <textarea name="things_to_carry" rows="8" defaultValue={t.ThingsToCarry} placeholder={'Clothing:\nFull-sleeve T-shirt\nRain jacket\nFootwear:\nTrekking shoes with good grip'} />
          <span className="hint">Optional groups: a line ending in a colon, like <code>Clothing:</code>, starts a group.</span><UseDefault field="things_to_carry" defaults={d.Defaults} /></label>
        <label>FAQs
          <textarea name="faqs" rows="8" defaultValue={t.FAQs} />
          <span className="hint">First line is the question, the next lines are the answer. Leave a blank line between questions.</span>
          <UseDefault field="faqs" defaults={d.Defaults} />
        </label>
        <label>Cover photo
          <select name="photo" defaultValue={t.Photo}>
            <option value="">— none (uses a mountain graphic) —</option>
            {(d.Photos || []).map((p) => <option key={p}>{p}</option>)}
          </select>
          <span className="hint">Drop images into the <code>photos</code> folder next to the app, then refresh this page.</span>
        </label>
        <div className="row">
          <label>Latitude <input name="lat" defaultValue={t.Lat || ''} placeholder="e.g. 13.13" inputMode="decimal" /></label>
          <label>Longitude <input name="lng" defaultValue={t.Lng || ''} placeholder="e.g. 75.27" inputMode="decimal" /></label>
          <label>Finder tags <input name="tags" defaultValue={t.Tags} placeholder="waterfalls, grasslands, sunrise, forest, summit, rocky" />
            <span className="hint">Used by the trek finder. Comma separated.</span></label>
        </div>
        <label>Elevation profile (optional)
          <textarea name="elevation" rows="4" defaultValue={t.Elevation} placeholder={'0 820\n2.5 1100\n5 1520'} />
          <span className="hint">One point per line: distance in km, then altitude in m. Take these from a GPS track (Strava, Wikiloc). Leave empty to hide the chart.</span>
        </label>
        <label>Pickup points (one per line) <textarea name="pickup_points" rows="4" defaultValue={t.PickupPoints} placeholder={'Majestic, Shanthala Silks | 11:00 PM\nIndiranagar Metro | 11:30 PM\nSilk Board Junction | 1:00 AM'} />
          <span className="hint">Add the time after a <code>|</code>. Trekkers see the times on the trek page and pick one when booking.</span></label>
        <label className="check"><input type="checkbox" name="active" defaultChecked={t.Active} /> Visible to trekkers</label>
        <button className="btn" disabled={busy}>Save trek</button>
      </form>

      {!isNew && (
        <>
          <h2>Batches</h2>
          <div className="table-wrap">
            <table className="table">
              <tbody>
                <tr><th>Date</th><th>Seats taken</th><th>Requests</th><th>Lead &amp; seats</th><th /></tr>
                {d.Batches && d.Batches.length
                  ? d.Batches.map((b) => <BatchRow key={`${b.ID}-${b.LeadID}-${b.Seats}`} b={b} leads={d.Leads} today={d.Today} />)
                  : <tr><td colSpan={5} className="muted">No batches yet.</td></tr>}
              </tbody>
            </table>
          </div>

          <form onSubmit={batchForm.submit(`/admin/treks/${t.ID}/batches`)} className="form panel">
            <strong>Add a batch</strong>
            {batchForm.error && <div className="flash err">{batchForm.error}</div>}
            <div className="row">
              <label>Start date <input type="date" name="start_date" min={d.Today} required /></label>
              <label>Seats <input type="number" name="seats" defaultValue="20" min="1" required /></label>
              <label>Trek lead
                <select name="lead_id" defaultValue="0">
                  <option value="0">— assign later —</option>
                  {d.Leads.map((l) => <option key={l.ID} value={l.ID}>{l.Name}</option>)}
                </select>
              </label>
            </div>
            <button className="btn" disabled={batchForm.busy}>Add as draft</button>
          </form>
        </>
      )}
    </>
  );
}
