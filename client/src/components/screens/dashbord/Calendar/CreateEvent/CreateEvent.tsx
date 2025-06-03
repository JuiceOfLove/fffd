import React, { useState } from "react";
import { observer } from "mobx-react-lite";
import { useNavigate } from "react-router";
import EventService from "../../../../../services/EventService";

const CreateEvent: React.FC = observer(() => {
  const navigate = useNavigate();
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [startTime, setStartTime] = useState("");
  const [endTime, setEndTime] = useState("");

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await EventService.createEvent({
        title,
        description,
        start_time: new Date(startTime).toISOString(),
        end_time: new Date(endTime).toISOString(),
      });
      navigate("/dashboard/calendar");
    } catch (err) {
      console.error("Ошибка создания события:", err);
    }
  };

  return (
    <div>
      <h2>Создать событие</h2>
      <form onSubmit={handleSubmit}>
        <div>
          <label>Название:</label>
          <input
            value={title}
            onChange={e => setTitle(e.target.value)}
            required
          />
        </div>
        <div>
          <label>Описание:</label>
          <textarea
            value={description}
            onChange={e => setDescription(e.target.value)}
          />
        </div>
        <div>
          <label>Начало:</label>
          <input
            type="datetime-local"
            onChange={e => setStartTime(e.target.value)}
            required
          />
        </div>
        <div>
          <label>Окончание:</label>
          <input
            type="datetime-local"
            onChange={e => setEndTime(e.target.value)}
            required
          />
        </div>
        <button type="submit">Создать</button>
      </form>
    </div>
  );
});

export default CreateEvent;