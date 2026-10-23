import { createContext, useContext, useEffect, useRef, useState } from 'react';
import { Link, useLocation } from 'react-router-dom';
import Icon from '../components/Icon.jsx';
import { useSession } from '../lib/session.jsx';
import { useSubmit } from '../lib/useSubmit.js';
import { initial, slug, whatsappTo } from '../lib/util.js';

// The WhatsApp button's message can be changed by a page (the trek page does).
const WaCtx = createContext(() => {});
export const useWhatsAppMessage = (text) => {
  const set = useContext(WaCtx);
  useEffect(() => {
    set(text || '');
    return () => set('');
  }, [text, set]);
};

// New page: back to the top, or to the #section named in the URL once it exists.
function ScrollManager() {
  const { pathname, hash, key } = useLocation();
  useEffect(() => {
    if (!hash) {
      window.scrollTo(0, 0);
      return undefined;
    }
    let tries = 0;
    let timer;
    const find = () => {
      const el = document.getElementById(decodeURIComponent(hash.slice(1)));
      if (el) el.scrollIntoView();
      else if (tries++ < 40) timer = setTimeout(find, 50);
    };
    find();
    return () => clearTimeout(timer);
  }, [pathname, hash, key]);
  return null;
}

const logoFallback = (e) => {
  e.currentTarget.onerror = null;
  e.currentTarget.src = '/static/logo/mark-small.png';
  e.currentTarget.classList.remove('badge');
};

export default function Layout({ children }) {
  const { pathname } = useLocation();
  const s = useSession();
  const { submit } = useSubmit();
  const [menu, setMenu] = useState(false);
  const [wa, setWa] = useState('');
  const account = useRef(null);
  const site = s.Site;
  const year = new Date().getFullYear();

  useEffect(() => {
    setMenu(false);
    if (account.current) account.current.open = false;
  }, [pathname]);

  const on = (cond) => (cond ? 'on' : undefined);
  const isLead = s.Role === 'Trek lead';
  const waText = wa || "Hi! I'd like to know more about your treks.";

  return (
    <WaCtx.Provider value={setWa}>
      <ScrollManager />
      <header className={'site-header' + (pathname === '/' ? ' over-hero' : '')}>
        <div className="wrap nav">
          <Link className="brand" to="/" aria-label={`${s.Brand} home`}>
            <img className="logo badge" src="/photos/logo/badge.png" alt="" onError={logoFallback} />
            <span className="brand-name">
              <b><span className="bn-light">Beyond</span>Expeditions</b>
              {site && site.Tagline && <i>{site.Tagline}</i>}
            </span>
          </Link>
          <input type="checkbox" id="menu-toggle" className="menu-toggle" aria-label="Menu" checked={menu} onChange={(e) => setMenu(e.target.checked)} />
          <nav className="menu">
            <Link to="/" className={on(pathname === '/')}>Home</Link>
            <Link to="/#treks" className={on(pathname.startsWith('/treks'))}>Treks</Link>
            <Link to="/leads" className={on(pathname.startsWith('/leads') || pathname === '/lead-with-us')}>Trek Leads</Link>
            <Link to="/gallery" className={on(pathname === '/gallery')}>Gallery</Link>
            <Link to="/about" className={on(pathname === '/about')}>About</Link>
            <Link to="/contact" className={on(pathname === '/contact')}>Contact</Link>
          </nav>
          <div className="nav-right">
            {s.User ? (
              <details className="account" ref={account}>
                <summary aria-label="Account">{initial(s.User.Name)}</summary>
                <div className="account-menu">
                  <span className="acct-who">Signed in as <b>{s.User.Name}</b><em className={`role role-${slug(s.Role)}`}>{s.Role}</em></span>
                  {s.User.IsAdmin ? (
                    s.AdminURL ? <a href={s.AdminURL} target="_blank" rel="noopener noreferrer">Admin console ↗</a> : null
                  ) : (
                    <>
                      <Link to="/my">My bookings</Link>
                      {isLead && <Link to="/lead">Lead dashboard</Link>}
                      <Link to="/lead-with-us">{isLead ? 'My lead profile' : 'Become a trek lead'}</Link>
                    </>
                  )}
                  <form onSubmit={submit('/logout')}><button className="linklike">Log out</button></form>
                </div>
              </details>
            ) : (
              <Link className="btn small" to="/login">Log in</Link>
            )}
            <label htmlFor="menu-toggle" className="menu-btn" aria-hidden="true"><span /></label>
          </div>
        </div>
      </header>
      {s.Demo && <div className="demo-banner">DEMO SITE · sample data for testing. Real bookings are not taken here.</div>}

      {children}

      <footer className="site-footer">
        <div className="wrap footer-grid">
          <div className="footer-about">
            <Link className="brand" to="/">
              <img className="logo badge" src="/photos/logo/badge.png" alt={s.Brand} onError={logoFallback} />
            </Link>
            <p>{site && site.Description}</p>
          </div>
          <div>
            <h4>Explore</h4>
            <Link to="/#treks">All treks</Link>
            <Link to="/leads">Trek leads</Link>
            <Link to="/gallery">Gallery</Link>
            <Link to="/about">About us</Link>
            <Link to="/lead-with-us">Lead treks with us</Link>
            <Link to="/signup">Create an account</Link>
          </div>
          <div>
            <h4>Policies</h4>
            <Link to="/cancellation-policy">Cancellation &amp; refunds</Link>
            <Link to="/terms">Terms of service</Link>
            <Link to="/privacy">Privacy policy</Link>
          </div>
          <div>
            <h4>Contact</h4>
            {site && (
              <div className="socials">
                <a href={`https://instagram.com/${site.Instagram}`} target="_blank" rel="noopener noreferrer" aria-label="Instagram"><Icon name="igbrand" /></a>
                <a href={whatsappTo(site.WhatsApp, 'Hi! I have a question about your treks.')} target="_blank" rel="noopener noreferrer" aria-label="WhatsApp"><Icon name="wabrand" /></a>
                <a href={`mailto:${site.Email}`} aria-label="Email"><Icon name="gmailbrand" /></a>
              </div>
            )}
            <Link to="/contact#enquire">Send a question</Link>
            <span>{site && site.City}</span>
          </div>
        </div>
        <div className="wrap footer-bottom">
          <span>© {year} {s.Brand}. All rights reserved.</span>
        </div>
      </footer>
      {site && (
        <a className="wa-dot" href={whatsappTo(site.WhatsApp, waText)} target="_blank" rel="noopener noreferrer" title="Enquire on WhatsApp" aria-label="Enquire now on WhatsApp">
          <Icon name="waicon" /><span>Enquire now!</span>
        </a>
      )}
    </WaCtx.Provider>
  );
}
