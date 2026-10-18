import { Route, Routes } from 'react-router-dom';
import Layout from './Layout.jsx';
import Home from './pages/Home.jsx';
import Trek from './pages/Trek.jsx';
import { Leads, LeadProfile } from './pages/Leads.jsx';
import LeadApply from './pages/LeadApply.jsx';
import LeadDashboard from './pages/LeadDashboard.jsx';
import Gallery from './pages/Gallery.jsx';
import { About, Contact, Terms, Privacy, Cancellation } from './pages/Static.jsx';
import { Login, Signup } from './pages/Auth.jsx';
import { My, BookingSent, Pay } from './pages/Bookings.jsx';
import NotFound from './pages/NotFound.jsx';

export default function App() {
  return (
    <Layout>
      <Routes>
        <Route path="/" element={<Home />} />
        <Route path="/treks/:slug" element={<Trek />} />
        <Route path="/leads" element={<Leads />} />
        <Route path="/leads/:slug" element={<LeadProfile />} />
        <Route path="/lead-with-us" element={<LeadApply />} />
        <Route path="/lead" element={<LeadDashboard />} />
        <Route path="/gallery" element={<Gallery />} />
        <Route path="/about" element={<About />} />
        <Route path="/contact" element={<Contact />} />
        <Route path="/terms" element={<Terms />} />
        <Route path="/privacy" element={<Privacy />} />
        <Route path="/cancellation-policy" element={<Cancellation />} />
        <Route path="/login" element={<Login />} />
        <Route path="/signup" element={<Signup />} />
        <Route path="/my" element={<My />} />
        <Route path="/my/bookings/:id/sent" element={<BookingSent />} />
        <Route path="/pay/:id/:token" element={<Pay />} />
        <Route path="*" element={<NotFound />} />
      </Routes>
    </Layout>
  );
}
