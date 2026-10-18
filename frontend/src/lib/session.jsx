import { createContext, useCallback, useContext, useMemo, useState } from 'react';

// What every API response carries about the site and the signed-in person, plus
// the flash message the server attached (?msg= / ?err= codes).
const Ctx = createContext(null);

export function SessionProvider({ children }) {
  const [s, setS] = useState({ User: null, Role: '', Site: null, Brand: '', Msg: '', Err: '', Demo: false, AdminURL: '', PublicURL: '' });
  const update = useCallback((data) => {
    if (!data) return;
    setS((prev) => ({
      ...prev,
      User: data.User ?? null,
      Role: data.Role || '',
      Site: data.Site || prev.Site,
      Brand: data.Brand || prev.Brand,
      Demo: !!data.Demo,
      AdminURL: data.AdminURL || '',
      PublicURL: data.PublicURL || '',
      Msg: data.Msg || '',
      Err: data.Err || '',
    }));
  }, []);
  const value = useMemo(() => ({ ...s, update }), [s, update]);
  return <Ctx.Provider value={value}>{children}</Ctx.Provider>;
}

export const useSession = () => useContext(Ctx);
