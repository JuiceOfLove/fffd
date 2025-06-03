import styles from "./DashboardHome.module.css";

const DashboardHome = () => {
  return (
    <div className={styles.container}>
      <h1>Добро пожаловать в FP Dashboard</h1>
      <p>Это ваша рабочая зона для управления семейными событиями и задачами.</p>
    </div>
  );
};

export default DashboardHome;