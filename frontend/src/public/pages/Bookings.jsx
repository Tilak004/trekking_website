import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import Icon from '../../components/Icon.jsx';
import { PageState } from '../../components/common.jsx';
import { useSession } from '../../lib/session.jsx';
import { usePage, useTitle } from '../../lib/usePage.js';
import { useSubmit } from '../../lib/useSubmit.js';
import { firstWord, fmtDate, inr, stars, whatsappTo } from '../../lib/util.js';
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
            {k.Pickup ? ` · Pickup: ${k.Pickup}` : ''}
          </p>
        </div>
        <div className="booking-side">
          {k.Held ? <span className="status pending">Seat held · <HoldTimer seconds={k.HoldSecondsLeft} /> left</span>
            : k.Claimed ? <span className="status pending">Checking your payment</span>
            : k.Status === 'cancelled' && k.RefundReason === 'expired' ? <span className="status cancelled">Hold ended</span>
            : <span className={`status ${k.Status}`}>{k.Status}</span>}
          {k.Status === 'cancelled'
            ? (k.RefundedAt ? <span className="status completed">Refunded ₹{inr(k.RefundAmount)}</span>
              : k.RefundAmount ? <span className="status requested">Refund ₹{inr(k.RefundAmount)} on its way</span> : null)
            : k.PaidAt ? <span className="status paid">Paid</span>
            : k.Status === 'approved' && k.PayToken && !k.Claimed ? <Link className="btn small" to={k.PayPath}>Pay ₹{inr(k.Amount)}</Link> : null}
          {k.GroupLink && k.Status === 'approved' && (
            <a className="wa-link" href={k.GroupLink} target="_blank" rel="noopener noreferrer"><Icon name="waicon" /> Join the trek's WhatsApp group</a>
          )}
          {k.Status === 'completed' && k.ChannelLink && (
            <a className="wa-link" href={k.ChannelLink} target="_blank" rel="noopener noreferrer"><Icon name="waicon" /> Follow our WhatsApp channel</a>
          )}
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
        <p className="small"><Link to="/my">Go to my bookings →</Link></p>
      </section>
    </PageShell>
  );
}

// "3h 12m" / "12m 05s" until the seat hold ends. Counts down on its own and
// reloads the page when it reaches zero so the page shows the released seat.
export function HoldTimer({ seconds }) {
  const [left, setLeft] = useState(seconds);
  useEffect(() => {
    setLeft(seconds);
    const end = Date.now() + seconds * 1000;
    const t = setInterval(() => {
      const s = Math.max(0, Math.round((end - Date.now()) / 1000));
      setLeft(s);
      if (s === 0) {
        clearInterval(t);
        window.dispatchEvent(new Event('beyond:refresh'));
      }
    }, 1000);
    return () => clearInterval(t);
  }, [seconds]);
  const h = Math.floor(left / 3600);
  const m = Math.floor((left % 3600) / 60);
  const s = left % 60;
  return <strong className="hold-timer">{h > 0 ? `${h}h ${String(m).padStart(2, '0')}m` : `${m}m ${String(s).padStart(2, '0')}s`}</strong>;
}

// Where the trekker sends the 12-digit UPI reference after paying.
function UtrForm({ k }) {
  const { submit, busy, error } = useSubmit();
  return (
    <form onSubmit={submit(`${k.PayPath}/utr`)} className="utr-form">
      <strong>Paid? Send us the UPI reference</strong>
      <p className="small muted">In your UPI app, open the payment and copy the 12-digit <b>UTR / UPI reference number</b>.</p>
      {error && <div className="flash err">{error}</div>}
      <div className="utr-row">
        <input name="utr" inputMode="numeric" autoComplete="off" required pattern="[0-9 ]{12,16}" maxLength={16} placeholder="123456789012" aria-label="UPI reference (UTR)" />
        <button className="btn" disabled={busy}>I've paid</button>
      </div>
    </form>
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
  const summary = `${k.TrekName} · ${fmtDate(k.StartDate)} · ${k.People} ${k.People === 1 ? 'person' : 'people'}`;
  if (k.Status === 'cancelled' || k.Status === 'rejected') {
    const expired = k.RefundReason === 'expired';
    return (
      <PageShell>
        <section className="pay panel">
          <p className="eyebrow">Booking #{k.ID}</p>
          <h1>{expired ? 'Your seat hold ended' : 'This booking is cancelled'}</h1>
          <p className="muted">{expired
            ? `We didn't receive a payment within 4 hours, so the seat on ${k.TrekName} was released for other trekkers.`
            : `The booking for ${k.TrekName} (${fmtDate(k.StartDate)}) is no longer active.`}</p>
          <Link className="btn big block" to={`/treks/${k.TrekSlug}`}>Book again</Link>
          {expired && Site && (
            <a className="wa-inline" href={whatsappTo(Site.BookingsWA, `Hi! I paid for booking #${k.ID} (${k.TrekName}, ${fmtDate(k.StartDate)}) but my seat hold had ended.`)} target="_blank" rel="noopener noreferrer">
              <Icon name="waicon" /> Already paid? Message us on WhatsApp
            </a>
          )}
        </section>
      </PageShell>
    );
  }
  if (k.Claimed) {
    return (
      <PageShell>
        <section className="pay panel">
          <p className="eyebrow">Booking #{k.ID}</p>
          <h1>Payment received, we're checking it</h1>
          <p className="muted">Thanks {firstWord(k.UserName)}! We've got your UPI reference <code>{k.Utr}</code> for ₹{inr(k.Amount)}. We'll confirm your seat on <strong>{k.TrekName}</strong> on WhatsApp as soon as we see the payment.</p>
          <p className="small muted">{summary}</p>
          <p className="small"><Link to="/my">Go to my bookings →</Link></p>
        </section>
      </PageShell>
    );
  }
  return (
    <PageShell>
      <section className="pay panel">
        {k.PaidAt ? (
          <>
            <p className="eyebrow">Booking #{k.ID}</p>
            <h1>Payment received ✓</h1>
            <p className="muted">Thanks {firstWord(k.UserName)}! Your seat on <strong>{k.TrekName}</strong> ({fmtDate(k.StartDate)}) is confirmed. See you on the trail.</p>
            {k.GroupLink && (
              <a className="btn big block" href={k.GroupLink} target="_blank" rel="noopener noreferrer">Join the {k.TrekName} WhatsApp group</a>
            )}
          </>
        ) : (
          <>
            <p className="eyebrow">Booking #{k.ID}</p>
            <h1>Pay ₹{inr(k.Amount)}</h1>
            <p className="muted">{summary}</p>
            {k.Held && (
              <p className="hold-banner">Your seat is held for <HoldTimer seconds={k.HoldSecondsLeft} />. Pay and send us the UPI reference before then, or the seat is released.</p>
            )}

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

            {k.HoldUntil ? <UtrForm k={k} /> : null}
            {Site && (
              <a className="wa-inline" href={whatsappTo(Site.BookingsWA, paidMsg)} target="_blank" rel="noopener noreferrer">
                <Icon name="waicon" /> {k.HoldUntil ? 'Trouble paying? Message us on WhatsApp' : 'Paid? Send us the screenshot on WhatsApp'}
              </a>
            )}
            <p className="small muted">{k.HoldUntil
              ? "We'll confirm your seat on WhatsApp once we see your payment."
              : "We'll mark your booking paid once we see it. Please pay only after we've confirmed your seat."}</p>
          </>
        )}
      </section>
    </PageShell>
  );
}
