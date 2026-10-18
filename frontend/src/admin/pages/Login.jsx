import { useLocation } from 'react-router-dom';
import { useSession } from '../../lib/session.jsx';
import { usePage, useTitle } from '../../lib/usePage.js';
import { useSubmit } from '../../lib/useSubmit.js';

export default function Login() {
  usePage();
  const { search } = useLocation();
  const { Brand } = useSession();
  const { submit, busy, error } = useSubmit();
  useTitle('Admin login');
  const next = new URLSearchParams(search).get('next') || '';
  return (
    <section className="c-login-card">
      <img src="/photos/logo/badge.png" alt="" onError={(e) => { e.currentTarget.onerror = null; e.currentTarget.src = '/static/logo/mark-small.png'; }} />
      <h1>{Brand}</h1>
      <p className="muted">Admin console. Sign in with an admin account.</p>
      {error && <div className="flash err">{error}</div>}
      <form onSubmit={submit('/login')} className="form">
        <input type="hidden" name="next" value={next} />
        <label>Email <input type="email" name="email" autoComplete="username" autoCapitalize="none" spellCheck="false" required autoFocus /></label>
        <label>Password <input type="password" name="password" autoComplete="current-password" required /></label>
        <label className="check small"><input type="checkbox" onChange={(e) => { e.currentTarget.form.password.type = e.currentTarget.checked ? 'text' : 'password'; }} /> Show password</label>
        <button className="btn big" disabled={busy}>Sign in</button>
      </form>
    </section>
  );
}
