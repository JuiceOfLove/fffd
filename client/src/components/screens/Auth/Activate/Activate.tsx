import { useEffect, useState } from "react";
import { useParams, useNavigate } from "react-router";
import axios from "axios";
import { API_URL } from "../../../../http";
import styles from "./Activate.module.css";

const Activate = () => {
  const { link } = useParams<{ link: string }>();
  const navigate = useNavigate();
  const [message, setMessage] = useState("Подтверждение аккаунта...");

  useEffect(() => {
    const activateAccount = async () => {
      try {
        const response = await axios.get(`${API_URL}/auth/activate/${link}`);
        setMessage(response.data.message || "Ваш аккаунт успешно активирован");
        // Перенаправляем через 3 секунды (либо сразу отправляем на страницу логина)
        setTimeout(() => {
          navigate("/auth/login");
        }, 3000);
      } catch (error: any) {
        console.error("Ошибка активации:", error.response?.data || error.message);
        setMessage("Ошибка активации аккаунта");
      }
    };

    if (link) {
      activateAccount();
    }
  }, [link, navigate]);

  return (
    <div className={styles.container}>
      <div className={styles.card}>
        <h2>{message}</h2>
      </div>
    </div>
  );
};

export default Activate;