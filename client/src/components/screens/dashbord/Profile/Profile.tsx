import { useContext } from "react";
import { observer } from "mobx-react-lite";
import { Context } from "../../../../main";

import styles from "./Profile.module.css"; // Если нужен CSS-модуль, либо можно использовать обычный CSS

const Profile = () => {
  const { store } = useContext(Context);

  // Если данные о пользователе еще не загружены, можно показывать loader
  if (!store.user) {
    return <div>Загрузка профиля...</div>;
  }

  const { name, email, role, isActivated, family_id, created_at } = store.user;

  return (
    <div className={styles.profileContainer}>
      <h1>Профиль пользователя</h1>
      <div className={styles.profileItem}>
        <strong>Имя:</strong> <span>{name}</span>
      </div>
      <div className={styles.profileItem}>
        <strong>Email:</strong> <span>{email}</span>
      </div>
      <div className={styles.profileItem}>
        <strong>Роль:</strong> <span>{role}</span>
      </div>
      <div className={styles.profileItem}>
        <strong>Статус активации:</strong> <span>{isActivated ? "Активирован" : "Не активирован"}</span>
      </div>
      <div className={styles.profileItem}>
        <strong>ID семьи:</strong> <span>{family_id}</span>
      </div>
      <div className={styles.profileItem}>
        <strong>Дата регистрации:</strong>{" "}
        <span>{new Date(created_at).toLocaleString()}</span>
      </div>
      {/* Можно добавить дополнительные поля или кнопку для редактирования */}
    </div>
  );
};

export default observer(Profile);
