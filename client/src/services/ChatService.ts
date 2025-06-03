import { IChatMessage } from '../types/chat';

export type MsgHandler  = (msg: IChatMessage) => void;
export type DelHandler  = (id : number)       => void;
export type PresenceHandler = (ids: number[]) => void;

class ChatService {
  private ws: WebSocket | null = null;
  private msgHandlers: MsgHandler[] = [];
  private delHandlers: DelHandler[] = [];
  private presHandlers: PresenceHandler[] = [];

  connect(token: string) {
    if (this.ws) { this.ws.close(); this.ws = null; }

    const proto = location.protocol === 'https:' ? 'wss' : 'ws';
    const host  = location.port === '5173' ? 'localhost:8080' : location.host;
    const url   = `${proto}://${host}/api/chat/ws?token=${token}`;

    console.log('ChatService connecting to', url);
    this.ws = new WebSocket(url);

    this.ws.onopen  = () => console.log('ChatService WS open');
    this.ws.onerror = ev => console.error('ChatService WS error', ev);
    this.ws.onclose = ev => console.warn ('ChatService WS closed', ev);

    this.ws.onmessage = evt => {
      const { type, data } = JSON.parse(evt.data);
      if (type === 'message')  this.msgHandlers.forEach(cb => cb(data));
      if (type === 'presence') this.presHandlers.forEach(cb => cb(data));
      if (type === 'delete')   this.delHandlers.forEach(cb => cb(data));
    };
  }

  /* subscriptions */
  onMessage(cb: MsgHandler)        { this.msgHandlers.push(cb); }
  onDelete (cb: DelHandler)        { this.delHandlers.push(cb); }
  onPresence(cb: PresenceHandler)  { this.presHandlers.push(cb); }

  /* send new text / image / reply */
  send(text: string, replyTo?: number, mediaB64?: string) {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) return;
    const msg: any = { content: text };
    if (replyTo)  msg.reply_to = replyTo;
    if (mediaB64) msg.media    = mediaB64;
    this.ws.send(JSON.stringify(msg));
  }

  /* soft-delete own message */
  delete(id: number) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({ delete_id: id }));
    }
  }

  disconnect() {
    this.ws?.close();
    this.ws = null;
    this.msgHandlers = [];
    this.delHandlers = [];
    this.presHandlers = [];
  }
}

export default new ChatService();
