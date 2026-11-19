// app.js draws the ants of lem-in moving through the colony.
//
// The Go program writes the colony and the turns into the page as JSON
// (see web.go). This file reads that JSON and draws it with SVG. Three jobs:
//
//   1. buildFrames   : for every turn, in which room is every ant?
//   2. project       : turn the room coordinates into positions in the drawing.
//   3. init          : draw, and animate between two turns (play, pause, step,
//                      speed). The ant is placed between two rooms, so one
//                      number p ("progress") describes everything:
//                      p = 2.4 means "turn 2 is done, turn 3 is 40% played".
//
// buildFrames, project and buildHeadings need no browser, so they are
// exported at the bottom and can be tested with node.
(function () {
  'use strict';

  // ---------- 1. state of every turn ----------

  // buildFrames returns pos, where pos[t][i] is the index (in data.rooms) of
  // the room of ant i+1 after t turns. pos[0] is the start: every ant is in
  // the start room. Like the Go code, only positions are kept.
  function buildFrames(data) {
    const index = new Map(data.rooms.map((room, i) => [room.name, i]));
    const start = index.get(data.start);
    const end = index.get(data.end);
    const pos = [new Int32Array(data.ants).fill(start)];
    for (const turn of data.turns) {
      const next = Int32Array.from(pos[pos.length - 1]);
      for (const move of turn) {
        next[move.a - 1] = index.get(move.r);
      }
      pos.push(next);
    }
    return { index, start, end, pos };
  }

  // ---------- 2. coordinates ----------

  // project scales the room coordinates to fit width x height, keeping the
  // proportions (a square colony stays square), and centres the result.
  function project(rooms, width, height, pad) {
    let minX = Infinity, maxX = -Infinity, minY = Infinity, maxY = -Infinity;
    for (const r of rooms) {
      minX = Math.min(minX, r.x); maxX = Math.max(maxX, r.x);
      minY = Math.min(minY, r.y); maxY = Math.max(maxY, r.y);
    }
    const w = maxX - minX;
    const h = maxY - minY;
    // || 1 avoids dividing by zero when all rooms share one x or one y.
    const scale = Math.min((width - 2 * pad) / (w || 1), (height - 2 * pad) / (h || 1));
    const offX = (width - w * scale) / 2 - minX * scale;
    const offY = (height - h * scale) / 2 - minY * scale;
    return rooms.map((r) => ({ x: r.x * scale + offX, y: r.y * scale + offY }));
  }

  // buildHeadings returns, for every turn, the direction (an angle in
  // radians) each ant faces. An ant faces the way of its last move and keeps
  // that direction while it rests. Before its first move it looks at the end.
  function buildHeadings(frames, pts, ants) {
    const first = Math.atan2(
      pts[frames.end].y - pts[frames.start].y,
      pts[frames.end].x - pts[frames.start].x
    );
    const heads = [new Float32Array(ants).fill(first)];
    for (let t = 0; t + 1 < frames.pos.length; t++) {
      const next = Float32Array.from(heads[t]);
      for (let i = 0; i < ants; i++) {
        const a = frames.pos[t][i];
        const b = frames.pos[t + 1][i];
        if (a !== b) {
          next[i] = Math.atan2(pts[b].y - pts[a].y, pts[b].x - pts[a].x);
        }
      }
      heads.push(next);
    }
    return heads;
  }

  // ease makes an ant start and stop gently instead of at a constant speed.
  function ease(f) {
    return f * f * (3 - 2 * f);
  }

  // ---------- 3. drawing and animation ----------

  const SVG = 'http://www.w3.org/2000/svg';
  const WIDTH = 1000;
  const HEIGHT = 620;
  const TURNS_PER_SECOND = 1; // at speed 1x

  function node(name, attrs, parent) {
    const e = document.createElementNS(SVG, name);
    for (const key in attrs) e.setAttribute(key, attrs[key]);
    if (parent) parent.appendChild(e);
    return e;
  }

  function init(data) {
    const frames = buildFrames(data);
    const pts = project(data.rooms, WIDTH, HEIGHT, 70);
    const heads = buildHeadings(frames, pts, data.ants);
    const total = data.turns.length;
    const pos = frames.pos;

    document.getElementById('summary').textContent =
      data.ants + (data.ants === 1 ? ' ant, ' : ' ants, ') +
      data.rooms.length + ' rooms, ' + data.links.length + ' tunnels';
    document.getElementById('turns').textContent = total;
    document.getElementById('ants-total').textContent = data.ants;

    // Rooms get smaller when there are many of them.
    const radius = Math.max(4, Math.min(22, 320 / Math.sqrt(data.rooms.length)));
    const antScale = Math.max(0.6, Math.min(2, radius / 9));
    const showNames = data.rooms.length <= 80;

    // Tunnels.
    const linkLayer = document.getElementById('links');
    const linkEls = new Map();
    for (const [a, b] of data.links) {
      const i = frames.index.get(a);
      const j = frames.index.get(b);
      const line = node('line', {
        class: 'link', x1: pts[i].x, y1: pts[i].y, x2: pts[j].x, y2: pts[j].y,
      }, linkLayer);
      linkEls.set(Math.min(i, j) + '|' + Math.max(i, j), line);
    }

    // Rooms. Start and end hold many ants, so they show a number. Every other
    // room holds one ant at most, which is drawn on top of the room.
    const roomLayer = document.getElementById('rooms');
    const roomEls = [];
    let startText, endText;
    data.rooms.forEach((room, i) => {
      const kind = i === frames.start ? ' start' : i === frames.end ? ' end' : '';
      const circle = node('circle', {
        class: 'room' + kind, cx: pts[i].x, cy: pts[i].y, r: radius,
      }, roomLayer);
      node('title', {}, circle).textContent = room.name;
      roomEls.push(circle);
      if (showNames) {
        node('text', { class: 'name', x: pts[i].x, y: pts[i].y + radius + 16 }, roomLayer)
          .textContent = room.name;
      }
      if (kind) {
        const text = node('text', { class: 'count', x: pts[i].x, y: pts[i].y }, roomLayer);
        node('text', {
          class: 'tag' + kind, x: pts[i].x, y: pts[i].y - radius - 8,
        }, roomLayer).textContent = kind.trim();
        if (i === frames.start) startText = text; else endText = text;
      }
    });

    // Ants. A pool of elements is reused each frame: <g> moves the ant, <use>
    // turns it, <text> shows its number when there are few ants.
    const antLayer = document.getElementById('ants');
    const showNumbers = data.ants <= 24;
    const pool = [];
    let inUse = 0;
    function antEl(k) {
      if (!pool[k]) {
        const g = node('g', {}, antLayer);
        const use = node('use', { href: '#ant-a' }, g);
        const num = node('text', { class: 'ant-num', y: -14 * antScale }, g);
        pool[k] = { g, use, num, shown: false };
      }
      return pool[k];
    }
    function placeAnt(k, i, x, y, angle, walking, now) {
      const el = antEl(k);
      el.g.setAttribute('transform', 'translate(' + x.toFixed(1) + ' ' + y.toFixed(1) + ')');
      el.use.setAttribute('transform',
        'rotate(' + (angle * 180 / Math.PI).toFixed(1) + ') scale(' + antScale + ')');
      // Two leg poses, swapped a few times per second, make the ant walk.
      el.use.setAttribute('href', walking && Math.floor(now / 110) % 2 ? '#ant-b' : '#ant-a');
      if (showNumbers) el.num.textContent = i + 1;
      if (!el.shown) { el.g.removeAttribute('display'); el.shown = true; }
    }

    // Controls and state.
    const seek = document.getElementById('seek');
    seek.max = Math.max(total, 1);
    const playBtn = document.getElementById('play');
    const playIcon = document.getElementById('play-icon');
    const turnOut = document.getElementById('turn');
    const arrivedOut = document.getElementById('arrived');
    const speedIn = document.getElementById('speed');
    const speedOut = document.getElementById('speed-out');

    let p = 0;          // progress, from 0 to total
    let playing = false;
    let goal = null;    // when set, p moves to this value and then stops
    let speed = 1;
    let last = null;
    let dirty = true;
    let busyNow = new Uint8Array(data.rooms.length);
    let activeNow = new Set();

    function render(now) {
      const t = Math.min(Math.floor(p), total);
      const f = t < total ? p - t : 0;
      const moving = f > 0;
      const e = ease(f);
      const busy = new Uint8Array(data.rooms.length);
      const active = new Set();
      let used = 0, atStart = 0, atEnd = 0;

      for (let i = 0; i < data.ants; i++) {
        const a = pos[t][i];
        const b = t < total ? pos[t + 1][i] : a;
        if (a !== b && moving) {
          active.add(Math.min(a, b) + '|' + Math.max(a, b));
          const x = pts[a].x + (pts[b].x - pts[a].x) * e;
          const y = pts[a].y + (pts[b].y - pts[a].y) * e;
          placeAnt(used++, i, x, y, heads[t + 1][i], true, now);
        } else if (a === frames.start) {
          atStart++;
        } else if (a === frames.end) {
          atEnd++;
        } else {
          busy[a] = 1;
          placeAnt(used++, i, pts[a].x, pts[a].y, heads[t][i], false, now);
        }
      }
      for (let k = used; k < inUse; k++) {
        pool[k].g.setAttribute('display', 'none');
        pool[k].shown = false;
      }
      inUse = used;

      // Only touch the classes that changed.
      for (let r = 0; r < busy.length; r++) {
        if (busy[r] !== busyNow[r]) roomEls[r].classList.toggle('busy', busy[r] === 1);
      }
      busyNow = busy;
      for (const key of activeNow) if (!active.has(key)) linkEls.get(key).classList.remove('active');
      for (const key of active) if (!activeNow.has(key)) linkEls.get(key).classList.add('active');
      activeNow = active;

      startText.textContent = atStart;
      endText.textContent = atEnd;
      arrivedOut.textContent = atEnd;
      turnOut.textContent = Math.ceil(p);
      seek.value = p;
    }

    function setPlaying(on) {
      playing = on;
      playBtn.setAttribute('aria-label', on ? 'Pause' : 'Play');
      playIcon.setAttribute('d', on ? 'M6 5h4v14H6zM14 5h4v14h-4z' : 'M8 5v14l11-7z');
    }

    function frame(now) {
      const dt = last === null ? 0 : Math.min((now - last) / 1000, 0.1);
      last = now;
      const move = TURNS_PER_SECOND * speed * dt;
      if (goal !== null) {
        if (Math.abs(goal - p) <= move) { p = goal; goal = null; }
        else p += goal > p ? move : -move;
        dirty = true;
      } else if (playing) {
        p += move;
        if (p >= total) { p = total; setPlaying(false); }
        dirty = true;
      }
      if (dirty) { render(now); dirty = false; }
      requestAnimationFrame(frame);
    }

    function togglePlay() {
      if (playing) { setPlaying(false); return; }
      if (p >= total) p = 0;
      goal = null;
      setPlaying(true);
    }
    // A step moves to the next whole turn with the same animation, then stops.
    function step(direction) {
      setPlaying(false);
      const base = goal !== null ? goal : p;
      const target = direction > 0 ? Math.floor(base) + 1 : Math.ceil(base) - 1;
      goal = Math.min(Math.max(target, 0), total);
    }
    function restart() { setPlaying(false); goal = null; p = 0; dirty = true; }
    function setSpeed(v) {
      speed = Math.min(Math.max(Math.round(v * 10) / 10, 0.1), 4);
      speedIn.value = speed;
      speedOut.textContent = speed + 'x';
    }

    playBtn.addEventListener('click', togglePlay);
    document.getElementById('forward').addEventListener('click', () => step(1));
    document.getElementById('back').addEventListener('click', () => step(-1));
    document.getElementById('restart').addEventListener('click', restart);
    speedIn.addEventListener('input', () => setSpeed(parseFloat(speedIn.value)));
    seek.addEventListener('input', () => {
      setPlaying(false); goal = null; p = parseFloat(seek.value); dirty = true;
    });
    document.addEventListener('keydown', (ev) => {
      // A focused slider or button already handles its own keys.
      if (ev.target.tagName === 'INPUT' || (ev.target.tagName === 'BUTTON' && ev.key === ' ')) return;
      switch (ev.key) {
        case ' ': ev.preventDefault(); togglePlay(); break;
        case 'ArrowRight': ev.preventDefault(); step(1); break;
        case 'ArrowLeft': ev.preventDefault(); step(-1); break;
        case 'Home': ev.preventDefault(); restart(); break;
        case '+': case '=': setSpeed(speed + 0.1); break;
        case '-': setSpeed(speed - 0.1); break;
        default: return;
      }
    });

    // After a slider is used, give the keyboard back to the shortcuts.
    for (const slider of [seek, speedIn]) {
      slider.addEventListener('change', () => slider.blur());
    }

    // Tests and screenshots can drive the page through this hook.
    window.lemin = { seekTo: (v) => { setPlaying(false); goal = null; p = v; dirty = true; } };

    requestAnimationFrame(frame);
  }

  if (typeof document !== 'undefined') {
    init(JSON.parse(document.getElementById('colony').textContent));
  }
  if (typeof module !== 'undefined') {
    module.exports = { buildFrames, project, buildHeadings, ease };
  }
})();