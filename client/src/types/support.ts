export interface IMyTicket {
  id: number;
  subject: string;
  status: "new" | "active" | "closed";
  status_label: string;       // REST возвращает "status_label" (уровень BE)
  last_message_at: string;    // ISO-строка
  operator_id: number | null;
}

export interface IOperatorTicket {
  id: number;
  subject: string;
  user_id: number;
  user_name: string;
  last_message_at: string;    // ISO-строка
  status: "new" | "active" | "closed";
  operator_id: number | null;
}

// Интерфейс, который возвращает getTicketInfo()
export interface ITicketInfo {
  id: number;
  subject: string;
  status: "new" | "active" | "closed";
  user_id: number;
  user_name: string;
  operator_id: number | null;
  operator_name: string | null;
  last_message_at: string;
}