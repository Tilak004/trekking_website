import { PageState } from '../../components/common.jsx';
import PhotoGrid from '../../components/PhotoGrid.jsx';
import { usePage, useTitle } from '../../lib/usePage.js';
import PageShell from '../PageShell.jsx';

export default function Gallery() {
  const page = usePage();
  const photos = (page.data && page.data.Gallery) || [];
  useTitle('Gallery');

  if (!page.data) return <PageShell><PageState page={page} /></PageShell>;

  return (
    <PageShell>
      <header className="page-head">
        <h1>Gallery</h1>
        <p className="muted">Snapshots from our treks. Tap any photo to see it full size.</p>
      </header>
      {photos.length ? <PhotoGrid photos={photos} /> : <div className="gallery"><p className="muted">Photos coming soon.</p></div>}
    </PageShell>
  );
}
