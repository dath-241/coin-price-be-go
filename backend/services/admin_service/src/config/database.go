package config

import (
    "context"
    "log"
    "time"
    "os"
    "fmt"
    
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

var (
    DB     *mongo.Database
    client *mongo.Client // Lưu trữ client để quản lý kết nối
)

// ConnectDatabase kết nối đến MongoDB và trả về database
func ConnectDatabase() error {

    // Lấy giá trị từ các biến môi trường
    mongoURI := os.Getenv("MONGO_URI")
    dbName := os.Getenv("MONGO_DB_NAME")
    //os.Getenv("")

    if mongoURI == "" || dbName == "" {
        log.Fatal("Required environment variables are missing!")
    }

    // Tạo context với timeout
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Tạo client MongoDB
    var err error
    client, err = mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
    if err != nil {
        return fmt.Errorf("failed to connect to MongoDB: %v", err)
    }

    // Kiểm tra kết nối
    err = client.Ping(ctx, nil)
    if err != nil {
        return err
    }

    DB = client.Database(dbName)

    log.Println("Connected to MongoDB!")
    return nil
}

// ConnectDatabaseWithRetry thử kết nối với MongoDB nhiều lần
func ConnectDatabaseWithRetry(maxRetries int, delay time.Duration) error {
    for i := 0; i < maxRetries; i++ {
        err := ConnectDatabase()
        if err != nil {
            log.Printf("Failed to connect to MongoDB (attempt %d/%d): %v", i+1, maxRetries, err)
            time.Sleep(delay) // Chờ trước khi thử lại
        } else {
            return nil // Kết nối thành công
        }
    }
    return fmt.Errorf("failed to connect to MongoDB after %d attempts", maxRetries)
}

// DisconnectDatabase đóng kết nối với MongoDB
func DisconnectDatabase() {
    if client != nil {
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()
        log.Println("Attempting to disconnect from MongoDB...")
        if err := client.Disconnect(ctx); err != nil {
            log.Printf("Error disconnecting from MongoDB: %v\n", err)
        } else {
            log.Println("Disconnected from MongoDB.")
        }
    } else {
        log.Println("Client is nil, skipping disconnect.")
    }
}