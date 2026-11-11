import { useEffect, useState } from 'react';
import { photoURL } from '../lib/util.js';

// A grid of photos; tapping one opens it full size with previous / next.
export default function PhotoGrid({ photos, alt = 'Trek photo', url = photoURL }) {
  const [open, setOpen] = useState(null); // index of the photo shown full size

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

  const go = (n) => setOpen((n + photos.length) % photos.length);

  return (
    <>
      <div className="gallery">
        {photos.map((p, i) => (
          <a href={url(p)} key={p} onClick={(e) => { e.preventDefault(); setOpen(i); }}>
            <img src={url(p)} alt={alt} loading="lazy" />
          </a>
        ))}
      </div>

      {open !== null && (
        <div className="lightbox" onClick={(e) => { if (e.target === e.currentTarget) setOpen(null); }}>
          <button className="lb-close" aria-label="Close" onClick={() => setOpen(null)}>×</button>
          <button className="lb-prev" aria-label="Previous photo" onClick={() => go(open - 1)}>‹</button>
          <img src={url(photos[open])} alt={alt} />
          <button className="lb-next" aria-label="Next photo" onClick={() => go(open + 1)}>›</button>
        </div>
      )}
    </>
  );
}
