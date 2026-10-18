import { Link, useLocation } from 'react-router-dom';
import { Flash } from '../components/common.jsx';
import { useSession } from '../lib/session.jsx';
import { useSubmit } from '../lib/useSubmit.js';

const logoFallback = (e) => {
  e.currentTarget.onerror = null;
  e.currentTarget.src = '/static/logo/mark-small.png';
};

function Nav({ to, exact, children }) {
  const { pathname } = useLocation();
  const on = exact ? pathname === to : pathname.startsWith(to);
  return <Link to={to} className={on ? 'on' : undefined}>{children}</Link>;
}

// The console's frame: a sidebar when signed in, a bare page for the login.
export default function Shell({ children }) {
  const s = useSession();
  const { pathname } = useLocation();
  const { submit } = useSubmit();

  if (pathname === '/login' || !s.User) {
    return <main className="c-login">{children}</main>;
  }

  return (
    <div className="c-shell">
      <aside className="c-side">
        <Link className="c-brand" to="/admin">
          <img src="/photos/logo/badge.png" alt="" onError={logoFallback} />
          <span><b>{s.Brand}</b><i>Admin console</i></span>
        </Link>
        <nav className="c-nav" aria-label="Admin">
          <Nav to="/admin" exact>Dashboard</Nav>
          <Nav to="/admin/events">Events</Nav>
          <Nav to="/admin/registrations">Registrations</Nav>
          <Nav to="/admin/enquiries">Enquiries</Nav>
          <Nav to="/admin/analytics">Analytics</Nav>
          <span className="c-sep">Manage</span>
          <Nav to="/admin/treks">Treks</Nav>
          <Nav to="/admin/leads">Trek leads</Nav>
          <Nav to="/admin/users">Users</Nav>
        </nav>
        <div className="c-foot">
          <span>{s.User.Name}</span>
          <a href={s.PublicURL} target="_blank" rel="noopener noreferrer">View public site ↗</a>
          <form onSubmit={submit('/logout')}><button className="linklike">Log out</button></form>
        </div>
      </aside>
      <main className="c-main">
        <Flash msg={s.Msg} err={s.Err} />
        {children}
      </main>
    </div>
  );
}
