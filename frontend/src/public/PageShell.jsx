import { Flash } from '../components/common.jsx';
import { useSession } from '../lib/session.jsx';

// <main class="wrap"> with the server's flash message, preceded by an optional hero.
export default function PageShell({ hero, children }) {
  const { Msg, Err } = useSession();
  return (
    <>
      {hero}
      <main className="wrap">
        <Flash msg={Msg} err={Err} />
        {children}
      </main>
    </>
  );
}
