package repository

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"web_backend/internal/app/ds"
	minioClient "web_backend/internal/app/minioClient"
	"web_backend/internal/app/serializer"
)

func (r *Repository) GetDrugs() ([]ds.Drug, error) {
	var drugs []ds.Drug
	err := r.db.Where("is_deleted = ?", false).Find(&drugs).Error
	if err != nil {
		return nil, err
	}
	return drugs, nil
}

func (r *Repository) GetDrug(id int) (*ds.Drug, error) {
	var drug ds.Drug
	err := r.db.Where("drug_id = ? AND is_deleted = ?", id, false).First(&drug).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: препарат с id %d", ErrNotFound, id)
		}
		return nil, err
	}
	return &drug, nil
}

func (r *Repository) GetDrugsByTitle(title string) ([]ds.Drug, error) {
	var drugs []ds.Drug
	err := r.db.Where("(title ILIKE ? OR description ILIKE ?) AND is_deleted = ?",
		"%"+title+"%", "%"+title+"%", false).Find(&drugs).Error
	if err != nil {
		return nil, err
	}
	return drugs, nil
}

func (r *Repository) CreateDrug(j serializer.DrugJSON) (ds.Drug, error) {
	drug := serializer.DrugFromJSON(j)
	err := r.db.Create(&drug).Scan(&drug).Error
	if err != nil {
		return ds.Drug{}, err
	}
	return drug, nil
}

func (r *Repository) AddDrugPhoto(ctx *gin.Context, drugID int, file *multipart.FileHeader) (*ds.Drug, error) {
	drug, err := r.GetDrug(drugID)
	if err != nil {
		return nil, err
	}
	if drug.PhotoURL != "" {
		_ = minioClient.DeleteObject(ctx, r.mc, minioClient.GetMediaBucket(), drug.PhotoURL)
	}
	fileName, err := minioClient.UploadImage(ctx, r.mc, minioClient.GetMediaBucket(), file, drug.DrugID)
	if err != nil {
		return nil, err
	}
	if err := r.db.Model(&ds.Drug{}).Where("drug_id = ?", drugID).Update("photo_url", fileName).Error; err != nil {
		return nil, err
	}
	drug.PhotoURL = fileName
	return drug, nil
}

func (r *Repository) AddDrugVideo(ctx *gin.Context, drugID int, file *multipart.FileHeader) (*ds.Drug, error) {
	drug, err := r.GetDrug(drugID)
	if err != nil {
		return nil, err
	}
	if drug.Video != "" {
		_ = minioClient.DeleteObject(ctx, r.mc, minioClient.GetMediaBucket(), drug.Video)
	}
	fileName, err := minioClient.UploadVideo(ctx, r.mc, minioClient.GetMediaBucket(), file, drug.DrugID)
	if err != nil {
		return nil, err
	}
	if err := r.db.Model(&ds.Drug{}).Where("drug_id = ?", drugID).Update("video", fileName).Error; err != nil {
		return nil, err
	}
	drug.Video = fileName
	return drug, nil
}

func (r *Repository) DeleteDrug(id int) error {
	drug, err := r.GetDrug(id)
	if err != nil {
		return err
	}
	if drug.PhotoURL != "" {
		_ = minioClient.DeleteObject(context.Background(), r.mc, minioClient.GetMediaBucket(), drug.PhotoURL)
	}
	if drug.Video != "" {
		_ = minioClient.DeleteObject(context.Background(), r.mc, minioClient.GetMediaBucket(), drug.Video)
	}
	return r.db.Model(&ds.Drug{}).Where("drug_id = ?", id).Update("is_deleted", true).Error
}
