import { Outlet, Link } from 'react-router';
import { useContext } from 'react';
import { observer } from 'mobx-react-lite';
import { Context } from '../../../main';
import styles from './Layout.module.css';

const Layout = () => {
  const { store } = useContext(Context);

  return (
    <div className={styles.container}>
      <header className={styles.header}>
        <div className={styles.logo}>
            <Link to="/">Lumivy</Link>
        </div>
        <nav className={styles.nav}>
          {store.isAuth ? (
            <>
              {/* Если авторизован – ссылка на дашборд */}
              <Link to="/dashboard" className={styles.navLink}>
                Платформа
              </Link>
              <span
                className={styles.navLink}
                onClick={() => store.logout()}
                style={{ cursor: 'pointer' }}
              >
                Выйти
              </span>
            </>
          ) : (
            <>
              <Link to="/auth/login" className={styles.navLink}>Войти</Link>
              <Link to="/auth/register" className={styles.navLink}>Регистрация</Link>
            </>
          )}
        </nav>
      </header>
      <main className={styles.main}>
        <Outlet />
      </main>
      <footer className={styles.footer}>
        <p>© 2025 Lumivy. Все права защищены.</p>
      </footer>
    </div>
  );
};

export default observer(Layout);