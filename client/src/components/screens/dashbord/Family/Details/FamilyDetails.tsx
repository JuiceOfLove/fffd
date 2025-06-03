import { useState, useEffect } from "react";
import FamilyService from "../../../../../services/FamilyService";
import styles from "./FamilyDetails.module.css";

import { IFamilyResponse } from "../../../../../types/family";

const FamilyDetails = () => {
  const [data, setData] = useState<IFamilyResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [inviteEmail, setInviteEmail] = useState("");
  const [statusMsg, setStatusMsg] = useState<{ text: string; type: "success" | "error" } | null>(null);

  useEffect(() => {
    const fetchFamilyData = async () => {
      try {
        const res = await FamilyService.getFamilyDetails();
        setData(res);
      } catch (error) {
        console.error("Ошибка загрузки данных семьи", error);
      } finally {
        setLoading(false);
      }
    };
    fetchFamilyData();
  }, []);

  const handleInvite = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!inviteEmail.trim()) return;
    try {
      await FamilyService.inviteMember(inviteEmail.trim());
      setStatusMsg({ text: "Приглашение отправлено", type: "success" });
      setInviteEmail("");
    } catch (err: any) {
      setStatusMsg({ text: err.response?.data?.error || "Ошибка при отправке приглашения", type: "error" });
    }
  };

  if (loading) {
    return (
      <div className={styles.loaderContainer}>
        <div className={styles.loader}>Загрузка данных семьи…</div>
      </div>
    );
  }

  if (!data) {
    return <div className={styles.errorMsg}>Данные семьи не найдены</div>;
  }

  return (
    <div className={styles.familyContainer}>
      <h1 className={styles.familyTitle}>{data.family.name}</h1>

      <section className={styles.membersSection}>
        <h2 className={styles.sectionTitle}>Члены семьи</h2>
        <ul className={styles.memberList}>
          {data.members.map((member) => (
            <li key={member.id} className={styles.memberItem}>
              <span className={styles.memberName}>{member.name}</span>
              <span className={styles.memberEmail}>{member.email}</span>
            </li>
          ))}
        </ul>
      </section>

      <section className={styles.inviteSection}>
        <h2 className={styles.inviteTitle}>Пригласить нового члена</h2>
        <form onSubmit={handleInvite} className={styles.inviteForm}>
          <input
            type="email"
            placeholder="Email приглашённого"
            value={inviteEmail}
            onChange={(e) => setInviteEmail(e.target.value)}
            className={styles.inviteInput}
            required
          />
          <button type="submit" className={styles.inviteButton}>
            Отправить
          </button>
        </form>
        {statusMsg && (
          <p className={statusMsg.type === "success" ? styles.successMsg : styles.errorMsg}>
            {statusMsg.text}
          </p>
        )}
      </section>
    </div>
  );
};

export default FamilyDetails;