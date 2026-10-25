package web

const DuolingoBoardHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Chess Match</title>
  <style>
    :root {
      --duo-green: #58cc02;
      --duo-green-border: #58a700;
      --duo-yellow: #ffc800;
      --duo-yellow-border: #e5a400;
      --duo-blue: #1cb0f6;
      --duo-blue-border: #1899d6;
      --duo-dark-bg: #131f24;
      --duo-card-bg: #1b272d;
      --duo-card-border: #2b3940;
      --duo-cell-light: #23343c;
      --duo-cell-dark: #17242a;
      --duo-text-main: #ffffff;
      --duo-text-muted: #8e9da4;
    }

    * {
      box-sizing: border-box;
      margin: 0;
      padding: 0;
      user-select: none;
      -webkit-font-smoothing: antialiased;
      -moz-osx-font-smoothing: grayscale;
    }

    body {
      background-color: var(--duo-dark-bg);
      color: var(--duo-text-main);
      font-family: "DIN Next Rounded LT Pro", "Nunito", -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
      display: flex;
      flex-direction: column;
      justify-content: center;
      align-items: center;
      min-height: 100vh;
      overflow: hidden;
    }

    .header-bar {
      display: flex;
      align-items: center;
      justify-content: flex-end;
      width: min(82vw, 82vh);
      margin-bottom: 16px;
    }

    .duo-btn {
      display: inline-flex;
      align-items: center;
      justify-content: center;
      text-decoration: none;
      font-weight: 800;
      font-size: 13px;
      letter-spacing: 0.5px;
      text-transform: uppercase;
      border-radius: 14px;
      padding: 10px 20px;
      cursor: pointer;
      color: #ffffff;
      background-color: #58cc02;
      border: none;
      box-shadow: 0 4px 0 #58a700;
      transition: transform 0.12s cubic-bezier(0.2, 0, 0, 1), box-shadow 0.12s cubic-bezier(0.2, 0, 0, 1), filter 0.12s ease;
      transform: translateZ(0);
    }
    .duo-btn:hover {
      filter: brightness(1.08);
      background-color: #62de02;
    }
    .duo-btn:active {
      transform: translateY(3px) translateZ(0);
      box-shadow: 0 1px 0 #58a700;
    }

    .board-container {
      width: min(82vw, 82vh);
      height: min(82vw, 82vh);
      position: relative;
      background-color: var(--duo-card-bg);
      border: 4px solid var(--duo-card-border);
      border-bottom-width: 6px;
      border-radius: 20px;
      overflow: hidden;
      box-shadow: 0 18px 48px rgba(0,0,0,0.65);
    }

    .board-grid {
      width: 100%;
      height: 100%;
      display: grid;
      grid-template-columns: repeat(8, 1fr);
      grid-template-rows: repeat(8, 1fr);
    }

    .cell { position: relative; width: 100%; height: 100%; }
    .cell.light { background-color: var(--duo-cell-light); }
    .cell.dark  { background-color: var(--duo-cell-dark); }
    .cell.highlight-from { background-color: #1a4d2e !important; box-shadow: inset 0 0 0 3px var(--duo-green); }
    .cell.highlight-to { background-color: #4a3e1a !important; box-shadow: inset 0 0 0 3px var(--duo-yellow); }

    .pieces-layer { position: absolute; top: 0; left: 0; width: 100%; height: 100%; pointer-events: none; }

    .piece-node {
      position: absolute;
      width: 12.5%;
      height: 12.5%;
      top: 0;
      left: 0;
      display: flex;
      align-items: center;
      justify-content: center;
      transition: transform 0.22s cubic-bezier(0.2, 0, 0, 1), opacity 0.18s ease;
      z-index: 10;
      will-change: transform;
    }

    .piece-node img {
      width: 88%;
      height: 88%;
      object-fit: contain;
      filter: drop-shadow(0 4px 6px rgba(0,0,0,0.45));
      pointer-events: none;
    }

    @keyframes fearfulShiver {
      0%   { transform: translate(0, 0) rotate(0deg); }
      15%  { transform: translate(-1.2px, 0.4px) rotate(-0.6deg); }
      30%  { transform: translate(1px, -0.6px) rotate(0.5deg); }
      45%  { transform: translate(-0.8px, -0.4px) rotate(-0.4deg); }
      60%  { transform: translate(1.1px, 0.5px) rotate(0.5deg); }
      75%  { transform: translate(-0.6px, 0.6px) rotate(-0.3deg); }
      90%  { transform: translate(0.7px, -0.3px) rotate(0.4deg); }
      100% { transform: translate(0, 0) rotate(0deg); }
    }

    .piece-node.king-in-check img {
      animation: fearfulShiver 0.55s infinite ease-in-out;
      transform-origin: 50% 80%;
      will-change: transform;
    }

    .info-bar {
      margin-top: 16px;
      display: flex;
      flex-wrap: wrap;
      justify-content: center;
      gap: 14px;
      width: min(82vw, 82vh);
    }

    .duo-badge {
      background-color: var(--duo-card-bg);
      padding: 9px 18px;
      border-radius: 14px;
      border: 2px solid var(--duo-card-border);
      border-bottom: 4px solid #18242a;
      box-shadow: 0 3px 0 #18242a;
      font-size: 13px;
      font-weight: 800;
      letter-spacing: 0.5px;
      color: var(--duo-text-muted);
      display: inline-flex;
      align-items: center;
      gap: 8px;
    }

    .duo-badge-val {
      color: #ffffff;
      font-weight: 800;
    }
    .duo-badge-val.green { color: var(--duo-green); }
    .duo-badge-val.yellow { color: var(--duo-yellow); }
    .duo-badge-val.blue { color: var(--duo-blue); }
  </style>
</head>
<body>
<div class="header-bar">
  <a href="/history" class="duo-btn">Match History</a>
</div>
<div class="board-container" id="boardBox">
  <div class="board-grid" id="gridBg"></div>
  <div class="pieces-layer" id="piecesLayer"></div>
</div>
<div class="info-bar">
  <div class="duo-badge">Turn: <span id="uiMove" class="duo-badge-val green">White</span></div>
  <div class="duo-badge">Playing as: <span id="uiMyTeam" class="duo-badge-val green">Detecting...</span></div>
</div>
<script>
  const gridBg = document.getElementById('gridBg');
  const piecesLayer = document.getElementById('piecesLayer');
  const uiMove = document.getElementById('uiMove');
  const uiMyTeam = document.getElementById('uiMyTeam');

  const PIECE_TYPES = ['wP','wN','wB','wR','wQ','wK','wK2','bP','bN','bB','bR','bQ','bK','bK2'];
  for (const pt of PIECE_TYPES) {
    const img = new Image();
    img.src = '/figures/' + pt + '.png';
  }

  function formatTeam(team) {
    if (!team) return '—';
    const lower = String(team).trim().toLowerCase();
    if (lower === 'white' || lower.startsWith('w') || lower.startsWith('бел')) return 'White';
    if (lower === 'black' || lower.startsWith('b') || lower.startsWith('чер') || lower.startsWith('чёр')) return 'Black';
    return team;
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

  let activePieces = [];
  let nextPieceId = 1;
  let currentState = 1;
  let currentMyTeam = '';
  let isMyCheckActive = false;
  let isOppCheckActive = false;

  function applyCheckAnimation() {
    const myTeamLower = String(currentMyTeam).toLowerCase();
    const myIsWhite = myTeamLower.startsWith('w') || myTeamLower.startsWith('бел');
    const myIsBlack = myTeamLower.startsWith('b') || myTeamLower.startsWith('чер') || myTeamLower.startsWith('чёр');

    for (const ap of activePieces) {
      const isKing = (ap.type === 'wK' || ap.type === 'bK');
      if (!isKing) continue;

      const isMyKing = (myIsWhite && ap.type === 'wK') || (myIsBlack && ap.type === 'bK');
      const isOppKing = (myIsWhite && ap.type === 'bK') || (myIsBlack && ap.type === 'wK');

      let inDanger = false;
      if (isMyKing && isMyCheckActive) inDanger = true;
      if (isOppKing && isOppCheckActive) inDanger = true;

      // Fallback if team is not yet detected: if check active, highlight matching king
      if (!myIsWhite && !myIsBlack && (isMyCheckActive || isOppCheckActive)) {
        inDanger = true;
      }

      const img = ap.el.querySelector('img');
      if (inDanger) {
        if (!ap.el.classList.contains('king-in-check')) {
          ap.el.classList.add('king-in-check');
        }
        const crackedSrc = '/figures/' + ap.type + '2.png';
        if (img && !img.src.endsWith(ap.type + '2.png')) {
          img.src = crackedSrc;
        }
      } else {
        ap.el.classList.remove('king-in-check');
        const normalSrc = '/figures/' + ap.type + '.png';
        if (img && !img.src.endsWith(ap.type + '.png')) {
          img.src = normalSrc;
        }
      }
    }
  }

  function updateBoard(matrix, state) {
    if (!matrix || matrix.length < 8) return;
    if (state !== undefined) currentState = state;

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

    // Pass 1: Exact matches (same piece type at same square)
    for (const item of targetItems) {
      const idx = activePieces.findIndex(ap => !ap.matched && ap.type === item.type && ap.r === item.r && ap.c === item.c);
      if (idx !== -1) {
        item.matched = true;
        const ap = activePieces[idx];
        ap.matched = true;
        ap.missingCount = 0;
        ap.el.style.opacity = '1';
        nextPieces.push(ap);
      }
    }

    // Pass 2: Piece movements (same type moving to new square)
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
        ap.missingCount = 0;
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
        missingCount: 0,
        matched: true
      });
    }

    // Pass 4: Pieces no longer matched on board
    const maxMissingThreshold = (currentState === 0) ? 35 : 18;

    for (const ap of activePieces) {
      if (!ap.matched) {
        const isOccupiedByAnother = nextPieces.some(np => np.r === ap.r && np.c === ap.c && np.id !== ap.id);
        ap.missingCount = (ap.missingCount || 0) + 1;

        if (isOccupiedByAnother || ap.missingCount >= maxMissingThreshold) {
          if (ap.el && ap.el.parentNode) {
            ap.el.parentNode.removeChild(ap.el);
          }
        } else {
          ap.matched = true;
          ap.el.style.opacity = (currentState === 0) ? '0.9' : '0.75';
          nextPieces.push(ap);
        }
      }
    }

    activePieces = nextPieces;
    for (const ap of activePieces) ap.matched = false;
    applyCheckAnimation();
  }

  function updateHighlights(from, to) {
    for (let r = 0; r < 8; r++) {
      for (let c = 0; c < 8; c++) {
        const cell = document.getElementById('sq_' + r + '_' + c);
        if (!cell) continue;
        const isFrom = (from && from[0] === r && from[1] === c && (from[0] !== -1 || from[1] !== -1));
        const isTo = (to && to[0] === r && to[1] === c && (to[0] !== -1 || to[1] !== -1));
        
        if (isFrom) {
          if (!cell.classList.contains('highlight-from')) cell.classList.add('highlight-from');
        } else {
          cell.classList.remove('highlight-from');
        }

        if (isTo) {
          if (!cell.classList.contains('highlight-to')) cell.classList.add('highlight-to');
        } else {
          cell.classList.remove('highlight-to');
        }
      }
    }
  }

  function connect() {
    const wsProto = location.protocol === 'https:' ? 'wss:' : 'ws:';
    const ws = new WebSocket(wsProto + '//' + location.host + '/ws');

    ws.onmessage = (evt) => {
      try {
        const msg = JSON.parse(evt.data);
        if (msg.state !== undefined) {
          currentState = msg.state;
        }
        if (msg.board) {
          updateBoard(msg.board, msg.state);
          updateHighlights(msg.from, msg.to);
        }
        if (msg.move !== undefined) uiMove.textContent = formatTeam(msg.move);
        if (msg.my_team !== undefined) {
          currentMyTeam = msg.my_team;
          uiMyTeam.textContent = formatTeam(msg.my_team);
        }
        if (msg.is_check !== undefined || msg.is_my_check !== undefined || msg.is_opp_check !== undefined) {
          isMyCheckActive = !!(msg.is_my_check !== undefined ? msg.is_my_check : msg.is_check);
          isOppCheckActive = !!msg.is_opp_check;
          applyCheckAnimation();
        }
      } catch (e) {
        console.error(e);
      }
    };

    ws.onclose = () => setTimeout(connect, 1000);
    ws.onerror = () => ws.close();
  }

  connect();
</script>
</body>
</html>
`
