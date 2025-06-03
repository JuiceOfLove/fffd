// src/components/screens/dashbord/Calendar/CalendarPage/CalendarPage.tsx
import React, { useEffect, useState } from "react";
import { observer } from "mobx-react-lite";
import { useParams } from "react-router";
import EventService from "../../../../../services/EventService";
import { IEvent } from "../../../../../types/EventTypes";
import styles from "./CalendarPage.module.css";

/** Возвращает начало суток выбранного дня */
function startOfDay(date: Date) {
  return new Date(date.getFullYear(), date.getMonth(), date.getDate(), 0, 0, 0);
}

/** Возвращает конец суток выбранного дня */
function endOfDay(date: Date) {
  return new Date(date.getFullYear(), date.getMonth(), date.getDate(), 23, 59, 59);
}

/** Проверяем, пересекает ли событие (ev) конкретный день (date) */
function isEventOnDate(ev: IEvent, date: Date) {
  const evStart = new Date(ev.start_time);
  const evEnd = new Date(ev.end_time);
  const dayStart = startOfDay(date);
  const dayEnd = endOfDay(date);
  return evStart <= dayEnd && evEnd >= dayStart;
}

const monthNames = [
  "Январь",
  "Февраль",
  "Март",
  "Апрель",
  "Май",
  "Июнь",
  "Июль",
  "Август",
  "Сентябрь",
  "Октябрь",
  "Ноябрь",
  "Декабрь",
];

const CalendarPage: React.FC = observer(() => {
  const { id } = useParams();
  const calendarId = id ? parseInt(id, 10) : 1;

  // Текущий месяц/год
  const [currentMonth, setCurrentMonth] = useState<number>(new Date().getMonth());
  const [currentYear, setCurrentYear] = useState<number>(new Date().getFullYear());

  const [events, setEvents] = useState<IEvent[]>([]);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string>("");

  // Модалка дня
  const [popupOpen, setPopupOpen] = useState<boolean>(false);
  const [selectedDate, setSelectedDate] = useState<Date | null>(null);
  const [dayEvents, setDayEvents] = useState<IEvent[]>([]);
  const [showCreateForm, setShowCreateForm] = useState<boolean>(false);

  // Поля формы «обычного» события
  const [allDay, setAllDay] = useState<boolean>(true);
  const [title, setTitle] = useState<string>("");
  const [description, setDescription] = useState<string>("");
  const [startTime, setStartTime] = useState<string>("");
  const [endTime, setEndTime] = useState<string>("");

  // Модалка «интервального» события
  const [intervalModalOpen, setIntervalModalOpen] = useState<boolean>(false);
  const [intervalTitle, setIntervalTitle] = useState<string>("");
  const [intervalDesc, setIntervalDesc] = useState<string>("");
  const [intervalStart, setIntervalStart] = useState<string>("");
  const [intervalEnd, setIntervalEnd] = useState<string>("");
  const [intervalColor, setIntervalColor] = useState<string>("#ff5252");

  // Диапазон лет для селекта
  const yearOptions = Array.from(
    { length: 11 },
    (_, i) => new Date().getFullYear() - 5 + i
  );

  useEffect(() => {
    loadEvents();
  }, [currentMonth, currentYear, calendarId]);

  async function loadEvents() {
    try {
      setLoading(true);
      setError("");
      const data = await EventService.getEventsForCalendar(
        calendarId,
        currentMonth + 1,
        currentYear
      );
      setEvents(data);
    } catch (e: any) {
      console.error("Ошибка загрузки:", e.response?.data || e.message);
      setError(e.response?.data?.error || "Ошибка при загрузке");
    } finally {
      setLoading(false);
    }
  }

  function getDaysInMonth(month: number, year: number): number {
    return new Date(year, month + 1, 0).getDate();
  }
  function getFirstDayOfMonth(month: number, year: number): number {
    return new Date(year, month, 1).getDay(); // 0 = Sunday
  }

  const daysInMonth = getDaysInMonth(currentMonth, currentYear);
  const firstDayIndex = getFirstDayOfMonth(currentMonth, currentYear);

  function prevMonth() {
    let m = currentMonth - 1;
    let y = currentYear;
    if (m < 0) {
      m = 11;
      y--;
    }
    setCurrentMonth(m);
    setCurrentYear(y);
    closeDayPopup();
  }
  function nextMonth() {
    let m = currentMonth + 1;
    let y = currentYear;
    if (m > 11) {
      m = 0;
      y++;
    }
    setCurrentMonth(m);
    setCurrentYear(y);
    closeDayPopup();
  }

  // Когда меняем месяц через селект
  function onMonthChange(e: React.ChangeEvent<HTMLSelectElement>) {
    setCurrentMonth(parseInt(e.target.value, 10));
    closeDayPopup();
  }
  // Когда меняем год через селект
  function onYearChange(e: React.ChangeEvent<HTMLSelectElement>) {
    setCurrentYear(parseInt(e.target.value, 10));
    closeDayPopup();
  }

  function handleDayClick(dayNum: number) {
    const date = new Date(currentYear, currentMonth, dayNum);
    const daily = events.filter((ev) => isEventOnDate(ev, date));
    setDayEvents(daily);
    setSelectedDate(date);
    setPopupOpen(true);
    setShowCreateForm(false);

    const yyyy = date.getFullYear();
    const mm = String(date.getMonth() + 1).padStart(2, "0");
    const dd = String(date.getDate()).padStart(2, "0");
    setAllDay(true);
    setTitle("");
    setDescription("");
    setStartTime(`${yyyy}-${mm}-${dd}T00:00`);
    setEndTime(`${yyyy}-${mm}-${dd}T23:59`);
  }

  function closeDayPopup() {
    setPopupOpen(false);
    setSelectedDate(null);
    setDayEvents([]);
    setShowCreateForm(false);
  }

  function getDayIndicators(dayNumber: number) {
    const date = new Date(currentYear, currentMonth, dayNumber);
    const dayEvs = events.filter((ev) => isEventOnDate(ev, date));
    const intervalColors: string[] = [];
    let hasNormalEvent = false;
    dayEvs.forEach((ev) => {
      if (ev.color && ev.color.trim() !== "") {
        if (!intervalColors.includes(ev.color)) {
          intervalColors.push(ev.color);
        }
      } else {
        hasNormalEvent = true;
      }
    });
    return { intervalColors, hasNormalEvent };
  }

  function openIntervalModal() {
    setIntervalModalOpen(true);
    setIntervalTitle("");
    setIntervalDesc("");
    setIntervalStart("");
    setIntervalEnd("");
    setIntervalColor("#ff5252");
  }
  function closeIntervalModal() {
    setIntervalModalOpen(false);
  }
  async function handleIntervalSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!intervalStart || !intervalEnd) return;
    try {
      await EventService.createEvent({
        calendar_id: calendarId,
        title: intervalTitle,
        description: intervalDesc,
        start_time: new Date(intervalStart).toISOString(),
        end_time: new Date(intervalEnd).toISOString(),
        color: intervalColor,
      });
      await loadEvents();
      closeIntervalModal();
    } catch (err) {
      console.error("Ошибка создания интервального события:", err);
    }
  }

  async function handleAddDayEvent(e: React.FormEvent) {
    e.preventDefault();
    if (!selectedDate) return;
    let startIso = new Date(startTime).toISOString();
    let endIso = new Date(endTime).toISOString();
    if (allDay) {
      const s = new Date(selectedDate);
      s.setHours(0, 0, 0, 0);
      startIso = s.toISOString();
      const eDay = new Date(selectedDate);
      eDay.setHours(23, 59, 59, 999);
      endIso = eDay.toISOString();
    }
    try {
      await EventService.createEvent({
        calendar_id: calendarId,
        title,
        description,
        start_time: startIso,
        end_time: endIso,
        color: "",
      });
      await loadEvents();
      const updated = events.filter((ev) =>
        selectedDate ? isEventOnDate(ev, selectedDate) : false
      );
      setDayEvents(updated);
      setTitle("");
      setDescription("");
      setShowCreateForm(false);
    } catch (err) {
      console.error("Ошибка создания события:", err);
    }
  }

  async function handleCompleteEvent(evId: number) {
    try {
      const updated = await EventService.completeEvent(evId);
      setEvents((prev) =>
        prev.map((ev) => (ev.id === updated.id ? updated : ev))
      );
      setDayEvents((prev) =>
        prev.map((ev) => (ev.id === updated.id ? updated : ev))
      );
    } catch (err) {
      console.error("Ошибка завершения события:", err);
    }
  }

  const today = new Date();
  const todayDay = today.getDate();
  const todayMonth = today.getMonth();
  const todayYear = today.getFullYear();

  // Собираем массив ячеек календаря
  const calendarCells: Array<number | null> = [];
  for (let i = 0; i < firstDayIndex; i++) {
    calendarCells.push(null);
  }
  for (let d = 1; d <= daysInMonth; d++) {
    calendarCells.push(d);
  }

  if (loading) {
    return <div className={styles.loading}>Загрузка...</div>;
  }

  return (
    <div className={styles.calendarContainer}>
      {/* ---------- Селект месяц/год и кнопки ---------- */}
      <div className={styles.header}>
        <div className={styles.navControls}>
          <button onClick={prevMonth} className={styles.navButton}>
            ←
          </button>
          <select
            value={currentMonth}
            onChange={onMonthChange}
            className={styles.selectMonth}
          >
            {monthNames.map((m, idx) => (
              <option value={idx} key={m}>
                {m}
              </option>
            ))}
          </select>
          <select
            value={currentYear}
            onChange={onYearChange}
            className={styles.selectYear}
          >
            {yearOptions.map((y) => (
              <option value={y} key={y}>
                {y}
              </option>
            ))}
          </select>
          <button onClick={nextMonth} className={styles.navButton}>
            →
          </button>
        </div>
        <button
          onClick={openIntervalModal}
          className={styles.createIntervalBtn}
        >
          + Интервальное
        </button>
      </div>

      {error && <div className={styles.error}>{error}</div>}

      {/* ---------- Дни недели ---------- */}
      <div className={styles.weekdays}>
        {["Вс", "Пн", "Вт", "Ср", "Чт", "Пт", "Сб"].map((wd) => (
          <div key={wd} className={styles.weekdayCell}>
            {wd}
          </div>
        ))}
      </div>

      {/* ---------- Сетка календаря ---------- */}
      <div className={styles.grid}>
        {calendarCells.map((val, idx) => {
          if (val === null) {
            return <div key={idx} className={styles.emptyCell} />;
          }
          const isToday =
            val === todayDay &&
            currentMonth === todayMonth &&
            currentYear === todayYear;
          const { intervalColors, hasNormalEvent } = getDayIndicators(val);
          return (
            <div
              key={idx}
              className={`${styles.dayCell} ${isToday ? styles.today : ""}`}
              onClick={() => handleDayClick(val)}
            >
              <span className={styles.dayNumber}>{val}</span>
              {intervalColors.map((clr, i) => (
                <span
                  key={i}
                  className={styles.intervalDot}
                  style={{ backgroundColor: clr }}
                />
              ))}
              {hasNormalEvent && <span className={styles.normalDot} />}
            </div>
          );
        })}
      </div>

      {/* ---------- Модалка «Интервальное событие» ---------- */}
      {intervalModalOpen && (
        <div className={styles.modalBackdrop}>
          <div className={styles.modal}>
            <h3 className={styles.modalTitle}>
              Новое интервальное событие
            </h3>
            <form
              onSubmit={handleIntervalSubmit}
              className={styles.modalForm}
            >
              <label>
                <span>Название:</span>
                <input
                  value={intervalTitle}
                  onChange={(e) => setIntervalTitle(e.target.value)}
                  required
                  className={styles.modalInput}
                />
              </label>
              <label>
                <span>Описание:</span>
                <textarea
                  value={intervalDesc}
                  onChange={(e) => setIntervalDesc(e.target.value)}
                  className={styles.modalTextarea}
                />
              </label>
              <label>
                <span>Начало (дата/время):</span>
                <input
                  type="datetime-local"
                  value={intervalStart}
                  onChange={(e) => setIntervalStart(e.target.value)}
                  required
                  className={styles.modalInput}
                />
              </label>
              <label>
                <span>Окончание (дата/время):</span>
                <input
                  type="datetime-local"
                  value={intervalEnd}
                  onChange={(e) => setIntervalEnd(e.target.value)}
                  required
                  className={styles.modalInput}
                />
              </label>
              <label>
                <span>Цвет:</span>
                <input
                  type="color"
                  value={intervalColor}
                  onChange={(e) => setIntervalColor(e.target.value)}
                  className={styles.modalColorInput}
                />
              </label>
              <div className={styles.modalButtons}>
                <button type="submit" className={styles.saveBtn}>
                  Создать
                </button>
                <button
                  type="button"
                  onClick={closeIntervalModal}
                  className={styles.cancelBtn}
                >
                  Отмена
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* ---------- Модалка «День» (список + новая запись) ---------- */}
      {popupOpen && selectedDate && (
        <div className={styles.dayPopupBackdrop}>
          <div className={styles.dayPopupWindow}>
            <button
              className={styles.closeBtn}
              onClick={closeDayPopup}
            >
              ✕
            </button>
            <h3 className={styles.popupTitle}>
              События:{" "}
              {`${selectedDate.getDate()}.${selectedDate.getMonth() + 1}.${selectedDate.getFullYear()}`}
            </h3>
            <div className={styles.eventsScrollArea}>
              {dayEvents.length === 0 ? (
                <p className={styles.noEvents}>Событий нет</p>
              ) : (
                <ul className={styles.eventList}>
                  {dayEvents.map((ev) => (
                    <li key={ev.id} className={styles.eventItem}>
                      <div className={styles.eventHeader}>
                        <strong>{ev.title}</strong>
                        {ev.color && ev.color.trim() !== "" && (
                          <span
                            className={styles.eventColorDot}
                            style={{ backgroundColor: ev.color }}
                          />
                        )}
                      </div>
                      <div className={styles.eventTime}>
                        {`${new Date(ev.start_time).toLocaleString()} – ${new Date(
                          ev.end_time
                        ).toLocaleString()}`}
                      </div>
                      <div className={styles.eventMeta}>
                        Автор: {ev.created_by}{" "}
                        {ev.is_completed ? (
                          <span className={styles.completedLabel}>
                            (Выполнено)
                          </span>
                        ) : (
                          <button
                            onClick={() => handleCompleteEvent(ev.id)}
                            className={styles.completeBtn}
                          >
                            Завершить
                          </button>
                        )}
                      </div>
                    </li>
                  ))}
                </ul>
              )}
            </div>
            <hr className={styles.separator} />
            {!showCreateForm && (
              <div className={styles.addEventWrapper}>
                <button
                  onClick={() => setShowCreateForm(true)}
                  className={styles.addEventBtn}
                >
                  + Добавить событие
                </button>
              </div>
            )}
            {showCreateForm && (
              <>
                <h4 className={styles.subTitle}>
                  Добавить событие на этот день
                </h4>
                <form
                  onSubmit={handleAddDayEvent}
                  className={styles.modalForm}
                >
                  <label>
                    <span>Название:</span>
                    <input
                      value={title}
                      onChange={(e) => setTitle(e.target.value)}
                      required
                      className={styles.modalInput}
                    />
                  </label>
                  <label>
                    <span>Описание:</span>
                    <textarea
                      value={description}
                      onChange={(e) => setDescription(e.target.value)}
                      className={styles.modalTextarea}
                    />
                  </label>
                  <label className={styles.allDayLabel}>
                    <input
                      type="checkbox"
                      checked={allDay}
                      onChange={(e) => setAllDay(e.target.checked)}
                    />{" "}
                    Весь день
                  </label>
                  {!allDay && (
                    <>
                      <label>
                        <span>Начало:</span>
                        <input
                          type="datetime-local"
                          value={startTime}
                          onChange={(e) =>
                            setStartTime(e.target.value)
                          }
                          className={styles.modalInput}
                        />
                      </label>
                      <label>
                        <span>Окончание:</span>
                        <input
                          type="datetime-local"
                          value={endTime}
                          onChange={(e) =>
                            setEndTime(e.target.value)
                          }
                          className={styles.modalInput}
                        />
                      </label>
                    </>
                  )}
                  <div className={styles.modalButtons}>
                    <button type="submit" className={styles.saveBtn}>
                      Сохранить
                    </button>
                    <button
                      type="button"
                      onClick={() => setShowCreateForm(false)}
                      className={styles.cancelBtn}
                    >
                      Отмена
                    </button>
                  </div>
                </form>
              </>
            )}
          </div>
        </div>
      )}
    </div>
  );
});

export default CalendarPage;
