export interface IEvent {
    id: number;
    family_id: number;
    title: string;
    description: string;
    start_time: string; // ISO-строка
    end_time: string;   // ISO-строка
    created_by: number;
    is_completed: boolean;
    color?: string | null; // <-- добавлено для цвета
    created_at: string;
    updated_at: string;
  }

  export interface ICreateEventRequest {
    calendar_id: number;   // <-- новое поле
    title: string;
    description: string;
    start_time: string;   // ISO
    end_time: string;
    color?: string;
  }