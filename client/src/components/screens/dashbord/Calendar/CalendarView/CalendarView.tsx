import React, { useEffect, useState } from "react";
import { observer } from "mobx-react-lite";
import EventService from "../../../../../services/EventService";
import { IEvent } from "../../../../../types/EventTypes";
import styles from "./CalendarView.module.css";

function getDaysInMonth(month: number, year: number) {
  return new Date(year, month + 1, 0).getDate();
}
function getFirstDayOfMonth(month: number, year: number) {
  return new Date(year, month, 1).getDay();
}

const CalendarView: React.FC = observer(() => {
  // 0-11
  const [currentMonth, setCurrentMonth] = useState<number>(new Date().getMonth());
  const [currentYear, setCurrentYear] = useState<number>(new Date().getFullYear());
  const [events, setEvents] = useState<IEvent[]>([]);
  const [selectedDay, setSelectedDay] = useState<number | null>(null);
  const [dayEvents, setDayEvents] = useState<IEvent[]>([]);

  useEffect(() => {
    (async function loadEvents() {
      try {
        // currentMonth + 1 => 1..12
        const data = await EventService.getEventsForMonth(currentMonth + 1, currentYear);
        setEvents(data);
      } catch (err) {
        console.error("Ошибка загрузки событий:", err);
      }
    })();
  }, [currentMonth, currentYear]);

  const hasEvents = (day: number) => {
    // находим все события, у которых day/month/year совпадает
    const dateString = new Date(currentYear, currentMonth, day).toDateString();
    return events.some(ev => {
      const evDate = new Date(ev.start_time).toDateString();
      return evDate === dateString;
    });
  };

  const getEventsForDay = (day: number) => {
    const dateString = new Date(currentYear, currentMonth, day).toDateString();
    return events.filter(ev => {
      const evDate = new Date(ev.start_time).toDateString();
      return evDate === dateString;
    });
  };

  const handleDayClick = (day: number) => {
    setSelectedDay(day);
    setDayEvents(getEventsForDay(day));
  };

  const prevMonth = () => {
    let newMonth = currentMonth - 1;
    let newYear = currentYear;
    if (newMonth < 0) {
      newMonth = 11;
      newYear = currentYear - 1;
    }
    setCurrentMonth(newMonth);
    setCurrentYear(newYear);
    setSelectedDay(null);
  };

  const nextMonth = () => {
    let newMonth = currentMonth + 1;
    let newYear = currentYear;
    if (newMonth > 11) {
      newMonth = 0;
      newYear = currentYear + 1;
    }
    setCurrentMonth(newMonth);
    setCurrentYear(newYear);
    setSelectedDay(null);
  };

  // для подсветки сегодняшнего дня
  const today = new Date();
  const todayDay = today.getDate();
  const todayMonth = today.getMonth();
  const todayYear = today.getFullYear();

  const daysInMonth = getDaysInMonth(currentMonth, currentYear);
  const firstDayIndex = getFirstDayOfMonth(currentMonth, currentYear); // 0=вс

  const calendarCells: Array<number | null> = [];
  for (let i = 0; i < firstDayIndex; i++) {
    calendarCells.push(null);
  }
  for (let d = 1; d <= daysInMonth; d++) {
    calendarCells.push(d);
  }

  // завершить событие
  const handleCompleteEvent = async (eventId: number) => {
    try {
      await EventService.completeEvent(eventId);
      // Перезагружаем события
      const data = await EventService.getEventsForMonth(currentMonth + 1, currentYear);
      setEvents(data);

      // Обновляем список dayEvents, чтобы сразу отобразилось
      if (selectedDay) {
        setDayEvents(getEventsForDay(selectedDay));
      }
    } catch (err) {
      console.error("Ошибка завершения события:", err);
    }
  };

  return (
    <div className={styles.calendarContainer}>
      <div className={styles.header}>
        <button onClick={prevMonth}>{"<"}</button>
        <span className={styles.monthLabel}>
          {currentYear}-{String(currentMonth + 1).padStart(2, "0")}
        </span>
        <button onClick={nextMonth}>{">"}</button>
      </div>

      <div className={styles.weekdays}>
        <div>Sun</div>
        <div>Mon</div>
        <div>Tue</div>
        <div>Wed</div>
        <div>Thu</div>
        <div>Fri</div>
        <div>Sat</div>
      </div>

      <div className={styles.grid}>
        {calendarCells.map((val, idx) => {
          if (val === null) {
            return <div key={idx} className={styles.emptyCell} />;
          }
          const isToday =
            val === todayDay && currentMonth === todayMonth && currentYear === todayYear;
          const dayHasEvents = hasEvents(val);

          return (
            <div
              key={idx}
              className={`${styles.dayCell} ${isToday ? styles.today : ""}`}
              onClick={() => handleDayClick(val!)}
            >
              <span>{val}</span>
              {dayHasEvents && <div className={styles.eventDot} />}
            </div>
          );
        })}
      </div>

      {selectedDay && (
        <div className={styles.eventList}>
          <h3>
            События {selectedDay}.{currentMonth + 1}.{currentYear}
          </h3>
          {dayEvents.length === 0 ? (
            <p>Нет событий</p>
          ) : (
            <ul>
              {dayEvents.map(ev => (
                <li key={ev.id}>
                  <strong>{ev.title}</strong>{" "}
                  <span>
                    ({new Date(ev.start_time).toLocaleTimeString()} –{" "}
                    {new Date(ev.end_time).toLocaleTimeString()})
                  </span>
                  {" | "}
                  {ev.is_completed ? (
                    <span style={{ color: "green" }}>Выполнено</span>
                  ) : (
                    <button onClick={() => handleCompleteEvent(ev.id)}>
                      Выполнить
                    </button>
                  )}
                </li>
              ))}
            </ul>
          )}
        </div>
      )}
    </div>
  );
});

export default CalendarView;