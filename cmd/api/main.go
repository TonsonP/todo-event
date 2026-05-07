package main

import (
	"context"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/samber/mo"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	taskhttp "todoe/domain/task/adapter/http"
	taskapplication "todoe/domain/task/application"
	taskdomain "todoe/domain/task/domain"
	taskworker "todoe/domain/task/worker"

	auditadapter "todoe/internal/audit/adapter"
	"todoe/internal/event"

	taskadapter "todoe/domain/task/adapter"
	healthadapter "todoe/internal/health/adapter"
	healthhttp "todoe/internal/health/adapter/http"
	healthapp "todoe/internal/health/application"
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

	writeBus := event.NewAppendBus()
	readBus := event.NewUpsertBus()

	taskRepo := taskadapter.NewMongoRepository(clientIO)

	taskProjector := taskworker.NewTaskProjector(readBus, taskRepo)
	taskProjector.Register(writeBus)

	taskService := taskapplication.NewService(writeBus, readBus)
	taskHandler := taskhttp.NewHandler(taskService)

	auditRepo := auditadapter.NewMongoRepository(clientIO)
	auditHandler := auditadapter.NewAuditHandler(auditRepo)

	writeBus.Subscribe(taskdomain.EventCreated, auditHandler)
	writeBus.Subscribe(taskdomain.EventStatusChanged, auditHandler)

	app := fiber.New()

	app.Get("/health", healthHandler.CheckHealth)

	app.Post("/tasks", taskHandler.Create)
	app.Get("/tasks", taskHandler.List)
	app.Get("/tasks/:id", taskHandler.Detail)
	app.Patch("/tasks/:id/status", taskHandler.ChangeStatus)
	app.Get("/debug/read-bus", func(c *fiber.Ctx) error {
		return c.JSON(readBus.All())
	})

	log.Fatal(app.Listen(":3000"))
}
