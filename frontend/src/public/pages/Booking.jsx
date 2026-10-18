import { useState } from 'react';
import { useSubmit } from '../../lib/useSubmit.js';
import { digitsOnly, inr } from '../../lib/util.js';

// Step-by-step booking request. No account needed: guests give a name and a
// WhatsApp number. It posts the same fields the server always took, so all
// validation still happens on the server.

const digits = (s) => digitsOnly(s).slice(-10);

// Two stages: (1) date + number of people, (2) contact details, pickup and a
// note, then send.
export default function Booking({ d }) {
  const { submit, busy, error } = useSubmit();
  const firstOpen = d.batches.find((b) => b.left > 0);
  const [step, setStep] = useState(0);
  const [f, setF] = useState({
    batch: firstOpen ? firstOpen.id : null,
    people: 1,
    name: d.name || '',
    phone: d.phone || '',
    pickup: d.pickups[0] || '',
    note: '',
  });
  const [touched, setTouched] = useState(false);
  const set = (k) => (e) => setF({ ...f, [k]: e.currentTarget.value });
  const batch = d.batches.find((b) => b.id === f.batch);
  const maxPeople = Math.min(10, batch ? batch.left : 10);
  const years = [...new Set(d.batches.map((b) => b.year).filter(Boolean))].join(' / ');

  const problems = [
    !batch && 'Pick a date.',
    !d.loggedIn && (!f.name.trim() || digits(f.phone).length !== 10) && 'Enter your name and a 10-digit WhatsApp number.',
  ];
  const next = () => {
    setTouched(true);
    if (problems[0]) return;
    setTouched(false);
    setStep(1);
  };
  const send = submit(`/treks/${d.slug}/book`);

  return (
    <form className="bk" onSubmit={(e) => {
      if (problems.some(Boolean)) {
        e.preventDefault();
        setTouched(true);
        return undefined;
      }
      return send(e);
    }}>
      <input type="hidden" name="batch_id" value={f.batch || ''} />
      <input type="hidden" name="people" value={f.people} />
      <input type="hidden" name="pickup" value={f.pickup} />
      <input type="hidden" name="note" value={f.note} />
      {!d.loggedIn && <input type="hidden" name="name" value={f.name} />}
      {!d.loggedIn && <input type="hidden" name="phone" value={f.phone} />}

      <div className="bk-pane" key={step}>
        {step === 0 && (
          <div>
            <p className="bk-label">Select date to register{years && ` — ${years}`}</p>
            <div className="bk-dates">
              {d.batches.map((b) => (
                <button type="button" key={b.id} disabled={b.left <= 0}
                  className={b.id === f.batch ? 'bk-date on' : 'bk-date'}
                  onClick={() => setF({ ...f, batch: b.id, people: Math.min(f.people, b.left) })}>
                  <b>{b.label}</b>
                  <small>{b.left <= 0 ? 'Full' : `${b.left} seats left`}{b.lead ? ` · with ${b.lead}` : ''}</small>
                </button>
              ))}
            </div>
            <p className="bk-label">Number of people</p>
            <div className="bk-stepper">
              <button type="button" aria-label="Fewer" disabled={f.people <= 1} onClick={() => setF({ ...f, people: Math.max(1, f.people - 1) })}>−</button>
              <b>{f.people}</b>
              <button type="button" aria-label="More" disabled={f.people >= maxPeople} onClick={() => setF({ ...f, people: Math.min(maxPeople, f.people + 1) })}>+</button>
            </div>
            {d.price > 0 && <p className="bk-total">₹{inr(d.price * f.people)} <span>total for {f.people} {f.people === 1 ? 'person' : 'people'} · pay only after we confirm</span></p>}
          </div>
        )}

        {step === 1 && (
          <div>
            <button type="button" className="bk-back" onClick={() => setStep(0)}>
              ← {batch && batch.label} · {f.people} {f.people === 1 ? 'person' : 'people'} <u>Change</u>
            </button>
            {!d.loggedIn && (
              <>
                <label>Your name <input value={f.name} onChange={set('name')} maxLength={100} autoComplete="name" /></label>
                <label>WhatsApp number <input type="tel" inputMode="tel" value={f.phone} onChange={set('phone')} placeholder="98765 43210" autoComplete="tel" /></label>
              </>
            )}
            {d.pickups.length > 0 && (
              <>
                <p className="bk-label">Pickup point</p>
                <div className="chips">
                  {d.pickups.map((p) => (
                    <button type="button" key={p} className={p === f.pickup ? 'chip on' : 'chip'} onClick={() => setF({ ...f, pickup: p })}>{p}</button>
                  ))}
                </div>
              </>
            )}
            <label>Anything else? (optional) <textarea rows="2" value={f.note} onChange={set('note')} maxLength={1000}
              placeholder="Friends joining, medical notes, questions…" /></label>
            {d.price > 0 && <p className="bk-total">₹{inr(d.price * f.people)} <span>for {f.people} · pay only after we confirm</span></p>}
          </div>
        )}
      </div>

      {touched && problems[step] && <p className="bk-err">{problems[step]}</p>}
      {error && <p className="bk-err">{error}</p>}

      <div className="bk-nav bk-actions">
        {step === 0
          ? <button type="button" className="btn big" onClick={next}>Book now</button>
          : <button type="submit" className="btn big" disabled={busy}>Send booking request</button>}
      </div>
      <p className="bk-hint">No payment now. We'll WhatsApp you to confirm your seat.</p>
    </form>
  );
}
