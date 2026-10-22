import { Link } from 'react-router-dom';
import { fmtDate, initial, inr, photoURL, rating, stars } from '../lib/util.js';

export function Flash({ msg, err }) {
  return (
    <>
      {msg ? <div className="flash ok">{msg}</div> : null}
      {err ? <div className="flash err">{err}</div> : null}
    </>
  );
}

export function Avatar({ person, big }) {
  const cls = big ? 'avatar big' : 'avatar';
  if (person.Photo) return <img className={`${cls} photo`} src={photoURL(person.Photo)} alt={person.Name} />;
  return <div className={cls}>{initial(person.Name)}</div>;
}

export const Rating = ({ n }) => <span className="rating">{stars(n)}</span>;

export function Reviews({ reviews }) {
  if (!reviews || !reviews.length) return <p className="muted">No reviews yet.</p>;
  return reviews.map((r, i) => (
    <div className="review" key={i}>
      <div className="review-head">
        <strong>{r.UserName}</strong>
        <span className="muted small">{r.TrekName}</span>
      </div>
      <div className="small">
        Trek <Rating n={r.TrekRating} />
        {r.LeadRating ? <> · Lead <Rating n={r.LeadRating} /></> : null}
      </div>
      {r.Comment ? <p>{r.Comment}</p> : null}
      {r.Photo ? <img className="review-photo" src={photoURL(r.Photo)} alt="Trek photo" loading="lazy" /> : null}
    </div>
  ));
}

const Mountains = () => (
  <svg viewBox="0 0 400 200" preserveAspectRatio="none" aria-hidden="true">
    <path d="M0 200 90 90l50 55 80-110 70 85 40-35 70 115z" fill="rgba(255,255,255,.12)" />
    <path d="M0 200 70 140l60 40 90-75 80 60 50-30 50 65z" fill="rgba(255,255,255,.18)" />
  </svg>
);

// A trek card from the full Trek record (the server's "trekcard" template).
export function TrekCard({ t }) {
  return (
    <Link className="card trek-card" to={`/treks/${t.Slug}`}>
      <div className="card-img" style={t.Photo ? { backgroundImage: `url('${photoURL(t.Photo)}')` } : undefined}>
        {!t.Photo && <Mountains />}
        <span className="tag diff">{t.Difficulty}</span>
        {t.TrailStatus === 'closed' ? <span className="tag open closed">Trail closed</span> : t.NextBatch && <span className="tag open">Booking open</span>}
        {t.NextBatch && <span className="card-price">₹{inr(t.Price)}</span>}
      </div>
      <div className="card-body">
        <p className="eyebrow">{t.Region}</p>
        <h3>{t.Name}</h3>
        {t.Lead && (
          <span className="card-lead">
            {t.Lead.photo ? <img src={photoURL(t.Lead.photo)} alt="" /> : <i>{initial(t.Lead.name)}</i>}
            Led by {t.Lead.name}
            {t.Lead.verified && <b title="Verified lead"> ✓</b>}
          </span>
        )}
        <p className="card-summary">{t.Summary}</p>
        <div className="facts">
          {t.Duration && <span>{t.Duration}</span>}
          {t.DistanceKm ? <span>{t.DistanceKm} km</span> : null}
          {t.MaxAltitude ? <span>{t.MaxAltitude} m</span> : null}
        </div>
        <div className="card-bottom">
          {t.NextBatch ? <span>Next: <strong>{fmtDate(t.NextBatch)}</strong></span> : <span className="soon-tag">Opening soon</span>}
          {t.Ratings ? <span className="rating">★ {rating(t.AvgRating)}</span> : null}
        </div>
      </div>
    </Link>
  );
}

export function Loading() {
  return <p className="muted" style={{ padding: '40px 0' }}>Loading…</p>;
}

export function ErrorBox({ message }) {
  return <div className="flash err">{message}</div>;
}

// Shows the standard states for a page's data, or null when the data is ready.
export function PageState({ page }) {
  if (page.loading) return <Loading />;
  if (page.error) return <ErrorBox message={page.error} />;
  return null;
}
