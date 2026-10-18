import { Link } from 'react-router-dom';
import Icon from '../../components/Icon.jsx';
import { PageState } from '../../components/common.jsx';
import { useSession } from '../../lib/session.jsx';
import { usePage, useTitle } from '../../lib/usePage.js';
import { useSubmit } from '../../lib/useSubmit.js';
import { whatsappTo } from '../../lib/util.js';
import PageShell from '../PageShell.jsx';

function Hero() {
  return (
    <section className="banner short" style={{ backgroundImage: "url('/photos/gallery/WhatsApp Image 2026-10-15 at 5.30.49 PM.jpeg')" }}>
      <div className="wrap banner-inner">
        <p className="eyebrow light">For trek leads</p>
        <h1>Lead treks. Earn on weekends.</h1>
        <p className="banner-sub">Love the trails and know how to look after a group? Build your trek-lead profile, get matched with trekkers and get paid for every trek you lead.</p>
      </div>
    </section>
  );
}

function Status({ form, brand, site }) {
  const msg = {
    pending: ['Application under review', "We'll review your profile and get back to you in a few days."],
    approved: ['Your trek-lead profile is live', 'Keep it up to date. Every trek you lead and every review adds to it.'],
    removed: ['Profile removed', 'Your trek-lead profile has been taken off the platform. Contact us if you think this is a mistake.'],
  }[form.Status] || ['Not approved this time', 'Add more detail about your trekking experience and save to apply again.'];
  return (
    <div className={`lead-status ${form.Status}`}>
      <strong>{msg[0]}</strong>
      <span>{msg[1]}</span>
      {form.Status === 'pending' && site && (
        <a className="wa-link" href={whatsappTo(site.WhatsApp, `Hi! I've applied to be a trek lead on ${brand}. My name is ${form.Name}.`)} target="_blank" rel="noopener noreferrer">
          <Icon name="waicon" /> Let us know on WhatsApp
        </a>
      )}
      {form.Status === 'approved' && (
        <>
          <Link to="/lead"><strong>Open your lead dashboard →</strong></Link>
          <Link to={`/leads/${form.Slug}`}>View your public profile →</Link>
        </>
      )}
    </div>
  );
}

export default function LeadApply() {
  const page = usePage();
  const { User, Site, Brand } = useSession();
  const { submit, busy, error } = useSubmit();
  const d = page.data;
  useTitle('Lead treks with us');
  if (!d) return <PageShell hero={<Hero />}><PageState page={page} /></PageShell>;
  const f = d.Form || {};

  return (
    <PageShell hero={<Hero />}>
      <div className="two-col">
        <div>
          {d.HasLead && <Status form={f} brand={Brand} site={Site} />}

          {!User ? (
            <div className="panel">
              <h2>Apply to lead treks</h2>
              <p className="muted">You'll need a free account to apply, so we can keep you updated on your application.</p>
              <p><Link className="btn" to="/signup">Create an account</Link> <Link className="btn ghost" to="/login?next=/lead-with-us">Log in</Link></p>
            </div>
          ) : (
            <form key={`${f.ID}-${f.Status}`} onSubmit={submit('/lead-with-us')} encType="multipart/form-data" className="form panel">
              <h2>{d.HasLead ? 'Your trek-lead profile' : 'Apply to lead treks'}</h2>
              {error && <div className="flash err">{error}</div>}
              <div className="row">
                <label>Full name <input name="name" defaultValue={f.Name} required maxLength={100} /></label>
                <label>City <input name="city" defaultValue={f.City} required maxLength={80} placeholder="e.g. Bangalore" /></label>
              </div>
              <div className="row">
                <label>Phone (WhatsApp) <input name="phone" type="tel" defaultValue={f.Phone} required maxLength={30} />
                  <span className="hint">Only our team sees this.</span></label>
                <label>Profile photo <input name="photo" type="file" accept="image/jpeg,image/png,image/webp" />
                  <span className="hint">{f.Photo ? 'You have a photo. Upload a new one to replace it.' : 'A clear photo of you on a trek. JPG/PNG, under 5 MB.'}</span></label>
              </div>
              <label>Headline <input name="headline" defaultValue={f.Headline} required maxLength={150} placeholder="e.g. Weekend trekker · 30+ Western Ghats treks" /></label>
              <label>About you
                <textarea name="bio" rows="5" required maxLength={3000} defaultValue={f.Bio} placeholder="How you got into trekking, what kind of treks you love, and how you look after a group." /></label>
              <div className="row">
                <label>Treks you've completed (one per line)
                  <textarea name="achievements" rows="6" required defaultValue={f.Achievements} placeholder={'Kudremukh\nKumara Parvatha\nHampta Pass'} /></label>
                <label>Certifications (one per line, optional)
                  <textarea name="certifications" rows="6" defaultValue={f.Certifications} placeholder={'Wilderness first aid\nBasic Mountaineering Course (BMC)'} /></label>
              </div>
              {!d.HasLead && (
                <label className="check"><input type="checkbox" name="agree" required /> I confirm these details are true and that I'm 18 or older.</label>
              )}
              <button className="btn" disabled={busy}>{d.HasLead ? 'Save profile' : 'Submit application'}</button>
            </form>
          )}
        </div>

        <aside className="panel">
          <h3>How it works</h3>
          <ol className="mini-steps">
            <li><strong>Apply</strong><span>Tell us about your trekking experience.</span></li>
            <li><strong>Get verified</strong><span>We check your ID and experience, and may do a trek together first.</span></li>
            <li><strong>Go live</strong><span>Your profile appears on the site and you're assigned treks.</span></li>
            <li><strong>Lead &amp; earn</strong><span>Get paid for every trek, and grow your rating with every group.</span></li>
          </ol>
          <h3>What we look for</h3>
          <ul className="tight-list">
            <li>Several treks completed, ideally in the Western Ghats</li>
            <li>Calm, friendly and comfortable leading a group</li>
            <li>First-aid training is a big plus</li>
            <li>Available on weekends</li>
          </ul>
        </aside>
      </div>
    </PageShell>
  );
}
