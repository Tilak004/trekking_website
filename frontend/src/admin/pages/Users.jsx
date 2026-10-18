import { PageState } from '../../components/common.jsx';
import { usePage, useTitle } from '../../lib/usePage.js';

export default function Users() {
  const page = usePage();
  useTitle('Users');
  const d = page.data;
  if (!d) return <PageState page={page} />;
  const isGuest = (u) => u.Email.endsWith('@guest.beyondexpeditions.local');
  return (
    <>
      <h1>Users</h1>
      <div className="table-wrap">
        <table className="table">
          <tbody>
            <tr><th>Name</th><th>Email</th><th>Phone</th><th>Bookings</th><th>Joined</th><th>Account</th></tr>
            {(d.Users || []).map((u) => (
              <tr key={u.ID}>
                <td>{u.Name}</td>
                <td>{isGuest(u) ? <span className="muted">–</span> : u.Email}</td>
                <td>{u.Phone}</td>
                <td>{u.Bookings}</td>
                <td className="small">{u.CreatedAt}</td>
                <td>{isGuest(u) ? <span className="status">Guest</span> : <span className="status approved">Registered</span>}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </>
  );
}
