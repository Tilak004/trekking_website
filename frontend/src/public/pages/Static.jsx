import { Link } from 'react-router-dom';
import Icon from '../../components/Icon.jsx';
import { useSession } from '../../lib/session.jsx';
import { usePage, useTitle } from '../../lib/usePage.js';
import { useSubmit } from '../../lib/useSubmit.js';
import { whatsappTo } from '../../lib/util.js';
import PageShell from '../PageShell.jsx';

// Pages that are mostly fixed text. They still ask the server for the site's
// contact details (and the signed-in user) by loading their own /api/<path>.
function useStatic(title) {
  const page = usePage();
  const s = useSession();
  useTitle(title);
  return s.Site;
}

export function About() {
  useStatic('About us');
  const { Brand } = useSession();
  const hero = (
    <section className="banner short" style={{ backgroundImage: "url('/photos/gallery/WhatsApp Image 2026-10-15 at 5.31.06 PM.jpeg')" }}>
      <div className="wrap banner-inner">
        <h1>Hi, we're {Brand}!</h1>
        <p className="banner-sub">Three friends who'd rather be on a mountain than anywhere else, and who'd love to take you along.</p>
      </div>
    </section>
  );
  return (
    <PageShell hero={hero}>
      <section className="about-story">
        <div>
          <h2>How it started</h2>
          <p>One of us has finished a full Ironman. Another races 24-hour adventure races and spends most weekends in the Ghats. The third has walked just about every major trail in Karnataka, and crossed Hampta Pass in the Himalayas.</p>
          <p>Somewhere along the way we noticed that the trek lead makes or breaks the day: the right pace, a calm voice on a slippery descent, knowing exactly where to stop for the view. So we started Beyond Expeditions to give every trekker that kind of day.</p>
        </div>
        <div className="about-values">
          <div><strong>Safety, always</strong><span>Every batch has an experienced lead and a first-aid kit. If the weather turns, we turn back. No summit is worth it.</span></div>
          <div><strong>Small groups</strong><span>We keep batches small, so nobody gets left behind (or stuck in a queue for photos!).</span></div>
          <div><strong>No surprises</strong><span>Everything included is listed on the trek page. What you see is what you pay.</span></div>
          <div><strong>Leave no trace</strong><span>We carry back everything we carry in, and usually a little more.</span></div>
        </div>
      </section>

      <div className="photo-strip">
        <img src="/photos/gallery/WhatsApp Image 2026-10-15 at 5.30.50 PM (1).jpeg" alt="Rolling grassland in the Western Ghats" loading="lazy" />
        <img src="/photos/gallery/WhatsApp Image 2026-10-15 at 5.30.42 PM.jpeg" alt="Cliffs above the clouds" loading="lazy" />
        <img src="/photos/gallery/WhatsApp Image 2026-10-15 at 5.41.26 PM.jpeg" alt="Sunrise over a sea of clouds" loading="lazy" />
      </div>

      <section className="section">
        <div className="section-head">
          <h2>How booking works</h2>
        </div>
        <ol className="steps">
          <li><strong>Pick a trek</strong><span>Choose a trail and a date that suits you.</span></li>
          <li><strong>Send a request</strong><span>It's free. Tell us your pickup point and any medical notes.</span></li>
          <li><strong>We confirm</strong><span>We call or WhatsApp you to confirm your seat and share payment details.</span></li>
          <li><strong>Trek!</strong><span>Show up, meet your group and lead, and enjoy the hills.</span></li>
        </ol>
      </section>

      <section className="cta-band">
        <h2>Are you a trek lead?</h2>
        <p>Love the trails and looking after people? Come lead with us. Your profile grows with every group you take up the mountain.</p>
        <Link className="btn big" to="/lead-with-us">Lead with us</Link>
      </section>
    </PageShell>
  );
}

export function Contact() {
  const site = useStatic('Contact');
  const { User, Msg } = useSession();
  const { submit, busy, error } = useSubmit();
  const done = (Msg || '').startsWith("Thanks! We've got your enquiry");
  return (
    <PageShell>
      <header className="page-head">
        <h1>Contact us</h1>
        <p className="muted">Got a question about a trek, planning a group trip, or want to lead with us? WhatsApp is the fastest way to reach us. Or leave your number below and we'll call you back.</p>
      </header>
      {site && (
        <div className="contact-grid">
          <a className="contact-card" href={whatsappTo(site.WhatsApp, 'Hi! I have a question about your treks.')} target="_blank" rel="noopener noreferrer"><Icon name="wabrand" /><strong>WhatsApp</strong><em>Usually replies within a few hours</em></a>
          <a className="contact-card" href={`https://instagram.com/${site.Instagram}`} target="_blank" rel="noopener noreferrer"><Icon name="igbrand" /><strong>Instagram</strong><em>@{site.Instagram} · photos from every trek</em></a>
          <a className="contact-card" href={`mailto:${site.Email}`}><Icon name="gmailbrand" /><strong>Email</strong><em>Best for group bookings and longer questions</em></a>
        </div>
      )}

      <section className="contact-form panel" id="enquire">
        <h2>Ask us a question</h2>
        {done ? (
          <p><b>✓ Got it, thanks!</b> We'll call or WhatsApp you soon.</p>
        ) : (
          <form onSubmit={submit('/contact')} className="form">
            {error && <div className="flash err">{error}</div>}
            <div className="row">
              <label>Your name <input name="name" required maxLength={100} autoComplete="name" defaultValue={User ? User.Name : ''} /></label>
              <label>Phone (WhatsApp) <input name="phone" type="tel" required inputMode="tel" placeholder="98765 43210" autoComplete="tel" defaultValue={User ? User.Phone : ''} /></label>
            </div>
            <label>Email (optional) <input name="email" type="email" maxLength={200} autoComplete="email" autoCapitalize="none" spellCheck="false" /></label>
            <label>Your question <textarea name="message" rows="4" maxLength={2000} placeholder="Group booking, custom dates, which trek suits you…" /></label>
            <input className="hp" name="website" tabIndex={-1} autoComplete="off" aria-hidden="true" />
            <button className="btn" disabled={busy}>Send question</button>
          </form>
        )}
      </section>
      {site && <p className="muted small">Based in {site.City}. Treks depart from Bangalore most weekends. See the <Link to="/cancellation-policy">cancellation policy</Link>.</p>}
    </PageShell>
  );
}

const year = () => new Date().getFullYear();

export function Terms() {
  const site = useStatic('Terms of service');
  const { Brand } = useSession();
  return (
    <PageShell>
      <article className="legal">
        <h1>Terms of service</h1>
        <p className="muted">Last updated: {year()}</p>

        <h2>1. About these terms</h2>
        <p>These terms apply when you use this website or join a trek organised by {Brand} ("we", "us"). By sending a booking request you agree to them.</p>

        <h2>2. Bookings</h2>
        <p>A booking request is not a confirmed booking. Your seat is confirmed only when we confirm it with you and receive payment. Prices are per person and include only the items listed under "Included" on the trek page.</p>

        <h2>3. Fitness and health</h2>
        <p>Trekking is physically demanding. You confirm that you are fit enough for the trek you choose and that you have told us about any medical conditions that could affect you on the trail. Participants under 18 must be accompanied by a parent or guardian.</p>

        <h2>4. Safety and conduct</h2>
        <p>You agree to follow your trek lead's instructions at all times. Alcohol, smoking and littering are not allowed on our treks. We may ask anyone who puts themselves or the group at risk to stop trekking.</p>

        <h2>5. Risk</h2>
        <p>Trekking in the outdoors carries risks, including injury from slips, falls, weather and wildlife. We take every reasonable precaution, but by joining a trek you accept these risks.</p>

        <h2 id="insurance">6. Insurance</h2>
        <p><strong>Our treks do not include insurance of any kind.</strong> The trek price does not cover travel, personal accident, medical or evacuation insurance, and we do not insure participants. If you are injured or fall ill on a trek, any medical, hospital, rescue or extra transport costs are yours to pay. Our trek lead will help you get to the nearest road or medical help.</p>
        <p>If you'd like cover, you can buy personal accident or travel insurance on your own before the trek. Check that it covers trekking and hiking.</p>

        <h2>7. Cancellations</h2>
        <p>Cancellations and refunds follow our <Link to="/cancellation-policy">cancellation &amp; refund policy</Link>.</p>

        <h2>8. Photos</h2>
        <p>We may take photos and videos during treks and use them on our website and social media. If you'd rather not appear, just tell your trek lead.</p>

        <h2>9. Contact</h2>
        <p>Questions about these terms? {site && <a href={`mailto:${site.Email}`}>Email us</a>}.</p>
      </article>
    </PageShell>
  );
}

export function Privacy() {
  const site = useStatic('Privacy policy');
  return (
    <PageShell>
      <article className="legal">
        <h1>Privacy policy</h1>
        <p className="muted">Last updated: {year()}</p>

        <h2>What we collect</h2>
        <p>When you create an account or request a booking, we collect your name, email, phone number, age, emergency contact, pickup point, and any medical or fitness notes you choose to share.</p>

        <h2>Why we collect it</h2>
        <ul>
          <li>To confirm and manage your bookings and contact you about them</li>
          <li>To keep you safe on the trail. Your trek lead sees your emergency contact and medical notes</li>
          <li>To show your reviews (with your name) on trek and trek lead pages</li>
        </ul>

        <h2>What we don't do</h2>
        <p>We don't sell your data or share it with advertisers. We only share it with the trek lead for your batch, and with authorities if required by law or in an emergency.</p>

        <h2>Cookies</h2>
        <p>We use a single cookie to keep you logged in. We don't use tracking or advertising cookies.</p>

        <h2>Your choices</h2>
        <p>You can ask us to correct or delete your data at any time by {site && <a href={`mailto:${site.Email}`}>emailing us</a>}.</p>
      </article>
    </PageShell>
  );
}

export function Cancellation() {
  const site = useStatic('Cancellation & refund policy');
  return (
    <PageShell>
      <article className="legal">
        <h1>Cancellation &amp; refund policy</h1>
        <p className="muted">Last updated: {year()}</p>

        <h2>If you cancel</h2>
        <div className="table-wrap">
          <table className="table">
            <tbody>
              <tr><th>When you cancel</th><th>Refund</th></tr>
              <tr><td>7 or more days before the trek</td><td>Full refund, minus a ₹200 processing fee</td></tr>
              <tr><td>3 to 6 days before the trek</td><td>50% refund, or move to another date for free</td></tr>
              <tr><td>Less than 3 days before the trek</td><td>No refund, but you can move to another date once</td></tr>
            </tbody>
          </table>
        </div>
        <p>Cancelling a booking <em>request</em> before we've confirmed it is always free. To cancel a confirmed booking, use <Link to="/my">My Bookings</Link> or message us on WhatsApp.</p>

        <h2>If we cancel</h2>
        <p>We may cancel or reschedule a trek because of heavy rain, forest department closures, unsafe conditions or too few participants. If that happens, you can choose a full refund or a free move to another date or trek.</p>

        <h2>During the trek</h2>
        <p>Your trek lead may turn the group back, or ask a participant not to continue, if conditions or someone's health make it unsafe. Once a trek has started, no refund is due in that case, but we'll always help you get back safely.</p>

        <h2>If you can't finish the trek</h2>
        <p>There is no refund, partial refund or date transfer if you leave the trek or can't complete it for your own reasons, for example:</p>
        <ul>
          <li>you can't keep up with the group or feel the trek is too hard</li>
          <li>illness, injury, dehydration or any other health problem</li>
          <li>you decide to stop</li>
          <li>you are asked to leave for breaking our rules, such as drinking alcohol, smoking or putting others at risk</li>
        </ul>
        <p>Any extra transport, medical or evacuation costs to leave early are yours to pay. Our trek price does not include insurance; see <Link to="/terms#insurance">Insurance</Link>.</p>

        <h2>How refunds are paid</h2>
        <p>Refunds go back to the original payment method within 7 working days of approval.</p>

        <h2>Questions</h2>
        <p>{site && <a href={whatsappTo(site.WhatsApp, 'Hi! I have a question about cancelling my booking.')} target="_blank" rel="noopener noreferrer">Message us on WhatsApp</a>}, {site && <a href={`mailto:${site.Email}`}>email us</a>} or <Link to="/contact#enquire">send a question</Link> and we'll get back to you.</p>
      </article>
    </PageShell>
  );
}
