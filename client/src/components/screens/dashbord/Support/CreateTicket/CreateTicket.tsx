import { useState } from "react";
import { useNavigate } from "react-router";
import SupportService from "../../../../../services/SupportService";
import styles from "./CreateTicket.module.css";

const CreateTicket: React.FC = () => {
  const [subject, setSubject] = useState("");
  const [message, setMessage] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError]   = useState<string | null>(null);
  const navigate            = useNavigate();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!subject.trim() || !message.trim()) return;

    try {
      setLoading(true);
      setError(null);

      const { ticket_id } = await SupportService.createTicket({
        subject,
        content: message,
      });

      navigate(`/dashboard/support/ticket/${ticket_id}`);
    } catch (err: any) {
      const msg = err.response?.data?.error ?? "Ошибка при создании тикета";
      setError(msg);
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className={styles.container}>
      <h2 className={styles.title}>Новый тикет в поддержку</h2>

      <form onSubmit={handleSubmit} className={styles.form}>
        <label className={styles.label}>
          Тема тикета:
          <input
            className={styles.input}
            placeholder="Кратко опишите проблему"
            required
            value={subject}
            onChange={e => setSubject(e.target.value)}
          />
        </label>

        <label className={styles.label}>
          Сообщение:
          <textarea
            className={styles.textarea}
            placeholder="Опишите детали…"
            required
            value={message}
            onChange={e => setMessage(e.target.value)}
          />
        </label>

        {error && <div className={styles.error}>{error}</div>}

        <button className={styles.submitBtn} disabled={loading}>
          {loading ? "Отправка…" : "Создать тикет"}
        </button>
      </form>
    </div>
  );
};

export default CreateTicket;
