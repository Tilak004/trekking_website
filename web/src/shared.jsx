// Small helpers shared by the islands.

export const inr = (n) => n.toLocaleString('en-IN');

export const photoURL = (p) => '/photos/' + p.split('/').map(encodeURIComponent).join('/');

// The same card markup as the server's "trekcard" template, so CSS applies to both.
export function TrekCard({ t, badge }) {
  const open = !!t.next;
  return (
    <a class="card trek-card" href={`/treks/${t.slug}`}>
      <div class="card-img" style={t.photo ? { backgroundImage: `url('${photoURL(t.photo)}')` } : null}>
        {!t.photo && (
          <svg viewBox="0 0 400 200" preserveAspectRatio="none" aria-hidden="true">
            <path d="M0 200 90 90l50 55 80-110 70 85 40-35 70 115z" fill="rgba(255,255,255,.12)" />
            <path d="M0 200 70 140l60 40 90-75 80 60 50-30 50 65z" fill="rgba(255,255,255,.18)" />
          </svg>
        )}
        <span class="tag diff">{t.difficulty}</span>
        {t.trail === 'closed' ? <span class="tag open closed">Trail closed</span> : open && <span class="tag open">Booking open</span>}
        {open && t.price > 0 && <span class="card-price">₹{inr(t.price)}</span>}
        {badge && <span class="tag match">{badge}</span>}
      </div>
      <div class="card-body">
        <p class="eyebrow">{t.region}</p>
        <h3>{t.name}</h3>
        {t.lead && (
          <span class="card-lead">
            {t.lead.photo ? <img src={photoURL(t.lead.photo)} alt="" /> : <i>{t.lead.name[0]}</i>}
            Led by {t.lead.name}{t.lead.verified && <b title="Verified lead"> ✓</b>}
          </span>
        )}
        <div class="card-bottom">
          {open ? <span>Next: <strong>{t.next}</strong></span> : <span class="soon-tag">Opening soon</span>}
          {t.ratings > 0 && <span class="rating">★ {t.rating.toFixed(1)}</span>}
        </div>
      </div>
    </a>
  );
}
