(() => {
  'use strict';
  const $ = (selector, root = document) => root.querySelector(selector);
  const $$ = (selector, root = document) => [...root.querySelectorAll(selector)];
  const pageTitle = $('#pageTitle');
  const pageEyebrow = $('#pageEyebrow');
  const sidebar = $('#sidebar');
  const mobileScrim = $('#mobileScrim');
  const toastRegion = $('#toastRegion');
  let currentPage = 'home';
  let boardFlipped = false;
  let selectedPiece = null;
  let soundEnabled = true;
  let learningTimer = null;

  function icon(id) {
    return '<svg aria-hidden="true"><use href="#' + id + '"></use></svg>';
  }

  function showToast(message) {
    const toast = document.createElement('div');
    toast.className = 'toast';
    toast.textContent = message;
    toastRegion.appendChild(toast);
    window.setTimeout(() => toast.remove(), 3200);
  }

  function closeMobileNav() {
    sidebar.classList.remove('open');
    mobileScrim.hidden = true;
  }

  function navigate(page, updateHash = true) {
    const target = $('#page-' + page);
    if (!target) return;
    currentPage = page;
    $$('.page').forEach((item) => item.classList.toggle('active', item === target));
    $$('.nav-item[data-page]').forEach((item) => item.classList.toggle('active', item.dataset.page === page));
    pageTitle.textContent = target.dataset.title || '';
    pageEyebrow.textContent = target.dataset.eyebrow || '';
    if (updateHash) history.replaceState(null, '', '#' + page);
    closeMobileNav();
    window.scrollTo({ top: 0, behavior: matchMedia('(prefers-reduced-motion: reduce)').matches ? 'auto' : 'smooth' });
  }

  $$('.nav-item[data-page]').forEach((button) => button.addEventListener('click', () => navigate(button.dataset.page)));
  $$('[data-go]').forEach((button) => button.addEventListener('click', () => navigate(button.dataset.go)));
  $('#mobileMenu').addEventListener('click', () => { sidebar.classList.add('open'); mobileScrim.hidden = false; });
  $('#sidebarClose').addEventListener('click', closeMobileNav);
  mobileScrim.addEventListener('click', closeMobileNav);

  function applyTheme(theme) {
    const dark = theme === 'dark' || (theme === 'system' && matchMedia('(prefers-color-scheme: dark)').matches);
    document.documentElement.dataset.theme = dark ? 'dark' : 'light';
    $('.theme-icon use').setAttribute('href', dark ? '#i-sun' : '#i-moon');
    $('meta[name="theme-color"]').setAttribute('content', dark ? '#141916' : '#f3eee4');
  }

  const savedTheme = localStorage.getItem('xiangqi-demo-theme') || 'light';
  applyTheme(savedTheme);
  if ($('#themeSelect')) $('#themeSelect').value = savedTheme;
  $('#themeToggle').addEventListener('click', () => {
    const next = document.documentElement.dataset.theme === 'dark' ? 'light' : 'dark';
    localStorage.setItem('xiangqi-demo-theme', next);
    if ($('#themeSelect')) $('#themeSelect').value = next;
    applyTheme(next);
    showToast(next === 'dark' ? '已切换为深色主题' : '已切换为浅色主题');
  });
  $('#themeSelect').addEventListener('change', (event) => {
    localStorage.setItem('xiangqi-demo-theme', event.target.value);
    applyTheme(event.target.value);
  });
  matchMedia('(prefers-color-scheme: dark)').addEventListener?.('change', () => {
    if ((localStorage.getItem('xiangqi-demo-theme') || 'light') === 'system') applyTheme('system');
  });

  const miniPieces = [
    ['黑','将',4,0],['黑','车',0,0],['黑','马',7,2],['黑','炮',3,3],['黑','卒',4,4],
    ['红','帅',4,9],['红','车',8,8],['红','马',2,7],['红','炮',5,6],['红','兵',4,5]
  ];
  miniPieces.forEach(([color, name, file, rank]) => {
    const el = document.createElement('span');
    el.className = 'mini-piece ' + (color === '红' ? 'red' : 'black');
    el.textContent = name;
    el.style.left = (file * 12.5) + '%';
    el.style.top = (rank * 11.11) + '%';
    $('#miniBoard').appendChild(el);
  });

  $$('[data-choice-group]').forEach((group) => {
    group.addEventListener('click', (event) => {
      const choice = event.target.closest('button[data-value]');
      if (!choice) return;
      $$('button[data-value]', group).forEach((button) => button.classList.toggle('active', button === choice));
      if (group.dataset.choiceGroup === 'mode') {
        const labels = { standard: '标准引擎', library: '棋谱库优先', style: '棋风模仿' };
        $('#summaryMode').textContent = labels[choice.dataset.value];
      }
      if (group.dataset.choiceGroup === 'side') {
        const labels = { red: '红方', black: '黑方', random: '随机执色' };
        $('.summary-versus>div:first-child small').textContent = labels[choice.dataset.value];
      }
    });
  });

  const difficultyProfiles = [
    ['入门','认识棋局','0.2–0.5 秒','3 路','较高','适合熟悉规则，会在合理着法中保留明显容错。'],
    ['入门','轻松起步','0.3–0.7 秒','3 路','较高','偏向清晰直观的走法，避免无意义送子。'],
    ['休闲','从容对弈','0.5–1 秒','4 路','中等','具备基础战术意识，候选着变化较丰富。'],
    ['休闲','有来有回','0.8–1.5 秒','4 路','中等','会把握简单战机，同时保留适度随机性。'],
    ['进阶','谨慎谋划','1–2 秒','4 路','较低','开始关注中局结构和子力协调。'],
    ['进阶','沉稳应战','1.5–3 秒','4 路','较低','会进行更深入的局面判断，在多个合理候选着中保持少量变化。'],
    ['高手','精确计算','2–4 秒','5 路','很低','缩小候选评分带，重视战术与局面转换。'],
    ['高手','深度布局','3–6 秒','5 路','很低','更深入地搜索复杂变化，较少主动放弃优势。'],
    ['大师','强力挑战','5–8 秒','6 路','极低','使用更高搜索预算，优先选择高质量候选着。'],
    ['大师','极致棋力','8–12 秒','8 路','极低','接近当前配置的最高搜索资源，不标注未经校准的 Elo。']
  ];
  const range = $('#difficultyRange');
  function renderDifficulty() {
    const value = Number(range.value);
    const profile = difficultyProfiles[value - 1];
    $('#difficultyLevel').textContent = profile[0] + ' · ' + value + ' 级';
    $('#difficultyName').textContent = profile[1];
    $('#difficultyTime').textContent = profile[2];
    $('#difficultyPv').textContent = profile[3];
    $('#difficultyRandom').textContent = profile[4];
    $('#difficultyDescription').textContent = profile[5];
    $('#summaryDifficulty').textContent = profile[0] + ' ' + value;
    range.setAttribute('aria-label', 'AI 难度，当前 ' + value + ' 级');
    $$('.difficulty-numbers span').forEach((item, index) => item.classList.toggle('active', index === value - 1));
  }
  range.addEventListener('input', renderDifficulty);
  renderDifficulty();
  $('#startMatch').addEventListener('click', () => { navigate('match'); showToast('对局已创建：你执红，AI 难度 ' + range.value + ' 级'); });

  const initialPieces = [
    ['black','车',0,0],['black','马',1,0],['black','象',2,0],['black','士',3,0],['black','将',4,0],['black','士',5,0],['black','象',6,0],['black','马',7,0],['black','车',8,0],
    ['black','炮',1,2],['black','炮',7,2],['black','卒',0,3],['black','卒',2,3],['black','卒',4,3],['black','卒',6,3],['black','卒',8,3],
    ['red','兵',0,6],['red','兵',2,6],['red','兵',4,6],['red','兵',6,6],['red','兵',8,6],['red','炮',1,7],['red','炮',7,7],
    ['red','车',0,9],['red','马',1,9],['red','相',2,9],['red','仕',3,9],['red','帅',4,9],['red','仕',5,9],['red','相',6,9],['red','马',7,9],['red','车',8,9]
  ];
  const board = $('#xiangqiBoard');
  function screenPosition(file, rank) {
    const f = boardFlipped ? 8 - file : file;
    const r = boardFlipped ? 9 - rank : rank;
    return { left: (4.5 + f * 11.375) + '%', top: (4.5 + r * 10.11) + '%' };
  }
  function positionElement(el, file, rank) {
    const pos = screenPosition(file, rank);
    el.style.left = pos.left;
    el.style.top = pos.top;
  }
  function addBoardOverlay() {
    const svg = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
    svg.setAttribute('viewBox', '0 0 800 900');
    svg.setAttribute('aria-hidden', 'true');
    svg.style.cssText = 'position:absolute;inset:4.5%;width:91%;height:91%;z-index:0;pointer-events:none;overflow:visible';
    svg.innerHTML = '<g fill="none" stroke="var(--board-line)" stroke-width="2"><path d="M300 0L500 200M500 0L300 200M300 700L500 900M500 700L300 900"/></g>';
    board.appendChild(svg);
  }
  addBoardOverlay();
  initialPieces.forEach(([color, name, file, rank], index) => {
    const piece = document.createElement('button');
    piece.className = 'board-piece ' + color + (index === 20 ? ' last' : '');
    piece.textContent = name;
    piece.dataset.file = file;
    piece.dataset.rank = rank;
    piece.dataset.name = name;
    piece.dataset.color = color;
    piece.setAttribute('aria-label', (color === 'red' ? '红方' : '黑方') + name + '，位置 ' + file + ',' + rank);
    positionElement(piece, Number(file), Number(rank));
    board.appendChild(piece);
  });

  function clearHints() {
    $$('.move-hint', board).forEach((hint) => hint.remove());
    $$('.board-piece', board).forEach((piece) => piece.classList.remove('selected'));
  }
  function candidateMoves(piece) {
    const file = Number(piece.dataset.file), rank = Number(piece.dataset.rank), color = piece.dataset.color, name = piece.dataset.name;
    const moves = [];
    const push = (f, r) => { if (f >= 0 && f <= 8 && r >= 0 && r <= 9) moves.push([f, r]); };
    if (name === '马') [[-2,-1],[-2,1],[-1,-2],[-1,2],[1,-2],[1,2],[2,-1],[2,1]].slice(0,4).forEach(([df,dr]) => push(file+df,rank+dr));
    else if (name === '车' || name === '炮') { push(file,rank+(color==='red'?-1:1)); push(file,rank+(color==='red'?-2:2)); push(file-1,rank); push(file+1,rank); }
    else { push(file,rank+(color==='red'?-1:1)); if (name==='兵'||name==='卒') { push(file-1,rank); push(file+1,rank); } else { push(file-1,rank); push(file+1,rank); } }
    return moves.slice(0,4);
  }
  board.addEventListener('click', (event) => {
    const piece = event.target.closest('.board-piece');
    const hint = event.target.closest('.move-hint');
    if (hint && selectedPiece) {
      $$('.board-piece', board).forEach((item) => item.classList.remove('last'));
      selectedPiece.dataset.file = hint.dataset.file;
      selectedPiece.dataset.rank = hint.dataset.rank;
      selectedPiece.classList.add('last');
      positionElement(selectedPiece, Number(hint.dataset.file), Number(hint.dataset.rank));
      showToast('已演示落子：' + selectedPiece.dataset.name + '移动到新位置');
      clearHints(); selectedPiece = null; return;
    }
    if (!piece) { clearHints(); selectedPiece = null; return; }
    clearHints(); selectedPiece = piece; piece.classList.add('selected');
    candidateMoves(piece).forEach(([file,rank]) => {
      const hintEl = document.createElement('button');
      hintEl.className = 'move-hint'; hintEl.dataset.file = file; hintEl.dataset.rank = rank;
      hintEl.setAttribute('aria-label', '演示移动到 ' + file + ',' + rank);
      positionElement(hintEl, file, rank); board.appendChild(hintEl);
    });
  });
  $('#flipBoard').addEventListener('click', () => {
    boardFlipped = !boardFlipped;
    $$('.board-piece', board).forEach((piece) => positionElement(piece, Number(piece.dataset.file), Number(piece.dataset.rank)));
    $$('.move-hint', board).forEach((hint) => positionElement(hint, Number(hint.dataset.file), Number(hint.dataset.rank)));
    board.setAttribute('aria-label', '中国象棋棋盘，' + (boardFlipped ? '黑方' : '红方') + '视角');
    showToast(boardFlipped ? '已切换为黑方视角' : '已切换为红方视角');
  });
  $('#undoMove').addEventListener('click', () => showToast('悔棋请求已提交；正式版将取消旧 AI 搜索并等待服务端确认'));
  $('#soundToggle').addEventListener('click', () => { soundEnabled = !soundEnabled; showToast(soundEnabled ? '落子音效已开启' : '落子音效已关闭'); });
  $('#resignButton').addEventListener('click', () => $('#confirmDialog').showModal());
  $('#confirmResign').addEventListener('click', () => setTimeout(() => { navigate('history'); showToast('本局已结束并保存到历史对局'); }, 50));

  $$('[data-match-tab]').forEach((tab) => tab.addEventListener('click', () => {
    $$('[data-match-tab]').forEach((item) => { const active = item === tab; item.classList.toggle('active', active); item.setAttribute('aria-selected', active); });
    $$('.match-tab').forEach((panel) => panel.classList.toggle('active', panel.id === 'match-tab-' + tab.dataset.matchTab));
  }));

  $('#uploadTrigger').addEventListener('click', () => $('#fileInput').click());
  $('#fileInput').addEventListener('change', (event) => {
    const files = [...event.target.files];
    if (!files.length) return;
    $('#importTitle').textContent = '刚刚上传的批次';
    $('#importStatus').textContent = '解析中';
    $('#importStatus').className = 'tag neutral';
    $('.import-progress span').style.width = '24%';
    showToast('已接收 ' + files.length + ' 个文件，正在模拟安全校验');
    setTimeout(() => { $('.import-progress span').style.width = '100%'; $('#importStatus').textContent = '已完成'; $('#importStatus').className = 'tag success'; $('#importSuccess').textContent = String(42 + files.length); }, 1300);
    const row = document.createElement('tr');
    row.innerHTML = '<td><strong>' + files[0].name.replace(/[<>]/g,'') + '</strong><small>刚刚导入的演示棋谱</small></td><td>待识别</td><td><span class="tag neutral">解析中</span></td><td>未分类</td><td>刚刚</td><td><button class="icon-button">' + icon('i-chevron') + '</button></td>';
    $('#recordTableBody').prepend(row);
  });

  $('#buildLearning').addEventListener('click', () => $('#learningDialog').showModal());
  $('#runLearningJob').addEventListener('click', () => {
    if (learningTimer) return;
    const stages = ['安全校验','逐着规则验证','局面着法统计','棋风特征提取','质量检查','构建完成'];
    let progress = 0;
    learningTimer = setInterval(() => {
      progress = Math.min(100, progress + 4);
      $('#jobProgress').style.width = progress + '%';
      $('#jobPercent').textContent = progress + '%';
      $('#jobStage').textContent = stages[Math.min(stages.length - 1, Math.floor(progress / 20))];
      if (progress >= 100) { clearInterval(learningTimer); learningTimer = null; $('#runLearningJob').textContent = '构建完成'; showToast('学习版本 v4 已完成模拟构建，等待质量确认'); }
    }, 110);
  });

  const historyItems = [
    ['win','胜','AI · 休闲 4','中炮对屏风马 · 执红','标准引擎','42 回合','38%','今日 09:42'],
    ['loss','负','AI · 进阶 6','仙人指路 · 执黑','棋风模仿','56 回合','26%','昨日 21:18'],
    ['draw','和','AI · 进阶 5','飞相局 · 执红','棋谱优先','68 回合','44%','7 月 13 日'],
    ['win','胜','AI · 入门 2','顺炮直车 · 执黑','标准引擎','31 回合','0%','7 月 12 日'],
    ['win','胜','AI · 高手 7','中炮急进中兵 · 执红','棋谱优先','49 回合','33%','7 月 10 日']
  ];
  $('#historyList').innerHTML = historyItems.map((item) => '<button class="history-row" data-go="analysis"><span class="result-badge ' + item[0] + '">' + item[1] + '</span><span><strong>' + item[2] + '</strong><small>' + item[3] + '</small></span><span><strong>' + item[4] + '</strong><small>AI 模式</small></span><span><strong>' + item[5] + '</strong><small>对局长度</small></span><span><strong>' + item[6] + '</strong><small>学习命中</small></span><span class="tag success">已复盘</span><svg class="row-chevron"><use href="#i-chevron"/></svg></button>').join('');
  $$('.history-row').forEach((row) => row.addEventListener('click', () => navigate('analysis')));

  $$('[data-settings-tab]').forEach((tab) => tab.addEventListener('click', () => {
    $$('[data-settings-tab]').forEach((item) => item.classList.toggle('active', item === tab));
    $$('.settings-tab').forEach((panel) => panel.classList.toggle('active', panel.id === 'settings-' + tab.dataset.settingsTab));
  }));

  $('button.icon-button:not([aria-label])').forEach((button) => button.setAttribute('aria-label', '更多操作'));

  const hashPage = location.hash.replace('#','');
  navigate($('#page-' + hashPage) ? hashPage : 'home', false);
})();
