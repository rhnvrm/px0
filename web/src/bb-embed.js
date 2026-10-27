import { S } from './state.js';
import { setBBHandler } from './selbar.js';
import { setTheme } from './theme.js';
import { openFile } from './tabs.js';
import { triggerRefresh } from './gitstream.js';

export function initBBEmbed() {
 const origin = S.meta?.embedOrigin;
 if (!origin || window.parent === window) return;
 document.body.classList.add('bb-embed');
 const button = document.querySelector('[data-sel="bb-review"]');
 if (button) button.hidden = false;
 setBBHandler(info => window.parent.postMessage({type:'px0:selection', path:info.path, l1:info.l1, l2:info.l2, text:info.text.slice(0, 24000)}, origin));
 window.addEventListener('message', event => {
  if (event.source !== window.parent || event.origin !== origin || !event.data || typeof event.data !== 'object') return;
  const data = event.data;
  if (data.type === 'bb:theme' && (data.theme === 'dark' || data.theme === 'light')) setTheme(data.theme, false);
  if (data.type === 'bb:refresh') triggerRefresh();
  if (data.type === 'bb:open' && typeof data.path === 'string' && Number.isInteger(data.line) && data.line > 0) openFile(data.path, {line:data.line});
 });
 window.parent.postMessage({type:'px0:ready'}, origin);
}
