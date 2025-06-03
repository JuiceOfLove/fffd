import { useEffect, useState, useRef, useContext } from "react";
import { useParams, useNavigate } from "react-router";
import { Context } from "../../../../../main";
import SupportService, { ITicketMessage } from "../../../../../services/SupportService";
import { IChatMessage } from "../../../../../types/chat";
import SupportSocketService from "../../../../../services/SupportSocketService";
import styles from "./TicketChat.module.css";

import OnlineList   from "../../chat/OnlineList/OnlineList";
import MessageList  from "../../chat/MessageList/MessageList";
import MessageInput from "../../chat/MessageInput/MessageInput";

/**
 * Чат внутри одного тикета поддержки.
 * Логика подключения к WebSocket такая:
 *  – пользователь или admin   → сразу открываем WS;
 *  – оператор                 → открываем WS только после назначения.
 */
const TicketChat: React.FC = () => {
  const { store } = useContext(Context);
  const navigate  = useNavigate();
  const { id }    = useParams<{ id: string }>();

  /* ──────────────── guard: корректный ID ────────────────*/
  const ticketId = Number(id);
  useEffect(() => {
    if (!id || isNaN(ticketId)) {
      navigate("/dashboard/support");
    }
  }, [id, ticketId, navigate]);

  /* ──────────────── state ───────────────────────────────*/
  const [ticketInfo, setTicketInfo] = useState<SupportService.ITicketInfo|null>(null);
  const [messages,   setMessages  ] = useState<ITicketMessage[]>([]);
  const [onlineOps,  setOnlineOps ] = useState<number[]>([]);
  const [replyTo,    setReplyTo   ] = useState<ITicketMessage|null>(null);
  const [loading,    setLoading   ] = useState(true);
  const scrollRef = useRef<HTMLDivElement>(null);
  const rawToken   = localStorage.getItem("token") ?? "";

  /* ──────────────── initial load ────────────────────────*/
  useEffect(() => {
    if (!store.isAuth || !rawToken) {
      navigate("/auth/login");
      return;
    }

    let wsOpened = false;

    (async () => {
      try {
        // 1. инфо о тикете
        const info = await SupportService.getTicketInfo(ticketId);
        setTicketInfo(info);

        // 2. история сообщений
        const history = await SupportService.getTicketMessages(ticketId);
        setMessages(history);

        // 3. решаем, можно ли сразу подключаться к WS
        const isOwner      = store.user!.id === info.user_id;
        const isAdmin      = store.user!.role === "admin";
        const isThisOp     = store.user!.role === "operator" && info.operator_id === store.user!.id;

        if (isOwner || isAdmin || isThisOp) {
          SupportSocketService.connect(ticketId, rawToken);
          wsOpened = true;
          SupportSocketService.onMessage((m) => setMessages((prev) => [...prev, m]));
          SupportSocketService.onPresence(setOnlineOps);
        }
      } catch (err) {
        console.error("Ошибка при загрузке тикета", err);
        navigate("/dashboard/support");
      } finally {
        setLoading(false);
      }
    })();

    return () => {
      if (wsOpened) SupportSocketService.disconnect();
    };
  }, [ticketId]);

  /* ──────────────── автоскролл ──────────────────────────*/
  useEffect(() => {
    scrollRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  /* ──────────────── render ──────────────────────────────*/
  if (loading || !ticketInfo) {
    return <div className={styles.loading}>Загрузка тикета…</div>;
  }

  const isOperator  = store.user!.role === "operator";
  const isOwner     = store.user!.id  === ticketInfo.user_id;
  const canClose    = ticketInfo.status === "active" && (isOperator || isOwner);

  /* ——— handlers ——— */
  const handleAssign = async () => {
    try {
      await SupportService.assignTicket(ticketId);
      setTicketInfo((prev) => prev && { ...prev, status: "active", operator_id: store.user!.id });

      // открываем WS после назначения
      SupportSocketService.connect(ticketId, rawToken);
      SupportSocketService.onMessage((m) => setMessages((p) => [...p, m]));
      SupportSocketService.onPresence(setOnlineOps);
    } catch (err) {
      console.error("Не удалось взять тикет", err);
    }
  };

  const handleClose = async () => {
    try {
      await SupportService.closeTicket(ticketId);
      setTicketInfo((prev) => prev && { ...prev, status: "closed" });
      SupportSocketService.disconnect();
    } catch (err) {
      console.error("Не удалось закрыть тикет", err);
    }
  };

  const onAction = (act: "reply" | "delete", msg: IChatMessage) => {
    if (act === "reply") setReplyTo(msg as ITicketMessage);
    if (act === "delete") SupportSocketService.deleteMessage(msg.id);
  };

  return (
    <div className={styles.page}>
      {isOperator && (
        <aside className={styles.sidebar}>
          <h3 className={styles.title}>Операторы онлайн</h3>
          <OnlineList online={onlineOps} users={[]} />
        </aside>
      )}

      <section className={styles.main}>
        {/* header */}
        <div className={styles.header}>
          <h2 className={styles.subject}>{ticketInfo.subject}</h2>
          <div className={styles.controls}>
            {isOperator && ticketInfo.status === "new" && (
              <button onClick={handleAssign} className={styles.assignBtn}>Взять в работу</button>
            )}
            {canClose && (
              <button onClick={handleClose} className={styles.closeBtn}>Закрыть тикет</button>
            )}
            <span className={`${styles.status} ${styles[ticketInfo.status]}`}>
              {ticketInfo.status === "new" ? "Новый" : ticketInfo.status === "active" ? "В работе" : "Закрыт"}
            </span>
          </div>
        </div>

        {/* messages */}
        <div className={styles.messages}>
          <MessageList
            messages={messages.map((m) => ({ ...m, user_id: m.sender_id }))}
            currentUserId={store.user!.id}
            users={[]}
            onAction={onAction}
          />
          <div ref={scrollRef} />
        </div>

        {/* input */}
        {ticketInfo.status !== "closed" && (
          <MessageInput replyTo={replyTo} onCancelReply={() => setReplyTo(null)} />
        )}
      </section>
    </div>
  );
};

export default TicketChat;