package repositories

import (
	"nekozanedex/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ChapterRepository interface {
	Create(chapter *models.Chapter) error
	FindByID(id uuid.UUID) (*models.Chapter, error)
	FindByStoryAndNumber(storyID uuid.UUID, chapterNumber int) (*models.Chapter, error)
	Update(chapter *models.Chapter) error
	Delete(id uuid.UUID) error
	GetByStory(storyID uuid.UUID, published bool) ([]models.Chapter, error)
	IncrementViewCount(id uuid.UUID) error
	GetScheduledChapters() ([]models.Chapter, error)
}

type chapterRepository struct {
	db *gorm.DB
}

//New Chapter Repository - Tạo Chapter Repository
func NewChapterRepository(db *gorm.DB) ChapterRepository {
	return &chapterRepository{db: db}
}

//Create Chapter - Tạo Chapter
func (r *chapterRepository) Create(chapter *models.Chapter) error {
	return r.db.Create(chapter).Error
}

//Find Chapter By ID - Tìm Chapter Theo ID
func (r *chapterRepository) FindByID(id uuid.UUID) (*models.Chapter, error) {
	var chapter models.Chapter
	err := r.db.Preload("Story").First(&chapter, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &chapter, nil
}

//Find Chapter By Story And Number - Tìm Chapter Theo Story Và Số Trang
func (r *chapterRepository) FindByStoryAndNumber(storyID uuid.UUID, chapterNumber int) (*models.Chapter, error) {
	var chapter models.Chapter
	err := r.db.First(&chapter, "story_id = ? AND chapter_number = ? AND is_published = ?", 
		storyID, chapterNumber, true).Error
	if err != nil {
		return nil, err
	}
	return &chapter, nil
}

//Update Chapter - Cập Nhật Chapter
func (r *chapterRepository) Update(chapter *models.Chapter) error {
	return r.db.Save(chapter).Error
}

//Delete Chapter - Xóa Chapter
func (r *chapterRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Chapter{}, "id = ?", id).Error
}

//Get By Story - Lấy Theo Story
func (r *chapterRepository) GetByStory(storyID uuid.UUID, published bool) ([]models.Chapter, error) {
	var chapters []models.Chapter
	query := r.db.Where("story_id = ?", storyID)
	if published {
		query = query.Where("is_published = ?", true)
	}
	err := query.Order("chapter_number ASC").Find(&chapters).Error
	return chapters, err
}

//Increment View Count - Tăng Lượt Xem
func (r *chapterRepository) IncrementViewCount(id uuid.UUID) error {
	return r.db.Model(&models.Chapter{}).Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

//Get Scheduled Chapters - Lấy Chương Được Lên Kệ
func (r *chapterRepository) GetScheduledChapters() ([]models.Chapter, error) {
	var chapters []models.Chapter
	err := r.db.Where("is_published = ? AND scheduled_at <= NOW()", false).Find(&chapters).Error
	return chapters, err
}