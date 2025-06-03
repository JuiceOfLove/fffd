import { IChatMessage } from "../types/chat";

export type MsgHandler = (msg: IChatMessage) => void;
export type DelHandler = (id: number) => void;
export type PresenceHandler = (ids: number[]) => void;

class SupportSocketService {
  private wsTicket: WebSocket | null = null;

  private ticketMsgHandlers: MsgHandler[] = [];
  private ticketDelHandlers: DelHandler[] = [];
  private ticketPresHandlers: PresenceHandler[] = [];

  private safeWrite(ws: WebSocket, payload: any) {
    try {
      ws.send(JSON.stringify(payload));
    } catch (err) {
      console.warn("SupportSocketService WS write error:", err);
    }
  }

  connect(ticketId: number, token: string) {
    if (this.wsTicket) {
      this.wsTicket.close();
      this.wsTicket = null;
    }

    const proto = location.protocol === "https:" ? "wss" : "ws";
    const host = location.port === "5173" ? "localhost:8080" : location.host;
    const url = `${proto}://${host}/api/support/ws/${ticketId}?token=${token}`;

    console.log("SupportSocketService connecting to", url);
    this.wsTicket = new WebSocket(url);

    this.wsTicket.addEventListener("open", () => {
      console.log(`SupportSocketService WS (ticket ${ticketId}) open`);
    });
    this.wsTicket.addEventListener("error", (ev) => {
      console.error("SupportSocketService WS (ticket) error", ev);
    });
    this.wsTicket.addEventListener("close", (ev) => {
      console.warn("SupportSocketService WS (ticket) closed", ev);
    });

    this.wsTicket.addEventListener("message", (evt) => {
      const envelope = JSON.parse(evt.data);
      const { event, data } = envelope;

      if (event === "support:ticket_message") {
        this.ticketMsgHandlers.forEach((cb) => cb(data));
      }
      if (event === "support:ticket_delete") {
        this.ticketDelHandlers.forEach((cb) => cb(data.message_id));
      }
      if (event === "support:ticket_presence") {
        this.ticketPresHandlers.forEach((cb) => cb(data));
      }
    });
  }

  disconnect() {
    if (this.wsTicket) {
      this.wsTicket.close();
      this.wsTicket = null;
    }
    this.ticketMsgHandlers = [];
    this.ticketDelHandlers = [];
    this.ticketPresHandlers = [];
  }

  onMessage(cb: MsgHandler) {
    this.ticketMsgHandlers.push(cb);
  }

  onDelete(cb: DelHandler) {
    this.ticketDelHandlers.push(cb);
  }

  onPresence(cb: PresenceHandler) {
    this.ticketPresHandlers.push(cb);
  }

  sendMessage(text: string, replyTo?: number, mediaB64?: string) {
    if (!this.wsTicket || this.wsTicket.readyState !== WebSocket.OPEN) return;
    const msg: any = { content: text };
    if (replyTo) msg.reply_to = replyTo;
    if (mediaB64) msg.media = mediaB64;
    this.safeWrite(this.wsTicket, msg);
  }

  deleteMessage(messageId: number) {
    if (!this.wsTicket || this.wsTicket.readyState !== WebSocket.OPEN) return;
    this.safeWrite(this.wsTicket, { delete_id: messageId });
  }
}

export default new SupportSocketService();