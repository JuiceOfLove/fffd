import React, { useEffect, useState, useContext } from "react";
import { IMyTicket } from "../../../../../types/support";
import SupportService from "../../../../../services/SupportService";
import { useNavigate } from "react-router";
import styles from "./MyTicketsList.module.css";
import { Context } from "../../../../../main";

const MyTicketsList: React.FC = () => {
  const { store } = useContext(Context);
  const [tickets, setTickets] = useState<IMyTicket[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const navigate = useNavigate();

  useEffect(() => {
    (async () => {
      try {
        const data = await SupportService.getMyTickets();
        setTickets(data);
      } catch (e) {
        console.error("Ошибка при загрузке моих тикетов", e);
      } finally {
        setLoading(false);
      }
    })();
  }, []);

  if (loading) {
    return <div className={styles.loading}>Загрузка…</div>;
  }

  return (
    <div className={styles.container}>
      {/* Кнопка «Новый тикет» */}
      <div className={styles.header}>
        <h2 className={styles.title}>Мои тикеты</h2>
        <button
          className={styles.newButton}
          onClick={() => navigate("/dashboard/support/new")}
        >
          + Новый тикет
        </button>
      </div>

      {tickets.length === 0 ? (
        <div className={styles.empty}>У вас пока нет тикетов поддержки</div>
      ) : (
        <ul className={styles.list}>
          {tickets.map((t) => (
            <li
              key={t.id}
              className={`${styles.ticketItem} ${styles[t.status]}`}
              onClick={() => navigate(`/dashboard/support/ticket/${t.id}`)}
            >
              <div className={styles.subject}>{t.subject}</div>
              <div className={styles.meta}>
                <span className={styles.date}>
                  {new Date(t.last_message_at).toLocaleString()}
                </span>
                <span className={styles.status}>{t.status_label}</span>
              </div>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
};

export default MyTicketsList;