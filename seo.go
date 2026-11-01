package main

import (
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"regexp"
	"strings"
)

// Search engines and AI answer engines (Google, ChatGPT, Perplexity) read three
// things from us: robots.txt (what they may crawl), sitemap.xml (every public
// page) and the structured data (schema.org JSON-LD) in each page's <head>,
// which tells them who we are, what each trek is and the answers to its FAQs.

// noindexPaths are private pages: account, booking and payment screens.
var noindexPaths = []string{"/my", "/login", "/signup", "/pay/", "/lead"}

func noindex(p string) bool {
	for _, n := range noindexPaths {
		if p == n || strings.HasPrefix(p, strings.TrimSuffix(n, "/")+"/") {
			return true
		}
	}
	return false
}

// staticPages are the public pages that aren't trek or lead profiles, with the
// title and description they show in search results.
var staticPages = []struct{ Path, Title, Desc string }{
	{"/leads", "Trek leads", "Meet the trek leads who guide our weekend treks from Bangalore: experienced, verified and trail tested."},
	{"/gallery", "Gallery", "Photos from our weekend treks around Bangalore and Karnataka."},
	{"/about", "About us", "Beyond Expeditions runs small-group weekend treks from Bangalore, with pickups in the city and leads who know every turn of the trail."},
	{"/contact", "Contact", "Questions about a trek? Call, WhatsApp or write to Beyond Expeditions."},
	{"/lead-with-us", "Lead with us", "Lead treks with Beyond Expeditions. Apply to guide small groups on weekend treks from Bangalore."},
	{"/terms", "Terms", "Terms of booking a trek with Beyond Expeditions."},
	{"/privacy", "Privacy", "How Beyond Expeditions handles your personal data."},
	{"/cancellation-policy", "Cancellation policy", "Cancellation and refund rules for treks booked with Beyond Expeditions."},
}

// knownPage reports whether a path is a page the React app has, so anything
// else is answered with a real 404 instead of an empty 200.
var knownPage = regexp.MustCompile(`^/($|treks/[^/]+$|leads(/[^/]+)?$|lead-with-us$|lead$|gallery$|about$|contact$|terms$|privacy$|cancellation-policy$|login$|signup$|my$|my/bookings/\d+/sent$|pay/\d+/[^/]+$)`)

// googleVerification is the Search Console ownership file: Google fetches
// /<name> and expects exactly "google-site-verification: <name>".
const googleVerification = "google1c53086d0db301f5.html"

func (a *App) googleVerify(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, "google-site-verification: "+googleVerification)
}

func (a *App) robotsTxt(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if a.adminMode {
		fmt.Fprint(w, "User-agent: *\nDisallow: /\n")
		return
	}
	fmt.Fprint(w, "User-agent: *\nAllow: /\nDisallow: /api/\nDisallow: /my\nDisallow: /pay/\nDisallow: /login\nDisallow: /signup\n\n")
	fmt.Fprintf(w, "Sitemap: %s/sitemap.xml\n", baseURL(r))
}

func (a *App) sitemapXML(w http.ResponseWriter, r *http.Request) {
	if a.adminMode {
		http.NotFound(w, r)
		return
	}
	base := baseURL(r)
	paths := []string{"/"}
	if treks, err := a.listTreks(true); err == nil {
		for _, t := range treks {
			paths = append(paths, "/treks/"+t.Slug)
		}
	}
	for _, p := range staticPages {
		paths = append(paths, p.Path)
	}
	if leads, err := a.listLeads(true); err == nil {
		for _, l := range leads {
			paths = append(paths, "/leads/"+l.Slug)
		}
	}
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	fmt.Fprint(w, `<?xml version="1.0" encoding="UTF-8"?>`+"\n"+`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">`+"\n")
	for _, p := range paths {
		fmt.Fprintf(w, "  <url><loc>%s</loc></url>\n", html.EscapeString(base+p))
	}
	fmt.Fprint(w, "</urlset>\n")
}

// organization describes the business; every page carries it.
func organization(base string) map[string]any {
	return map[string]any{
		"@type":       "TravelAgency",
		"@id":         base + "/#org",
		"name":        site.Name,
		"slogan":      site.Tagline,
		"description": site.Description,
		"url":         base + "/",
		"logo":        base + "/static/logo/mark.png",
		"telephone":   site.Phone,
		"sameAs":      []string{"https://www.instagram.com/" + site.Instagram + "/"},
		"address":     map[string]any{"@type": "PostalAddress", "addressLocality": "Bangalore", "addressRegion": "Karnataka", "addressCountry": "IN"},
		"areaServed":  "Karnataka, India",
	}
}

// faqBlocks splits a trek's FAQs (question on the first line, answer below,
// blank line between them) the same way the trek page does.
func faqBlocks(s string) [][2]string {
	var out [][2]string
	for _, b := range strings.Split(strings.ReplaceAll(s, "\r\n", "\n"), "\n\n") {
		q, ans, _ := strings.Cut(strings.TrimSpace(b), "\n")
		if q = strings.TrimSpace(q); q != "" && strings.TrimSpace(ans) != "" {
			out = append(out, [2]string{q, strings.TrimSpace(ans)})
		}
	}
	return out
}

func faqPage(faqs [][2]string) map[string]any {
	var qs []any
	for _, f := range faqs {
		qs = append(qs, map[string]any{"@type": "Question", "name": f[0], "acceptedAnswer": map[string]any{"@type": "Answer", "text": f[1]}})
	}
	return map[string]any{"@type": "FAQPage", "mainEntity": qs}
}

// homeFAQs are the questions people search for before booking a trek. They are
// shown on the home page and published as FAQPage data, which is what Google and
// AI answer engines quote. Answers only repeat what the site's policies say.
func (a *App) homeFAQs() [][2]string {
	answer := func(faq string) string { _, ans, _ := strings.Cut(faq, "\n"); return ans }
	var names []string
	if treks, err := a.listTreks(true); err == nil {
		for _, t := range treks {
			names = append(names, t.Name)
		}
	}
	faqs := [][2]string{}
	if len(names) > 0 {
		faqs = append(faqs, [2]string{"Which weekend treks do you run from Bangalore?",
			"Treks in the Western Ghats and the granite hills around Bangalore. Right now: " + strings.Join(names, ", ") + ". Each trek page has the route, difficulty, itinerary and next dates."})
	}
	return append(faqs,
		[2]string{"How do I book a trek and pay?", answer(bookingFAQ)},
		[2]string{"Where do you pick up from in Bangalore?", answer(newBoardFAQ)},
		[2]string{"I've never trekked before. Can I join?", "Yes. Every trek page shows its difficulty. Easy and easy–moderate treks suit beginners who can walk for a few hours with short breaks. For moderate and difficult treks, regular walks or jogging for a few weeks beforehand help a lot."},
		[2]string{"How big are the groups?", "Small. We cap every batch, so it feels like a trip with friends, not a crowd on a hill."},
		[2]string{"Who leads the treks?", "Trek leads who have walked these hills many times over. They set a pace everyone can enjoy, make sure nobody gets left behind, and teach the basics of map reading, navigation and using a compass along the way. Every lead has a profile on our Leads page."},
		[2]string{"What should I carry?", "Trekking shoes with good grip, 2 litres of water, a rain jacket in the monsoon (June to November), light snacks, a cap and sunscreen, a small backpack, a government photo ID and any personal medicines. Each trek page has its full list."},
		[2]string{"What if I need to cancel?", "7 or more days before the trek: full refund minus a ₹200 processing fee. 3 to 6 days before: 50% refund, or move to another date for free. Less than 3 days before: no refund, but you can move to another date once. Cancelling before you've paid is always free."},
		[2]string{"What happens if it rains heavily or the trail is closed?", "Your safety comes first. If we cancel or reschedule because of heavy rain, a forest department closure or unsafe conditions, you choose a full refund or a free move to another date or trek."},
		[2]string{"Is insurance included?", "No. Our price does not include travel, accident or medical insurance. If you want cover, you can buy personal accident insurance on your own before the trek."},
	)
}

// structuredData builds the JSON-LD for a page.
func (a *App) structuredData(r *http.Request, base string) []byte {
	graph := []any{organization(base)}
	p := r.URL.Path
	if p == "/" {
		graph = append(graph, map[string]any{"@type": "WebSite", "@id": base + "/#site", "name": site.Name, "url": base + "/", "publisher": map[string]any{"@id": base + "/#org"}},
			faqPage(a.homeFAQs()))
	}
	if slug, ok := strings.CutPrefix(p, "/treks/"); ok {
		if t, err := a.trekWhere(`t.slug = ? AND t.active = 1`, slug); err == nil {
			trip := map[string]any{
				"@type":       "TouristTrip",
				"name":        t.Name,
				"description": t.Summary,
				"url":         base + p,
				"provider":    map[string]any{"@id": base + "/#org"},
				"touristType": "Trekkers",
			}
			if t.Photo != "" {
				trip["image"] = base + "/photos/" + t.Photo
			}
			if t.Region != "" {
				trip["itinerary"] = map[string]any{"@type": "Place", "name": t.Name, "address": t.Region}
			}
			if t.Price > 0 {
				trip["offers"] = map[string]any{"@type": "Offer", "price": t.Price, "priceCurrency": "INR", "url": base + p}
			}
			graph = append(graph, trip, map[string]any{
				"@type": "BreadcrumbList",
				"itemListElement": []any{
					map[string]any{"@type": "ListItem", "position": 1, "name": "Treks", "item": base + "/"},
					map[string]any{"@type": "ListItem", "position": 2, "name": t.Name, "item": base + p},
				},
			})
			if faqs := faqBlocks(t.FAQs); len(faqs) > 0 {
				graph = append(graph, faqPage(faqs))
			}
		}
	}
	b, _ := json.Marshal(map[string]any{"@context": "https://schema.org", "@graph": graph})
	// keep "</script>" in any text from closing the tag early
	return []byte(strings.ReplaceAll(string(b), "</", `<\/`))
}
