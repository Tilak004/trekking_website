import icons from '../lib/iconData.js';

// SVG icons copied from the Go layout. Pass name="compass" | "users" | ... or a
// brand logo: igbrand, wabrand, gmailbrand, phonebrand, waicon.
export default function Icon({ name }) {
  const ic = icons[name];
  if (!ic) return null;
  return <svg viewBox={ic.viewBox} aria-hidden="true" dangerouslySetInnerHTML={{ __html: ic.inner }} />;
}
