import { Link, useParams } from 'react-router-dom';
import { Avatar, PageState } from '../../components/common.jsx';
import { useSession } from '../../lib/session.jsx';
import { usePage, useTitle } from '../../lib/usePage.js';
import { useSubmit } from '../../lib/useSubmit.js';
import { fmtDate, lines, rating } from '../../lib/util.js';

function StatusButton({ id, status, className, children, confirm }) {
  const { submit } = useSubmit();
  return (
    <form onSubmit={submit(`/admin/leads/${id}/status`, confirm ? { confirm } : undefined)} className="inline">
      <input type="hidden" name="status" value={status} />
      <button className={className}>{children}</button>
    </form>
  );
}

function AssignForm({ lead, batches }) {
  const { submit } = useSubmit();
  return (
    <form onSubmit={submit(`/admin/leads/${lead.ID}/assign`)} className="la-assign">
      <select name="batch_id" required defaultValue="">
        <option value="">Assign to a trek date…</option>
        {batches.filter((b) => b.LeadID !== lead.ID).map((b) => (
          <option key={b.ID} value={b.ID}>{b.TrekName} · {fmtDate(b.StartDate)}{b.LeadName ? ` (now: ${b.LeadName})` : ''}</option>
        ))}
      </select>
      <button className="btn small act-approved">Assign</button>
    </form>
  );
}

function LeadItem({ l, batches, publicURL }) {
  const achievements = lines(l.Achievements);
  const certs = lines(l.Certifications);
  const mine = batches.filter((b) => b.LeadID === l.ID);
  return (
    <details className="lead-admin" id={`lead-${l.ID}`}>
      <summary>
        <Avatar person={l} />
        <span className="la-name"><strong>{l.Name}</strong>{l.Verified && <> <span className="verified sm">✓</span></>}<br /><span className="muted small">{l.Headline}{l.City ? ` · ${l.City}` : ''}</span></span>
        <span className={`status ${l.Status}`}>{l.Status}</span>
        <span className="la-stats small muted">{l.TreksLed} treks · {l.TrekkersGuided} trekkers · {l.Ratings ? `★ ${rating(l.AvgRating)}` : 'no ratings'}</span>
      </summary>
      <div className="la-body">
        <div className="la-profile">
          {l.Bio && <p>{l.Bio}</p>}
          {achievements.length > 0 && <><h4>Treks &amp; achievements</h4><ul className="tight-list">{achievements.map((x, i) => <li key={i}>{x}</li>)}</ul></>}
          {certs.length > 0 && <><h4>Certifications</h4><ul className="tight-list">{certs.map((x, i) => <li key={i}>{x}</li>)}</ul></>}
          <p className="small muted">
            {l.Phone && <>Phone: <a href={`tel:${l.Phone}`}>{l.Phone}</a> · </>}
            {l.AppliedAt && <>Applied {l.AppliedAt} · </>}
            <a href={`${publicURL}/leads/${l.Slug}`} target="_blank" rel="noopener noreferrer">Public profile</a> · <Link to={`/admin/leads/${l.ID}`}>Edit profile</Link>
          </p>
        </div>
        <div className="la-side">
          <div className="la-actions">
            {l.Status !== 'approved' && l.Status !== 'removed' && <StatusButton id={l.ID} status="approved" className="btn small act-approved">Approve</StatusButton>}
            {l.Status === 'pending' && <StatusButton id={l.ID} status="rejected" className="btn small act-rejected">Reject</StatusButton>}
            {l.Status !== 'removed' && l.Status !== 'rejected' && (
              <StatusButton id={l.ID} status="removed" className="btn small act-cancelled"
                confirm={`Remove ${l.Name} as a trek lead?\n\nTheir profile is hidden, they lose dashboard access, their upcoming trek dates become unassigned and their proposed dates are dropped.`}>Remove</StatusButton>
            )}
          </div>
          {l.Status === 'approved' && (
            <>
              <h4>Leading</h4>
              {mine.length > 0
                ? <ul className="tight-list">{mine.map((b) => <li key={b.ID}>{b.TrekName} · {fmtDate(b.StartDate)}</li>)}</ul>
                : <p className="muted small">No upcoming treks assigned.</p>}
              {batches.length > 0 && <AssignForm lead={l} batches={batches} />}
            </>
          )}
        </div>
      </div>
    </details>
  );
}

export function Leads() {
  const page = usePage();
  const { PublicURL } = useSession();
  useTitle('Trek leads');
  const d = page.data;
  if (!d) return <PageState page={page} />;
  const batches = d.Batches || [];
  return (
    <>
      <div className="head-row">
        <h1>Trek leads</h1>
        <Link className="btn" to="/admin/leads/new">+ New lead</Link>
      </div>
      <p className="muted small">Open a lead to see their full profile. Approve new applications, and assign approved leads to upcoming trek dates.</p>
      <div className="lead-admin-list">
        {d.Leads && d.Leads.length
          ? d.Leads.map((l) => <LeadItem key={l.ID} l={l} batches={batches} publicURL={PublicURL} />)
          : <p className="muted">No leads yet.</p>}
      </div>
      <p className="muted small">Treks led and trekkers guided count only bookings marked <em>completed</em>.</p>
    </>
  );
}

export function LeadForm() {
  const { id } = useParams();
  const page = usePage();
  const { PublicURL } = useSession();
  const { submit, busy, error } = useSubmit();
  const d = page.data;
  useTitle(d && (d.Lead.ID ? `Edit ${d.Lead.Name}` : 'New lead'));
  if (page.notFound) return <p className="muted">Lead not found. <Link to="/admin/leads">Back to leads</Link></p>;
  if (!d) return <PageState page={page} />;
  const l = d.Lead;
  const isNew = !l.ID;
  return (
    <>
      <p className="crumb">
        <Link to="/admin/leads">← All leads</Link>
        {!isNew && <> · <a href={`${PublicURL}/leads/${l.Slug}`} target="_blank" rel="noopener noreferrer">View public profile</a></>}
      </p>
      <h1>{isNew ? 'New trek lead' : `Edit ${l.Name}`}</h1>
      <form key={l.ID || 'new'} onSubmit={submit(isNew ? '/admin/leads/new' : `/admin/leads/${id}`)} className="form panel">
        {(error || d.Err) && <div className="flash err">{error || d.Err}</div>}
        <div className="row">
          <label>Name <input name="name" defaultValue={l.Name} required maxLength={100} /></label>
          <label>URL slug <input name="slug" defaultValue={l.Slug} placeholder="auto from name" maxLength={100} /></label>
        </div>
        {l.UserID ? <p className="muted small">Applied through the website{l.AppliedAt ? ` on ${l.AppliedAt}` : ''} · status: <span className={`status ${l.Status}`}>{l.Status}</span></p> : null}
        <label>City <input name="city" defaultValue={l.City} maxLength={80} /></label>
        <label>Headline <input name="headline" defaultValue={l.Headline} maxLength={150} placeholder="e.g. Ironman finisher · Endurance coach" /></label>
        <label>Bio <textarea name="bio" rows="5" defaultValue={l.Bio} /></label>
        <label>Achievements (one per line) <textarea name="achievements" rows="4" defaultValue={l.Achievements} /></label>
        <label>Certifications (one per line) <textarea name="certifications" rows="3" defaultValue={l.Certifications} /></label>
        <label>Profile photo
          <select name="photo" defaultValue={l.Photo}>
            <option value="">— none (shows their initial) —</option>
            {(d.Photos || []).map((p) => <option key={p}>{p}</option>)}
          </select>
          <span className="hint">Put lead photos in the <code>photos/leads</code> folder, then refresh this page.</span>
        </label>
        <label>Phone (internal only, not shown publicly) <input name="phone" defaultValue={l.Phone} maxLength={30} /></label>
        <label className="check"><input type="checkbox" name="verified" defaultChecked={l.Verified} /> Verified (ID and experience checked) — shows a ✓ badge</label>
        <label className="check"><input type="checkbox" name="active" defaultChecked={l.Active} /> Show public profile (only for approved leads)</label>
        <button className="btn" disabled={busy}>Save lead</button>
      </form>
    </>
  );
}
