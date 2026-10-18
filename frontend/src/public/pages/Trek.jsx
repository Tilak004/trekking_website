import { Link, useParams } from 'react-router-dom';
import { PageState, Reviews } from '../../components/common.jsx';
import { usePage, useTitle } from '../../lib/usePage.js';
import { useSubmit } from '../../lib/useSubmit.js';
import { blocks, fmtDate, inr, lines, paras, photoURL, rating } from '../../lib/util.js';
import { useSession } from '../../lib/session.jsx';
import { useWhatsAppMessage } from '../Layout.jsx';
import PageShell from '../PageShell.jsx';
import NotFound from './NotFound.jsx';
import Booking from './Booking.jsx';
import Elevation from './Elevation.jsx';

function Hero({ t }) {
  return (
    <>
      <section className="banner" style={t.Photo ? { backgroundImage: `url('${photoURL(t.Photo)}')` } : undefined}>
        <div className="wrap banner-inner">
          <p className="crumb"><Link to="/#treks">← All treks</Link></p>
          <span className="tag diff">{t.Difficulty}</span>
          <h1>{t.Name}</h1>
          <p className="banner-sub">{t.Region}</p>
          <a className="btn banner-cta" href="#book">Book now</a>
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

function EnquiryForm({ t, user }) {
  const { submit, busy, error } = useSubmit();
  const guestEmail = user && /@guest\.beyondexpeditions\.local$/.test(user.Email);
  return (
    <div className="panel enquire-panel" id="enquire">
      <h3>Enquire now</h3>
      <p className="muted small">Not ready to book, or have a question? Leave your details and we'll call or WhatsApp you about {t.Name}.</p>
      <form onSubmit={submit(`/treks/${t.Slug}/enquire`)} className="form">
        {error && <div className="flash err">{error}</div>}
        <label>Your name <input name="name" required maxLength={100} autoComplete="name" defaultValue={user ? user.Name : ''} /></label>
        <label>Phone (WhatsApp) <input name="phone" type="tel" required inputMode="tel" placeholder="98765 43210" autoComplete="tel" defaultValue={user ? user.Phone : ''} /></label>
        <label>Email (optional) <input name="email" type="email" maxLength={200} autoComplete="email" autoCapitalize="none" spellCheck="false" defaultValue={user && !guestEmail ? user.Email : ''} /></label>
        <label>Your question <textarea name="message" rows="3" maxLength={2000} placeholder="Fitness needed, what to wear, group bookings…" /></label>
        <input className="hp" name="website" tabIndex={-1} autoComplete="off" aria-hidden="true" />
        <button className="btn block" disabled={busy}>Send enquiry</button>
      </form>
    </div>
  );
}

export default function Trek() {
  const { slug } = useParams();
  const page = usePage();
  const { User } = useSession();
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
  const carry = lines(t.ThingsToCarry);
  const batches = d.Batches || [];
  const existing = d.Existing;

  return (
    <PageShell hero={<Hero t={t} />}>
      {!t.Active && <div className="flash err">This trek is inactive and hidden from trekkers (you're seeing it as admin).</div>}

      <nav className="section-nav">
        <a href="#overview">Overview</a>
        {itinerary.length > 0 && <a href="#itinerary">Itinerary</a>}
        {(incl.length > 0 || excl.length > 0) && <a href="#inclusions">Inclusions</a>}
        {carry.length > 0 && <a href="#carry">What to carry</a>}
        <a href="#book">Book</a>
        {faqs.length > 0 && <a href="#faqs">FAQs</a>}
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

          {(incl.length > 0 || excl.length > 0) && (
            <section id="inclusions" className="trek-section">
              <h2>What's included</h2>
              <div className="incl-grid">
                {incl.length > 0 && (
                  <div className="incl-box yes">
                    <h3>Included</h3>
                    <ul className="checklist">{incl.map((x, i) => <li key={i}>{x}</li>)}</ul>
                  </div>
                )}
                {excl.length > 0 && (
                  <div className="incl-box no">
                    <h3>Not included</h3>
                    <ul className="checklist">{excl.map((x, i) => <li key={i}>{x}</li>)}</ul>
                  </div>
                )}
              </div>
            </section>
          )}

          {carry.length > 0 && (
            <section id="carry" className="trek-section">
              <h2>What to carry</h2>
              <ul className="carry-list">{carry.map((x, i) => <li key={i}>{x}</li>)}</ul>
            </section>
          )}

          <section id="book" className="trek-section book-section">
            <span id="dates" />
            <h2>Book or enquire</h2>
            <div className="book-grid">
              <div className="panel book-panel">
                <h3>Book now</h3>
                {t.Price > 0 && <p className="bp-price">₹{inr(t.Price)} <span className="bp-badge">per person</span></p>}
                {existing ? (
                  <>
                    <p>You already have a <span className={`status ${existing.Status}`}>{existing.Status}</span> booking for {fmtDate(existing.StartDate)}.</p>
                    <p><Link to="/my">View my bookings →</Link></p>
                  </>
                ) : batches.length === 0 ? (
                  <>
                    <p className="bk-label">Select date to register</p>
                    <div className="bp-none">
                      <svg viewBox="0 0 24 24" aria-hidden="true"><circle cx="12" cy="12" r="9" fill="none" stroke="currentColor" strokeWidth="2" /><path d="M12 7.5v5.5M12 16.5v.01" stroke="currentColor" strokeWidth="2.2" strokeLinecap="round" /></svg>
                      {' '}No fixed departures right now
                    </div>
                    <div className="bk-actions"><button className="btn big" disabled>No departures available</button></div>
                    <p className="bk-hint">Online booking opens once dates are announced. Send an enquiry and we'll tell you first.</p>
                  </>
                ) : (
                  <Booking key={slug} d={d.BookingData} />
                )}
                <ul className="assurances">
                  <li>No payment to request a seat</li>
                  <li>Seat confirmed personally on WhatsApp</li>
                  <li><Link to="/cancellation-policy">Clear cancellation &amp; refund policy</Link></li>
                </ul>
              </div>
              <EnquiryForm t={t} user={User} />
            </div>
          </section>

          {faqs.length > 0 && (
            <section id="faqs" className="trek-section">
              <h2>Frequently asked questions</h2>
              <div className="faqs">
                {faqs.map((b, i) => <details key={i}><summary>{b.Title}</summary><p>{b.Body}</p></details>)}
              </div>
            </section>
          )}

          <section id="reviews" className="trek-section">
            <h2>Reviews</h2>
            <Reviews reviews={d.Reviews} />
          </section>
        </div>
      </div>
    </PageShell>
  );
}

