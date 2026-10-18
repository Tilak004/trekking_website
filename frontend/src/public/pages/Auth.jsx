import { Link, useLocation } from 'react-router-dom';
import { usePage, useTitle } from '../../lib/usePage.js';
import { useSubmit } from '../../lib/useSubmit.js';
import PageShell from '../PageShell.jsx';

const toggle = (e) => {
  e.currentTarget.form.password.type = e.currentTarget.checked ? 'text' : 'password';
};

export function Login() {
  const page = usePage();
  const { search } = useLocation();
  const { submit, busy, error } = useSubmit();
  useTitle('Log in');
  const next = new URLSearchParams(search).get('next') || '';
  const d = page.data;
  return (
    <PageShell>
      <div className="auth panel">
        <h1>Log in</h1>
        {error && <div className="flash err">{error}</div>}
        <form onSubmit={submit('/login')} className="form">
          <input type="hidden" name="next" value={next} />
          <label>Email <input type="email" name="email" defaultValue={d ? d.Email : ''} required autoFocus autoComplete="username" autoCapitalize="none" spellCheck="false" /></label>
          <label>Password <input type="password" name="password" required autoComplete="current-password" /></label>
          <label className="check small"><input type="checkbox" onChange={toggle} /> Show password</label>
          <button className="btn block" disabled={busy}>Log in</button>
        </form>
        <p className="muted small">No account yet? <Link to="/signup">Sign up</Link></p>
      </div>
    </PageShell>
  );
}

export function Signup() {
  const page = usePage();
  const { submit, busy, error } = useSubmit();
  useTitle('Sign up');
  const f = (page.data && page.data.Form) || {};
  return (
    <PageShell>
      <div className="auth panel">
        <h1>Create your account</h1>
        {error && <div className="flash err">{error}</div>}
        <form onSubmit={submit('/signup')} className="form">
          <label>Full name <input name="name" defaultValue={f.Name} required maxLength={100} autoFocus autoComplete="name" /></label>
          <label>Email <input type="email" name="email" defaultValue={f.Email} required maxLength={200} autoComplete="email" autoCapitalize="none" spellCheck="false" /></label>
          <label>Phone (WhatsApp) <input type="tel" name="phone" defaultValue={f.Phone} required maxLength={30} inputMode="tel" autoComplete="tel" /></label>
          <label>Password <input type="password" name="password" minLength={8} maxLength={72} required autoComplete="new-password" /></label>
          <label className="check small"><input type="checkbox" onChange={toggle} /> Show password</label>
          <button className="btn block" disabled={busy}>Sign up</button>
        </form>
        <p className="muted small">Already have an account? <Link to="/login">Log in</Link></p>
      </div>
    </PageShell>
  );
}
