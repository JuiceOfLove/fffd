import { useState } from "react";
import { useNavigate } from "react-router";
import FamilyService from "../../../../../services/FamilyService";
import styles from "./CreateFamily.module.css";

const CreateFamily = () => {
  const [familyName, setFamilyName] = useState("");
  const [inviteEmails, setInviteEmails] = useState<string[]>([]);
  const [message, setMessage] = useState<{ text: string; type: "success" | "error" } | null>(null);
  const navigate = useNavigate();

  const handleAddEmailInput = () => {
    setInviteEmails([...inviteEmails, ""]);
  };

  const handleEmailChange = (index: number, value: string) => {
    const newEmails = [...inviteEmails];
    newEmails[index] = value;
    setInviteEmails(newEmails);
  };

  const handleRemoveEmail = (index: number) => {
    const newEmails = inviteEmails.filter((_, i) => i !== index);
    setInviteEmails(newEmails);
  };

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    try {
      const familyResponse = await FamilyService.createFamily(familyName);
      setMessage({ text: "Семья создана успешно", type: "success" });

      if (inviteEmails.length > 0) {
        for (const email of inviteEmails) {
          if (email.trim() !== "") {
            await FamilyService.inviteMember(email.trim());
          }
        }
        setMessage({ text: "Семья создана и приглашения отправлены", type: "success" });
      }

      setTimeout(() => {
        navigate("/dashboard/family");
      }, 2000);
    } catch (error: any) {
      console.error("Ошибка создания семьи:", error);
      setMessage({ text: error.response?.data?.error || "Ошибка создания семьи", type: "error" });
    }
  };

  return (
    <div className={styles.container}>
      <div className={styles.card}>
        <h2 className={styles.title}>Создание семьи</h2>
        <p className={styles.subtitle}>
          Введите название семьи и добавьте email для приглашения.
        </p>
        <form onSubmit={handleSubmit} className={styles.form}>
          <div className={styles.inputGroup}>
            <label htmlFor="familyName">Название семьи</label>
            <input
              id="familyName"
              type="text"
              placeholder="Название семьи"
              value={familyName}
              onChange={(e) => setFamilyName(e.target.value)}
              required
            />
          </div>

          <div className={styles.inviteContainer}>
            <h3 className={styles.inviteTitle}>Пригласить членов семьи</h3>
            {inviteEmails.map((email, index) => (
              <div key={index} className={styles.inviteGroup}>
                <input
                  type="email"
                  placeholder="Email приглашённого"
                  value={email}
                  onChange={(e) => handleEmailChange(index, e.target.value)}
                  required
                />
                <button
                  type="button"
                  className={styles.removeButton}
                  onClick={() => handleRemoveEmail(index)}
                >
                  &times;
                </button>
              </div>
            ))}
            <button
              type="button"
              className={styles.addButton}
              onClick={handleAddEmailInput}
            >
              + Добавить ещё
            </button>
          </div>

          <button type="submit" className={styles.submitButton}>
            Создать семью
          </button>
        </form>

        {message && (
          <p className={message.type === "success" ? styles.message : styles.errorMsg}>
            {message.text}
          </p>
        )}
      </div>
    </div>
  );
};

export default CreateFamily;