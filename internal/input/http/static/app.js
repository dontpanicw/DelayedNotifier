const API = '/api/notifications';
const listEl = document.getElementById('list');
const emptyEl = document.getElementById('empty');
const form = document.getElementById('form');
const formError = document.getElementById('form-error');
const submitBtn = document.getElementById('submit-btn');

function setFormError(msg) {
  formError.textContent = msg || '';
  formError.style.display = msg ? 'block' : 'none';
}

function formatDate(iso) {
  if (!iso) return '—';
  try {
    const d = new Date(iso);
    return d.toLocaleString('ru-RU', { dateStyle: 'short', timeStyle: 'short' });
  } catch (_) {
    return iso;
  }
}

function statusClass(s) {
  if (!s) return '';
  return 'status-' + (s.replace(/\s+/g, '_'));
}

function renderItem(m) {
  const li = document.createElement('li');
  li.innerHTML = `
    <div>
      <div class="notif-text">${escapeHtml(m.text || '')}</div>
      <div class="notif-meta">
        <span class="notif-id">${escapeHtml(m.id || '')}</span><br>
        ${formatDate(m.scheduled_at)} · user_id: ${m.user_id ?? '—'} · chat_id: ${m.telegram_chat_id ?? '—'}
      </div>
    </div>
    <span class="status ${statusClass(m.status)}">${escapeHtml(m.status || '')}</span>
  `;
  return li;
}

function escapeHtml(s) {
  const div = document.createElement('div');
  div.textContent = s;
  return div.innerHTML;
}

async function loadList() {
  try {
    const res = await fetch(API);
    if (!res.ok) throw new Error(res.statusText);
    const data = await res.json();
    listEl.innerHTML = '';
    if (!data || data.length === 0) {
      emptyEl.textContent = 'Нет уведомлений';
      emptyEl.style.display = 'block';
      return;
    }
    emptyEl.style.display = 'none';
    data.forEach(m => listEl.appendChild(renderItem(m)));
  } catch (e) {
    emptyEl.textContent = 'Ошибка загрузки: ' + e.message;
    emptyEl.style.display = 'block';
  }
}

function toRFC3339Local(dateInput) {
  const d = new Date(dateInput);
  const pad = n => String(n).padStart(2, '0');
  const tz = -d.getTimezoneOffset();
  const sign = tz >= 0 ? '+' : '-';
  const h = Math.floor(Math.abs(tz) / 60);
  const m = Math.abs(tz) % 60;
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}${sign}${pad(h)}:${pad(m)}`;
}

form.addEventListener('submit', async (e) => {
  e.preventDefault();
  setFormError('');
  submitBtn.classList.add('loading');
  const text = form.text.value.trim();
  const scheduledAt = form.scheduled_at.value;
  const userId = parseInt(form.user_id.value, 10);
  const telegramChatId = parseInt(form.telegram_chat_id.value, 10);
  const body = {
    text,
    scheduled_at: scheduledAt ? new Date(scheduledAt).toISOString() : null,
    user_id: userId,
    telegram_chat_id: telegramChatId
  };
  if (!body.scheduled_at) {
    setFormError('Укажите время отправки');
    submitBtn.classList.remove('loading');
    return;
  }
  try {
    const res = await fetch(API, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body)
    });
    if (!res.ok) {
      const text = await res.text();
      setFormError(text || res.statusText || 'Ошибка создания');
      return;
    }
    form.text.value = '';
    form.scheduled_at.value = '';
    loadList();
  } catch (err) {
    setFormError(err.message || 'Ошибка сети');
  } finally {
    submitBtn.classList.remove('loading');
  }
});

loadList();
setInterval(loadList, 10000);
