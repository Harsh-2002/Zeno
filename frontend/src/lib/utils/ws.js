const MSG_DATA = 0x00;
const MSG_RESIZE = 0x01;
const MSG_SESSION = 0x02;

export { MSG_DATA, MSG_RESIZE, MSG_SESSION };

export function createWebSocket() {
  const proto = location.protocol === 'https:' ? 'wss:' : 'ws:';
  const ws = new WebSocket(`${proto}//${location.host}/ws`);
  ws.binaryType = 'arraybuffer';
  return ws;
}

export function sendData(ws, text) {
  if (ws.readyState !== WebSocket.OPEN) return;
  const enc = new TextEncoder().encode(text);
  const msg = new Uint8Array(enc.length + 1);
  msg[0] = MSG_DATA;
  msg.set(enc, 1);
  ws.send(msg);
}

export function sendBinary(ws, data) {
  if (ws.readyState !== WebSocket.OPEN) return;
  const buf = new Uint8Array(data.length + 1);
  buf[0] = MSG_DATA;
  for (let i = 0; i < data.length; i++) buf[i + 1] = data.charCodeAt(i) & 0xff;
  ws.send(buf);
}

export function sendResize(ws, cols, rows) {
  if (ws.readyState !== WebSocket.OPEN) return;
  const payload = JSON.stringify({ cols, rows });
  const enc = new TextEncoder().encode(payload);
  const msg = new Uint8Array(enc.length + 1);
  msg[0] = MSG_RESIZE;
  msg.set(enc, 1);
  ws.send(msg);
}

export function sendSessionConnect(ws, sessionId) {
  if (ws.readyState !== WebSocket.OPEN) return;
  const payload = JSON.stringify({ action: 'connect', sessionID: sessionId || '' });
  const enc = new TextEncoder().encode(payload);
  const msg = new Uint8Array(enc.length + 1);
  msg[0] = MSG_SESSION;
  msg.set(enc, 1);
  ws.send(msg);
}

export function parseMessage(data) {
  const bytes = new Uint8Array(data);
  if (bytes.length === 0) return null;
  return { type: bytes[0], payload: bytes.slice(1) };
}
