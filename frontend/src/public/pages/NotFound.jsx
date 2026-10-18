import { Link } from 'react-router-dom';
import { useTitle } from '../../lib/usePage.js';
import PageShell from '../PageShell.jsx';

export default function NotFound() {
  useTitle('Page not found');
  return (
    <PageShell>
      <section className="notfound">
        <p className="eyebrow">Error 404</p>
        <h1>Looks like you've wandered off the trail.</h1>
        <p className="muted">The page you're looking for doesn't exist or has moved.</p>
        <Link className="btn big" to="/">Back to all treks</Link>
      </section>
    </PageShell>
  );
}
