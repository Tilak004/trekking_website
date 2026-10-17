import { render } from 'preact';
import { useState } from 'preact/hooks';
import { inr } from './shared.jsx';

// Step-by-step booking request. No account needed: guests give a name and a
// WhatsApp number. It posts the same fields as the plain form it replaces, so
// all validation still happens on the server.

const STEPS = ['Date', 'Details', 'Confirm'];
const digits = (s) => s.replace(/\D/g, '').slice(-10);

function Booking({ d }) {
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

  const problems = [
    !batch && 'Pick a date.',
    !d.loggedIn && (!f.name.trim() || digits(f.phone).length !== 10) && 'Enter your name and a 10-digit WhatsApp number.',
    null,
  ];
  const next = () => {
    setTouched(true);
    if (problems[step]) return;
    setTouched(false);
    setStep(step + 1);
  };

  return (
    <form class="bk" method="post" action={`/treks/${d.slug}/book`}
      onSubmit={(e) => { if (problems.some(Boolean)) { e.preventDefault(); setTouched(true); } }}>
      <input type="hidden" name="batch_id" value={f.batch || ''} />
      <input type="hidden" name="people" value={f.people} />
      <input type="hidden" name="pickup" value={f.pickup} />
      <input type="hidden" name="note" value={f.note} />
      {!d.loggedIn && <input type="hidden" name="name" value={f.name} />}
      {!d.loggedIn && <input type="hidden" name="phone" value={f.phone} />}

      <ol class="bk-steps">
        {STEPS.map((s, i) => (
          <li class={i < step ? 'done' : i === step ? 'on' : ''}>
            <button type="button" disabled={i > step} onClick={() => setStep(i)}>{i < step ? '✓' : i + 1}</button>
            <span>{s}</span>
          </li>
        ))}
      </ol>

      <div class="bk-pane" key={step}>
        {step === 0 && (
          <div>
            <h4>Pick your date</h4>
            <div class="bk-dates">
              {d.batches.map((b) => (
                <button type="button" disabled={b.left <= 0}
                  class={b.id === f.batch ? 'bk-date on' : 'bk-date'} onClick={() => setF({ ...f, batch: b.id, people: Math.min(f.people, b.left) })}>
                  <b>{b.label}</b>
                  <small>{b.left <= 0 ? 'Full' : `${b.left} seats left`}{b.lead ? ` · with ${b.lead}` : ''}</small>
                </button>
              ))}
            </div>
          </div>
        )}

        {step === 1 && (
          <div>
            <h4>Your details</h4>
            {!d.loggedIn && (
              <>
                <label>Your name <input value={f.name} onInput={set('name')} maxLength={100} autocomplete="name" /></label>
                <label>WhatsApp number <input type="tel" inputmode="tel" value={f.phone} onInput={set('phone')} placeholder="98765 43210" autocomplete="tel" /></label>
              </>
            )}
            <div class="bk-people">
              <span>People</span>
              <button type="button" aria-label="Fewer" onClick={() => setF({ ...f, people: Math.max(1, f.people - 1) })}>−</button>
              <b>{f.people}</b>
              <button type="button" aria-label="More" onClick={() => setF({ ...f, people: Math.min(maxPeople, f.people + 1) })}>+</button>
            </div>
            {d.pickups.length > 0 && (
              <>
                <p class="bk-label">Pickup point</p>
                <div class="chips">
                  {d.pickups.map((p) => (
                    <button type="button" class={p === f.pickup ? 'chip on' : 'chip'} onClick={() => setF({ ...f, pickup: p })}>{p}</button>
                  ))}
                </div>
              </>
            )}
          </div>
        )}

        {step === 2 && (
          <div>
            <h4>All good?</h4>
            <dl class="bk-review">
              <dt>Date</dt><dd>{batch && batch.label}</dd>
              <dt>Name</dt><dd>{f.name}</dd>
              {!d.loggedIn && <><dt>WhatsApp</dt><dd>{f.phone}</dd></>}
              <dt>People</dt><dd>{f.people}</dd>
              {f.pickup && <><dt>Pickup</dt><dd>{f.pickup}</dd></>}
            </dl>
            {d.price > 0 && <p class="bk-total">₹{inr(d.price * f.people)} <span>for {f.people} · pay only after we confirm</span></p>}
            <label>Anything else? (optional) <textarea rows="2" value={f.note} onInput={set('note')} maxLength={1000}
              placeholder="Friends joining, medical notes, questions…" /></label>
          </div>
        )}
      </div>

      {touched && problems[step] && <p class="bk-err">{problems[step]}</p>}

      <div class="bk-nav">
        {step > 0 && <button type="button" class="btn ghost" onClick={() => setStep(step - 1)}>Back</button>}
        {step < STEPS.length - 1
          ? <button type="button" class="btn" onClick={next}>Continue</button>
          : <button type="submit" class="btn">Send booking request</button>}
      </div>
      <p class="bk-hint">No account needed. We'll WhatsApp you to confirm your seat and add you to the trek group.</p>
    </form>
  );
}

export function mount(el, d) {
  if (!d || !d.batches || !d.batches.length) return;
  el.textContent = '';
  render(<Booking d={d} />, el);
}
