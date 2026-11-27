import { useEffect } from 'react';
import { Route, Routes, useLocation } from 'react-router-dom';
import { trackPageView } from '../lib/analytics.js';
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

// Reports each page the visitor opens to Google Analytics, a moment after it loads so
// the page title is the new page's.
function PageViews() {
  const { pathname } = useLocation();
  useEffect(() => {
    const t = setTimeout(() => trackPageView(pathname), 400);
    return () => clearTimeout(t);
  }, [pathname]);
  return null;
}

export default function App() {
  return (
    <Layout>
      <PageViews />
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
