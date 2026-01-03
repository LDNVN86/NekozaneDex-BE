package database

import (
	"fmt"
	"log"
	"nekozanedex/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func ConnectDB(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable Timezone=Asia/Ho-_Chi_Minh", cfg.Host, cfg.User, cfg.Password, cfg.DBName, cfg.Port)	
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil,fmt.Errorf("Kết Nối Database Thất Bại: %w", err)
	}
	log.Println("Đã Kết Nối Database Thành Công")
	return db, nil
}