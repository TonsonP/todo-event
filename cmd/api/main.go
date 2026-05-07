package main

import (
	"context"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/samber/mo"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	healthadapter "todoe/internal/health/adapter"
	healthhttp "todoe/internal/health/adapter/http"
	healthapp "todoe/internal/health/application"

	taskadapter "todoe/domain/task/adapter"
	taskhttp "todoe/domain/task/adapter/http"
	taskapplication "todoe/domain/task/application"
	taskdomain "todoe/domain/task/domain"
	auditadapter "todoe/internal/audit/adapter"
	"todoe/internal/event"
	csvadapter "todoe/internal/sla/adapter"
)

func main() {
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://root:root@localhost:27017"
	}

	clientIO := mo.NewIOEither(func() (*mongo.Client, error) {
		return mongo.Connect(options.Client().ApplyURI(mongoURI))
	})

	healthRepo := healthadapter.NewMongoRepository(clientIO)
	defer healthRepo.Disconnect(context.Background())

	healthService := healthapp.NewService(healthRepo)
	healthHandler := healthhttp.NewHandler(healthService)

	bus := event.NewEventBus()
	auditRepo := auditadapter.NewMongoRepository(clientIO)
	// auditHandler := auditadapter.NewAuditHandler(auditRepo)
	auditHandler := auditadapter.NewAuditHandlerWithRetry(auditRepo, "audit_fallback.jsonl")

	mongoTaskRepo := taskadapter.NewMongoRepository(clientIO)
	taskFallbackRepo := taskadapter.NewFileFallbackRepository("task_fallback.jsonl")
	taskRepo := taskadapter.NewResilientRepository(mongoTaskRepo, taskFallbackRepo)

	taskService := taskapplication.NewService(taskRepo, bus)
	taskHandler := taskhttp.NewHandler(taskService)

	csvRepo := csvadapter.NewCSVRepository("tasks_audit.csv")
	csvHandler := csvadapter.NewCSVHandler(csvRepo)

	bus.Subscribe(taskdomain.EventCreated, auditHandler)
	bus.Subscribe(taskdomain.EventStatusChanged, auditHandler)
	bus.Subscribe(taskdomain.EventStatusCompleted, auditHandler)
	bus.Subscribe(taskdomain.EventStatusCompleted, csvHandler)
	// bus.Subscribe(taskdomain.EventStatusCompleted, taskHandler)

	app := fiber.New()
	app.Get("/health", healthHandler.CheckHealth)
	app.Post("/tasks", taskHandler.Create)
	app.Get("/tasks", taskHandler.List)
	app.Get("/tasks/:id", taskHandler.Detail)
	app.Patch("/tasks/:id/status", taskHandler.ChangeStatus)

	log.Fatal(app.Listen(":3000"))
}
