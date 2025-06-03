import React from "react";
import { Link } from "react-router";
import styles from "./PaymentSuccess.module.css";

const PaymentSuccess: React.FC = () => {
  return (
    <div className={styles.successContainer}>
      <h2>Оплата прошла успешно!</h2>
      <p>Спасибо за покупку подписки.</p>
      <p>Вы можете вернуться на <Link to="/dashboard">главную страницу</Link>.</p>
    </div>
  );
};

export default PaymentSuccess;
