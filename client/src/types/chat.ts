export interface IChatMessage {
  id:          number;
  /** family chat → user_id,  ticket chat → sender_id → поэтому делаем оба */
  user_id:     number;        // семейный чат
  sender_id?:  number;        // тикеты (optional, чтобы не ломать старый код)

  content?:    string;
  media_url?:  string;
  reply_to_id?: number | null;
  created_at:  string;
  deleted_at?: string | null;
}