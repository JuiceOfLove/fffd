import { useState, useContext, useEffect } from "react";
import { useNavigate } from "react-router";
import { observer } from "mobx-react-lite";
import { Context } from "../../../../main";
import styles from "./Login.module.css";

const Login = () => {
  const { store } = useContext(Context);
  const navigate = useNavigate();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");

  useEffect(() => {
    if (store.isAuth) {
      navigate("/dashboard");
    }
  }, [store.isAuth, navigate]);

  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    await store.login(email, password);
    // Если успешный логин, navigate сработает в useEffect
  };

  return (
    <div className={styles.container}>
      <div className={styles.card}>
        <h2 className={styles.title}>Вход</h2>
        <p className={styles.subtitle}>Добро пожаловать, управляйте, контролируйте, выполняйте</p>
        <form onSubmit={onSubmit} className={styles.form}>
          <div className={styles.inputGroup}>
            <label htmlFor="email">Email</label>
            <input
              id="email"
              type="email"
              placeholder="mail@example.com"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
            />
          </div>
          <div className={styles.inputGroup}>
            <label htmlFor="password">Пароль</label>
            <input
              id="password"
              type="password"
              placeholder="Введите пароль"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
            />
          </div>
          <button type="submit" className={styles.button}>Войти</button>
        </form>
      </div>
    </div>
  );
};

export default observer(Login);