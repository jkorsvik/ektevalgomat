// /rommet: the twelve dimensions drawn as three axes.
//
// Hand-written WebGL, no library, and no network call of any kind: the geometry is
// inlined in a JSON block by the server, so this file only reads the DOM and draws.
// That is a property the site tests enforce (web/network_guard_test.go), and it is
// why there is no three.js here for what is 144 line segments and 108 points.
//
// The numbers are not invented here. Every coordinate comes from the payload, which
// the server computes from the dataset on each request. If a position changes, this
// picture changes with it, and a Go test fails if the page and the projection ever
// disagree.
(function () {
  'use strict';

  var canvas = document.getElementById('rommet-canvas');
  var status = document.getElementById('rommet-status');
  var dataEl = document.getElementById('rommet-data');
  if (!canvas || !dataEl) return;

  var say = function (msg) {
    if (status) status.textContent = msg;
  };

  var space;
  try {
    space = JSON.parse(dataEl.textContent);
  } catch (e) {
    say('Kunne ikke lese tallene bak tegningen. Tabellen under har dem.');
    return;
  }
  if (!space || !space.clouds || !space.clouds.length) {
    say('Ingen geometri å tegne. Tabellen under har tallene.');
    return;
  }

  // Palette keyed by party id, mirrored by .rom-dot-<id> in rommet.css so the
  // table and the drawing cannot disagree about which colour is which party.
  var PALETTE = {
    ap: [0.843, 0.149, 0.239],
    frp: [0.122, 0.294, 0.600],
    h: [0.184, 0.435, 0.816],
    krf: [0.890, 0.710, 0.020],
    mdg: [0.184, 0.490, 0.196],
    rodt: [0.557, 0.106, 0.137],
    sp: [0.420, 0.557, 0.137],
    sv: [0.757, 0.275, 0.184],
    v: [0.059, 0.561, 0.541]
  };
  function colourOf(id) {
    return PALETTE[id] || [0.75, 0.75, 0.75];
  }

  // Normalise the cloud to the unit ball so the camera distance is fixed and the
  // view cannot silently zoom differently as the data grows.
  var maxAbs = 0;
  space.clouds.forEach(function (c) {
    c.points.forEach(function (p) {
      for (var a = 0; a < 3; a++) {
        var v = Math.abs(c.centre[a] + p[a]);
        if (v > maxAbs) maxAbs = v;
      }
    });
    for (var a = 0; a < 3; a++) {
      var w = Math.abs(c.centre[a]);
      if (w > maxAbs) maxAbs = w;
    }
  });
  if (maxAbs <= 0) maxAbs = 1;

  // Three draw lists. springs: the vote-agreement bands, which the reader can turn
  // off because they answer a different question from the geometry. lines: the
  // spokes from a party's centre to each dimension's contribution. points: those
  // contributions plus the centre. Interleaved x,y,z,r,g,b,a.
  var lineVerts = [];
  var pointVerts = [];
  var springVerts = [];
  var frameVerts = [];

  function pushLine(buf, x1, y1, z1, x2, y2, z2, rgb, alpha) {
    buf.push(x1 / maxAbs, y1 / maxAbs, z1 / maxAbs, rgb[0], rgb[1], rgb[2], alpha);
    buf.push(x2 / maxAbs, y2 / maxAbs, z2 / maxAbs, rgb[0], rgb[1], rgb[2], alpha);
  }

  // Vote springs, in their own buffer so they can be hidden without rebuilding the
  // geometry: they are a claim about voting, not about position, and a reader
  // comparing positions should be able to remove them.
  (space.springs || []).forEach(function (s) {
    var a = null, b = null;
    space.clouds.forEach(function (c) {
      if (c.id === s.a) a = c;
      if (c.id === s.b) b = c;
    });
    if (!a || !b) return;
    // Agreement runs 0..1 in the data but never reaches 1, so stretch the visible
    // range over what actually occurs instead of over a fictional 0..1 span.
    var t = (s.agreement - 0.4) / 0.5;
    if (t < 0) t = 0;
    if (t > 1) t = 1;
    pushLine(springVerts,
      a.centre[0], a.centre[1], a.centre[2],
      b.centre[0], b.centre[1], b.centre[2],
      [0.549, 0.561, 0.600], 0.10 + 0.45 * t);
  });

  // The reference frame. Without it the cloud has no orientation: a reader can see
  // that parties are apart but cannot say along what. Each axis is drawn from the
  // origin to a fixed length, so the frame does not move when the data grows, and
  // its label carries the share of the variation that axis keeps.
  var FRAME_LEN = 1.25;
  var AXIS_RGB = [[0.40, 0.44, 0.50], [0.55, 0.49, 0.42], [0.44, 0.52, 0.46]];
  for (var ai = 0; ai < 3; ai++) {
    var tip = [0, 0, 0];
    tip[ai] = FRAME_LEN * maxAbs;
    pushLine(frameVerts, 0, 0, 0, tip[0], tip[1], tip[2], AXIS_RGB[ai], 0.75);
    // A short cross at the tip, in the two other directions, so the line reads as an
    // axis with an end rather than as one more spoke through the cloud.
    for (var bi = 0; bi < 3; bi++) {
      if (bi === ai) continue;
      var a0 = [0, 0, 0], a1 = [0, 0, 0];
      a0[ai] = FRAME_LEN * maxAbs; a1[ai] = FRAME_LEN * maxAbs;
      a0[bi] = -0.05 * maxAbs; a1[bi] = 0.05 * maxAbs;
      pushLine(frameVerts, a0[0], a0[1], a0[2], a1[0], a1[1], a1[2], AXIS_RGB[ai], 0.75);
    }
  }

  space.clouds.forEach(function (c) {
    var rgb = colourOf(c.id);
    c.points.forEach(function (off) {
      var x = c.centre[0] + off[0];
      var y = c.centre[1] + off[1];
      var z = c.centre[2] + off[2];
      // The spoke shows which dimension pulls the party where; the offset is exact,
      // it is the party's own contribution minus its average pull.
      pushLine(lineVerts, c.centre[0], c.centre[1], c.centre[2], x, y, z, rgb, 0.40);
      pointVerts.push(x / maxAbs, y / maxAbs, z / maxAbs, rgb[0], rgb[1], rgb[2], 0.95);
    });
    // The centre itself, brighter, so the party is findable inside its own cloud.
    pointVerts.push(
      c.centre[0] / maxAbs, c.centre[1] / maxAbs, c.centre[2] / maxAbs,
      rgb[0], rgb[1], rgb[2], 1.0);
  });

  var gl = canvas.getContext('webgl', { antialias: true, alpha: false });
  if (!gl) {
    say('Nettleseren gir ikke WebGL her. Tabellen under har de samme tallene.');
    return;
  }

  function compile(type, src) {
    var sh = gl.createShader(type);
    gl.shaderSource(sh, src);
    gl.compileShader(sh);
    if (!gl.getShaderParameter(sh, gl.COMPILE_STATUS)) {
      throw new Error(gl.getShaderInfoLog(sh) || 'shader');
    }
    return sh;
  }

  var vs = compile(gl.VERTEX_SHADER, [
    'attribute vec3 aPos;',
    'attribute vec4 aCol;',
    'uniform mat4 uMVP;',
    'uniform float uSize;',
    'varying vec4 vCol;',
    'void main() {',
    '  vCol = aCol;',
    '  gl_Position = uMVP * vec4(aPos, 1.0);',
    '  gl_PointSize = uSize;',
    '}'
  ].join('\n'));

  var fs = compile(gl.FRAGMENT_SHADER, [
    'precision mediump float;',
    'varying vec4 vCol;',
    'void main() {',
    '  vec2 d = gl_PointCoord - vec2(0.5);',
    '  float r = dot(d, d);',
    '  if (r > 0.25) discard;',
    '  gl_FragColor = vec4(vCol.rgb, vCol.a * (1.0 - r * 2.0));',
    '}'
  ].join('\n'));

  var prog = gl.createProgram();
  gl.attachShader(prog, vs);
  gl.attachShader(prog, fs);
  gl.linkProgram(prog);
  if (!gl.getProgramParameter(prog, gl.LINK_STATUS)) {
    say('Tegningen kunne ikke startes. Tabellen under har de samme tallene.');
    return;
  }
  gl.useProgram(prog);

  var aPos = gl.getAttribLocation(prog, 'aPos');
  var aCol = gl.getAttribLocation(prog, 'aCol');
  var uMVP = gl.getUniformLocation(prog, 'uMVP');
  var uSize = gl.getUniformLocation(prog, 'uSize');

  function bufferFrom(verts) {
    var b = gl.createBuffer();
    gl.bindBuffer(gl.ARRAY_BUFFER, b);
    gl.bufferData(gl.ARRAY_BUFFER, new Float32Array(verts), gl.STATIC_DRAW);
    return { buf: b, count: verts.length / 7 };
  }
  var frame = bufferFrom(frameVerts);
  var lines = bufferFrom(lineVerts);
  var points = bufferFrom(pointVerts);
  var springs = bufferFrom(springVerts);

  // The dimension highlight: one spoke per party for the chosen dimension, drawn
  // from the party's centre to that dimension's own contribution. Rebuilt when the
  // reader picks another dimension, because the geometry is per dimension.
  function spikesFor(dim) {
    var verts = [];
    if (dim < 0) return verts;
    space.clouds.forEach(function (c) {
      var off = c.points[dim];
      if (!off) return;
      var rgb = colourOf(c.id);
      pushLine(verts,
        c.centre[0], c.centre[1], c.centre[2],
        c.centre[0] + off[0], c.centre[1] + off[1], c.centre[2] + off[2],
        rgb, 0.95);
    });
    return verts;
  }
  var spikeBuf = { buf: gl.createBuffer(), count: 0 };
  function setSpikes(dim) {
    var verts = spikesFor(dim);
    gl.bindBuffer(gl.ARRAY_BUFFER, spikeBuf.buf);
    gl.bufferData(gl.ARRAY_BUFFER, new Float32Array(verts), gl.DYNAMIC_DRAW);
    spikeBuf.count = verts.length / 7;
  }

  // The ring around the selected party's centre: one point, drawn larger than the
  // contributions so it stays findable inside its own cloud.
  var focusRing = { buf: gl.createBuffer(), count: 0 };
  function setFocus(id) {
    selected = id || null;
    if (!selected) { focusRing.count = 0; return; }
    var c0 = null;
    space.clouds.forEach(function (c) { if (c.id === selected) c0 = c; });
    if (!c0) { focusRing.count = 0; return; }
    var verts = [c0.centre[0] / maxAbs, c0.centre[1] / maxAbs, c0.centre[2] / maxAbs,
                 1.0, 1.0, 1.0, 1.0];
    gl.bindBuffer(gl.ARRAY_BUFFER, focusRing.buf);
    gl.bufferData(gl.ARRAY_BUFFER, new Float32Array(verts), gl.DYNAMIC_DRAW);
    focusRing.count = 1;
  }

  // Axis labels: HTML, because WebGL cannot write text, positioned from the same
  // matrix the frame was drawn with. They are aria-hidden: the percentages are in
  // the table below, and a screen reader should not read a floating duplicate.
  var axesEl = document.getElementById('rommet-axes');
  var axisLabels = [];
  function tenths(v) {
    return (Math.round(v * 1000) / 10).toFixed(1).replace('.', ',') + ' %';
  }
  if (axesEl) {
    for (var k = 0; k < 3; k++) {
      var sp = document.createElement('span');
      sp.className = 'rom-axis-label rom-axis-label-' + (k + 1);
      var av = (space.axes || [])[k];
      sp.textContent = 'Akse ' + (k + 1) + (av ? ' \u00b7 ' + tenths(av.variance) : '');
      axesEl.appendChild(sp);
      axisLabels.push(sp);
    }
  }
  function placeAxisLabels() {
    if (!axisLabels.length || !canvas.clientWidth) return;
    var sx = canvas.width / canvas.clientWidth;
    var sy = canvas.height / canvas.clientHeight;
    for (var k = 0; k < 3; k++) {
      var lbl = axisLabels[k];
      var tip = [0, 0, 0];
      tip[k] = FRAME_LEN * maxAbs;
      var pt = projectPoint(tip[0], tip[1], tip[2]);
      if (!pt) { lbl.hidden = true; continue; }
      lbl.hidden = false;
      lbl.style.left = Math.round(pt.x / sx) + 'px';
      lbl.style.top = Math.round(pt.y / sy) + 'px';
    }
  }

  // state the reader drives
  var dist = 2.35;
  var springsOn = true;
  var dimSel = -1;
  var selected = null;
  var lastMVP = null;
  var lastW = 0, lastH = 0;

  gl.enable(gl.BLEND);
  gl.blendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA);
  gl.clearColor(0.98, 0.98, 0.97, 1.0);
  gl.enable(gl.DEPTH_TEST);

  // --- small matrix helpers, column-major like WebGL expects -------------------
  function perspective(fovy, aspect, near, far) {
    var f = 1 / Math.tan(fovy / 2), nf = 1 / (near - far);
    return new Float32Array([
      f / aspect, 0, 0, 0,
      0, f, 0, 0,
      0, 0, (far + near) * nf, -1,
      0, 0, 2 * far * near * nf, 0
    ]);
  }
  function mul(a, b) {
    var o = new Float32Array(16);
    for (var c = 0; c < 4; c++) {
      for (var r = 0; r < 4; r++) {
        o[c * 4 + r] = a[r] * b[c * 4] + a[4 + r] * b[c * 4 + 1] +
                       a[8 + r] * b[c * 4 + 2] + a[12 + r] * b[c * 4 + 3];
      }
    }
    return o;
  }
  function rotY(t) {
    var c = Math.cos(t), s = Math.sin(t);
    return new Float32Array([c, 0, -s, 0, 0, 1, 0, 0, s, 0, c, 0, 0, 0, 0, 1]);
  }
  function rotX(t) {
    var c = Math.cos(t), s = Math.sin(t);
    return new Float32Array([1, 0, 0, 0, 0, c, s, 0, 0, -s, c, 0, 0, 0, 0, 1]);
  }

  var yaw = 0.6, pitch = -0.25, drag = false, lastX = 0, lastY = 0;
  var downX = 0, downY = 0;

  function draw() {
    var dpr = Math.min(window.devicePixelRatio || 1, 2);
    var w = canvas.clientWidth || canvas.width;
    var h = canvas.clientHeight || canvas.height;
    if (canvas.width !== Math.round(w * dpr) || canvas.height !== Math.round(h * dpr)) {
      canvas.width = Math.round(w * dpr);
      canvas.height = Math.round(h * dpr);
    }
    gl.viewport(0, 0, canvas.width, canvas.height);
    gl.clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT);

    var proj = perspective(0.9, canvas.width / canvas.height, 0.1, 100);
    // Camera at a distance the reader controls: the cloud is normalised, so nothing
    // here depends on how many rows the dataset holds.
    var view = mul(rotX(pitch), rotY(yaw));
    var t = new Float32Array([1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, -dist, 1]);
    // proj * translate * rotate. The rotation must apply FIRST so the field spins
    // about its own centre; multiplying proj*view*t instead spun a translated
    // object around the world origin and threw the whole thing into the corner of
    // the clip volume, which looks exactly like a canvas that draws nothing.
    var m = mul(proj, mul(t, view));
    lastMVP = m;
    lastW = canvas.width;
    lastH = canvas.height;

    gl.uniformMatrix4fv(uMVP, false, m);

    gl.enableVertexAttribArray(aPos);
    gl.enableVertexAttribArray(aCol);
    gl.uniform1f(uSize, 1);

    // The frame first, and thinner than the spokes, so it reads as the box the
    // parties sit in rather than as one more party.
    gl.bindBuffer(gl.ARRAY_BUFFER, frame.buf);
    gl.vertexAttribPointer(aPos, 3, gl.FLOAT, false, 28, 0);
    gl.vertexAttribPointer(aCol, 4, gl.FLOAT, false, 28, 12);
    gl.drawArrays(gl.LINES, 0, frame.count);

    gl.bindBuffer(gl.ARRAY_BUFFER, lines.buf);
    gl.vertexAttribPointer(aPos, 3, gl.FLOAT, false, 28, 0);
    gl.vertexAttribPointer(aCol, 4, gl.FLOAT, false, 28, 12);
    gl.drawArrays(gl.LINES, 0, lines.count);

    if (springsOn && springs.count) {
      gl.bindBuffer(gl.ARRAY_BUFFER, springs.buf);
      gl.vertexAttribPointer(aPos, 3, gl.FLOAT, false, 28, 0);
      gl.vertexAttribPointer(aCol, 4, gl.FLOAT, false, 28, 12);
      gl.drawArrays(gl.LINES, 0, springs.count);
    }

    if (dimSel >= 0 && spikeBuf.count) {
      gl.bindBuffer(gl.ARRAY_BUFFER, spikeBuf.buf);
      gl.vertexAttribPointer(aPos, 3, gl.FLOAT, false, 28, 0);
      gl.vertexAttribPointer(aCol, 4, gl.FLOAT, false, 28, 12);
      gl.drawArrays(gl.LINES, 0, spikeBuf.count);
    }

    gl.bindBuffer(gl.ARRAY_BUFFER, points.buf);
    gl.vertexAttribPointer(aPos, 3, gl.FLOAT, false, 28, 0);
    gl.vertexAttribPointer(aCol, 4, gl.FLOAT, false, 28, 12);
    gl.uniform1f(uSize, 6.5 * dpr);
    gl.drawArrays(gl.POINTS, 0, points.count);

    // The selected party's centre, drawn last and larger so it reads through the
    // cloud it sits inside.
    if (selected) {
      var c0 = null;
      space.clouds.forEach(function (c) { if (c.id === selected) c0 = c; });
      if (c0) {
        gl.bindBuffer(gl.ARRAY_BUFFER, focusRing.buf);
        gl.vertexAttribPointer(aPos, 3, gl.FLOAT, false, 28, 0);
        gl.vertexAttribPointer(aCol, 4, gl.FLOAT, false, 28, 12);
        gl.uniform1f(uSize, 16 * dpr);
        gl.drawArrays(gl.POINTS, 0, focusRing.count);
      }
    }
    placeAxisLabels();
  }

  function loop() {
    if (!drag) yaw += 0.0022;
    draw();
    window.requestAnimationFrame(loop);
  }

  // Pointer events: no pointer capture library, and touch works through the same
  // path because pointerdown covers both.
  canvas.addEventListener('pointerdown', function (e) {
    drag = true;
    lastX = e.clientX;
    lastY = e.clientY;
    downX = e.clientX;
    downY = e.clientY;
    if (canvas.setPointerCapture) canvas.setPointerCapture(e.pointerId);
  });
  canvas.addEventListener('pointermove', function (e) {
    if (!drag) return;
    yaw += (e.clientX - lastX) * 0.008;
    pitch += (e.clientY - lastY) * 0.006;
    if (pitch > 1.4) pitch = 1.4;
    if (pitch < -1.4) pitch = -1.4;
    lastX = e.clientX;
    lastY = e.clientY;
  });
  function release() { drag = false; }
  canvas.addEventListener('pointerup', release);
  canvas.addEventListener('pointercancel', release);
  canvas.addEventListener('pointerleave', release);
  // Keyboard: the canvas is not focusable by default, so make it reachable and let
  // the arrow keys rotate it. A mouse-only picture is not a picture everyone can use.
  canvas.tabIndex = 0;
  canvas.addEventListener('keydown', function (e) {
    var step = 0.12;
    if (e.key === 'ArrowLeft') { yaw -= step; e.preventDefault(); }
    else if (e.key === 'ArrowRight') { yaw += step; e.preventDefault(); }
    else if (e.key === 'ArrowUp') { pitch = Math.max(-1.4, pitch - step); e.preventDefault(); }
    else if (e.key === 'ArrowDown') { pitch = Math.min(1.4, pitch + step); e.preventDefault(); }
  });

  // --- exploration: pick a party, highlight a dimension, choose the distance ----

  var focusEl = document.getElementById('rommet-focus');
  var dimEl = document.getElementById('rommet-dim');
  var springsEl = document.getElementById('rommet-springs');
  var zoomEl = document.getElementById('rommet-zoom');

  function fmt(v) {
    var s = (Math.round(v * 100) / 100).toFixed(2).replace('.', ',');
    return s.replace('-', '\u2212');
  }

  // Which fitted axis a dimension leans on most. A dimension's contribution has a
  // component on all three axes, so ranking by the axis it loads heaviest on is the
  // only ranking that reads as one number per party.
  function dominantAxis(dim) {
    var best = 0, bestAbs = -1;
    (space.axes || []).forEach(function (a, k) {
      var w = Math.abs((a.loadings || [])[dim] || 0);
      if (w > bestAbs) { bestAbs = w; best = k; }
    });
    return best;
  }

  function agreementWith(id) {
    var out = null;
    (space.springs || []).forEach(function (s) {
      if (s.a !== id && s.b !== id) return;
      var other = s.a === id ? s.b : s.a;
      var v = s.agreement;
      if (!out) out = { best: { id: other, a: v }, worst: { id: other, a: v } };
      if (v > out.best.a) out.best = { id: other, a: v };
      if (v < out.worst.a) out.worst = { id: other, a: v };
    });
    return out;
  }

  function nameOf(id) {
    var n = id;
    space.clouds.forEach(function (c) { if (c.id === id && c.short) n = c.short; });
    return n;
  }

  function renderFocus() {
    if (!focusEl) return;
    focusEl.textContent = '';
    var c0 = null;
    space.clouds.forEach(function (c) { if (c.id === selected) c0 = c; });
    if (!c0) {
      var p = document.createElement('p');
      p.className = 'rom-focus-empty';
      p.textContent = 'Klikk et punkt for å se partiets egne tall.';
      focusEl.appendChild(p);
      return;
    }
    var h = document.createElement('p');
    h.className = 'rom-focus-head';
    h.textContent = (c0.short || c0.id) + ' · spredning ' + fmt(c0.spread) +
      (c0.answers ? ' · ' + c0.answers + ' svar' : '');
    focusEl.appendChild(h);

    var ag = agreementWith(c0.id);
    if (ag) {
      var ap = document.createElement('p');
      ap.className = 'rom-focus-line';
      ap.textContent = 'Mest enig med ' + nameOf(ag.best.id) + ' (' +
        Math.round(ag.best.a * 100) + ' %), minst med ' + nameOf(ag.worst.id) + ' (' +
        Math.round(ag.worst.a * 100) + ' %).';
      focusEl.appendChild(ap);
    }

    var list = document.createElement('ol');
    list.className = 'rom-focus-list';
    if (dimSel >= 0) {
      var axis = dominantAxis(dimSel);
      var intro = document.createElement('p');
      intro.className = 'rom-focus-line';
      intro.textContent = '«' + (space.dims[dimSel] || {}).name + '» lener seg mest på akse ' +
        (axis + 1) + '. Rangert etter hvor mye den skyver partiet langs den:';
      focusEl.appendChild(intro);
      var ranked = space.clouds.map(function (c) {
        return { id: c.id, v: (c.points[dimSel] || [0, 0, 0])[axis] };
      }).sort(function (a, b) { return b.v - a.v; });
      ranked.forEach(function (r) {
        var li = document.createElement('li');
        li.textContent = nameOf(r.id) + ' ' + fmt(r.v);
        if (r.id === c0.id) li.className = 'rom-focus-mine';
        list.appendChild(li);
      });
    } else {
      var intro2 = document.createElement('p');
      intro2.className = 'rom-focus-line';
      intro2.textContent = 'Dimensjonene som drar dette partiet lengst fra dets eget snitt:';
      focusEl.appendChild(intro2);
      var mine = c0.points.map(function (off, j) {
        return { j: j, len: Math.sqrt(off[0] * off[0] + off[1] * off[1] + off[2] * off[2]) };
      }).sort(function (a, b) { return b.len - a.len; }).slice(0, 3);
      mine.forEach(function (m) {
        var li = document.createElement('li');
        li.textContent = (space.dims[m.j] || {}).name + ' ' + fmt(m.len);
        list.appendChild(li);
      });
    }
    focusEl.appendChild(list);
  }

  // Screen position of a world point under the matrix the last frame drew with.
  function projectPoint(x, y, z) {
    if (!lastMVP) return null;
    var v = [x / maxAbs, y / maxAbs, z / maxAbs, 1];
    var out = [0, 0, 0, 0];
    for (var r = 0; r < 4; r++) {
      out[r] = lastMVP[r] * v[0] + lastMVP[4 + r] * v[1] +
               lastMVP[8 + r] * v[2] + lastMVP[12 + r] * v[3];
    }
    if (out[3] <= 0) return null;
    return {
      x: (out[0] / out[3] * 0.5 + 0.5) * lastW,
      y: (1 - (out[1] / out[3] * 0.5 + 0.5)) * lastH
    };
  }

  function pickAt(clientX, clientY) {
    var rect = canvas.getBoundingClientRect();
    var sx = (clientX - rect.left) * (canvas.width / rect.width);
    var sy = (clientY - rect.top) * (canvas.height / rect.height);
    var best = null, bestD = 26 * (canvas.width / Math.max(rect.width, 1));
    space.clouds.forEach(function (c) {
      var p = projectPoint(c.centre[0], c.centre[1], c.centre[2]);
      if (!p) return;
      var d = Math.hypot(p.x - sx, p.y - sy);
      if (d < bestD) { bestD = d; best = c.id; }
    });
    return best;
  }


  // --- the spectrum of one dimension ------------------------------------------
  //
  // The three-axis space is a summary: it keeps shared variation and mixes the
  // dimensions. A reader asking "where does the spectrum run on THIS question"
  // would otherwise have to read a mixture, so the strip plots the dimension
  // itself, standardised per dimension by the same projection, one dot per party.
  var spectrumEl = document.getElementById('rommet-spectrum');
  var spectrumNote = document.getElementById('rommet-spectrum-note');
  var SVGNS = 'http://www.w3.org/2000/svg';

  function svgEl(name, attrs) {
    var el = document.createElementNS(SVGNS, name);
    for (var key in attrs) {
      if (Object.prototype.hasOwnProperty.call(attrs, key)) {
        el.setAttribute(key, String(attrs[key]));
      }
    }
    return el;
  }
  function ink(rgb) {
    return 'rgb(' + Math.round(rgb[0] * 255) + ',' + Math.round(rgb[1] * 255) +
      ',' + Math.round(rgb[2] * 255) + ')';
  }

  function renderSpectrum(dim) {
    if (!spectrumEl) return;
    while (spectrumEl.firstChild) spectrumEl.removeChild(spectrumEl.firstChild);
    var dimName = dim >= 0 ? ((space.dims || [])[dim] || {}).name : '';
    if (dim < 0) {
      spectrumEl.setAttribute('hidden', 'hidden');
      if (spectrumNote) {
        spectrumNote.textContent = 'Velg en dimensjon over, og partiene stilles opp ' +
          'etter hvor de faktisk står på akkurat den.';
      }
      return;
    }
    spectrumEl.removeAttribute('hidden');

    // Standardised scores, so the axis reads as "how many spreads from the mean"
    // rather than in the unit of whichever statements happen to be coded there.
    var rows = space.clouds.map(function (c) {
      return { id: c.id, short: c.short || c.id, v: (c.scores || [])[dim] };
    }).filter(function (r) { return typeof r.v === 'number'; });
    if (!rows.length) {
      spectrumEl.setAttribute('hidden', 'hidden');
      if (spectrumNote) spectrumNote.textContent = 'Denne dimensjonen har ingen kodede posisjoner.';
      return;
    }
    var lo = Infinity, hi = -Infinity;
    rows.forEach(function (r) {
      if (r.v < lo) lo = r.v;
      if (r.v > hi) hi = r.v;
    });
    if (hi - lo < 1e-9) { hi = lo + 1; }

    var W = 620, PAD = 46, BASE = 44;
    var x = function (v) { return PAD + (v - lo) / (hi - lo) * (W - 2 * PAD); };

    spectrumEl.setAttribute('viewBox', '0 0 ' + W + ' 78');
    spectrumEl.appendChild(svgEl('line', {
      x1: PAD, y1: BASE, x2: W - PAD, y2: BASE, class: 'rom-spectrum-axis'
    }));
    // The mean, marked because the scale is standardised and zero is meaningful.
    var zero = (lo <= 0 && hi >= 0) ? x(0) : null;
    if (zero !== null) {
      spectrumEl.appendChild(svgEl('line', {
        x1: zero, y1: BASE - 7, x2: zero, y2: BASE + 7, class: 'rom-spectrum-zero'
      }));
    }

    // Parties that share a position are drawn as one stack with one label, because
    // an overlap is a finding: on this dimension Ap, SV and R are the same point,
    // and three dots on top of each other would hide that rather than say it.
    var sorted = rows.slice().sort(function (a, b) { return a.v - b.v; });
    var groups = [];
    sorted.forEach(function (r) {
      var last = groups[groups.length - 1];
      if (last && x(r.v) - x(last[0].v) < 16) last.push(r);
      else groups.push([r]);
    });
    groups.forEach(function (g) {
      var cx = x(g[0].v);
      var n = g.length;
      g.forEach(function (r, i) {
        var dot = svgEl('circle', {
          cx: cx, cy: BASE - (n > 1 ? (i - (n - 1) / 2) * 9 : 0), r: 7,
          fill: ink(colourOf(r.id)), class: 'rom-spectrum-dot'
        });
        if (r.id === selected) dot.setAttribute('class', 'rom-spectrum-dot rom-spectrum-mine');
        var t = svgEl('title', {});
        t.textContent = r.short + ' ' + fmt(r.v);
        dot.appendChild(t);
        spectrumEl.appendChild(dot);
      });
      var lab = svgEl('text', {
        x: cx, y: BASE - 14 - (n > 1 ? (n - 1) * 4.5 : 0), 'text-anchor': 'middle',
        class: 'rom-spectrum-name' +
          (g.some(function (r) { return r.id === selected; }) ? ' rom-spectrum-name-mine' : '')
      });
      lab.textContent = g.map(function (r) { return r.short; }).join(' \u00b7 ');
      spectrumEl.appendChild(lab);
    });

    [[lo, PAD, 'start'], [hi, W - PAD, 'end']].forEach(function (end) {
      var t = svgEl('text', {
        x: end[1], y: BASE + 22, 'text-anchor': end[2], class: 'rom-spectrum-scale'
      });
      t.textContent = fmt(end[0]);
      spectrumEl.appendChild(t);
    });

    if (spectrumNote) {
      spectrumNote.textContent = '\u00ab' + dimName + '\u00bb: hvert parti plassert etter ' +
        'sin egen posisjon på denne dimensjonen, standardisert så tallet er antall ' +
        'spredninger fra gjennomsnittet. ' + rows.length + ' partier.';
    }
  }

  if (dimEl) {
    dimEl.addEventListener('change', function () {
      dimSel = parseInt(dimEl.value, 10);
      if (isNaN(dimSel)) dimSel = -1;
      setSpikes(dimSel);
      renderSpectrum(dimSel);
      renderFocus();
    });
  }
  if (springsEl) {
    springsEl.addEventListener('change', function () {
      springsOn = springsEl.checked;
    });
  }
  if (zoomEl) {
    // The slider is the accessible control and the wheel is the quick one; both
    // write the same distance, so neither can leave the other out of date.
    zoomEl.addEventListener('input', function () {
      dist = parseInt(zoomEl.value, 10) / 10;
    });
  }
  canvas.addEventListener('wheel', function (e) {
    dist += e.deltaY > 0 ? 0.12 : -0.12;
    if (dist < 1.4) dist = 1.4;
    if (dist > 4.6) dist = 4.6;
    if (zoomEl) zoomEl.value = String(Math.round(dist * 10));
    e.preventDefault();
  }, { passive: false });

  canvas.addEventListener('pointerup', function (e) {
    // A drag is a rotation, not a click. Only a pointer that barely moved selects.
    if (Math.abs(e.clientX - downX) > 4 || Math.abs(e.clientY - downY) > 4) return;
    var id = pickAt(e.clientX, e.clientY);
    setFocus(id);
    renderFocus();
    renderSpectrum(dimSel);
  });

  setSpikes(-1);
  setFocus(null);
  renderSpectrum(-1);

  // Draw one frame BEFORE announcing success. Announcing first meant the page
  // said it had drawn while the canvas was blank, because the message was set
  // before any pixel existed.
  try {
    draw();
  } catch (e) {
    say('Tegningen feilet: ' + (e && e.message ? e.message : e) +
        '. Tabellen under har de samme tallene.');
    return;
  }
  if (gl.getError() !== gl.NO_ERROR) {
    say('Tegningen ga en WebGL-feil. Tabellen under har de samme tallene.');
    return;
  }

  say('Tegningen viser ' + space.clouds.length + ' partier og ' +
      (space.springs || []).length + ' stemmepar. Tre akser beholder ' +
      (space.explained * 100).toFixed(1).replace('.', ',') +
      ' prosent av variasjonen. Dra for å rotere, rull for å zoome, klikk et punkt ' +
      'for partiets egne tall.');

  loop();
})();
