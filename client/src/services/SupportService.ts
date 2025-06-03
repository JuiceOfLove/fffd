import $api from "../http";
import { IMyTicket, IOperatorTicket, ITicketInfo } from "../types/support";
import { IChatMessage } from "../types/chat";

class SupportService {
  /*───────────────────────── тикеты пользователя ─────────────────────────*/

  static async createTicket(payload: { subject: string; content: string })
    : Promise<{ ticket_id: number }> {
    const { data } = await $api.post("/support/tickets", payload);
    return data;                    // { ticket_id }
  }

  static async getMyTickets(): Promise<IMyTicket[]> {
    const { data } = await $api.get("/support/tickets/my");
    return data;
  }

  static async getTicketInfo(ticketId: number): Promise<ITicketInfo> {
    const { data } = await $api.get(`/support/tickets/${ticketId}`);
    return data;
  }

  static async getTicketMessages(ticketId: number): Promise<IChatMessage[]> {
    const { data } = await $api.get(`/support/tickets/${ticketId}/messages`);
    /* сервер отдаёт sender_id → переименуем в user_id для UI */
    return data.map((m: any) => ({ ...m, user_id: m.sender_id }));
  }

  /*───────────────────────── действия оператора ──────────────────────────*/

  static async assignTicket(id: number) { await $api.post(`/support/tickets/${id}/assign`); }
  static async closeTicket (id: number) { await $api.post(`/support/tickets/${id}/close`); }

  /*───────────────────────── операторские списки ─────────────────────────*/

  static async getOperatorTickets(status: "new" | "active" | "closed")
    : Promise<IOperatorTicket[]> {
    const { data } = await $api.get("/support/tickets/operator/list",
                                    { params: { status } });
    return data;
  }

  static async deleteTicketMessage(ticketId: number, msgId: number) {
    await $api.delete(`/support/tickets/${ticketId}/messages/${msgId}`);
  }
}

export default SupportService;
