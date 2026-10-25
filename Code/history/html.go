package history

const HistoryViewerHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Match History Viewer</title>
  <style>
    * { box-sizing: border-box; margin: 0; padding: 0; user-select: none; }
    body { background-color: #0b1418; display: flex; height: 100vh; overflow: hidden; font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, monospace; color: #fff; }
    
    .sidebar { width: 320px; background-color: #141f23; border-right: 2px solid #23343c; display: flex; flex-direction: column; }
    .sidebar-header { padding: 16px; border-bottom: 1px solid #23343c; display: flex; align-items: center; justify-content: space-between; }
    .sidebar-header h2 { font-size: 16px; color: #58cc02; font-weight: 800; display: flex; align-items: center; gap: 8px; }
    .back-btn { text-decoration: none; background: #23343c; color: #93a1a7; padding: 4px 10px; border-radius: 6px; font-size: 12px; font-weight: bold; transition: all 0.2s; }
    .back-btn:hover { background: #58cc02; color: #0b1418; }
    .match-list { flex: 1; overflow-y: auto; padding: 10px; display: flex; flex-direction: column; gap: 8px; }
    .date-header { font-size: 12px; font-weight: bold; color: #58cc02; padding: 6px 4px 2px 4px; text-transform: uppercase; letter-spacing: 0.5px; }
    .match-item { background: #1a272c; border: 1px solid #23343c; border-radius: 8px; padding: 10px; cursor: pointer; transition: all 0.2s; }
    .match-item:hover { border-color: #58cc02; background: #203138; }
    .match-item.active { border-color: #58cc02; background: #1c3629; }
    .match-title { font-size: 13px; font-weight: bold; color: #fff; margin-bottom: 4px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
    .match-meta { font-size: 11px; color: #93a1a7; display: flex; justify-content: space-between; }
    .match-result { color: #ffc800; }
    
    .main-view { flex: 1; display: flex; flex-direction: column; align-items: center; justify-content: center; position: relative; padding: 16px; }
    .empty-state { color: #637580; font-size: 16px; font-weight: bold; text-align: center; }
    
    .viewer-container { display: none; flex-direction: column; align-items: center; width: 100%; max-width: min(80vw, 80vh); }
    .match-info-banner { width: 100%; margin-bottom: 12px; display: flex; align-items: center; justify-content: space-between; }
    .match-main-title { font-size: 18px; font-weight: 800; color: #fff; }
    .match-main-result { font-size: 13px; background: #141f23; border: 1px solid #23343c; padding: 4px 10px; border-radius: 8px; color: #ffc800; font-weight: bold; }
    
    .board-container { width: min(72vw, 72vh); height: min(72vw, 72vh); position: relative; background-color: #141f23; border: 4px solid #1a272c; border-radius: 12px; overflow: hidden; box-shadow: 0 16px 40px rgba(0,0,0,0.7); }
    .board-grid { width: 100%; height: 100%; display: grid; grid-template-columns: repeat(8, 1fr); grid-template-rows: repeat(8, 1fr); }
    .cell { position: relative; width: 100%; height: 100%; }
    .cell.light { background-color: #273941; }
    .cell.dark  { background-color: #17242a; }
    .cell.highlight-from { background-color: #1a4d2e !important; box-shadow: inset 0 0 0 3px #58cc02; }
    .cell.highlight-to { background-color: #4a3e1a !important; box-shadow: inset 0 0 0 3px #ffc800; }
    .pieces-layer { position: absolute; top: 0; left: 0; width: 100%; height: 100%; pointer-events: none; }
    .pieces-layer.no-transition .piece-node { transition: none !important; }
    .piece-node {
      position: absolute;
      width: 12.5%;
      height: 12.5%;
      top: 0;
      left: 0;
      display: flex;
      align-items: center;
      justify-content: center;
      z-index: 10;
      transition: transform 0.22s cubic-bezier(0.2, 0, 0, 1), opacity 0.18s ease;
      will-change: transform;
    }
    .piece-node img { width: 88%; height: 88%; object-fit: contain; filter: drop-shadow(0 3px 6px rgba(0,0,0,0.45)); pointer-events: none; }
    
    .controls { width: min(72vw, 72vh); margin-top: 14px; display: flex; flex-direction: column; gap: 10px; }
    .slider-row { display: flex; align-items: center; gap: 12px; width: 100%; }
    .timeline-slider {
      flex: 1;
      width: 100%;
      height: 12px;
      -webkit-appearance: none;
      appearance: none;
      background: #e5e5e5;
      border-radius: 999px;
      outline: none;
      cursor: pointer;
      border: none;
      position: relative;
    }
    .timeline-slider::-webkit-slider-runnable-track {
      width: 100%;
      height: 12px;
      border-radius: 999px;
      background: transparent;
      border: none;
    }
    .timeline-slider::-moz-range-track {
      width: 100%;
      height: 12px;
      border-radius: 999px;
      background: #e5e5e5;
      border: none;
    }
    .timeline-slider::-moz-range-progress {
      background-color: #58cc02;
      height: 12px;
      border-radius: 999px;
    }
    .timeline-slider::-webkit-slider-thumb {
      -webkit-appearance: none;
      appearance: none;
      width: 24px;
      height: 24px;
      border-radius: 50%;
      background: #ffffff;
      border: 2px solid #58cc02;
      border-bottom: 4px solid #58a700;
      box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
      cursor: pointer;
      margin-top: -6px;
      transition: transform 0.1s cubic-bezier(0.2, 0, 0, 1), border-bottom-width 0.1s ease, box-shadow 0.1s ease;
    }
    .timeline-slider::-webkit-slider-thumb:hover {
      transform: scale(1.12);
      background: #f7fff2;
    }
    .timeline-slider::-webkit-slider-thumb:active,
    .timeline-slider:active::-webkit-slider-thumb {
      transform: scale(0.96) translateY(2px);
      border-bottom: 2px solid #58a700;
      box-shadow: 0 1px 2px rgba(0, 0, 0, 0.2);
    }
    .timeline-slider::-moz-range-thumb {
      width: 24px;
      height: 24px;
      border-radius: 50%;
      background: #ffffff;
      border: 2px solid #58cc02;
      border-bottom: 4px solid #58a700;
      box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
      cursor: pointer;
      transition: transform 0.1s cubic-bezier(0.2, 0, 0, 1), border-bottom-width 0.1s ease, box-shadow 0.1s ease;
    }
    .timeline-slider::-moz-range-thumb:hover {
      transform: scale(1.12);
      background: #f7fff2;
    }
    .timeline-slider::-moz-range-thumb:active,
    .timeline-slider:active::-moz-range-thumb {
      transform: scale(0.96) translateY(2px);
      border-bottom: 2px solid #58a700;
      box-shadow: 0 1px 2px rgba(0, 0, 0, 0.2);
    }
    .btn-row { display: flex; justify-content: center; align-items: center; gap: 10px; }
    .ctrl-btn {
      background: #141f23;
      border: 1px solid #23343c;
      color: #fff;
      padding: 8px 16px;
      border-radius: 10px;
      cursor: pointer;
      font-size: 13px;
      font-weight: bold;
      transition: all 0.2s ease;
      display: inline-flex;
      align-items: center;
      gap: 6px;
      box-shadow: 0 4px 10px rgba(0,0,0,0.25);
    }
    .ctrl-btn:hover {
      background: #58cc02;
      color: #0b1418;
      border-color: #58cc02;
      transform: translateY(-1px);
    }
    .ctrl-btn:active {
      transform: translateY(0);
    }
    .frame-counter { font-size: 13px; color: #93a1a7; font-weight: bold; min-width: 80px; text-align: center; }
    
    .status-bar { margin-top: 8px; display: flex; justify-content: center; gap: 12px; font-size: 13px; font-weight: bold; color: #93a1a7; }
    .badge { background-color: #141f23; padding: 6px 12px; border-radius: 8px; border: 1px solid #23343c; }
    .active-val { color: #58cc02; }
    .active-move { color: #ffc800; }
  </style>
</head>
<body>
<div class="sidebar">
  <div class="sidebar-header">
    <h2>📜 Matches</h2>
    <a href="/" class="back-btn">⬅ Live</a>
  </div>
  <div class="match-list" id="matchList"></div>
</div>

<div class="main-view">
  <div class="empty-state" id="emptyState">Select a match to replay</div>
  <div class="viewer-container" id="viewerBox">
    <div class="match-info-banner">
      <div class="match-main-title" id="mTitle">Match</div>
      <div class="match-main-result" id="mResult">Result</div>
    </div>
    <div class="board-container">
      <div class="board-grid" id="gridBg"></div>
      <div class="pieces-layer" id="piecesLayer"></div>
    </div>
    <div class="controls">
      <div class="slider-row">
        <input type="range" class="timeline-slider" id="frameSlider" min="0" max="0" value="0">
      </div>
      <div class="btn-row">
        <button class="ctrl-btn" id="btnPrev">◀ Prev</button>
        <button class="ctrl-btn" id="btnPlay">▶ Play</button>
        <button class="ctrl-btn" id="btnNext">Next ▶</button>
        <span class="frame-counter" id="frameCount">0 / 0</span>
      </div>
      <div class="status-bar">
        <div class="badge">TURN: <span id="uiMove" class="active-val">-</span></div>
        <div class="badge">YOU: <span id="uiMyTeam" class="active-val">-</span></div>
        <div class="badge">MOVE: <span id="uiBestMove" class="active-move">-</span></div>
      </div>
    </div>
  </div>
</div>

<script>
  const gridBg = document.getElementById('gridBg');
  const piecesLayer = document.getElementById('piecesLayer');
  const frameSlider = document.getElementById('frameSlider');
  const frameCount = document.getElementById('frameCount');
  const btnPrev = document.getElementById('btnPrev');
  const btnPlay = document.getElementById('btnPlay');
  const btnNext = document.getElementById('btnNext');
  const uiMove = document.getElementById('uiMove');
  const uiMyTeam = document.getElementById('uiMyTeam');
  const uiBestMove = document.getElementById('uiBestMove');
  const mTitle = document.getElementById('mTitle');
  const mResult = document.getElementById('mResult');
  const emptyState = document.getElementById('emptyState');
  const viewerBox = document.getElementById('viewerBox');
  const matchList = document.getElementById('matchList');

  let currentMatch = null;
  let currentIdx = 0;
  let playInterval = null;
  let activePieces = [];
  let nextPieceId = 1;

  function updateSliderFill() {
    const min = parseFloat(frameSlider.min) || 0;
    const max = parseFloat(frameSlider.max) || 0;
    const val = parseFloat(frameSlider.value) || 0;
    const pct = (max > min) ? ((val - min) / (max - min)) * 100 : 0;
    frameSlider.style.background = 'linear-gradient(to right, #58cc02 0%, #58cc02 ' + pct + '%, #e5e5e5 ' + pct + '%, #e5e5e5 100%)';
  }

  function initGrid() {
    gridBg.innerHTML = '';
    for (let r = 0; r < 8; r++) {
      for (let c = 0; c < 8; c++) {
        const cell = document.createElement('div');
        const isLight = (r + c) % 2 === 0;
        cell.className = 'cell ' + (isLight ? 'light' : 'dark');
        cell.id = 'sq_' + r + '_' + c;
        gridBg.appendChild(cell);
      }
    }
  }
  initGrid();

  function renderFrame(idx, isStepJump = false) {
    if (!currentMatch || !currentMatch.frames || currentMatch.frames.length === 0) return;
    if (idx < 0) idx = 0;
    if (idx >= currentMatch.frames.length) idx = currentMatch.frames.length - 1;
    const prevIdx = currentIdx;
    currentIdx = idx;
    frameSlider.value = idx;
    updateSliderFill();
    frameCount.textContent = (idx + 1) + ' / ' + currentMatch.frames.length;

    const frame = currentMatch.frames[idx];
    uiMove.textContent = frame.move || '-';
    uiMyTeam.textContent = frame.my_team || currentMatch.my_team || '-';
    uiBestMove.textContent = frame.best_move || frame.status_text || '-';

    // Highlights
    for (let r = 0; r < 8; r++) {
      for (let c = 0; c < 8; c++) {
        const cell = document.getElementById('sq_' + r + '_' + c);
        if (!cell) continue;
        const isFrom = (frame.from && frame.from[0] === r && frame.from[1] === c && (frame.from[0] !== -1 || frame.from[1] !== -1));
        const isTo = (frame.to && frame.to[0] === r && frame.to[1] === c && (frame.to[0] !== -1 || frame.to[1] !== -1));
        cell.className = 'cell ' + ((r + c) % 2 === 0 ? 'light' : 'dark');
        if (isFrom) cell.classList.add('highlight-from');
        if (isTo) cell.classList.add('highlight-to');
      }
    }

    // Pieces rendering with smooth transitions
    updatePieces(frame.board, isStepJump, Math.abs(currentIdx - prevIdx) > 1);
  }

  function updatePieces(matrix, isStepJump, isBigJump) {
    if (!matrix || matrix.length < 8) return;

    if (isBigJump) {
      piecesLayer.classList.add('no-transition');
    } else {
      piecesLayer.classList.remove('no-transition');
    }

    const targetItems = [];
    for (let r = 0; r < 8; r++) {
      for (let c = 0; c < 8; c++) {
        const p = matrix[r][c];
        if (p && p !== '.' && p !== '?' && p !== '--') {
          targetItems.push({ r, c, type: p, matched: false });
        }
      }
    }

    const nextPieces = [];

    // Pass 1: Exact matches (same piece type staying at same square)
    for (const item of targetItems) {
      const idx = activePieces.findIndex(ap => !ap.matched && ap.type === item.type && ap.r === item.r && ap.c === item.c);
      if (idx !== -1) {
        item.matched = true;
        const ap = activePieces[idx];
        ap.matched = true;
        ap.el.style.opacity = '1';
        nextPieces.push(ap);
      }
    }

    // Pass 2: Piece movements (same type moving to new square, prioritize shortest distance)
    for (const item of targetItems) {
      if (item.matched) continue;
      let bestIdx = -1;
      let minDistance = 999;
      for (let i = 0; i < activePieces.length; i++) {
        const ap = activePieces[i];
        if (!ap.matched && ap.type === item.type) {
          const dist = Math.abs(ap.r - item.r) + Math.abs(ap.c - item.c);
          if (dist < minDistance) {
            minDistance = dist;
            bestIdx = i;
          }
        }
      }

      if (bestIdx !== -1) {
        item.matched = true;
        const ap = activePieces[bestIdx];
        ap.matched = true;
        ap.r = item.r;
        ap.c = item.c;
        ap.el.style.opacity = '1';
        ap.el.style.transform = 'translate(' + (item.c * 100) + '%, ' + (item.r * 100) + '%)';
        nextPieces.push(ap);
      }
    }

    // Pass 3: Brand new pieces (spawns / promotions)
    for (const item of targetItems) {
      if (item.matched) continue;
      const el = document.createElement('div');
      el.className = 'piece-node';
      el.style.transform = 'translate(' + (item.c * 100) + '%, ' + (item.r * 100) + '%)';
      el.style.opacity = '1';

      const img = document.createElement('img');
      img.src = '/figures/' + item.type + '.png';
      img.alt = item.type;
      el.appendChild(img);

      piecesLayer.appendChild(el);

      nextPieces.push({
        id: nextPieceId++,
        type: item.type,
        r: item.r,
        c: item.c,
        el: el,
        matched: true
      });
    }

    // Pass 4: Pieces that were captured or removed
    for (const ap of activePieces) {
      if (!ap.matched) {
        if (ap.el && ap.el.parentNode) {
          ap.el.parentNode.removeChild(ap.el);
        }
      }
    }

    activePieces = nextPieces;
    for (const ap of activePieces) ap.matched = false;

    if (isBigJump) {
      // Re-enable transitions on next animation frame
      requestAnimationFrame(() => {
        piecesLayer.classList.remove('no-transition');
      });
    }
  }

  function loadMatch(path, elem) {
    if (playInterval) {
      clearInterval(playInterval);
      playInterval = null;
      btnPlay.textContent = '▶ Play';
    }
    document.querySelectorAll('.match-item').forEach(el => el.classList.remove('active'));
    if (elem) elem.classList.add('active');

    fetch('/api/history/load?file=' + encodeURIComponent(path))
      .then(res => res.json())
      .then(data => {
        currentMatch = data;
        emptyState.style.display = 'none';
        viewerBox.style.display = 'flex';
        mTitle.textContent = currentMatch.title || currentMatch.id;
        mResult.textContent = currentMatch.result || 'Finished';

        frameSlider.max = Math.max(0, (currentMatch.frames ? currentMatch.frames.length - 1 : 0));
        activePieces = [];
        piecesLayer.innerHTML = '';
        currentIdx = 0;
        renderFrame(0, false);
      })
      .catch(err => console.error(err));
  }

  function fetchMatches() {
    fetch('/api/history/list')
      .then(res => res.json())
      .then(groups => {
        matchList.innerHTML = '';
        if (!groups || groups.length === 0) {
          matchList.innerHTML = '<div style="color: #637580; font-size: 13px; text-align: center; margin-top: 20px;">No match history found</div>';
          return;
        }

        groups.forEach(g => {
          const dateTitle = document.createElement('div');
          dateTitle.className = 'date-header';
          dateTitle.textContent = g.date;
          matchList.appendChild(dateTitle);

          g.matches.forEach(m => {
            const item = document.createElement('div');
            item.className = 'match-item';
            item.innerHTML = '<div class="match-title">' + (m.title || m.id) + '</div>' +
              '<div class="match-meta">' +
              '<span>' + (m.frames || 0) + ' frames</span>' +
              '<span class="match-result">' + (m.result || '') + '</span>' +
              '</div>';
            item.onclick = () => loadMatch(m.path, item);
            matchList.appendChild(item);
          });
        });
      })
      .catch(err => console.error(err));
  }

  btnPrev.onclick = () => {
    if (playInterval) { clearInterval(playInterval); playInterval = null; btnPlay.textContent = '▶ Play'; }
    renderFrame(currentIdx - 1);
  };

  btnNext.onclick = () => {
    if (playInterval) { clearInterval(playInterval); playInterval = null; btnPlay.textContent = '▶ Play'; }
    renderFrame(currentIdx + 1);
  };

  frameSlider.oninput = (e) => {
    if (playInterval) { clearInterval(playInterval); playInterval = null; btnPlay.textContent = '▶ Play'; }
    renderFrame(parseInt(e.target.value, 10));
  };

  btnPlay.onclick = () => {
    if (playInterval) {
      clearInterval(playInterval);
      playInterval = null;
      btnPlay.textContent = '▶ Play';
    } else {
      btnPlay.textContent = '⏸ Pause';
      playInterval = setInterval(() => {
        if (!currentMatch || !currentMatch.frames || currentIdx >= currentMatch.frames.length - 1) {
          clearInterval(playInterval);
          playInterval = null;
          btnPlay.textContent = '▶ Play';
          return;
        }
        renderFrame(currentIdx + 1);
      }, 350);
    }
  };

  document.addEventListener('keydown', (e) => {
    if (!currentMatch) return;
    if (e.key === 'ArrowLeft') {
      if (playInterval) { clearInterval(playInterval); playInterval = null; btnPlay.textContent = '▶ Play'; }
      renderFrame(currentIdx - 1);
    } else if (e.key === 'ArrowRight') {
      if (playInterval) { clearInterval(playInterval); playInterval = null; btnPlay.textContent = '▶ Play'; }
      renderFrame(currentIdx + 1);
    } else if (e.key === ' ' || e.code === 'Space') {
      e.preventDefault();
      btnPlay.click();
    }
  });

  fetchMatches();
</script>
</body>
</html>
`
