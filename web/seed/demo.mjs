// Builds a DEMO database full of sample trek leads, trekkers, trek dates,
// bookings and reviews, for testing and showing the site.
//
//   node seed/demo.mjs [source.db] [demo.db]     (defaults: ../beyond.db → ../beyond-demo.db)
//
// It copies the real database (so treks, photos and the admin login carry over)
// and adds sample people on top. Never point it at the live database: the
// sample leads and reviews are made up.
//
// All demo accounts use the password "demo1234" and @example.com emails.
// Phone numbers start with 00000 so they can't belong to real people.

import { DatabaseSync } from 'node:sqlite';
import { copyFileSync, existsSync, readdirSync, rmSync } from 'node:fs';
import { faker } from '@faker-js/faker/locale/en_IN';
import bcrypt from 'bcryptjs';

const SRC = process.argv[2] || '../beyond.db';
const OUT = process.argv[3] || '../beyond-demo.db';
if (/beyond\.db$/.test(OUT) && !/demo/.test(OUT)) throw new Error('refusing to write demo data into the live database');

for (const f of [OUT, OUT + '-wal', OUT + '-shm']) if (existsSync(f)) rmSync(f);
copyFileSync(SRC, OUT);
const db = new DatabaseSync(OUT);
db.exec('PRAGMA foreign_keys = ON');

faker.seed(2026); // same demo data every run
const pick = (a) => a[faker.number.int({ min: 0, max: a.length - 1 })];
const chance = (p) => faker.number.float() < p;
const hash = bcrypt.hashSync('demo1234', 10);
let phoneSeq = 0;
const fakePhone = () => `+91 00000 ${String(10000 + ++phoneSeq).slice(-5)}`;
const usedEmails = new Set(db.prepare('SELECT email FROM users').all().map((r) => r.email));

function person(sex) {
  const first = faker.person.firstName(sex);
  const last = faker.person.lastName();
  let base = `${first}.${last}`.toLowerCase().replace(/[^a-z.]/g, '');
  let email = `${base}@example.com`;
  for (let n = 2; usedEmails.has(email); n++) email = `${base}${n}@example.com`;
  usedEmails.add(email);
  return { name: `${first} ${last}`, email, phone: fakePhone() };
}

const insertUser = db.prepare(`INSERT INTO users (name, email, phone, password_hash, created_at) VALUES (?, ?, ?, ?, ?)`);
function addUser(p, joined) {
  return Number(insertUser.run(p.name, p.email, p.phone, hash, joined).lastInsertRowid);
}
const daysAgo = (n) => {
  const d = new Date();
  d.setDate(d.getDate() - n);
  return d.toISOString().slice(0, 10);
};
const daysAhead = (n) => daysAgo(-n);
const nextSat = (from) => {
  const d = new Date(from);
  while (d.getDay() !== 6) d.setDate(d.getDate() + 1);
  return d.toISOString().slice(0, 10);
};

// ---------------------------------------------------------------- trek leads
// Hand-written profiles for variety; faker supplies the names.
const LEADS = [
  { sex: 'male', city: 'Bangalore', status: 'approved', verified: true, headline: 'Ex-Army officer · 12 years in the mountains',
    bio: 'Served with the Madras Engineer Group and led winter patrols in Ladakh before settling in Bangalore. On weekends I take groups up the Ghats with the same rule I had in the army: nobody gets left behind.',
    achievements: 'Basic & Advanced Mountaineering Course (HMI Darjeeling)\nStok Kangri summit\n40+ Western Ghats treks led', certifications: 'Wilderness First Responder (WFR)\nHMI Advanced Mountaineering' },
  { sex: 'female', city: 'Bangalore', status: 'approved', verified: true, headline: 'Ultra-runner · Women-only weekend groups',
    bio: 'I run ultra-marathons and lead women-only batches on Kumara Parvatha and Tadiandamol. Most of my trekkers are first-timers, and my favourite moment is watching them reach a summit they thought was impossible.',
    achievements: 'Malnad Ultra 80K finisher\nKumara Parvatha 15+ times\nRuns a 200-member women trekking community', certifications: 'Wilderness first aid (St John Ambulance)' },
  { sex: 'male', city: 'Mysore', status: 'approved', verified: true, headline: 'Wildlife photographer · Shola forest specialist',
    bio: 'Ten years photographing the Western Ghats taught me where the Malabar trogons hide and when the Neelakurinji blooms. My treks move slowly so everyone gets the shot.',
    achievements: 'Published in Sanctuary Asia\nKudremukh & Netravati 25+ times\nBird-call identification walks', certifications: 'Basic Mountaineering Course (NIM Uttarkashi)\nFirst aid (Red Cross)' },
  { sex: 'female', city: 'Bangalore', status: 'approved', verified: true, headline: 'Yoga teacher · Sunrise & mindful treks',
    bio: 'I combine short sunrise climbs with a breathing and stretching session at the top. Perfect if you want a calm, unhurried morning away from the city.',
    achievements: '500-hour certified yoga teacher\nSkandagiri & Bananthimari sunrise treks every month', certifications: 'Wilderness first aid' },
  { sex: 'male', city: 'Bangalore', status: 'approved', verified: false, headline: 'Software engineer by week, trek lead by weekend',
    bio: 'I build payment systems at a fintech during the week and swap my laptop for trekking poles every Saturday. I love introducing colleagues and first-timers to the Ghats.',
    achievements: '30+ Karnataka treks\nKedarkantha & Brahmatal in the Himalayas', certifications: 'CPR & first aid (online, 2025)' },
  { sex: 'male', city: 'Mangalore', status: 'approved', verified: false, headline: 'Grew up below Kudremukh · local trail expert',
    bio: 'My village is at the foot of Kudremukh. I know every stream crossing, every leech patch and the best homestay meals in Kalasa.',
    achievements: 'Grew up trekking the Kudremukh range\nFluent in Kannada, Tulu and English', certifications: '' },
  { sex: 'female', city: 'Bangalore', status: 'approved', verified: false, headline: 'College trekking club captain · beginner-friendly',
    bio: 'I captain my college trekking club and organise easy treks for students. Expect music, games and a lot of laughter on the way up.',
    achievements: '20+ treks with the college club\nOrganised 6 group treks of 30 students', certifications: 'Basic first aid' },
  // waiting for approval
  { sex: 'male', city: 'Hubli', status: 'pending', verified: false, headline: 'Cyclist & trekker · North Karnataka trails',
    bio: 'I mostly cycle, but I have done most of the Ghats treks and want to lead groups on the quieter North Karnataka trails.',
    achievements: 'Bangalore–Goa cycle ride\n12 Western Ghats treks', certifications: '' },
  { sex: 'female', city: 'Bangalore', status: 'pending', verified: false, headline: 'Marathoner · Weekend sunrise treks',
    bio: 'Marathon runner for 6 years. I would love to lead early-morning treks for people who want to get fit outdoors.',
    achievements: 'TCS World 10K & Bengaluru Marathon finisher\nSkandagiri 10+ times', certifications: 'CPR certified' },
  { sex: 'male', city: 'Chikmagalur', status: 'pending', verified: false, headline: 'Coffee planter · Baba Budangiri regular',
    bio: 'I run a coffee estate in Chikmagalur and walk the Mullayanagiri ranges every week. Happy to host groups and lead estate-to-peak treks.',
    achievements: 'Mullayanagiri 50+ times\nHosts homestay groups', certifications: '' },
  // not approved
  { sex: 'male', city: 'Bangalore', status: 'rejected', verified: false, headline: 'Weekend trekker',
    bio: 'I like trekking and want to make some extra money on weekends.', achievements: 'Nandi Hills', certifications: '' },
  { sex: 'female', city: 'Pune', status: 'rejected', verified: false, headline: 'Travel blogger',
    bio: 'I write about travel and would like to lead treks for my followers.', achievements: 'Visited 15 states', certifications: '' },
];

const insertLead = db.prepare(`INSERT INTO leads (slug, name, headline, bio, achievements, phone, photo, active, user_id, status,
  certifications, city, verified, applied_at) VALUES (?, ?, ?, ?, ?, ?, '', ?, ?, ?, ?, ?, ?, ?)`);
const slugs = new Set(db.prepare('SELECT slug FROM leads').all().map((r) => r.slug));
const approvedLeads = [];
for (const L of LEADS) {
  const p = person(L.sex);
  const uid = addUser(p, daysAgo(faker.number.int({ min: 120, max: 600 })));
  let slug = p.name.toLowerCase().replace(/[^a-z]+/g, '-').replace(/^-|-$/g, '');
  for (let n = 2; slugs.has(slug); n++) slug = `${slug}-${n}`;
  slugs.add(slug);
  const id = Number(insertLead.run(slug, p.name, L.headline, L.bio, L.achievements, p.phone, L.status === 'approved' ? 1 : 0, uid,
    L.status, L.certifications, L.city, L.verified ? 1 : 0, daysAgo(faker.number.int({ min: 5, max: 300 })) + ' 10:00:00').lastInsertRowid);
  if (L.status === 'approved') approvedLeads.push({ id, name: p.name, verified: L.verified });
}

// ---------------------------------------------------------------- trekkers
const PROFILES = ['student', 'it', 'it', 'it', 'family', 'fitness', 'senior', 'photographer'];
const trekkers = [];
for (let i = 0; i < 70; i++) {
  const p = person(chance(0.45) ? 'female' : 'male');
  const kind = pick(PROFILES);
  const age = { student: [18, 23], it: [24, 38], family: [32, 46], fitness: [22, 40], senior: [50, 64], photographer: [25, 45] }[kind];
  trekkers.push({ id: addUser(p, daysAgo(faker.number.int({ min: 3, max: 700 }))), name: p.name, kind,
    age: faker.number.int({ min: age[0], max: age[1] }) });
}

// ---------------------------------------------------------------- dates & bookings
const treks = db.prepare(`SELECT id, slug, name, pickup_points, difficulty FROM treks WHERE active = 1`).all();
const insertBatch = db.prepare(`INSERT INTO batches (trek_id, lead_id, start_date, seats, status, proposed_at) VALUES (?, ?, ?, ?, ?, ?)`);
const insertBooking = db.prepare(`INSERT INTO bookings (user_id, batch_id, status, people, pickup, emergency_name, emergency_phone,
  age, medical, note, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`);
const insertReview = db.prepare(`INSERT INTO reviews (booking_id, trek_rating, lead_rating, comment, photo, created_at) VALUES (?, ?, ?, ?, ?, ?)`);

const MEDICAL = ['Mild asthma, carry an inhaler', 'Old knee injury, using trekking poles', 'Allergic to peanuts', 'Low blood pressure in the mornings', 'Type 1 diabetic, carry glucose'];
const NOTES = ['Two friends joining, names on WhatsApp', 'First trek ever, a bit nervous!', 'Can we get vegetarian breakfast?', 'Coming straight from the airport Friday night', 'Celebrating a birthday on this trek 🎉'];
const PRAISE = [
  'Perfectly organised from pickup to drop. {lead} kept the pace comfortable for everyone and the summit view was unreal.',
  'My first trek and I loved it. {lead} was patient, checked on everyone and even shared snacks at the top!',
  'The grasslands looked straight out of a postcard. Small group, no rush, great breakfast at the base.',
  'Safety briefing was thorough and {lead} knew the trail by heart. Will definitely book again.',
  'Woke up at 3 AM and it was worth every minute. Sea of clouds at the top!',
  'Loved that the group was small. Felt like trekking with friends rather than a big tour.',
  'Leeches everywhere in the monsoon but {lead} had salt and socks ready. Great experience.',
  'Pricing was honest, everything listed was included. No surprises at the base.',
  'Tough climb but super rewarding. {lead} motivated the whole group on the last stretch.',
  'Beautiful trail, well-timed breaks and amazing photos. Already planning the next one.',
];
const MEH = ['Trail was great but the bus left 40 minutes late.', 'Good trek, though breakfast could have been better.', 'Nice lead, but the group was a little too big for my liking.'];
const galleryPhotos = existsSync('../photos/gallery') ? readdirSync('../photos/gallery').filter((f) => /\.(jpe?g|png|webp)$/i.test(f)).map((f) => 'gallery/' + f) : [];

let counts = { batches: 0, bookings: 0, reviews: 0 };
for (const t of treks) {
  const pickups = t.pickup_points.split('\n').map((s) => s.trim()).filter(Boolean);
  const plan = [
    ...Array.from({ length: faker.number.int({ min: 2, max: 4 }) }, () => ({ date: nextSat(daysAgo(faker.number.int({ min: 20, max: 420 }))), status: 'live', past: true })),
    ...Array.from({ length: faker.number.int({ min: 1, max: 2 }) }, () => ({ date: nextSat(daysAhead(faker.number.int({ min: 10, max: 80 }))), status: 'live', past: false })),
  ];
  if (chance(0.4)) plan.push({ date: nextSat(daysAhead(faker.number.int({ min: 30, max: 90 }))), status: 'pending', past: false });

  for (const b of plan) {
    const lead = pick(approvedLeads);
    const seats = pick([12, 15, 15, 20]);
    const batchID = Number(insertBatch.run(t.id, lead.id, b.date, seats, b.status, b.status === 'pending' ? daysAgo(2) + ' 18:30:00' : '').lastInsertRowid);
    counts.batches++;
    if (b.status !== 'live') continue;

    let filled = 0;
    const want = b.past ? faker.number.int({ min: Math.floor(seats * 0.5), max: seats }) : faker.number.int({ min: 2, max: Math.floor(seats * 0.7) });
    const group = faker.helpers.shuffle([...trekkers]);
    for (const u of group) {
      if (filled >= want) break;
      const people = Math.min(pick([1, 1, 1, 2, 2, 3]), want - filled);
      const status = b.past ? (chance(0.92) ? 'completed' : 'cancelled') : (chance(0.55) ? 'approved' : 'requested');
      const kid = Number(insertBooking.run(u.id, batchID, status, people, pick(pickups) || '', faker.person.firstName(), fakePhone(),
        u.age, chance(0.15) ? pick(MEDICAL) : '', chance(0.2) ? pick(NOTES) : '', daysAgo(faker.number.int({ min: 1, max: 40 })) + ' 09:00:00').lastInsertRowid);
      counts.bookings++;
      if (status !== 'cancelled') filled += people;
      if (status === 'completed' && chance(0.7)) {
        const happy = chance(0.85);
        const trekRating = happy ? pick([4, 5, 5, 5]) : pick([3, 4]);
        const leadRating = happy ? pick([4, 5, 5]) : pick([3, 4]);
        const comment = (happy ? pick(PRAISE) : pick(MEH)).replaceAll('{lead}', lead.name.split(' ')[0]);
        const photo = galleryPhotos.length && chance(0.2) ? pick(galleryPhotos) : '';
        insertReview.run(kid, trekRating, leadRating, comment, photo, daysAgo(faker.number.int({ min: 1, max: 30 })) + ' 20:00:00');
        counts.reviews++;
      }
    }
  }
}

db.close();
console.log(`demo database written to ${OUT}`);
console.log(`  ${LEADS.length} trek leads (${approvedLeads.length} approved), ${trekkers.length} trekkers`);
console.log(`  ${counts.batches} trek dates, ${counts.bookings} bookings, ${counts.reviews} reviews`);
console.log('  login for any demo account: <email from Admin → Users> / demo1234');
