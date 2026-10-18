import { PageState } from '../../components/common.jsx';
import { useSession } from '../../lib/session.jsx';
import { usePage, useTitle } from '../../lib/usePage.js';
import { useSubmit } from '../../lib/useSubmit.js';

function Replied({ q }) {
  const { submit } = useSubmit();
  return (
    <form onSubmit={submit(`/admin/enquiries/${q.ID}/replied`)} className="inline">
      {q.RepliedAt ? (
        <>
          <input type="hidden" name="undo" value="1" />
          <span className="status completed">Replied</span> <button className="linklike small">Undo</button>
        </>
      ) : <button className="btn small act-approved">Mark replied</button>}
    </form>
  );
}

export default function Enquiries() {
  const page = usePage();
  const { Brand, PublicURL } = useSession();
  useTitle('Enquiries');
  const d = page.data;
  if (!d) return <PageState page={page} />;
  return (
    <>
      <div className="c-head"><h1>Enquiries</h1><span className="muted">{d.Open} waiting for a reply</span></div>
      <p className="muted small">People who asked about a trek from its page. Call or WhatsApp them, then mark it replied.</p>
      <div className="table-wrap">
        <table className="table">
          <tbody>
            <tr><th>Person</th><th>Trek</th><th>Question</th><th>Received</th><th /></tr>
            {d.Enquiries && d.Enquiries.length ? d.Enquiries.map((q) => {
              const text = `Hi ${q.Name}, this is ${Brand} about your enquiry on the ${q.TrekName} trek.`;
              return (
                <tr key={q.ID} className={q.RepliedAt ? 'done' : undefined}>
                  <td>
                    <strong>{q.Name}</strong><br />
                    <span className="small">
                      <a href={`tel:+91${q.Phone}`}>+91 {q.Phone}</a> · <a href={`https://wa.me/91${q.Phone}?text=${encodeURIComponent(text)}`} target="_blank" rel="noopener noreferrer">WhatsApp</a>
                      {q.Email ? <><br /><a href={`mailto:${q.Email}`}>{q.Email}</a></> : null}
                    </span>
                  </td>
                  <td>{q.TrekSlug ? <a href={`${PublicURL}/treks/${q.TrekSlug}`} target="_blank" rel="noopener noreferrer">{q.TrekName}</a> : <span className="muted">–</span>}</td>
                  <td className="small enq-msg">{q.Message || '—'}</td>
                  <td className="small">{q.CreatedAt}</td>
                  <td><Replied q={q} /></td>
                </tr>
              );
            }) : <tr><td colSpan={5} className="muted">No enquiries yet.</td></tr>}
          </tbody>
        </table>
      </div>
    </>
  );
}
