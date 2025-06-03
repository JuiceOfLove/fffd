import { useEffect, useState } from "react";
import AdminService from "../../../../services/AdminService";
import { IPayment } from "../../../../types/payment";
import styles from "./AdminPanel.module.css";

const AdminPanel = () => {
  const [payments, setPayments] = useState<IPayment[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchPayments = async () => {
      try {
        const data = await AdminService.getPaymentHistory();
        setPayments(data);
      } catch (err) {
        console.error("Ошибка загрузки транзакций:", err);
      } finally {
        setLoading(false);
      }
    };

    fetchPayments();
  }, []);

  if (loading) {
    return <div>Загрузка истории транзакций...</div>;
  }

  return (
    <div className={styles.adminContainer}>
      <h1>История транзакций</h1>
      {payments.length === 0 ? (
        <p>Транзакций не найдено</p>
      ) : (
        <table className={styles.table}>
          <thead>
            <tr>
              <th>ID</th>
              <th>Payment ID</th>
              <th>Семья</th>
              <th>Пользователь</th>
              <th>Сумма</th>
              <th>Статус</th>
              <th>Дата создания</th>
            </tr>
          </thead>
          <tbody>
            {payments.map((payment) => (
              <tr key={payment.id}>
                <td>{payment.id}</td>
                <td>{payment.payment_id}</td>
                <td>{payment.family_id}</td>
                <td>{payment.user_id}</td>
                <td>{payment.amount}</td>
                <td>{payment.status}</td>
                <td>{new Date(payment.created_at).toLocaleString()}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
};

export default AdminPanel;
