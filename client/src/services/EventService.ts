import $api from "../http";
import { IEvent, ICreateEventRequest } from "../types/EventTypes";

export default class EventService {
  // Создать событие (теперь c поддержкой color)
  static async createEvent(data: ICreateEventRequest): Promise<IEvent> {
    const res = await $api.post<{ event: IEvent }>("/calendar/events", data);
    return res.data.event;
  }

  // Получить все события
  static async getAllEvents(): Promise<IEvent[]> {
    const res = await $api.get<IEvent[]>("/calendar/events/all");
    return res.data;
  }

  // Получить события за конкретный месяц
  static async getEventsForMonth(month: number, year: number): Promise<IEvent[]> {
    const res = await $api.get<IEvent[]>("/calendar/events", { params: { month, year } });
    return res.data;
  }

  // Завершить событие
  static async completeEvent(id: number): Promise<IEvent> {
    const res = await $api.post<{ event: IEvent }>(`/calendar/events/${id}/complete`);
    return res.data.event;
  }

  // (Опционально) обновить событие
  static async updateEvent(id: number, data: ICreateEventRequest): Promise<IEvent> {
    const res = await $api.put<{ event: IEvent }>(`/calendar/events/${id}`, data);
    return res.data.event;
  }

  static async getEventsForCalendar(calendarId: number, month: number, year: number): Promise<IEvent[]> {
    const res = await $api.get<IEvent[]>(`/calendar/${calendarId}/events`, {
      params: { month, year },
    });
    return res.data;
  }
}