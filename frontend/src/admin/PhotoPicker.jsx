import { useRef, useState } from 'react';
import { post } from '../lib/api.js';
import { photoURL } from '../lib/util.js';

// Uploads image files to the server one by one: [{ path }] for the files that
// went through, and the first error message otherwise.
async function uploadFiles(files, folder) {
  const done = [];
  let error = '';
  for (const file of files) {
    if (!/^image\//.test(file.type)) {
      error = `${file.name} isn't an image.`;
      continue;
    }
    const body = new FormData();
    body.append('file', file);
    body.append('folder', folder);
    try {
      done.push(await post('/admin/photos/upload', body));
    } catch (e) {
      error = `${file.name}: ${e.message}`;
    }
  }
  return { done, error };
}

// A box you can drop image files onto, or click to choose them.
function DropZone({ multiple, onFiles, busy, children }) {
  const input = useRef(null);
  const [over, setOver] = useState(false);
  const take = (list) => {
    const files = [...list];
    if (files.length) onFiles(multiple ? files : files.slice(0, 1));
  };
  return (
    <div
      className={`dropzone${over ? ' over' : ''}${busy ? ' busy' : ''}`}
      role="button"
      tabIndex={0}
      onClick={() => input.current && input.current.click()}
      onKeyDown={(e) => (e.key === 'Enter' || e.key === ' ') && (e.preventDefault(), input.current && input.current.click())}
      onDragOver={(e) => { e.preventDefault(); setOver(true); }}
      onDragLeave={() => setOver(false)}
      onDrop={(e) => { e.preventDefault(); setOver(false); take(e.dataTransfer.files); }}
    >
      <input ref={input} type="file" accept="image/jpeg,image/png,image/webp" multiple={multiple} hidden
        onChange={(e) => { take(e.target.files); e.target.value = ''; }} />
      {children}
    </div>
  );
}

// One photo chosen for a form field (a trek's cover, a lead's profile): drag a file
// onto the box or click it to upload from your computer, or pick one already on the
// site. Submits the chosen path as <name>.
export function PhotoPicker({ name, value, options, folder = 'uploads', emptyLabel, hint }) {
  const [current, setCurrent] = useState(value || '');
  const [extra, setExtra] = useState([]);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  const all = [...new Set([...(options || []), ...extra, ...(current ? [current] : [])])];

  const onFiles = async (files) => {
    setBusy(true);
    setError('');
    const { done, error: err } = await uploadFiles(files, folder);
    setBusy(false);
    if (err) setError(err);
    if (done.length) {
      setExtra((x) => [...x, done[0].path]);
      setCurrent(done[0].path);
    }
  };

  return (
    <div className="photo-picker">
      <input type="hidden" name={name} value={current} />
      <DropZone onFiles={onFiles} busy={busy}>
        {current ? <img src={photoURL(current)} alt="" /> : <span className="dz-empty" aria-hidden="true">+</span>}
        <span className="dz-text">
          {busy ? 'Uploading…' : <><strong>Drag a photo here</strong> or click to upload from your computer</>}
          <small>JPG, PNG or WebP, up to 15 MB</small>
        </span>
      </DropZone>
      {error && <div className="flash err">{error}</div>}
      <label className="dz-existing">Or choose one already on the site
        <select value={current} onChange={(e) => setCurrent(e.target.value)}>
          <option value="">{emptyLabel}</option>
          {all.map((p) => <option key={p} value={p}>{p}</option>)}
        </select>
      </label>
      {hint && <span className="hint">{hint}</span>}
    </div>
  );
}

// A trek's own photo gallery (the "Photos" section on its page): drop several at
// once, remove ones you don't want. Changes apply straight away, without the Save button.
export function TrekPhotos({ slug, photos }) {
  const [list, setList] = useState(photos || []);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');

  const onFiles = async (files) => {
    setBusy(true);
    setError('');
    const { done, error: err } = await uploadFiles(files, `treks/${slug}`);
    setBusy(false);
    if (err) setError(err);
    if (done.length) setList((l) => [...l, ...done.map((r) => r.path)]);
  };
  const remove = async (p) => {
    if (!window.confirm('Remove this photo from the trek page?')) return;
    try {
      await post('/admin/photos/delete', { path: p });
      setList((l) => l.filter((x) => x !== p));
    } catch (e) {
      setError(e.message);
    }
  };

  return (
    <div className="photo-picker trek-photos">
      <DropZone multiple onFiles={onFiles} busy={busy}>
        <span className="dz-empty" aria-hidden="true">+</span>
        <span className="dz-text">
          {busy ? 'Uploading…' : <><strong>Drag photos here</strong> or click to choose several</>}
          <small>JPG, PNG or WebP, up to 15 MB each. Shown in file-name order.</small>
        </span>
      </DropZone>
      {error && <div className="flash err">{error}</div>}
      {list.length > 0 && (
        <ul className="dz-grid">
          {list.map((p) => (
            <li key={p}>
              <img src={photoURL(p)} alt="" loading="lazy" />
              <button type="button" className="dz-remove" onClick={() => remove(p)} aria-label="Remove photo" title="Remove">×</button>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
