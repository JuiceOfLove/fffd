package controllers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"

	"diplom/config"
	"diplom/models"
)

/* ---------- Структуры для JSON (Event) ----------- */

// CreateEventInput — структура для создания нового события
type CreateEventInput struct {
	CalendarID  uint   `json:"calendar_id"` // <-- новое поле
	Title       string `json:"title"`
	Description string `json:"description"`
	StartTime   string `json:"start_time"`
	EndTime     string `json:"end_time"`
	Color       string `json:"color"`
}

// UpdateEventInput — структура для обновления существующего события
type UpdateEventInput struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	StartTime   string `json:"start_time"`
	EndTime     string `json:"end_time"`
	Color       string `json:"color"`
}

// MonthQuery — чтение query-параметров ?month=...&year=...
type MonthQuery struct {
	Month int `query:"month"`
	Year  int `query:"year"`
}

/* ---------- Handlers для Event ------------------ */

// CreateEvent создает новое событие в семье
func CreateEvent(c *fiber.Ctx) error {
    // JWT
    claims, ok := c.Locals("user").(jwt.MapClaims)
    if !ok {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Нет JWT claims"})
    }
    userIDFloat, ok := claims["user_id"].(float64)
    if !ok {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Неверный user_id"})
    }
    userID := uint(userIDFloat)

    // Находим пользователя
    var user models.User
    if err := config.DB.First(&user, userID).Error; err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Пользователь не найден"})
    }
    if user.FamilyID == 0 {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "У пользователя нет семьи"})
    }

    // Парсим входные данные
    var input CreateEventInput
    if err := c.BodyParser(&input); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Ошибка парсинга JSON"})
    }

    // Проверим, что такой календарь существует и принадлежит семье
    var cal models.Calendar
    if err := config.DB.First(&cal, input.CalendarID).Error; err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Календарь не найден"})
    }
    if cal.FamilyID != user.FamilyID {
        return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Нет доступа к календарю"})
    }

    // Конвертация времени
    start, err := time.Parse(time.RFC3339, input.StartTime)
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Некорректный формат start_time"})
    }
    end, err := time.Parse(time.RFC3339, input.EndTime)
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Некорректный формат end_time"})
    }

    event := models.Event{
        CalendarID:  input.CalendarID,
        FamilyID:    user.FamilyID,
        Title:       input.Title,
        Description: input.Description,
        StartTime:   start,
        EndTime:     end,
        CreatedBy:   userID,
        IsCompleted: false,
    }
    // Если color не пустой — сохраняем
    if input.Color != "" {
        event.Color = &input.Color
    }

    if err := config.DB.Create(&event).Error; err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Ошибка сохранения события"})
    }

    return c.JSON(fiber.Map{"event": event})
}

// UpdateEvent меняет существующее событие (например, период, цвет и т. д.)
func UpdateEvent(c *fiber.Ctx) error {
	claims, ok := c.Locals("user").(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Нет JWT claims"})
	}
	userID := uint(claims["user_id"].(float64))

	eventID := c.Params("id")
	var event models.Event
	if err := config.DB.First(&event, eventID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Событие не найдено"})
	}

	// Проверка, что пользователь принадлежит к той же семье
	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Пользователь не найден"})
	}
	if user.FamilyID != event.FamilyID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Нет доступа к событию"})
	}

	var input UpdateEventInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Ошибка парсинга JSON"})
	}

	start, err := time.Parse(time.RFC3339, input.StartTime)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Некорректный формат start_time"})
	}
	end, err := time.Parse(time.RFC3339, input.EndTime)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Некорректный формат end_time"})
	}

	event.Title = input.Title
	event.Description = input.Description
	event.StartTime = start
	event.EndTime = end

	if input.Color == "" {
		event.Color = nil
	} else {
		event.Color = &input.Color
	}

	if err := config.DB.Save(&event).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Ошибка обновления события"})
	}

	return c.JSON(fiber.Map{"event": event})
}

// GetAllEvents возвращает все события семьи (если хочется «все» сразу)
func GetAllEvents(c *fiber.Ctx) error {
	claims, ok := c.Locals("user").(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Нет JWT claims"})
	}
	userID := uint(claims["user_id"].(float64))

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Пользователь не найден"})
	}
	if user.FamilyID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "У пользователя нет семьи"})
	}

	var events []models.Event
	if err := config.DB.
		Where("family_id = ?", user.FamilyID).
		Order("start_time ASC").
		Find(&events).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Ошибка загрузки событий"})
	}

	return c.JSON(events)
}

// GetEventsForMonth — /events?month=X&year=Y
func GetEventsForMonth(c *fiber.Ctx) error {
	claims, ok := c.Locals("user").(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Нет JWT claims"})
	}
	userID := uint(claims["user_id"].(float64))

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Пользователь не найден"})
	}
	if user.FamilyID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "У пользователя нет семьи"})
	}

	month := c.QueryInt("month", 0)
	year := c.QueryInt("year", 0)
	if month < 1 || month > 12 || year < 1 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Нужны валидные month и year"})
	}

	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0) // +1 месяц

	var events []models.Event
	if err := config.DB.
		Where("family_id = ? AND start_time >= ? AND start_time < ?",
			user.FamilyID, startDate, endDate).
		Order("start_time ASC").
		Find(&events).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Ошибка при запросе"})
	}

	return c.JSON(events)
}

// CompleteEvent отмечает событие выполненным
func CompleteEvent(c *fiber.Ctx) error {
	claims, ok := c.Locals("user").(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Нет JWT claims"})
	}
	userID := uint(claims["user_id"].(float64))

	eventID := c.Params("id")
	var event models.Event
	if err := config.DB.First(&event, eventID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Событие не найдено"})
	}

	// Проверяем семью
	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Пользователь не найден"})
	}
	if user.FamilyID == 0 || user.FamilyID != event.FamilyID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Нет доступа к событию"})
	}

	event.IsCompleted = true
	if err := config.DB.Save(&event).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Ошибка обновления события"})
	}

	return c.JSON(fiber.Map{"message": "Событие выполнено", "event": event})
}

/*
================= НОВЫЕ СТРУКТУРЫ ================
   Для создания дополнительного календаря
*/

type CreateExtraCalendarInput struct {
	Title string `json:"title"`
}

// CreateExtraCalendar создает новый календарь в семье
// (требует JWT). При желании, можно проверить подписку.
func CreateExtraCalendar(c *fiber.Ctx) error {
	claims, ok := c.Locals("user").(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Нет JWT claims"})
	}
	userID := uint(claims["user_id"].(float64))

	// Ищем пользователя => familyID
	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Пользователь не найден"})
	}
	if user.FamilyID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "У пользователя нет семьи"})
	}

	// Если хотите проверять Premium — проверьте FamilySubscription

	var input CreateExtraCalendarInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Ошибка парсинга JSON"})
	}

	cal := models.Calendar{
		FamilyID: user.FamilyID,
		Title:    input.Title,
	}
	if err := config.DB.Create(&cal).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Ошибка создания календаря"})
	}

	return c.JSON(fiber.Map{"calendar": cal})
}

/*
================= СТАРЫЙ МЕТОД ===================

	GetCalendarsList — возвращает все календари семьи
*/
func GetCalendarsList(c *fiber.Ctx) error {
	claims, ok := c.Locals("user").(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Нет JWT claims"})
	}
	userID := uint(claims["user_id"].(float64))

	// Ищем пользователя
	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Пользователь не найден"})
	}
	if user.FamilyID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "У пользователя нет семьи"})
	}

	// Находим все календари семьи
	var cals []models.Calendar
	if err := config.DB.Where("family_id = ?", user.FamilyID).Find(&cals).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Ошибка загрузки календарей"})
	}

	return c.JSON(cals)
}

func GetEventsForCalendar(c *fiber.Ctx) error {
    claims, ok := c.Locals("user").(jwt.MapClaims)
    if !ok {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Нет JWT claims"})
    }
    userID := uint(claims["user_id"].(float64))

    // Находим пользователя
    var user models.User
    if err := config.DB.First(&user, userID).Error; err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Пользователь не найден"})
    }
    if user.FamilyID == 0 {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "У пользователя нет семьи"})
    }

    calendarID := c.Params("calendar_id")
    var cal models.Calendar
    if err := config.DB.First(&cal, calendarID).Error; err != nil {
        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Календарь не найден"})
    }
    if cal.FamilyID != user.FamilyID {
        return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Нет доступа к календарю"})
    }

    month := c.QueryInt("month", 0)
    year := c.QueryInt("year", 0)
    if month < 1 || month > 12 || year < 1 {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Нужны валидные month и year"})
    }

    startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
    endDate := startDate.AddDate(0, 1, 0)

    var events []models.Event
    if err := config.DB.
        Where("calendar_id = ? AND start_time >= ? AND start_time < ?", cal.ID, startDate, endDate).
        Order("start_time ASC").
        Find(&events).Error; err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Ошибка при запросе"})
    }

    return c.JSON(events)
}
