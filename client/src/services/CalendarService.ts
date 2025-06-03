import $api from "../http";

export default class CalendarService {
  static async getAllCalendars() {
    const response = await $api.get("/calendar/list");
    return response.data; // массив [{ id, title, family_id, ... }, ...]
  }

  static async createExtraCalendar(title: string) {
    // вызывем POST /calendar/create_extra
    const response = await $api.post("/calendar/create_extra", { title });
    return response.data.calendar; // объект { id, title, family_id, ...}
  }
}