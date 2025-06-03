package routes

import (
	"diplom/controllers"
	"diplom/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

func Setup(app *fiber.App) {
	api := app.Group("/api")

	// 1. CHAT семейный WebSocket
	api.Get("/chat/ws", websocket.New(controllers.ChatWebSocket))

	// 2. CHAT HTTP + JWT
	chat := api.Group("/chat", middleware.JWTProtected())
	chat.Get("/history", controllers.ChatHistory)

	// 3. AUTH
	auth := api.Group("/auth")
	auth.Post("/register", controllers.Register)
	auth.Post("/login",    controllers.Login)
	auth.Get("/activate/:link", controllers.Activate)
	auth.Post("/refresh",  controllers.Refresh)
	auth.Post("/logout",   controllers.Logout)

	// 4. FAMILY
	family := api.Group("/family", middleware.JWTProtected())
	family.Post("/create", controllers.CreateFamily)
	family.Post("/invite", controllers.InviteMember)
	api.Get("/family/accept/:token", controllers.AcceptInvitation)
	family.Get("/details", controllers.GetFamilyDetails)

	// 5. CALENDAR
	calendar := api.Group("/calendar", middleware.JWTProtected())
	calendar.Post("/events",              controllers.CreateEvent)
	calendar.Get("/events",               controllers.GetEventsForMonth)
	calendar.Get("/events/all",           controllers.GetAllEvents)
	calendar.Post("/events/:id/complete", controllers.CompleteEvent)
	calendar.Put("/events/:id",           controllers.UpdateEvent)
	calendar.Get("/list",                 controllers.GetCalendarsList)
	calendar.Post("/create_extra",        controllers.CreateExtraCalendar)
	calendar.Get("/:calendar_id/events",  controllers.GetEventsForCalendar)

	// 6. SUBSCRIPTION
	sub := api.Group("/subscription")
	sub.Post("/webhook", controllers.YooKassaWebhook)
	subAuth := sub.Group("", middleware.JWTProtected())
	subAuth.Post("/buy",   controllers.BuySubscription)
	subAuth.Get("/check",  controllers.CheckSubscription)

	// 7. ADMIN
	admin := api.Group("/admin", middleware.JWTProtected())
	admin.Get("/payments", controllers.GetPaymentHistory)

	// 8. SUPPORT (тикеты + чат)
	support := api.Group("/support", middleware.JWTProtected())
	// 8.1. Создать новый тикет
	support.Post("/tickets", controllers.CreateTicket)
	// 8.2. Список своих тикетов
	support.Get("/tickets/my", controllers.GetMyTickets)
	// 8.3. Информация по тикету
	support.Get("/tickets/:id", controllers.GetTicketInfo)
	// 8.4. Закрыть тикет
	support.Post("/tickets/:id/close", controllers.CloseTicket)
	// 8.5. Получить все сообщения тикета
	support.Get("/tickets/:id/messages", controllers.GetTicketMessages)
	// 8.6. Удалить сообщение
	support.Delete("/tickets/:id/messages/:msgId", controllers.DeleteTicketMessageHTTP)

	// 8.7. Оператор: список тикетов по статусу
	support.Get("/tickets/operator/list", controllers.ListOperatorTickets)
	// 8.8. Оператор: взять тикет в работу
	support.Post("/tickets/:id/assign", controllers.AssignTicket)

	// 8.9. WebSocket для заметок “новые тикеты” (только операторы)
	api.Get("/support/new/ws", websocket.New(controllers.SupportNewTicketsWS))
	// 8.10. WebSocket для чата по конкретному тикету
	api.Get("/support/ws/:id", websocket.New(controllers.SupportTicketChatWS))
}
