package web

const DuolingoBoardHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>ChessBot Live</title>
  <style>
    * { box-sizing: border-box; margin: 0; padding: 0; user-select: none; }
    body { background-color: #0b1418; display: flex; flex-direction: column; justify-content: center; align-items: center; min-height: 100vh; overflow: hidden; font-family: monospace; }
    .board-container { width: min(82vw, 82vh); height: min(82vw, 82vh); position: relative; background-color: #141f23; border: 4px solid #1a272c; box-shadow: 0 12px 36px rgba(0,0,0,0.7); }
    .board-grid { width: 100%; height: 100%; display: grid; grid-template-columns: repeat(8, 1fr); grid-template-rows: repeat(8, 1fr); }
    .cell { position: relative; width: 100%; height: 100%; }
    .cell.light { background-color: #273941; }
    .cell.dark  { background-color: #17242a; }
    .cell.highlight-from { background-color: #1a4d2e !important; box-shadow: inset 0 0 0 3px #58cc02; }
    .cell.highlight-to { background-color: #4a3e1a !important; box-shadow: inset 0 0 0 3px #ffc800; }
    .pieces-layer { position: absolute; top: 0; left: 0; width: 100%; height: 100%; pointer-events: none; }
    .piece-node { position: absolute; width: 12.5%; height: 12.5%; top: 0; left: 0; display: flex; align-items: center; justify-content: center; transition: transform 0.16s cubic-bezier(0.2, 0, 0, 1); z-index: 10; will-change: transform; }
    .piece-node img { width: 88%; height: 88%; object-fit: contain; filter: drop-shadow(0 3px 6px rgba(0,0,0,0.35)); pointer-events: none; }
    .info-bar { margin-top: 14px; color: #93a1a7; font-size: 16px; font-weight: bold; display: flex; flex-wrap: wrap; justify-content: center; gap: 16px; }
    .badge { background-color: #141f23; padding: 6px 12px; border-radius: 8px; border: 1px solid #23343c; }
    .active-val { color: #58cc02; }
    .active-move { color: #ffc800; }
    .gameover-val { color: #ff4757; }
  </style>
</head>
<body>
<div class="board-container" id="boardBox">
  <div class="board-grid" id="gridBg"></div>
  <div class="pieces-layer" id="piecesLayer"></div>
</div>
<div class="info-bar">
  <div class="badge">STATE: <span id="uiState" class="active-val">1</span></div>
  <div class="badge">MOVE: <span id="uiMove" class="active-val">White</span></div>
  <div class="badge">YOU: <span id="uiMyTeam" class="active-val">...</span></div>
  <div class="badge">STATUS: <span id="uiBestMove" class="active-move">-</span></div>
  <div class="badge" style="color: #637580;">CREATED BY: <span style="color: #00bcd4;">ytika</span></div>
</div>
<script>
  const gridBg = document.getElementById('gridBg');
  const piecesLayer = document.getElementById('piecesLayer');
  const uiState = document.getElementById('uiState');
  const uiMove = document.getElementById('uiMove');
  const uiMyTeam = document.getElementById('uiMyTeam');
  const uiBestMove = document.getElementById('uiBestMove');

  const PIECE_TYPES = ['wP','wN','wB','wR','wQ','wK','bP','bN','bB','bR','bQ','bK'];
  const preloadedImgs = {};
  for (const pt of PIECE_TYPES) {
    const img = new Image();
    img.src = '/figures/' + pt + '.png';
    preloadedImgs[pt] = img;
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

  function updateBoard(matrix) {
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

    for (const item of targetItems) {
      const idx = activePieces.findIndex(ap => !ap.matched && ap.type === item.type && ap.r === item.r && ap.c === item.c);
      if (idx !== -1) {
        item.matched = true;
        const ap = activePieces[idx];
        ap.matched = true;
        nextPieces.push(ap);
      }
    }

    for (const item of targetItems) {
      if (item.matched) continue;
      const idx = activePieces.findIndex(ap => !ap.matched && ap.type === item.type);
      if (idx !== -1) {
        item.matched = true;
        const ap = activePieces[idx];
        ap.matched = true;
        ap.r = item.r;
        ap.c = item.c;
        ap.el.style.transform = 'translate(' + (item.c * 100) + '%, ' + (item.r * 100) + '%)';
        nextPieces.push(ap);
      }
    }

    for (const item of targetItems) {
      if (item.matched) continue;
      const el = document.createElement('div');
      el.className = 'piece-node';
      el.style.transform = 'translate(' + (item.c * 100) + '%, ' + (item.r * 100) + '%)';

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

    for (const ap of activePieces) {
      if (!ap.matched) {
        if (ap.el.parentNode) ap.el.parentNode.removeChild(ap.el);
      }
    }

    activePieces = nextPieces;
    for (const ap of activePieces) ap.matched = false;
  }

  function updateHighlights(from, to) {
    for (let r = 0; r < 8; r++) {
      for (let c = 0; c < 8; c++) {
        const cell = document.getElementById('sq_' + r + '_' + c);
        if (!cell) continue;
        const isFrom = (from && from[0] === r && from[1] === c);
        const isTo = (to && to[0] === r && to[1] === c);
        
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
        if (msg.board) {
          updateBoard(msg.board);
          updateHighlights(msg.from, msg.to);
        }
        if (msg.state !== undefined) uiState.textContent = msg.state;
        if (msg.move !== undefined) uiMove.textContent = msg.move;
        if (msg.my_team !== undefined) uiMyTeam.textContent = msg.my_team;
        if (msg.best_move !== undefined) uiBestMove.textContent = msg.best_move;

        if (msg.is_game_over) {
          uiBestMove.className = 'active-move gameover-val';
        } else {
          uiBestMove.className = 'active-move';
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
