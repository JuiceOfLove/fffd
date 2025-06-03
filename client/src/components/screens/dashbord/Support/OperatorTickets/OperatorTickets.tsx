import React, { useEffect, useState, useContext } from "react";
import { useParams, useNavigate, useLocation } from "react-router";
import SupportService from "../../../../../services/SupportService";
import { IOperatorTicket } from "../../../../../types/support";
import styles from "./OperatorTickets.module.css";
import { Context } from "../../../../../main";

const OperatorTickets: React.FC = () => {
  const { store } = useContext(Context);
  const navigate = useNavigate();
  const location = useLocation();
  const { status } = useParams<{ status: "new" | "active" | "closed" }>();
  const [tickets, setTickets] = useState<IOperatorTicket[]>([]);
  const [loading, setLoading] = useState<boolean>(true);

  // Вычисляем текущий статус из URL, по умолчанию «new»
  const currentStatus = status || "new";

  useEffect(() => {
    if (store.user?.role !== "operator") {
      navigate("/dashboard/support");
      return;
    }

    const fetchTickets = async () => {
      setLoading(true);
      try {
        const data = await SupportService.getOperatorTickets(currentStatus);
        setTickets(data);
      } catch (e) {
        console.error("Ошибка загрузки тикетов оператора", e);
      } finally {
        setLoading(false);
      }
    };

    fetchTickets();
  }, [currentStatus, store.user, navigate]);

  const handleTabChange = (stat: "new" | "active" | "closed") => {
    navigate(`/dashboard/support/operator/${stat}`);
  };

  if (loading) {
    return <div className={styles.loading}>Загрузка…</div>;
  }

  return (
    <div className={styles.container}>
      {/* Вкладки */}
      <div className={styles.tabs}>
        <button
          className={`${styles.tabBtn} ${currentStatus === "new" ? styles.active : ""}`}
          onClick={() => handleTabChange("new")}
        >
          Новые
        </button>
        <button
          className={`${styles.tabBtn} ${currentStatus === "active" ? styles.active : ""}`}
          onClick={() => handleTabChange("active")}
        >
          В работе
        </button>
        <button
          className={`${styles.tabBtn} ${currentStatus === "closed" ? styles.active : ""}`}
          onClick={() => handleTabChange("closed")}
        >
          Закрытые
        </button>
      </div>

      {/* Список тикетов оператора */}
      {tickets.length === 0 ? (
        <div className={styles.empty}>Нет тикетов</div>
      ) : (
        <ul className={styles.list}>
          {tickets.map((t) => (
            <li
              key={t.id}
              className={styles.ticketItem}
              onClick={() => navigate(`/dashboard/support/operator/ticket/${t.id}`)}
            >
              <div className={styles.subject}>{t.subject}</div>
              <div className={styles.meta}>
                <span className={styles.userName}>Пользователь: {t.user_name}</span>
                <span className={styles.date}>
                  {new Date(t.last_message_at).toLocaleString()}
                </span>
              </div>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
};

export default OperatorTickets;