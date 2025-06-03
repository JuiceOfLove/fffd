import React, { useEffect, useRef, useState, useContext } from 'react';
import $api from '../../../../../http';
import { Context } from '../../../../../main';
import { IChatMessage } from '../../../../../types/chat';
import { IUser } from '../../../../../types/auth';
import FamilyService from '../../../../../services/FamilyService';
import ChatService from '../../../../../services/ChatService';

import OnlineList   from '../OnlineList/OnlineList';
import MessageList  from '../MessageList/MessageList';
import MessageInput from '../MessageInput/MessageInput';

import styles from './ChatPage.module.css';

const ChatPage: React.FC = () => {
  const { store } = useContext(Context);
  const [members,  setMembers]  = useState<IUser[]>([]);
  const [messages, setMessages] = useState<IChatMessage[]>([]);
  const [online,   setOnline]   = useState<number[]>([]);
  const [replyTo,  setReplyTo]  = useState<IChatMessage|null>(null);
  const scrollRef = useRef<HTMLDivElement>(null);

  /* init */
  useEffect(() => {
    FamilyService.getFamilyDetails().then(r => setMembers(r.members));
    $api.get<IChatMessage[]>('/chat/history').then(r => setMessages(r.data));

    const token = localStorage.getItem('token');
    token && ChatService.connect(token);

    ChatService.onMessage(m   => setMessages(p => [...p, m]));
    ChatService.onDelete (id  =>
      setMessages(p => p.map(m => m.id===id ? {...m, deleted_at: new Date().toISOString()} as any : m))
    );
    ChatService.onPresence(setOnline);

    return () => ChatService.disconnect();
  }, []);

  /* авто-проскролл к последнему */
  useEffect(() => {
    scrollRef.current?.scrollIntoView({behavior:'smooth'});
  }, [messages]);

  if (!store.user) return <div className={styles.loading}>Loading…</div>;

  return (
    <div className={styles.page}>
      <aside className={styles.sidebar}>
        <h3 className={styles.title}>Семейный чат</h3>
        <OnlineList online={online} users={members}/>
      </aside>

      <section className={styles.main}>
        <div className={styles.messages}>
          <MessageList
            messages={messages}
            currentUserId={store.user.id}
            users={members}
            onAction={(act,msg)=>{
              if (act==='reply')  setReplyTo(msg);
              if (act==='delete') ChatService.delete(msg.id);
            }}
          />
          <div ref={scrollRef}/>
        </div>

        <MessageInput
          replyTo={replyTo}
          onCancelReply={()=>setReplyTo(null)}
        />
      </section>
    </div>
  );
};

export default ChatPage;
