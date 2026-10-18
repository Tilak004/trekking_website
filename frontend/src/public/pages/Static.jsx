import { Link } from 'react-router-dom';
import Icon from '../../components/Icon.jsx';
import { useSession } from '../../lib/session.jsx';
import { usePage, useTitle } from '../../lib/usePage.js';
import { telLink } from '../../lib/util.js';
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
        <p className="eyebrow light">About us</p>
        <h1>Three athletes. One trail.</h1>
        <p className="banner-sub">We started {Brand} to share the Western Ghats the way we love them: properly prepared, at the right pace and in good company.</p>
      </div>
    </section>
  );
  return (
    <PageShell hero={hero}>
      <section className="about-story">
        <div>
          <p className="eyebrow">Our story</p>
          <h2>Built by people who race mountains</h2>
          <p>One of us has finished a full Ironman. One races 24-hour adventure races and runs the Ghats most weekends. One has completed every major trek in Karnataka and crossed Hampta Pass in the Himalayas.</p>
          <p>Between us we've seen how much a good trek lead changes the day: the right pace, a calm voice on a slippery descent, and knowing exactly where the best view is. That's what we want every trekker to get, every weekend.</p>
        </div>
        <div className="about-values">
          <div><strong>Safety first</strong><span>Every batch has an experienced lead and a first-aid kit. If the weather turns, we turn back.</span></div>
          <div><strong>Right-sized groups</strong><span>Limited seats per batch so nobody gets left behind.</span></div>
          <div><strong>Honest pricing</strong><span>What's included is listed on every trek. No surprises at the base village.</span></div>
          <div><strong>Leave no trace</strong><span>We carry back everything we carry in, and a little more.</span></div>
        </div>
      </section>

      <div className="photo-strip">
        <img src="/photos/gallery/WhatsApp Image 2026-10-15 at 5.30.50 PM (1).jpeg" alt="Rolling grassland in the Western Ghats" loading="lazy" />
        <img src="/photos/gallery/WhatsApp Image 2026-10-15 at 5.30.42 PM.jpeg" alt="Cliffs above the clouds" loading="lazy" />
        <img src="/photos/gallery/WhatsApp Image 2026-10-15 at 5.41.26 PM.jpeg" alt="Sunrise over a sea of clouds" loading="lazy" />
      </div>

      <section className="section">
        <div className="section-head">
          <p className="eyebrow">How it works</p>
          <h2>From request to summit</h2>
        </div>
        <ol className="steps">
          <li><strong>Pick a trek</strong><span>Choose a trail and a date that suits you.</span></li>
          <li><strong>Send a request</strong><span>It's free. Tell us your pickup point and any medical notes.</span></li>
          <li><strong>We confirm</strong><span>We call or WhatsApp you to confirm your seat and share payment details.</span></li>
          <li><strong>Trek</strong><span>Show up, meet your group and lead, and enjoy the Ghats.</span></li>
        </ol>
      </section>

      <section className="cta-band">
        <h2>Are you a trek lead?</h2>
        <p>We're building a home for great trek leads, with a profile that grows with every group you take up the mountain.</p>
        <Link className="btn big" to="/lead-with-us">Lead with us</Link>
      </section>
    </PageShell>
  );
}

export function Contact() {
  const site = useStatic('Contact');
  return (
    <PageShell>
      <header className="page-head">
        <p className="eyebrow">Get in touch</p>
        <h1>Contact us</h1>
        <p className="muted">Questions about a trek, a group booking or leading treks with us? Give us a call, or reach us on WhatsApp, Instagram or email using the icons at the bottom of the page.</p>
      </header>
      {site && (
        <>
          <div className="contact-grid phones">
            <a className="contact-card" href={`tel:${telLink(site.Phone)}`}><Icon name="phonebrand" /><strong>{site.Phone}</strong><em>Call or WhatsApp · 9 AM – 8 PM</em></a>
            {site.Phone2 && <a className="contact-card" href={`tel:${telLink(site.Phone2)}`}><Icon name="phonebrand" /><strong>{site.Phone2}</strong><em>Call · 9 AM – 8 PM</em></a>}
          </div>
          <p className="muted small">Based in {site.City}. Treks depart from Bangalore most weekends. See the <Link to="/cancellation-policy">cancellation policy</Link> and <Link to="/#treks">upcoming treks</Link>.</p>
        </>
      )}
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
        <p className="eyebrow">Policies</p>
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
        <p>Trekking in the outdoors carries risks, including injury from slips, falls, weather and wildlife. We take every reasonable precaution, but by joining a trek you accept these risks. We recommend personal accident insurance.</p>

        <h2>6. Cancellations</h2>
        <p>Cancellations and refunds follow our <Link to="/cancellation-policy">cancellation &amp; refund policy</Link>.</p>

        <h2>7. Photos</h2>
        <p>We may take photos and videos during treks and use them on our website and social media. If you'd rather not appear, just tell your trek lead.</p>

        <h2>8. Contact</h2>
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
        <p className="eyebrow">Policies</p>
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
        <p className="eyebrow">Policies</p>
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
        <p>Your trek lead may turn the group back, or ask a participant not to continue, if conditions or someone's health make it unsafe. No refund is due in that case, but we'll always help you get back safely.</p>

        <h2>How refunds are paid</h2>
        <p>Refunds go back to the original payment method within 7 working days of approval.</p>

        <h2>Questions</h2>
        <p>Write to us by email or call {site && <>{site.Phone}{site.Phone2 ? ` / ${site.Phone2}` : ''}</>}.{site && <> <a href={`mailto:${site.Email}`}>Email us</a></>}</p>
      </article>
    </PageShell>
  );
}
