import { useCallback, useEffect, useRef, useState } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { ApiError, get } from './api.js';
import { useSession } from './session.jsx';

// Loads the data for the current URL from the Go API.
//   const { data, loading, notFound, error, reload } = usePage();
// A 401 sends the person to the login page; a {"redirect"} answer is followed.
export function usePage(pathOverride) {
  const { pathname, search } = useLocation();
  const navigate = useNavigate();
  const session = useSession();
  const url = (pathOverride || pathname) + search;
  const [state, setState] = useState({ url: null, data: null, notFound: false, error: '' });
  const seq = useRef(0);

  const load = useCallback(async () => {
    const mine = ++seq.current;
    try {
      const data = await get(url);
      if (mine !== seq.current) return;
      if (data && data.redirect) {
        navigate(data.redirect, { replace: true });
        return;
      }
      session.update(data);
      setState({ url, data, notFound: false, error: '' });
    } catch (e) {
      if (mine !== seq.current) return;
      if (e instanceof ApiError && e.status === 401 && e.data && e.data.redirect) {
        navigate(e.data.redirect, { replace: true });
        return;
      }
      if (e instanceof ApiError && e.status === 404) {
        if (e.data) session.update(e.data);
        setState({ url, data: null, notFound: true, error: '' });
        return;
      }
      setState({ url, data: null, notFound: false, error: e.message || 'Something went wrong. Please try again.' });
    }
  }, [url]); // eslint-disable-line react-hooks/exhaustive-deps

  useEffect(() => {
    load();
  }, [load]);

  // Forms that redirect back to the page they are on ask for a fresh copy.
  useEffect(() => {
    window.addEventListener('beyond:refresh', load);
    return () => window.removeEventListener('beyond:refresh', load);
  }, [load]);

  const fresh = state.url === url;
  return {
    data: fresh ? state.data : null,
    loading: !fresh,
    notFound: fresh && state.notFound,
    error: fresh ? state.error : '',
    reload: load,
  };
}

// document.title for the page.
export function useTitle(title) {
  const { Brand } = useSession();
  useEffect(() => {
    if (title) document.title = `${title} · ${Brand || 'Beyond Expeditions'}`;
  }, [title, Brand]);
}
