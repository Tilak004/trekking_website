import { useEffect, useState } from 'react';
import { PageState } from '../../components/common.jsx';
import { usePage, useTitle } from '../../lib/usePage.js';
import { photoURL } from '../../lib/util.js';
import PageShell from '../PageShell.jsx';

export default function Gallery() {
  const page = usePage();
  const [open, setOpen] = useState(null); // index of the photo shown full size
  const photos = (page.data && page.data.Gallery) || [];
  useTitle('Gallery');

  useEffect(() => {
    if (open === null) return undefined;
    const onKey = (e) => {
      if (e.key === 'Escape') setOpen(null);
      if (e.key === 'ArrowLeft') setOpen((i) => (i - 1 + photos.length) % photos.length);
      if (e.key === 'ArrowRight') setOpen((i) => (i + 1) % photos.length);
    };
    document.addEventListener('keydown', onKey);
    return () => document.removeEventListener('keydown', onKey);
  }, [open, photos.length]);

  if (!page.data) return <PageShell><PageState page={page} /></PageShell>;
  const go = (n) => setOpen((n + photos.length) % photos.length);

  return (
    <PageShell>
      <header className="page-head">
        <h1>Gallery</h1>
        <p className="muted">Snapshots from our treks. Tap any photo to see it full size.</p>
      </header>
      <div className="gallery">
        {photos.length
          ? photos.map((p, i) => (
            <a href={photoURL(p)} key={p} onClick={(e) => { e.preventDefault(); setOpen(i); }}>
              <img src={photoURL(p)} alt="Trek photo" loading="lazy" />
            </a>
          ))
          : <p className="muted">Photos coming soon.</p>}
      </div>

      {open !== null && (
        <div className="lightbox" onClick={(e) => { if (e.target === e.currentTarget) setOpen(null); }}>
          <button className="lb-close" aria-label="Close" onClick={() => setOpen(null)}>×</button>
          <button className="lb-prev" aria-label="Previous photo" onClick={() => go(open - 1)}>‹</button>
          <img src={photoURL(photos[open])} alt="Trek photo" />
          <button className="lb-next" aria-label="Next photo" onClick={() => go(open + 1)}>›</button>
        </div>
      )}
    </PageShell>
  );
}
