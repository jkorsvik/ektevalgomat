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

  // Two draw lists: lines (cloud spokes and vote springs) and points (the dimension
  // contributions). Interleaved x,y,z,r,g,b,a.
  var lineVerts = [];
  var pointVerts = [];

  function pushLine(buf, x1, y1, z1, x2, y2, z2, rgb, alpha) {
    buf.push(x1 / maxAbs, y1 / maxAbs, z1 / maxAbs, rgb[0], rgb[1], rgb[2], alpha);
    buf.push(x2 / maxAbs, y2 / maxAbs, z2 / maxAbs, rgb[0], rgb[1], rgb[2], alpha);
  }

  // Vote springs first, so the party spokes draw over them.
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
    pushLine(lineVerts,
      a.centre[0], a.centre[1], a.centre[2],
      b.centre[0], b.centre[1], b.centre[2],
      [0.549, 0.561, 0.600], 0.10 + 0.45 * t);
  });

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
  var lines = bufferFrom(lineVerts);
  var points = bufferFrom(pointVerts);

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
    // Camera at a fixed distance: the cloud is normalised, so nothing here depends
    // on how many rows the dataset holds.
    var view = mul(rotX(pitch), rotY(yaw));
    var t = new Float32Array([1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, -2.35, 1]);
    // proj * translate * rotate. The rotation must apply FIRST so the field spins
    // about its own centre; multiplying proj*view*t instead spun a translated
    // object around the world origin and threw the whole thing into the corner of
    // the clip volume, which looks exactly like a canvas that draws nothing.
    var m = mul(proj, mul(t, view));

    gl.uniformMatrix4fv(uMVP, false, m);

    gl.bindBuffer(gl.ARRAY_BUFFER, lines.buf);
    gl.enableVertexAttribArray(aPos);
    gl.enableVertexAttribArray(aCol);
    gl.vertexAttribPointer(aPos, 3, gl.FLOAT, false, 28, 0);
    gl.vertexAttribPointer(aCol, 4, gl.FLOAT, false, 28, 12);
    gl.uniform1f(uSize, 1);
    gl.drawArrays(gl.LINES, 0, lines.count);

    gl.bindBuffer(gl.ARRAY_BUFFER, points.buf);
    gl.vertexAttribPointer(aPos, 3, gl.FLOAT, false, 28, 0);
    gl.vertexAttribPointer(aCol, 4, gl.FLOAT, false, 28, 12);
    gl.uniform1f(uSize, 6.5 * dpr);
    gl.drawArrays(gl.POINTS, 0, points.count);
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
      ' prosent av variasjonen. Dra, eller bruk piltastene, for å rotere.');

  loop();
})();
