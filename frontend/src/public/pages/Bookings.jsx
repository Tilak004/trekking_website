import { Link } from 'react-router-dom';
import Icon from '../../components/Icon.jsx';
import { PageState } from '../../components/common.jsx';
import { useSession } from '../../lib/session.jsx';
import { usePage, useTitle } from '../../lib/usePage.js';
import { useSubmit } from '../../lib/useSubmit.js';
import { firstWord, fmtDate, inr, stars, telLink, whatsappTo } from '../../lib/util.js';
import PageShell from '../PageShell.jsx';
import NotFound from './NotFound.jsx';

function ReviewForm({ k }) {
  const { submit, busy, error } = useSubmit();
  const rates = [1, 2, 3, 4, 5];
  return (
    <form onSubmit={submit(`/my/bookings/${k.ID}/review`)} className="form review-form" encType="multipart/form-data">
      <strong>How was it? Leave a review</strong>
      {error && <div className="flash err">{error}</div>}
      <div className="row">
        <label>Trek
          <select name="trek_rating" required defaultValue="">
            <option value="">Rate…</option>
            {rates.map((n) => <option key={n} value={n}>{stars(n)}</option>)}
          </select>
        </label>
        {k.HasLead && (
          <label>Trek lead ({k.LeadName})
            <select name="lead_rating" required defaultValue="">
              <option value="">Rate…</option>
              {rates.map((n) => <option key={n} value={n}>{stars(n)}</option>)}
            </select>
          </label>
        )}
      </div>
      <label>Add a photo from the trek (optional) <input type="file" name="photo" accept="image/jpeg,image/png,image/webp" /></label>
      <label>Comment <textarea name="comment" rows="2" maxLength={1000} placeholder="What stood out?" /></label>
      <button className="btn small" disabled={busy}>Submit review</button>
    </form>
  );
}

function Booking({ k }) {
  const { submit } = useSubmit();
  const confirmText = k.PaidAt
    ? `Cancel this booking? You will get ₹${inr(k.CustomerRefund)} back as per our cancellation policy (${k.RefundRule}).`
    : 'Cancel this booking?';
  return (
    <div className="booking">
      <div className="booking-main">
        <div>
          <h3><Link to={`/treks/${k.TrekSlug}`}>{k.TrekName}</Link></h3>
          <p className="muted small">
            {fmtDate(k.StartDate)} · {k.People} {k.People === 1 ? 'person' : 'people'}
            {k.Pickup ? ` · Pickup: ${k.Pickup}` : ''}{k.LeadName ? ` · Lead: ${k.LeadName}` : ''}
          </p>
        </div>
        <div className="booking-side">
          <span className={`status ${k.Status}`}>{k.Status}</span>
          {k.Status === 'cancelled'
            ? (k.RefundedAt ? <span className="status completed">Refunded ₹{inr(k.RefundAmount)}</span>
              : k.RefundAmount ? <span className="status requested">Refund ₹{inr(k.RefundAmount)} on its way</span> : null)
            : k.PaidAt ? <span className="status paid">Paid</span>
            : k.Status === 'approved' && k.PayToken ? <Link className="btn small" to={k.PayPath}>Pay ₹{inr(k.Amount)}</Link> : null}
          {k.Status === 'requested' && (
            <a className="wa-link" href={k.WALink} target="_blank" rel="noopener noreferrer"><Icon name="waicon" /> Send on WhatsApp</a>
          )}
          {(k.Status === 'requested' || k.Status === 'approved') && (
            <form onSubmit={submit(`/my/bookings/${k.ID}/cancel`, { confirm: confirmText })} className="inline">
              <button className="linklike danger">Cancel</button>
            </form>
          )}
        </div>
      </div>
      {k.Status === 'completed' && !k.Reviewed ? <ReviewForm k={k} />
        : k.Reviewed ? <p className="muted small">✓ You reviewed this trek. Thanks!</p> : null}
    </div>
  );
}

export function My() {
  const page = usePage();
  const { User } = useSession();
  useTitle('My Bookings');
  const d = page.data;
  if (!d) return <PageShell><PageState page={page} /></PageShell>;
  return (
    <PageShell>
      <header className="page-head">
        <h1>My bookings</h1>
        <p className="muted">Hi {User && User.Name}. Here are your trek requests and past treks.</p>
      </header>
      {d.Bookings && d.Bookings.length
        ? d.Bookings.map((k) => <Booking key={k.ID} k={k} />)
        : <p className="muted">No bookings yet. <Link to="/#treks">Find a trek →</Link></p>}
    </PageShell>
  );
}

export function BookingSent() {
  const page = usePage();
  const { Site } = useSession();
  useTitle('One last step');
  if (page.notFound) return <NotFound />;
  const d = page.data;
  if (!d) return <PageShell><PageState page={page} /></PageShell>;
  const k = d.Booking;
  return (
    <PageShell>
      <section className="sent panel">
        <p className="eyebrow">Request #{k.ID} saved</p>
        <h1>One last step: send it to us on WhatsApp</h1>
        <p className="muted">Tap the button below. WhatsApp opens with your booking details already typed in, so just hit send. We'll reply to confirm your seat.</p>

        <dl className="sent-summary">
          <div><dt>Trek</dt><dd>{k.TrekName}</dd></div>
          <div><dt>Date</dt><dd>{fmtDate(k.StartDate)}</dd></div>
          <div><dt>People</dt><dd>{k.People}</dd></div>
          {k.Pickup && <div><dt>Pickup</dt><dd>{k.Pickup}</dd></div>}
        </dl>

        <a className="btn-wa" href={k.WALink} target="_blank" rel="noopener noreferrer"><Icon name="waicon" /> Send on WhatsApp</a>
        {Site && <p className="small muted">Prefer to call? Ring us on <a href={`tel:${telLink(Site.Phone)}`}>{Site.Phone}</a>.</p>}
        <p className="small"><Link to="/my">Go to my bookings →</Link></p>
      </section>
    </PageShell>
  );
}

export function Pay() {
  const page = usePage();
  const { Site } = useSession();
  useTitle('Pay for your trek');
  if (page.notFound) return <NotFound />;
  const d = page.data;
  if (!d) return <PageShell><PageState page={page} /></PageShell>;
  const k = d.Booking;
  const paidMsg = `Hi! I've paid ₹${inr(k.Amount)} for booking #${k.ID} (${k.TrekName}, ${fmtDate(k.StartDate)}). Sharing the screenshot here.`;
  return (
    <PageShell>
      <section className="pay panel">
        {k.PaidAt ? (
          <>
            <p className="eyebrow">Booking #{k.ID}</p>
            <h1>Payment received ✓</h1>
            <p className="muted">Thanks {firstWord(k.UserName)}! Your seat on <strong>{k.TrekName}</strong> ({fmtDate(k.StartDate)}) is confirmed. See you on the trail.</p>
          </>
        ) : (
          <>
            <p className="eyebrow">Booking #{k.ID}</p>
            <h1>Pay ₹{inr(k.Amount)}</h1>
            <p className="muted">{k.TrekName} · {fmtDate(k.StartDate)} · {k.People} {k.People === 1 ? 'person' : 'people'}</p>

            <a className="btn big block pay-upi" href={k.UPI}>Pay with UPI app</a>
            <p className="small muted pay-apps">Opens Google Pay, PhonePe, Paytm or any UPI app with the amount filled in.</p>

            <div className="pay-qr">
              <img src={`${k.PayPath}/qr.png`} alt={`UPI QR code for ₹${inr(k.Amount)}`} width="200" height="200" />
              <div>
                <strong>On a laptop?</strong>
                <span className="small muted">Scan this QR with any UPI app.</span>
                <span className="small">UPI ID: <code>{Site && Site.UPI}</code></span>
                <span className="small">Amount: <strong>₹{inr(k.Amount)}</strong></span>
              </div>
            </div>

            {Site && (
              <a className="wa-inline" href={whatsappTo(Site.BookingsWA, paidMsg)} target="_blank" rel="noopener noreferrer">
                <Icon name="waicon" /> Paid? Send us the screenshot on WhatsApp
              </a>
            )}
            <p className="small muted">We'll mark your booking paid once we see it. Please pay only after we've confirmed your seat.</p>
          </>
        )}
      </section>
    </PageShell>
  );
}
