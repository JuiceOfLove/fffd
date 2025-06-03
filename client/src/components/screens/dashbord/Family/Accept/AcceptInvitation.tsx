import { useEffect, useState } from "react";
import { useParams, useNavigate } from "react-router";
import FamilyService from "../../../../../services/FamilyService";
import styles from "./AcceptInvitation.module.css";

const AcceptInvitation = () => {
  const { token } = useParams<{ token: string }>();
  const navigate = useNavigate();
  const [message, setMessage] = useState("Обработка приглашения…");
  const [success, setSuccess] = useState<boolean | null>(null);

  useEffect(() => {
    const acceptInvitation = async () => {
      try {
        const response = await FamilyService.acceptInvitation(token!);
        setMessage(response.message || "Вы успешно вступили в семью");
        setSuccess(true);
        setTimeout(() => {
          navigate("/dashboard/family");
        }, 2000);
      } catch (error: any) {
        console.error("Ошибка принятия приглашения:", error);
        setMessage(error.response?.data?.error || "Ошибка при вступлении в семью");
        setSuccess(false);
      }
    };

    if (token) {
      acceptInvitation();
    }
  }, [token, navigate]);

  return (
    <div className={styles.container}>
      <div className={styles.card}>
        <h2 className={styles.msgTitle}>{message}</h2>
      </div>
    </div>
  );
};

export default AcceptInvitation;