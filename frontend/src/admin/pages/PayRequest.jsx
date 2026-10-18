import { useEffect, useRef, useState } from 'react';
import { Link } from 'react-router-dom';
import { PageState } from '../../components/common.jsx';
import { useSession } from '../../lib/session.jsx';
import { usePage, useTitle } from '../../lib/usePage.js';
import { firstWord, fmtDate, inr } from '../../lib/util.js';

// Shows the UPI QR and the buttons to send the payment request to the trekker on
// WhatsApp.
export default function PayRequest() {
  const page = usePage();
  const { Site } = useSession();
  const d = page.data;
  useTitle(d && `Payment request #${d.Booking.ID}`);
  const [canShare, setCanShare] = useState(false);
  const [copied, setCopied] = useState('');
  const file = useRef(null);
  const qrURL = d ? `${d.Booking.PayPath}/qr.png` : '';

  useEffect(() => {
    if (!d) return undefined;
    let gone = false;
    (async () => {
      try {
        const blob = await (await fetch(qrURL)).blob();
        const f = new File([blob], `pay-booking-${d.Booking.ID}.png`, { type: 'image/png' });
        file.current = f;
        if (!gone && navigator.canShare && navigator.canShare({ files: [f] })) setCanShare(true);
      } catch { /* the buttons below still work without it */ }
    })();
    return () => { gone = true; };
  }, [d, qrURL]);

  if (!d) return <PageState page={page} />;
  const k = d.Booking;

  const share = () => navigator.share({ files: [file.current], text: d.Message }).catch(() => {});
  const copy = async () => {
    try {
      const blob = file.current || (await (await fetch(qrURL)).blob());
      await navigator.clipboard.write([new ClipboardItem({ 'image/png': blob })]);
      setCopied('✓ QR copied');
    } catch {
      setCopied('Copy failed: right-click the QR → Copy image');
    }
  };

  return (
    <section className="payreq panel">
      <p className="eyebrow">Booking #{k.ID} approved</p>
      <h1>Send the payment request to {firstWord(k.UserName)}</h1>
      <p className="muted small">{k.TrekName} · {fmtDate(k.StartDate)} · {k.People} {k.People === 1 ? 'person' : 'people'} · {k.UserPhone}</p>

      <div className="payreq-body">
        <img id="qr" src={qrURL} alt={`UPI QR for ₹${inr(k.Amount)}`} width="220" height="220" />
        <div className="payreq-info">
          <strong className="payreq-amt">₹{inr(k.Amount)}</strong>
          <span className="small">UPI ID <code>{Site && Site.UPI}</code></span>
          <span className="small muted">They can scan this QR with any UPI app and pay directly.</span>
        </div>
      </div>

      <div className="payreq-actions">
        {canShare && <button type="button" className="btn" onClick={share}>Share QR + message on WhatsApp</button>}
        <button type="button" className="btn" onClick={copy}>{copied || '1. Copy QR image'}</button>
        <a className="btn" href={d.WALink} target="_blank" rel="noopener noreferrer">2. Open WhatsApp chat</a>
      </div>
      <p className="small muted">
        {canShare
          ? `On your phone: tap "Share", choose WhatsApp and pick ${k.UserName}. Or use the steps above.`
          : <>Copy the QR, open the chat (the message is already typed), then press <b>Ctrl + V</b> to paste the QR and send.</>}
      </p>
      <p className="small"><Link to="/admin/registrations?status=approved">← Back to registrations</Link></p>
    </section>
  );
}
