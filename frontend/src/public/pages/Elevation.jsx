import { useMemo, useState } from 'react';

// Elevation profile: an area chart drawn from "km altitude" lines entered in the
// admin panel, with a hover/touch readout.

const W = 640, H = 170, PAD = { l: 44, r: 12, t: 12, b: 26 };

function parse(text) {
  return String(text || '')
    .split('\n')
    .map((l) => l.trim().split(/[\s,]+/).map(Number))
    .filter((p) => p.length >= 2 && p.every(Number.isFinite))
    .map(([km, m]) => ({ km, m }))
    .sort((a, b) => a.km - b.km);
}

function Profile({ pts }) {
  const [hover, setHover] = useState(null);
  const s = useMemo(() => {
    const maxKm = pts[pts.length - 1].km || 1;
    const lo = Math.min(...pts.map((p) => p.m));
    const hi = Math.max(...pts.map((p) => p.m));
    const span = Math.max(hi - lo, 50);
    const x = (km) => PAD.l + (km / maxKm) * (W - PAD.l - PAD.r);
    const y = (m) => PAD.t + (1 - (m - lo) / span) * (H - PAD.t - PAD.b);
    const line = pts.map((p, i) => `${i ? 'L' : 'M'}${x(p.km).toFixed(1)} ${y(p.m).toFixed(1)}`).join(' ');
    const area = `${line} L${x(maxKm)} ${H - PAD.b} L${x(0)} ${H - PAD.b}Z`;
    let gain = 0;
    for (let i = 1; i < pts.length; i++) gain += Math.max(0, pts[i].m - pts[i - 1].m);
    return { maxKm, lo, hi, x, y, line, area, gain };
  }, [pts]);

  const onMove = (e) => {
    const box = e.currentTarget.getBoundingClientRect();
    const km = ((((e.clientX - box.left) / box.width) * W - PAD.l) / (W - PAD.l - PAD.r)) * s.maxKm;
    let best = pts[0];
    for (const p of pts) if (Math.abs(p.km - km) < Math.abs(best.km - km)) best = p;
    setHover(best);
  };

  return (
    <div className="elev">
      <div className="elev-stats">
        <span><b>{s.maxKm} km</b> distance</span>
        <span><b>↑ {Math.round(s.gain).toLocaleString('en-IN')} m</b> total climb</span>
        <span><b>{s.hi.toLocaleString('en-IN')} m</b> highest point</span>
      </div>
      <svg viewBox={`0 0 ${W} ${H}`} className="elev-svg" onPointerMove={onMove} onPointerLeave={() => setHover(null)}
        role="img" aria-label={`Elevation profile from ${s.lo} m to ${s.hi} m over ${s.maxKm} km`}>
        <defs>
          <linearGradient id="elev-fill" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0" stopColor="var(--forest-500)" stopOpacity=".55" />
            <stop offset="1" stopColor="var(--forest-500)" stopOpacity=".05" />
          </linearGradient>
        </defs>
        <path d={s.area} fill="url(#elev-fill)" />
        <path d={s.line} fill="none" stroke="var(--forest-700)" strokeWidth="2.2" strokeLinejoin="round" />
        <text x={PAD.l - 6} y={s.y(s.hi) + 4} textAnchor="end">{s.hi} m</text>
        <text x={PAD.l - 6} y={s.y(s.lo) + 4} textAnchor="end">{s.lo} m</text>
        <text x={PAD.l} y={H - 8}>0 km</text>
        <text x={W - PAD.r} y={H - 8} textAnchor="end">{s.maxKm} km</text>
        {hover && (
          <g>
            <line x1={s.x(hover.km)} x2={s.x(hover.km)} y1={PAD.t} y2={H - PAD.b} stroke="var(--sun)" strokeDasharray="3 3" />
            <circle cx={s.x(hover.km)} cy={s.y(hover.m)} r="5" fill="var(--sun)" stroke="#fff" strokeWidth="2" />
            <text x={Math.min(Math.max(s.x(hover.km), PAD.l + 50), W - 60)} y={PAD.t + 10} textAnchor="middle" className="elev-tip">
              {hover.km} km · {hover.m.toLocaleString('en-IN')} m
            </text>
          </g>
        )}
      </svg>
    </div>
  );
}

export default function Elevation({ points }) {
  const pts = useMemo(() => parse(points), [points]);
  if (pts.length < 2) return null;
  return <div className="elev-wrap"><Profile pts={pts} /></div>;
}
