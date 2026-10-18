import { useCallback, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { ApiError, post } from './api.js';

// Submits a form to the Go API.
//   const { submit, busy, error } = useSubmit();
//   <form onSubmit={submit('/login')}>
// On {"redirect"} it navigates there. If the server sent the form page back with
// an error message (e.g. wrong password) that message is shown instead.
export function useSubmit() {
  const navigate = useNavigate();
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');

  const run = useCallback(
    async (path, form, { confirm: ask, onDone } = {}) => {
      if (ask && !window.confirm(ask)) return;
      setBusy(true);
      setError('');
      try {
        const res = await post(path, form);
        if (res && res.redirect) {
          const here = window.location.pathname + window.location.search;
          if (res.redirect === here) window.dispatchEvent(new Event('beyond:refresh'));
          else navigate(res.redirect);
        } else if (res && res.Err) {
          setError(res.Err);
        }
        if (onDone) onDone(res);
        return res;
      } catch (e) {
        if (e instanceof ApiError && e.status === 401 && e.data && e.data.redirect) navigate(e.data.redirect);
        else setError(e.message || 'Something went wrong. Please try again.');
      } finally {
        setBusy(false);
      }
    },
    [navigate],
  );

  // submit('/path') returns an onSubmit handler for a <form>.
  const submit = useCallback(
    (path, opts) => (e) => {
      e.preventDefault();
      return run(typeof path === 'function' ? path(e.currentTarget) : path, e.currentTarget, opts);
    },
    [run],
  );

  return { submit, run, busy, error, setError };
}
