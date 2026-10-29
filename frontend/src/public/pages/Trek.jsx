import { useMemo } from 'react';
import { Link, useParams } from 'react-router-dom';
import Icon from '../../components/Icon.jsx';
import { PageState, Reviews } from '../../components/common.jsx';
import { usePage, useTitle } from '../../lib/usePage.js';
import { useSubmit } from '../../lib/useSubmit.js';
import { blocks, carryGroups, fmtDate, inr, lines, mentions, paras, parsePickups, photoURL, rating, whatsappTo } from '../../lib/util.js';
import { useSession } from '../../lib/session.jsx';
import { useWhatsAppMessage } from '../Layout.jsx';
import PageShell from '../PageShell.jsx';
import NotFound from './NotFound.jsx';
import Booking from './Booking.jsx';
import Elevation from './Elevation.jsx';

function TrailStatus({ t }) {
  if (t.TrailStatus === 'open') {
    return (
      <p className="trail-status open"><i />Trail open{t.TrailNote ? ` · ${t.TrailNote}` : ''}{t.TrailUpdated ? <span> · checked {fmtDate(t.TrailUpdated)}</span> : null}</p>
    );
  }
  if (t.TrailStatus === 'closed') {
    return (
      <p className="trail-status closed"><i />Trail closed{t.TrailNote ? ` · ${t.TrailNote}` : ''}{t.TrailUpdated ? <span> · updated {fmtDate(t.TrailUpdated)}</span> : null}</p>
    );
  }
  return null;
}

function Hero({ t, bookable }) {
  return (
    <>
      <section className="banner" style={t.Photo ? { backgroundImage: `url('${photoURL(t.Photo)}')` } : undefined}>
        <div className="wrap banner-inner">
          <p className="crumb"><Link to="/#treks">← All treks</Link></p>
          <span className="tag diff">{t.Difficulty}</span>
          <h1>{t.Name}</h1>
          <p className="banner-sub">{t.Region}</p>
          <TrailStatus t={t} />
          {bookable
            ? <a className="btn banner-cta" href="#book">Book now</a>
            : <><span className="soon-tag banner-soon">Opening soon</span>{' '}<a className="btn banner-cta" href="#enquire">Enquire now</a></>}
        </div>
      </section>
      <div className="wrap">
        <div className="stat-strip">
          {t.Duration && <div><span>Duration</span><strong>{t.Duration}</strong></div>}
          {t.DistanceKm ? <div><span>Distance</span><strong>{t.DistanceKm} km</strong></div> : null}
          {t.MaxAltitude ? <div><span>Max altitude</span><strong>{t.MaxAltitude} m</strong></div> : null}
          {t.BestSeason && <div><span>Best season</span><strong>{t.BestSeason}</strong></div>}
          <div><span>Rating</span><strong>{t.Ratings ? <>★ {rating(t.AvgRating)} <small>({t.Ratings})</small></> : 'New'}</strong></div>
        </div>
      </div>
    </>
  );
}

const ENQUIRED = "Thanks! We've got your enquiry";

// After an enquiry is sent, offer to send the same text on WhatsApp too. The
// enquiry is remembered in sessionStorage by the form below.
function EnquiryDone({ t, site }) {
  const href = useMemo(() => {
    let e = null;
    try {
      e = JSON.parse(sessionStorage.getItem('enquiry') || 'null');
      sessionStorage.removeItem('enquiry');
    } catch { /* private mode: fall back to the short message */ }
    const text = e
      ? ['Hi Beyond Expeditions! I just sent an enquiry about the ' + t.Name + ' trek.', '', 'Name: ' + e.name, 'Phone: ' + e.phone, ...(e.message ? ['Question: ' + e.message] : [])].join('\n')
      : `Hi Beyond Expeditions! I just sent an enquiry about the ${t.Name} trek.`;
    return site ? whatsappTo(site.BookingsWA, text) : '#';
  }, [t.Name, site]);
  return (
    <div className="enq-done">
      <p><b>✓ Got it!</b> We'll call or WhatsApp you soon.</p>
      <p className="muted small">Want a quicker reply? Send it to us on WhatsApp too.</p>
      <a className="btn big block wa-btn" id="enquiry-wa" target="_blank" rel="noopener noreferrer" href={href}><Icon name="waicon" /> Also send on WhatsApp</a>
      <p className="small"><Link to={`/treks/${t.Slug}#enquire`}>Send another enquiry</Link></p>
    </div>
  );
}

function EnquiryForm({ t, user, msg, site }) {
  const { submit, busy, error } = useSubmit();
  const guestEmail = user && /@guest\.beyondexpeditions\.local$/.test(user.Email);
  const remember = (e) => {
    const v = (n) => e.currentTarget.elements.namedItem(n).value.trim();
    try { sessionStorage.setItem('enquiry', JSON.stringify({ name: v('name'), phone: v('phone'), message: v('message') })); } catch { /* ignore */ }
  };
  const send = submit(`/treks/${t.Slug}/enquire`);
  return (
    <div className="panel enquire-panel" id="enquire">
      <h3>Enquire now</h3>
      {msg.startsWith(ENQUIRED) ? <EnquiryDone t={t} site={site} /> : (
        <>
          <p className="muted small">Not ready to book, or have a question? Leave your details and we'll call or WhatsApp you about {t.Name}.</p>
          <form onSubmit={(e) => { remember(e); return send(e); }} className="form" id="enquiry-form">
            {error && <div className="flash err">{error}</div>}
            <label>Your name <input name="name" required maxLength={100} autoComplete="name" defaultValue={user ? user.Name : ''} /></label>
            <label>Phone (WhatsApp) <input name="phone" type="tel" required inputMode="tel" placeholder="98765 43210" autoComplete="tel" defaultValue={user ? user.Phone : ''} /></label>
            <label>Email (optional) <input name="email" type="email" maxLength={200} autoComplete="email" autoCapitalize="none" spellCheck="false" defaultValue={user && !guestEmail ? user.Email : ''} /></label>
            <label>Your question <textarea name="message" rows="3" maxLength={2000} placeholder="Fitness needed, what to wear, group bookings…" /></label>
            <input className="hp" name="website" tabIndex={-1} autoComplete="off" aria-hidden="true" />
            <button className="btn block" disabled={busy}>Send enquiry</button>
          </form>
        </>
      )}
    </div>
  );
}

export default function Trek() {
  const { slug } = useParams();
  const page = usePage();
  const { User, Msg, Site } = useSession();
  const d = page.data;
  useTitle(d && `${d.Trek.Name} Trek`);
  useWhatsAppMessage(d ? `Hi! I'm interested in the ${d.Trek.Name} trek.` : '');

  if (page.notFound) return <NotFound />;
  if (!d) return <PageShell><PageState page={page} /></PageShell>;
  const t = d.Trek;
  const itinerary = blocks(t.Itinerary);
  const faqs = blocks(t.FAQs);
  const incl = lines(t.Inclusions);
  const excl = lines(t.Exclusions);
  const carry = carryGroups(t.ThingsToCarry);
  const pickups = parsePickups(t.PickupPoints);
  const batches = d.Batches || [];
  const existing = d.Existing;
  const closed = t.TrailStatus === 'closed';
  // Bookable only when a departure is published and still has seats.
  const openBatches = ((d.BookingData && d.BookingData.batches) || []).filter((b) => b.left > 0);
  const bookable = openBatches.length > 0;

  return (
    <PageShell hero={<Hero t={t} bookable={bookable} />}>
      {!t.Active && <div className="flash err">This trek is inactive and hidden from trekkers (you're seeing it as admin).</div>}

      <nav className="section-nav">
        <a href="#overview">Overview</a>
        {itinerary.length > 0 && <a href="#itinerary">Itinerary</a>}
        {pickups.length > 0 && <a href="#pickup">Pickup</a>}
        <a href="#inclusions">Inclusions</a>
        {carry.length > 0 && <a href="#carry">What to carry</a>}
        <a href="#book">{bookable ? 'Book' : 'Enquire'}</a>
        <a href="#faqs">FAQs</a>
        <a href="#reviews">Reviews</a>
      </nav>

      <div className="two-col single">
        <div className="trek-main">
          <section id="overview" className="trek-section">
            <p className="lead">{t.Summary}</p>
            <div className="prose">{paras(t.Description).map((p, i) => <p key={i}>{p}</p>)}</div>
            {t.Elevation && <Elevation points={t.Elevation} />}
          </section>

          {itinerary.length > 0 && (
            <section id="itinerary" className="trek-section">
              <h2>Itinerary</h2>
              <ol className="timeline">
                {itinerary.map((b, i) => <li key={i}><strong>{b.Title}</strong>{b.Body && <p>{b.Body}</p>}</li>)}
              </ol>
            </section>
          )}

          {pickups.length > 0 && (
            <section id="pickup" className="trek-section">
              <h2>Pickup points</h2>
              <ul className="pickup-list">
                {pickups.map((p, i) => <li key={i}>{p.Time && <b>{p.Time}</b>}<span>{p.Place}</span></li>)}
              </ul>
              <p className="muted small">Please be at your pickup point 10 minutes early.</p>
            </section>
          )}

          <section id="inclusions" className="trek-section">
            <h2>What's included</h2>
            <div className="incl-grid">
              {incl.length > 0 && (
                <div className="incl-box yes">
                  <h3>Included</h3>
                  <ul className="checklist">{incl.map((x, i) => <li key={i}>{x}</li>)}</ul>
                </div>
              )}
              <div className="incl-box no">
                <h3>Not included</h3>
                <ul className="checklist">
                  {excl.map((x, i) => <li key={i}>{x}</li>)}
                  {!mentions(t.Exclusions, 'insurance') && <li>Insurance of any kind (travel, accident or medical)</li>}
                </ul>
              </div>
            </div>
          </section>

          {carry.length > 0 && (
            <section id="carry" className="trek-section">
              <h2>What to carry</h2>
              {carry.map((g, i) => (
                <div key={i}>
                  {g.Title && <h3 className="carry-title">{g.Title}</h3>}
                  <ul className="carry-list">{g.Items.map((x, j) => <li key={j}>{x}</li>)}</ul>
                </div>
              ))}
            </section>
          )}

          <section id="book" className="trek-section book-section">
            <span id="dates" />
            <h2>{bookable ? 'Book or enquire' : 'Opening soon'}</h2>
            <div className="book-grid">
              <div className="panel book-panel">
                <h3>{bookable ? 'Book now' : 'Dates coming soon'}</h3>
                {closed && (
                  <div className="bp-none">
                    <svg viewBox="0 0 24 24" aria-hidden="true"><circle cx="12" cy="12" r="9" fill="none" stroke="currentColor" strokeWidth="2" /><path d="M12 7.5v5.5M12 16.5v.01" stroke="currentColor" strokeWidth="2.2" strokeLinecap="round" /></svg>
                    {' '}The trail is closed right now{t.TrailNote ? ` (${t.TrailNote})` : ''}. We'll confirm your date only if it reopens.
                  </div>
                )}
                {t.Price > 0 && <p className="bp-price">₹{inr(t.Price)} <span className="bp-badge">per person</span></p>}
                {existing ? (
                  <>
                    <p>You already have a <span className={`status ${existing.Status}`}>{existing.Status}</span> booking for {fmtDate(existing.StartDate)}.</p>
                    <p><Link to="/my">View my bookings →</Link></p>
                  </>
                ) : !bookable ? (
                  <>
                    <div className="bp-none">
                      <svg viewBox="0 0 24 24" aria-hidden="true"><circle cx="12" cy="12" r="9" fill="none" stroke="currentColor" strokeWidth="2" /><path d="M12 7.5v5.5M12 16.5v.01" stroke="currentColor" strokeWidth="2.2" strokeLinecap="round" /></svg>
                      {' '}{batches.length ? 'All current departures are full' : 'No departures open yet'}
                    </div>
                    <p className="bk-hint">Send us an enquiry and we'll let you know as soon as seats open.</p>
                  </>
                ) : (
                  <Booking key={slug} d={d.BookingData} />
                )}
                <ul className="assurances">
                  <li>Seat held for 10 minutes while you pay</li>
                  <li>Confirmed on WhatsApp once you've paid</li>
                  <li><Link to="/cancellation-policy">Clear cancellation &amp; refund policy</Link></li>
                </ul>
              </div>
              <EnquiryForm t={t} user={User} msg={Msg || ''} site={Site} />
            </div>
          </section>

          <section id="faqs" className="trek-section">
            <h2>Frequently asked questions</h2>
            <div className="faqs">
              {faqs.map((b, i) => <details key={i}><summary>{b.Title}</summary><p>{b.Body}</p></details>)}
              {!mentions(t.FAQs, 'insurance') && (
                <details>
                  <summary>Is insurance included?</summary>
                  <p>No. Our price does not include travel, accident or medical insurance, and any medical or evacuation costs are yours. If you want cover, you can buy personal accident insurance on your own before the trek.</p>
                </details>
              )}
            </div>
          </section>

          <section id="reviews" className="trek-section">
            <h2>Reviews</h2>
            <Reviews reviews={d.Reviews} />
          </section>
        </div>
      </div>
    </PageShell>
  );
}
