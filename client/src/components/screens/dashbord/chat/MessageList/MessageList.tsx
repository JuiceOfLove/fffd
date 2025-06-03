import React, { useRef, useMemo } from 'react';
import { API_URL } from '../../../../../http';
import { IChatMessage } from '../../../../../types/chat';
import { IUser } from '../../../../../types/auth';

import styles from './MessageList.module.css';

interface Props {
  messages: IChatMessage[];
  currentUserId: number;
  users: IUser[];
  onAction: (a: 'reply' | 'delete', m: IChatMessage) => void;
}

/* убираем суффикс /api → получаем origin, на котором отдаются /uploads */
const FILES_ORIGIN = API_URL.replace(/\/api\/?$/, '');

const MessageList: React.FC<Props> = ({
  messages,
  currentUserId,
  users,
  onAction,
}) => {
  /* map user_id → name */
  const nameOf = useMemo(() => {
    const map = new Map<number, string>();
    users.forEach(u => map.set(u.id, u.name));
    return (id: number) => map.get(id) ?? `ID${id}`;
  }, [users]);

  /* refs для scrollTo */
  const refs = useRef<Record<number, HTMLDivElement | null>>({});

  const scrollTo = (id: number) =>
    refs.current[id]?.scrollIntoView({ behavior: 'smooth', block: 'center' });

  /* swipe mobile */
  const swipe = useRef<{ id: number; x: number } | null>(null);
  const onStart = (m: IChatMessage) => (e: React.TouchEvent) =>
    (swipe.current = { id: m.id, x: e.touches[0].clientX });
  const onMove = (m: IChatMessage) => (e: React.TouchEvent) => {
    if (!swipe.current || swipe.current.id !== m.id) return;
    if (swipe.current.x - e.touches[0].clientX > 60) {
      onAction('reply', m);
      swipe.current = null;
    }
  };

  /* context-menu desktop */
  const context = (m: IChatMessage) => (e: React.MouseEvent) => {
    e.preventDefault();
    const menu = document.createElement('div');
    menu.className = styles.menu;
    menu.style.left = `${e.clientX}px`;
    menu.style.top = `${e.clientY}px`;

    const item = (txt: string, cb: () => void) => {
      const d = document.createElement('div');
      d.className = styles.menuItem;
      d.textContent = txt;
      d.onclick = () => {
        cb();
        menu.remove();
      };
      menu.appendChild(d);
    };
    item('Ответить', () => onAction('reply', m));
    if (m.user_id === currentUserId) item('Удалить', () => onAction('delete', m));

    const close = () => {
      menu.remove();
      window.removeEventListener('click', close);
    };
    window.addEventListener('click', close);
    document.body.appendChild(menu);
  };

  /* render */
  return (
    <>
      {messages
        .filter(m => !m.deleted_at)
        .map(m => (
          <div
            key={m.id}
            ref={el => (refs.current[m.id] = el)}
            className={`${styles.msg} ${
              m.user_id === currentUserId ? styles.mine : styles.theirs
            }`}
            onTouchStart={onStart(m)}
            onTouchMove={onMove(m)}
            onContextMenu={context(m)}
          >
            {/* имя */}
            <div className={styles.author}>{nameOf(m.user_id)}</div>

            {/* превью ответа */}
            {m.reply_to_id && (
              <div
                className={styles.reply}
                onClick={() => scrollTo(m.reply_to_id!)}
              >
                <span className={styles.replyLine} />
                <span className={styles.replyText}>
                  {messages.find(x => x.id === m.reply_to_id)?.content || '…'}
                </span>
              </div>
            )}

            {/* СНАЧАЛА картинка, ПОТОМ подпись */}
            {m.media_url && (
              <div className={styles.media}>
                <img src={FILES_ORIGIN + m.media_url} alt="attachment" />
              </div>
            )}

            {m.content && <div className={styles.text}>{m.content}</div>}

            <div className={styles.meta}>
              {new Date(m.created_at).toLocaleTimeString()}
            </div>
          </div>
        ))}
    </>
  );
};

export default MessageList;
