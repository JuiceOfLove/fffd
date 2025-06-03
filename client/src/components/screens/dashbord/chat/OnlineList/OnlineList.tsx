import React from 'react';
import { IUser } from '../../../../../types/auth';
import styles from './OnlineList.module.css';

interface Props { online:number[]; users:IUser[] }

const OnlineList:React.FC<Props>=({online,users})=>(
  <ul className={styles.list}>
    {users.map(u=>(
      <li key={u.id} className={online.includes(u.id)?styles.on:styles.off}>
        <span className={styles.dot}/> {u.name}
      </li>
    ))}
  </ul>
);
export default OnlineList;
