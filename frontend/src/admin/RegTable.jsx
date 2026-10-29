import { Link } from 'react-router-dom';
import { useSubmit } from '../lib/useSubmit.js';
import { fmtDate, inr, whatsappTo } from '../lib/util.js';

// One small POST-button form. `fields` become hidden inputs; `back` returns the
// admin to the page they were on.
function Act({ path, fields, back, className, children, confirm, title }) {
  const { submit } = useSubmit();
  return (
    <form onSubmit={submit(path, confirm ? { confirm } : undefined)} className="inline">
      {Object.entries(fields || {}).map(([k, v]) => <input key={k} type="hidden" name={k} value={v} />)}
      <input type="hidden" name="back" value={back} />
      <button className={className} title={title}>{children}</button>
    </form>
  );
}

// One tap: marks the booking paid and opens WhatsApp to the trekker with the
// confirmation already written (the admin only has to press send there).
function ConfirmPay({ k, back }) {
  const { submit } = useSubmit();
  const send = submit(`/admin/bookings/${k.ID}/paid`);
  const onSubmit = (e) => {
    if (k.ConfirmWA) window.open(k.ConfirmWA, '_blank', 'noopener');
    return send(e);
  };
  return (
    <form onSubmit={onSubmit} className="inline">
      <input type="hidden" name="paid" value="1" />
      <input type="hidden" name="amount" value={k.Amount} />
      <input type="hidden" name="back" value={back} />
      <button className="btn small act-approved" title="Marks it paid and opens WhatsApp with the confirmation written">Confirm payment &amp; send WhatsApp</button>
    </form>
  );
}

function Actions({ k, back }) {
  const id = k.ID;
  if (k.Status === 'requested') {
    return (
      <>
        <Act path={`/admin/bookings/${id}/status`} fields={{ status: 'approved', send_pay: 1 }} back={back} className="btn small act-approved">Approve &amp; request payment</Act>
        <Act path={`/admin/bookings/${id}/status`} fields={{ status: 'rejected' }} back={back} className="btn small act-rejected">Reject</Act>
      </>
    );
  }
  if (k.Status === 'approved') {
    return (
      <>
        {k.Claimed && (
          <>
            <ConfirmPay k={k} back={back} />
            <Act path={`/admin/bookings/${id}/claim-reject`} back={back} className="linklike small danger"
              confirm={`Payment for #${id} not found? The claim is cleared and ${k.UserName} gets another 10 minutes to pay and send the right reference.`}>Not received</Act>
          </>
        )}
        {!k.PaidAt && !k.Claimed && (
          <>
            {!k.Held && <Link className="btn small ghost" to={`/admin/bookings/${id}/payreq`}>Send payment request</Link>}
            <Act path={`/admin/bookings/${id}/paid`} fields={{ paid: 1, amount: k.Amount }} back={back} className="btn small">Mark paid</Act>
          </>
        )}
        {k.TrekPassed ? (
          <Act path={`/admin/bookings/${id}/status`} fields={{ status: 'completed' }} back={back} className="btn small act-approved" title="The trek has happened: unlocks the trekker's review">Mark trek done</Act>
        ) : k.PaidAt ? (
          <details className="cancel-menu">
            <summary>Cancel booking</summary>
            <div className="cancel-opts">
              <CancelOption k={k} back={back} reason="customer"
                confirm={`Cancel booking #${id}? Refund ₹${inr(k.CustomerRefund)} (policy: ${k.RefundRule}).`}
                label={`Trekker cancelled · refund ₹${inr(k.CustomerRefund)}`} note={`${k.DaysToTrek} days to go · ${k.RefundRule}`} />
              <CancelOption k={k} back={back} reason="us"
                confirm={`Cancel booking #${id} from our side? Full refund ₹${inr(k.PaidAmount)}.`}
                label={`We cancelled · full refund ₹${inr(k.PaidAmount)}`} note="Weather, closures, too few trekkers" />
            </div>
          </details>
        ) : (
          <Act path={`/admin/bookings/${id}/cancel`} back={back} className="btn small act-cancelled" confirm={`Cancel booking #${id}? Nothing was paid, so there is no refund.`}>Cancel</Act>
        )}
        {k.PaidAt && k.ConfirmWA && (
          <a className="btn small act-approved" href={k.ConfirmWA} target="_blank" rel="noopener noreferrer" title="Opens WhatsApp to the trekker with the confirmation (and the trek group invite) written">Send confirmation</a>
        )}
        {k.PaidAt && !k.GroupInviteWA && <span className="muted small">Add the trek's WhatsApp group link to send an invite.</span>}
        {k.PaidAt && (
          <Act path={`/admin/bookings/${id}/paid`} fields={{ paid: 0 }} back={back} className="linklike small" confirm={`Mark booking #${id} as NOT paid?`}>Undo paid</Act>
        )}
      </>
    );
  }
  if (k.Status === 'completed' && k.ThankYouWA) {
    return <a className="btn small" href={k.ThankYouWA} target="_blank" rel="noopener noreferrer" title="Thanks, review link and the channel link">Send thank-you</a>;
  }
  if (k.Status === 'cancelled' && k.CancelNoticeWA) {
    return (
      <>
        <a className="btn small" href={k.CancelNoticeWA} target="_blank" rel="noopener noreferrer" title="Opens WhatsApp to the trekker with the cancellation and refund written">Send cancellation message</a>
        {k.RefundAmount > 0 && !k.RefundedAt && (
          <Act path={`/admin/bookings/${k.ID}/refunded`} back={back} className="btn small act-approved" confirm={`Have you sent the ₹${inr(k.RefundAmount)} refund to ${k.UserName}?`}>Mark refunded</Act>
        )}
      </>
    );
  }
  if (k.Status === 'cancelled') {
    if (k.RefundAmount && !k.RefundedAt) {
      return <Act path={`/admin/bookings/${id}/refunded`} back={back} className="btn small act-approved" confirm={`Have you sent the ₹${inr(k.RefundAmount)} refund to ${k.UserName}?`}>Mark refunded</Act>;
    }
    if (!k.PaidAt) return <Act path={`/admin/bookings/${id}/status`} fields={{ status: 'requested' }} back={back} className="btn small">Reopen</Act>;
  }
  if (k.Status === 'rejected') {
    return <Act path={`/admin/bookings/${id}/status`} fields={{ status: 'requested' }} back={back} className="btn small">Reopen</Act>;
  }
  return null;
}

function CancelOption({ k, back, reason, confirm, label, note }) {
  const { submit } = useSubmit();
  return (
    <form onSubmit={submit(`/admin/bookings/${k.ID}/cancel`, { confirm })}>
      <input type="hidden" name="reason" value={reason} />
      <input type="hidden" name="back" value={back} />
      <button className="btn small">{label}</button>
      <span className="muted small">{note}</span>
    </form>
  );
}

// The registrations table shared by the registrations page, the dashboard and
// each event page. Set showEvent to add the event column.
export default function RegTable({ bookings, back, showEvent }) {
  return (
    <div className="table-wrap">
      <table className="table reg-table">
        <tbody>
          <tr>
            <th>Trekker</th>{showEvent && <th>Event</th>}<th>Group</th><th>Pickup &amp; emergency</th><th>Health &amp; notes</th><th>Status</th><th>Actions</th>
          </tr>
          {bookings && bookings.length ? bookings.map((k) => (
            <tr key={k.ID}>
              <td>
                <strong>{k.UserName}</strong><br />
                <span className="small">
                  <a href={`tel:${k.UserPhone}`}>{k.UserPhone}</a> · <a href={whatsappTo(k.UserPhone, '')} target="_blank" rel="noopener noreferrer">WhatsApp</a>
                  {!k.UserEmail.startsWith('guest-')
                    ? <><br /><a href={`mailto:${k.UserEmail}`}>{k.UserEmail}</a></>
                    : <><br /><span className="muted">Guest booking</span></>}
                </span>
                <br /><span className="muted small">#{k.ID} · booked {k.CreatedAt}</span>
              </td>
              {showEvent && <td><Link to={`/admin/events/${k.BatchID}`}>{k.TrekName}</Link><br /><span className="small">{fmtDate(k.StartDate)}</span></td>}
              <td className="small"><strong>{k.People}</strong> {k.People === 1 ? 'person' : 'people'}{k.Age ? <><br />Age {k.Age}</> : null}</td>
              <td className="small">
                {k.Pickup || '—'}<br />
                <span className="muted">Emergency:</span> {k.EmergencyName || '—'}
                {k.EmergencyPhone ? <> · <a href={`tel:${k.EmergencyPhone}`}>{k.EmergencyPhone}</a></> : null}
              </td>
              <td className="small">
                {k.Medical ? <><span className="warn">Medical:</span> {k.Medical}</> : null}
                {k.Note ? <><br />Note: {k.Note}</> : null}
                {!k.Medical && !k.Note && <span className="muted">—</span>}
              </td>
              <td>
                {k.Claimed ? <span className="status pending">Payment claimed</span>
                  : k.Held ? <span className="status pending">Seat held</span>
                  : <span className={`status ${k.Status}`}>{k.Status}</span>}
                {k.Claimed && <><br /><span className="small">UTR <code>{k.Utr}</code></span></>}
                {k.Held && <><br /><span className="muted small">releases in {k.HoldSecondsLeft >= 3600 ? `${Math.floor(k.HoldSecondsLeft / 3600)}h ${Math.floor((k.HoldSecondsLeft % 3600) / 60)}m` : `${Math.ceil(k.HoldSecondsLeft / 60)}m`}</span></>}
                {k.Status === 'cancelled' && k.RefundReason === 'expired' && <><br /><span className="muted small">hold expired, never paid</span></>}
                {k.PaidAt ? <><br /><span className="status paid">Paid ₹{inr(k.PaidAmount)}</span></>
                  : (k.Status === 'approved' || k.Status === 'requested') ? <><br /><span className="muted small">Due ₹{inr(k.Amount)}</span></> : null}
                {k.Status === 'cancelled' && (
                  <>
                    {k.RefundedAt ? <><br /><span className="status completed">Refunded ₹{inr(k.RefundAmount)}</span></>
                      : k.RefundAmount ? <><br /><span className="status requested">Refund ₹{inr(k.RefundAmount)} pending</span></> : null}
                    {k.RefundReason && k.RefundReason !== 'expired' ? <><br /><span className="muted small">{k.RefundReason === 'us' ? 'Cancelled by us' : 'Cancelled by trekker'}</span></> : null}
                  </>
                )}
              </td>
              <td className="actions"><Actions k={k} back={back} /></td>
            </tr>
          )) : (
            <tr><td colSpan={showEvent ? 7 : 6} className="muted">No registrations here yet.</td></tr>
          )}
        </tbody>
      </table>
    </div>
  );
}
