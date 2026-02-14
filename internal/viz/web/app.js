// ─── Color Palette ──────────────────────────────────────────
const NODE_COLORS = {
  agent:       '#4fc3f7',
  tool:        '#81c784',
  mcp_server:  '#ce93d8',
  data_source: '#ffb74d',
  llm_endpoint:'#ffd54f',
};

const EDGE_COLORS = {
  tool_call:   '#81c784',
  mcp_call:    '#ce93d8',
  retrieval:   '#4fc3f7',
  write:       '#ffb74d',
  model_call:  '#ffd54f',
};

// ─── State ──────────────────────────────────────────────────
const state = {
  nodes: new Map(),      // id → node data
  edges: new Map(),      // key → edge data
  paused: false,
  pinned: null,
  filterText: '',
  showRaw: false,
  ws: null,
  eventBuffer: [],
  stats: {},
  graphDirty: false,
};

// ─── 3d-force-graph Setup ───────────────────────────────────
const container = document.getElementById('graph-container');

const Graph = ForceGraph3D()(container)
  .backgroundColor('#0a0a1a')
  .nodeColor(node => NODE_COLORS[node.type] || '#cccccc')
  .nodeVal(node => Math.max(1, Math.log2((node.activityCount || 1) + 1) * 2))
  .nodeOpacity(node => {
    if (state.filterText) {
      const filter = state.filterText.toLowerCase();
      const match = node.id.toLowerCase().includes(filter) ||
                    node.label.toLowerCase().includes(filter) ||
                    node.type.toLowerCase().includes(filter);
      return match ? 1.0 : 0.08;
    }
    return node.stale ? 0.3 : 1.0;
  })
  .nodeLabel(node => `${node.type}: ${node.label}`)
  .linkColor(link => EDGE_COLORS[link.type] || '#666666')
  .linkOpacity(link => Math.max(0.05, (link.opacity !== undefined ? link.opacity : 1.0) * 0.8))
  .linkDirectionalParticles(2)
  .linkDirectionalParticleWidth(0.8)
  .linkDirectionalParticleSpeed(0.006)
  .linkWidth(link => Math.max(0.3, (link.callCount || 1) * 0.3))
  .linkLabel(link => `${link.type} (${link.callCount || 1}x)`)
  .onNodeClick(node => {
    state.pinned = { type: 'node', id: node.id };
    showDetail(state.pinned);
  })
  .onBackgroundClick(() => {
    state.pinned = null;
    hideDetail();
  });

// Add starfield via underlying Three.js scene
const THREE = Graph.scene().constructor.__proto__.constructor;
(function createStarfield() {
  const scene = Graph.scene();
  const geo = new window.THREE.BufferGeometry();
  const count = 2000;
  const positions = new Float32Array(count * 3);
  for (let i = 0; i < count * 3; i++) {
    positions[i] = (Math.random() - 0.5) * 2000;
  }
  geo.setAttribute('position', new window.THREE.BufferAttribute(positions, 3));
  const mat = new window.THREE.PointsMaterial({ color: 0x444466, size: 0.5, sizeAttenuation: true });
  scene.add(new window.THREE.Points(geo, mat));
})();

// ─── Token Animation System ─────────────────────────────────
const TOKEN_ANIM = {
  DURATION: 2.5,         // total sprite lifetime (seconds)
  BOUNCE_END: 0.3,       // end of bounce-in phase
  FADE_START: 2.0,       // start of fade-out phase
  DRIFT_SPEED: 8,        // units/sec upward drift
  BOUNCE_OVERSHOOT: 2,   // units overshoot during bounce-in
  H_OFFSET: 2,           // horizontal offset for input/output pair
  V_STAGGER: 3,          // vertical stagger between stacked sprites
  MAX_PER_NODE: 3,       // max sprites per node before coalescing
  MAX_GLOBAL: 50,        // global sprite cap
  FONT_SIZE: 3,
  INPUT_COLOR: '#4fc3f7',  // cyan (agent color — input flows from agent)
  OUTPUT_COLOR: '#ffd54f', // amber (llm color — output flows from model)
};

// Sprite tracking
const spritesByNode = new Map(); // nodeId → [{sprite, spawnTime, baseY, isOutput}]
const allSprites = [];           // global flat list for cap enforcement

function formatTokenCount(n) {
  return n.toLocaleString('en-US');
}

function spawnTokenSprite(nodeId, count, isOutput) {
  if (!count || count <= 0) return;

  const nodeSprites = spritesByNode.get(nodeId) || [];

  // Coalescing: if >= MAX_PER_NODE active, update most recent instead of spawning
  if (nodeSprites.length >= TOKEN_ANIM.MAX_PER_NODE) {
    const recent = nodeSprites[nodeSprites.length - 1];
    recent.sprite.text = isOutput ? '+' + formatTokenCount(count) : formatTokenCount(count);
    recent.sprite.color = isOutput ? TOKEN_ANIM.OUTPUT_COLOR : TOKEN_ANIM.INPUT_COLOR;
    recent.spawnTime = performance.now() / 1000;
    return;
  }

  // Global cap: evict oldest if at limit
  if (allSprites.length >= TOKEN_ANIM.MAX_GLOBAL) {
    removeSprite(allSprites[0]);
  }

  // Find the destination node's current position in the graph
  const graphData = Graph.graphData();
  const node = graphData.nodes.find(n => n.id === nodeId);
  if (!node) return;

  const text = isOutput ? '+' + formatTokenCount(count) : formatTokenCount(count);
  const sprite = new SpriteText(text, TOKEN_ANIM.FONT_SIZE);
  sprite.color = isOutput ? TOKEN_ANIM.OUTPUT_COLOR : TOKEN_ANIM.INPUT_COLOR;
  sprite.fontFace = 'monospace';
  sprite.material.depthWrite = false;
  sprite.material.transparent = true;

  // Position: at node, offset horizontally for input/output
  const hOff = isOutput ? TOKEN_ANIM.H_OFFSET : -TOKEN_ANIM.H_OFFSET;

  // Vertical stagger: offset above highest active sprite on this node
  let baseY = node.y || 0;
  if (nodeSprites.length > 0) {
    const highest = nodeSprites.reduce((max, s) => Math.max(max, s.baseY), baseY);
    baseY = highest + TOKEN_ANIM.V_STAGGER;
  } else {
    baseY += 5; // small offset above node center
  }

  sprite.position.set(
    (node.x || 0) + hOff,
    baseY,
    node.z || 0
  );

  // Initial state for bounce-in: small and transparent
  sprite.scale.setScalar(0.5);
  sprite.material.opacity = 0;

  const now = performance.now() / 1000;
  const entry = { sprite, spawnTime: now, baseY, isOutput, nodeId };

  nodeSprites.push(entry);
  spritesByNode.set(nodeId, nodeSprites);
  allSprites.push(entry);

  Graph.scene().add(sprite);
}

function removeSprite(entry) {
  Graph.scene().remove(entry.sprite);
  entry.sprite.material.dispose();

  // Remove from per-node list
  const nodeSprites = spritesByNode.get(entry.nodeId);
  if (nodeSprites) {
    const idx = nodeSprites.indexOf(entry);
    if (idx >= 0) nodeSprites.splice(idx, 1);
    if (nodeSprites.length === 0) spritesByNode.delete(entry.nodeId);
  }

  // Remove from global list
  const gIdx = allSprites.indexOf(entry);
  if (gIdx >= 0) allSprites.splice(gIdx, 1);
}

// Elastic ease-out: overshoots then settles
function elasticOut(t) {
  if (t <= 0) return 0;
  if (t >= 1) return 1;
  return Math.pow(2, -10 * t) * Math.sin((t - 0.075) * (2 * Math.PI) / 0.3) + 1;
}

function animateTokenSprites() {
  requestAnimationFrame(animateTokenSprites);

  const now = performance.now() / 1000;
  // Iterate backwards for safe removal
  for (let i = allSprites.length - 1; i >= 0; i--) {
    const entry = allSprites[i];
    const elapsed = now - entry.spawnTime;

    if (elapsed >= TOKEN_ANIM.DURATION) {
      removeSprite(entry);
      continue;
    }

    const sprite = entry.sprite;

    if (elapsed < TOKEN_ANIM.BOUNCE_END) {
      // Phase 1: bounce-in (0–0.3s)
      const t = elapsed / TOKEN_ANIM.BOUNCE_END; // 0→1
      const eased = elasticOut(t);

      sprite.material.opacity = Math.min(1, t * 3); // ramp opacity fast
      const scale = 0.5 + eased * 0.7; // 0.5 → ~1.2 → 1.0 (elastic settles)
      sprite.scale.setScalar(scale);

      // Y: overshoot up then settle
      const yOffset = TOKEN_ANIM.BOUNCE_OVERSHOOT * (eased > 1 ? eased : eased);
      sprite.position.y = entry.baseY + yOffset;

    } else if (elapsed < TOKEN_ANIM.FADE_START) {
      // Phase 2: float up (0.3–2.0s)
      sprite.material.opacity = 1.0;
      sprite.scale.setScalar(1.0);
      const driftTime = elapsed - TOKEN_ANIM.BOUNCE_END;
      sprite.position.y = entry.baseY + TOKEN_ANIM.BOUNCE_OVERSHOOT + driftTime * TOKEN_ANIM.DRIFT_SPEED;

    } else {
      // Phase 3: fade out (2.0–2.5s)
      const fadeT = (elapsed - TOKEN_ANIM.FADE_START) / (TOKEN_ANIM.DURATION - TOKEN_ANIM.FADE_START);
      sprite.material.opacity = 1.0 - fadeT;
      sprite.scale.setScalar(1.0);
      const driftTime = elapsed - TOKEN_ANIM.BOUNCE_END;
      sprite.position.y = entry.baseY + TOKEN_ANIM.BOUNCE_OVERSHOOT + driftTime * TOKEN_ANIM.DRIFT_SPEED;
    }
  }
}

// Start the animation loop
animateTokenSprites();

// ─── Graph Data Sync ────────────────────────────────────────
function syncGraphData() {
  if (!state.graphDirty) return;
  state.graphDirty = false;

  const nodes = Array.from(state.nodes.values());
  const links = Array.from(state.edges.values()).map(e => ({
    source: e.src,
    target: e.dst,
    ...e,
  }));

  Graph.graphData({ nodes, links });
}

// Sync on a timer to batch updates
setInterval(syncGraphData, 200);

// ─── Node/Edge Management ───────────────────────────────────
function addOrUpdateNode(data) {
  const existing = state.nodes.get(data.id);
  if (existing) {
    Object.assign(existing, data);
  } else {
    state.nodes.set(data.id, { ...data });
  }
  state.graphDirty = true;
}

function addOrUpdateEdge(data) {
  const key = data.src + '|' + data.dst + '|' + data.type;
  const existing = state.edges.get(key);
  if (existing) {
    Object.assign(existing, data);
  } else {
    state.edges.set(key, { ...data, id: key });
  }
  state.graphDirty = true;
}

// ─── Detail Card ────────────────────────────────────────────
const detailCard = document.getElementById('detail-card');
const detailContent = document.getElementById('detail-content');

function showDetail(obj) {
  detailCard.style.display = 'block';
  if (obj.type === 'node') {
    const d = state.nodes.get(obj.id);
    if (!d) return;
    let html = `<div class="detail-type">${d.type.toUpperCase()}: ${escapeHtml(d.label)}</div>`;
    html += row('ID', d.id);
    html += row('Type', d.type);
    html += row('Activity', d.activityCount);
    html += row('Last Seen', d.lastSeen || '—');
    html += row('Stale', d.stale ? 'Yes' : 'No');
    if (state.showRaw && d.attributes) {
      html += '<div style="margin-top:8px;color:#888">Attributes:</div>';
      for (const [k, v] of Object.entries(d.attributes)) {
        html += row(k, JSON.stringify(v));
      }
    }
    detailContent.innerHTML = html;
  }
}

function hideDetail() {
  detailCard.style.display = 'none';
}

function row(key, value) {
  return `<div class="detail-row"><span class="detail-key">${escapeHtml(key)}</span><span class="detail-value">${escapeHtml(String(value))}</span></div>`;
}

function escapeHtml(s) {
  return s.replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;');
}

// ─── HUD Updates ────────────────────────────────────────────
function updateHUD() {
  document.getElementById('hud-nodes').textContent = state.nodes.size;
  document.getElementById('hud-edges').textContent = state.edges.size;

  if (state.stats.eventsPerSec !== undefined) {
    document.getElementById('hud-eps').textContent = state.stats.eventsPerSec.toFixed(1);
  }
  if (state.stats.errorCount !== undefined) {
    document.getElementById('hud-errors').textContent = state.stats.errorCount;
  }
  if (state.stats.avgLatencyMs !== undefined) {
    document.getElementById('hud-latency').textContent = state.stats.avgLatencyMs.toFixed(1) + 'ms';
  }

  const statusEl = document.getElementById('hud-status');
  if (state.paused) {
    statusEl.textContent = 'paused';
    statusEl.style.color = '#f7768e';
  } else if (state.ws && state.ws.readyState === WebSocket.OPEN) {
    statusEl.textContent = state.stats.mode || 'connected';
    statusEl.style.color = '#81c784';
  } else {
    statusEl.textContent = 'disconnected';
    statusEl.style.color = '#f7768e';
  }

  const cappedRow = document.getElementById('hud-capped-row');
  if (state.stats.capped) {
    cappedRow.style.display = 'flex';
  }

  const droppedRow = document.getElementById('hud-dropped-row');
  if (state.stats.dropped > 0) {
    droppedRow.style.display = 'flex';
    document.getElementById('hud-dropped').textContent = state.stats.dropped;
  }

  const samplingRow = document.getElementById('hud-sampling-row');
  samplingRow.style.display = state.stats.sampling ? 'flex' : 'none';
}

setInterval(updateHUD, 500);

// ─── WebSocket ──────────────────────────────────────────────
function connectWS() {
  const proto = location.protocol === 'https:' ? 'wss:' : 'ws:';
  const ws = new WebSocket(`${proto}//${location.host}/ws`);
  state.ws = ws;

  ws.onmessage = (event) => {
    const msg = JSON.parse(event.data);
    if (state.paused) {
      state.eventBuffer.push(msg);
      return;
    }
    handleMessage(msg);
  };

  ws.onclose = () => {
    setTimeout(connectWS, 2000);
  };

  ws.onerror = () => {
    ws.close();
  };
}

function handleMessage(msg) {
  switch (msg.type) {
    case 'snapshot':
      handleSnapshot(msg.data);
      break;
    case 'update':
      handleUpdate(msg.data);
      break;
    case 'stats':
      state.stats = msg.data;
      break;
  }
}

function handleSnapshot(data) {
  if (data.nodes) {
    for (const n of data.nodes) {
      addOrUpdateNode(n);
    }
  }
  if (data.edges) {
    for (const e of data.edges) {
      addOrUpdateEdge(e);
    }
  }
}

function handleUpdate(data) {
  if (data.nodes) {
    for (const n of data.nodes) {
      addOrUpdateNode(n);
    }
  }
  if (data.edges) {
    for (const e of data.edges) {
      addOrUpdateEdge(e);
      // Spawn floating token animations for model_call edges
      if (e.type === 'model_call' && e.attributes) {
        const inTok = e.attributes.input_tokens;
        const outTok = e.attributes.output_tokens;
        if (inTok && inTok > 0) spawnTokenSprite(e.dst, inTok, false);
        if (outTok && outTok > 0) spawnTokenSprite(e.dst, outTok, true);
      }
    }
  }
}

// ─── Controls ───────────────────────────────────────────────
document.getElementById('btn-pause').addEventListener('click', function() {
  state.paused = !state.paused;
  this.textContent = state.paused ? 'Resume' : 'Pause';
  if (!state.paused) {
    for (const msg of state.eventBuffer) {
      handleMessage(msg);
    }
    state.eventBuffer = [];
  }
});

document.getElementById('btn-reset').addEventListener('click', () => {
  Graph.cameraPosition({ x: 0, y: 0, z: 300 });
});

document.getElementById('search').addEventListener('input', (e) => {
  state.filterText = e.target.value;
  // Force re-render of node opacities
  Graph.nodeColor(Graph.nodeColor());
});

document.getElementById('chk-raw').addEventListener('change', (e) => {
  state.showRaw = e.target.checked;
  if (state.pinned) showDetail(state.pinned);
});

document.getElementById('detail-close').addEventListener('click', () => {
  state.pinned = null;
  hideDetail();
});

window.addEventListener('keydown', (e) => {
  if (e.key === 'Escape') {
    state.pinned = null;
    hideDetail();
  }
});

// ─── Init ───────────────────────────────────────────────────
connectWS();
