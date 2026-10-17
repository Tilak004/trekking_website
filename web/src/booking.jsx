import { render } from 'preact';
import { useState } from 'preact/hooks';
import { inr } from './shared.jsx';

// Step-by-step booking request. It posts the same fields as the plain form it
// replaces, so all validation still happens on the server.

const STEPS = ['Date', 'Group', 'Safety', 'Confirm'];

function Booking({ d }) {
  const firstOpen = d.batches.find((b) => b.left > 0);
  const [step, setStep] = useState(0);
  const [f, setF] = useState({
    batch: firstOpen ? firstOpen.id : null,
    people: 1,
    age: '',
    pickup: d.pickups[0] || '',
    emName: '',
    emPhone: '',
    medical: '',
    note: '',
  });
  const [touched, setTouched] = useState(false);
  const set = (k) => (e) => setF({ ...f, [k]: e.currentTarget.value });
  const batch = d.batches.find((b) => b.id === f.batch);
  const maxPeople = Math.min(10, batch ? batch.left : 10);

  const problems = [
    !batch && 'Pick a date.',
    (Number(f.age) < 5 || Number(f.age) > 100) && 'Enter your age.',
    (!f.emName.trim() || !f.emPhone.trim()) && 'Add an emergency contact name and phone.',
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
      <input type="hidden" name="age" value={f.age} />
      <input type="hidden" name="pickup" value={f.pickup} />
      <input type="hidden" name="emergency_name" value={f.emName} />
      <input type="hidden" name="emergency_phone" value={f.emPhone} />
      <input type="hidden" name="medical" value={f.medical} />
      <input type="hidden" name="note" value={f.note} />

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
            <h4>Who's coming?</h4>
            <div class="bk-people">
              <span>People</span>
              <button type="button" aria-label="Fewer" onClick={() => setF({ ...f, people: Math.max(1, f.people - 1) })}>−</button>
              <b>{f.people}</b>
              <button type="button" aria-label="More" onClick={() => setF({ ...f, people: Math.min(maxPeople, f.people + 1) })}>+</button>
            </div>
            <label>Your age <input type="number" min="5" max="100" value={f.age} onInput={set('age')} /></label>
            {d.price > 0 && <p class="bk-total">₹{inr(d.price * f.people)} <span>for {f.people} · pay only after we confirm</span></p>}
          </div>
        )}

        {step === 2 && (
          <div>
            <h4>Pickup &amp; safety</h4>
            {d.pickups.length > 0 && (
              <div class="chips">
                {d.pickups.map((p) => (
                  <button type="button" class={p === f.pickup ? 'chip on' : 'chip'} onClick={() => setF({ ...f, pickup: p })}>{p}</button>
                ))}
              </div>
            )}
            <label>Emergency contact name <input value={f.emName} onInput={set('emName')} maxLength={100} /></label>
            <label>Emergency contact phone <input type="tel" value={f.emPhone} onInput={set('emPhone')} maxLength={30} /></label>
            <label>Medical conditions / fitness notes <textarea rows="2" value={f.medical} onInput={set('medical')} maxLength={1000}
              placeholder="Asthma, knee issues, allergies… or leave blank" /></label>
          </div>
        )}

        {step === 3 && (
          <div>
            <h4>All good?</h4>
            <dl class="bk-review">
              <dt>Date</dt><dd>{batch && batch.label}</dd>
              <dt>People</dt><dd>{f.people}{f.age && ` · you're ${f.age}`}</dd>
              {f.pickup && <><dt>Pickup</dt><dd>{f.pickup}</dd></>}
              <dt>Emergency</dt><dd>{f.emName} ({f.emPhone})</dd>
            </dl>
            <label>Anything else? <textarea rows="2" value={f.note} onInput={set('note')} maxLength={1000}
              placeholder="Friends joining, questions…" /></label>
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
    </form>
  );
}

export function mount(el, d) {
  if (!d || !d.batches || !d.batches.length) return;
  el.textContent = '';
  render(<Booking d={d} />, el);
}
