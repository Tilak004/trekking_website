import { Navigate, Route, Routes } from 'react-router-dom';
import Shell from './Shell.jsx';
import Login from './pages/Login.jsx';
import Dashboard from './pages/Dashboard.jsx';
import { Events, EventPage } from './pages/Events.jsx';
import Registrations from './pages/Registrations.jsx';
import Analytics from './pages/Analytics.jsx';
import Enquiries from './pages/Enquiries.jsx';
import { Treks, TrekForm, TrekDefaults } from './pages/Treks.jsx';
import { Leads, LeadForm } from './pages/Leads.jsx';
import Users from './pages/Users.jsx';
import PayRequest from './pages/PayRequest.jsx';

export default function App() {
  return (
    <Shell>
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route path="/admin" element={<Dashboard />} />
        <Route path="/admin/events" element={<Events />} />
        <Route path="/admin/events/new" element={<EventPage />} />
        <Route path="/admin/events/:id" element={<EventPage />} />
        <Route path="/admin/registrations" element={<Registrations />} />
        <Route path="/admin/analytics" element={<Analytics />} />
        <Route path="/admin/enquiries" element={<Enquiries />} />
        <Route path="/admin/treks" element={<Treks />} />
        <Route path="/admin/treks/new" element={<TrekForm />} />
        <Route path="/admin/trek-defaults" element={<TrekDefaults />} />
        <Route path="/admin/treks/:id" element={<TrekForm />} />
        <Route path="/admin/leads" element={<Leads />} />
        <Route path="/admin/leads/new" element={<LeadForm />} />
        <Route path="/admin/leads/:id" element={<LeadForm />} />
        <Route path="/admin/users" element={<Users />} />
        <Route path="/admin/bookings/:id/payreq" element={<PayRequest />} />
        <Route path="/" element={<Navigate to="/admin" replace />} />
        <Route path="*" element={<p className="muted">Page not found.</p>} />
      </Routes>
    </Shell>
  );
}
