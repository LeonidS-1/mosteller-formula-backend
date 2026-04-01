package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"web_backend/internal/app/ds"
	"web_backend/internal/app/serializer"
)

// CalculatePediatricDose вычисляет детскую дозу по формуле Mosteller (площадь поверхности тела).
func CalculatePediatricDose(heightCm, weightKg int, dosePerM2Mg, maxDailyMg float64) float64 {
	if heightCm <= 0 || weightKg <= 0 {
		return 0
	}
	bsa := math.Sqrt(float64(heightCm*weightKg) / 3600.0)
	dose := bsa * dosePerM2Mg
	if dose > maxDailyMg {
		dose = maxDailyMg
	}
	return math.Round(dose*10) / 10
}

const (
	defaultDraftHeightCm = 120
	defaultDraftWeightKg = 22
	draftDoctorName      = "Смирнова Анна Владимировна"
)

func (r *Repository) GetPrescriptionDrugCount(creatorID uint) int64 {
	var prescriptionID uint
	err := r.db.Model(&ds.Prescription{}).
		Where("creator_id = ? AND status = ?", creatorID, "draft").
		Select("prescription_id").First(&prescriptionID).Error
	if err != nil {
		return 0
	}

	var count int64
	err = r.db.Model(&ds.PrescriptionDrug{}).
		Where("prescription_id = ?", prescriptionID).Count(&count).Error
	if err != nil {
		logrus.Error("prescription_drugs count:", err)
	}
	return count
}

func (r *Repository) GetActivePrescriptionID(creatorID uint) uint {
	var prescriptionID uint
	err := r.db.Model(&ds.Prescription{}).
		Where("creator_id = ? AND status = ?", creatorID, "draft").
		Select("prescription_id").First(&prescriptionID).Error
	if err != nil {
		return 0
	}
	return prescriptionID
}

func (r *Repository) CheckCurrentDraft(creatorID uint) (ds.Prescription, error) {
	if creatorID == 0 {
		return ds.Prescription{}, ErrNotAllowed
	}
	var p ds.Prescription
	res := r.db.Where("creator_id = ? AND status = ?", creatorID, "draft").Limit(1).Find(&p)
	if res.Error != nil {
		return ds.Prescription{}, res.Error
	}
	if res.RowsAffected == 0 {
		return ds.Prescription{}, ErrNoDraft
	}
	return p, nil
}

func (r *Repository) GetPrescriptionDraft(creatorID uint) (ds.Prescription, bool, error) {
	p, err := r.CheckCurrentDraft(creatorID)
	if errors.Is(err, ErrNoDraft) {
		p = ds.Prescription{
			Status:         "draft",
			CreatedAt:      time.Now(),
			CreatorID:      creatorID,
			DoctorFullName: draftDoctorName,
		}
		if err := r.db.Create(&p).Error; err != nil {
			return ds.Prescription{}, false, err
		}
		return p, true, nil
	}
	if err != nil {
		return ds.Prescription{}, false, err
	}
	return p, false, nil
}

func (r *Repository) GetModeratorAndCreatorLogin(p ds.Prescription) (string, string, error) {
	var creator ds.Users
	if err := r.db.Where("user_id = ?", p.CreatorID).First(&creator).Error; err != nil {
		return "", "", err
	}
	var moderatorLogin string
	if p.ModeratorID != nil && *p.ModeratorID != 0 {
		var moderator ds.Users
		if err := r.db.Where("user_id = ?", *p.ModeratorID).First(&moderator).Error; err != nil {
			return "", "", err
		}
		moderatorLogin = moderator.Login
	}
	return creator.Login, moderatorLogin, nil
}

func (r *Repository) GetCompletedDoseLineCount(prescriptionID uint) (int, error) {
	var count int64
	err := r.db.Model(&ds.PrescriptionDrug{}).
		Where("prescription_id = ? AND pediatric_dose_mg IS NOT NULL", prescriptionID).
		Count(&count).Error
	return int(count), err
}

func (r *Repository) GetAllPrescriptions(from, to time.Time, status string) ([]ds.Prescription, error) {
	var list []ds.Prescription
	sub := r.db.Where("status != ? AND status != ?", "deleted", "draft")
	if !from.IsZero() {
		sub = sub.Where("forming_date >= ?", from)
	}
	if !to.IsZero() {
		sub = sub.Where("forming_date < ?", to.Add(time.Hour*24))
	}
	if status != "" {
		sub = sub.Where("status = ?", status)
	}
	err := sub.Order("prescription_id").Find(&list).Error
	return list, err
}

func (r *Repository) GetSinglePrescription(id int) (ds.Prescription, error) {
	if id < 0 {
		return ds.Prescription{}, errors.New("неверное id")
	}
	var p ds.Prescription
	err := r.db.Where("prescription_id = ?", id).First(&p).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ds.Prescription{}, fmt.Errorf("%w: рецепт с id %d", ErrNotFound, id)
		}
		return ds.Prescription{}, err
	}
	if p.Status == "deleted" {
		return ds.Prescription{}, fmt.Errorf("%w: рецепт удалён", ErrNotAllowed)
	}
	return p, nil
}

func (r *Repository) GetPrescriptionItems(prescriptionID int) ([]ds.PrescriptionDrug, error) {
	var items []ds.PrescriptionDrug
	err := r.db.Where("prescription_id = ?", prescriptionID).
		Preload("Drug").
		Find(&items).Error
	return items, err
}

func (r *Repository) AddDrugToPrescription(drugID uint, creatorID uint) error {
	var rx ds.Prescription

	err := r.db.Where("creator_id = ? AND status = ?", creatorID, "draft").First(&rx).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		rx = ds.Prescription{
			Status:         "draft",
			CreatedAt:      time.Now(),
			CreatorID:      creatorID,
			DoctorFullName: draftDoctorName,
		}
		if err := r.db.Create(&rx).Error; err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	var drug ds.Drug
	if err := r.db.Where("drug_id = ? AND is_deleted = ?", drugID, false).First(&drug).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("%w: препарат с id %d", ErrNotFound, drugID)
		}
		return err
	}

	var count int64
	r.db.Model(&ds.PrescriptionDrug{}).
		Where("prescription_id = ? AND drug_id = ?", rx.PrescriptionID, drugID).
		Count(&count)

	if count > 0 {
		return fmt.Errorf("%w: препарат %d уже в рецепте %d", ErrAlreadyExists, drugID, rx.PrescriptionID)
	}

	dose := CalculatePediatricDose(defaultDraftHeightCm, defaultDraftWeightKg, drug.DosePerM2Mg, drug.MaxDailyMg)
	row := ds.PrescriptionDrug{
		PrescriptionID:  rx.PrescriptionID,
		DrugID:          drugID,
		HeightCm:        defaultDraftHeightCm,
		WeightKg:        defaultDraftWeightKg,
		PediatricDoseMg: &dose,
	}
	return r.db.Create(&row).Error
}

func (r *Repository) DeleteDrugFromPrescription(prescriptionID, drugID int) (ds.Prescription, error) {
	var rx ds.Prescription
	err := r.db.Where("prescription_id = ?", prescriptionID).First(&rx).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ds.Prescription{}, fmt.Errorf("%w: рецепт с id %d", ErrNotFound, prescriptionID)
		}
		return ds.Prescription{}, err
	}
	if rx.Status != "draft" {
		return ds.Prescription{}, fmt.Errorf("%w: можно удалять только из черновика", ErrNotAllowed)
	}
	err = r.db.Where("prescription_id = ? AND drug_id = ?", prescriptionID, drugID).
		Delete(&ds.PrescriptionDrug{}).Error
	if err != nil {
		return ds.Prescription{}, err
	}
	return rx, nil
}

func (r *Repository) EditDrugInPrescription(prescriptionID, drugID int, j serializer.PrescriptionDrugJSON) (ds.PrescriptionDrug, error) {
	var item ds.PrescriptionDrug
	err := r.db.Where("prescription_id = ? AND drug_id = ?", prescriptionID, drugID).
		First(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ds.PrescriptionDrug{}, fmt.Errorf("%w: препарат в рецепте", ErrNotFound)
		}
		return ds.PrescriptionDrug{}, err
	}

	var rx ds.Prescription
	if err := r.db.Where("prescription_id = ?", prescriptionID).First(&rx).Error; err != nil {
		return ds.PrescriptionDrug{}, err
	}
	if rx.Status != "draft" {
		return ds.PrescriptionDrug{}, fmt.Errorf("%w: можно редактировать только черновик", ErrNotAllowed)
	}

	height := j.HeightCm
	weight := j.WeightKg
	if height <= 0 {
		height = item.HeightCm
	}
	if weight <= 0 {
		weight = item.WeightKg
	}

	var drug ds.Drug
	if err := r.db.First(&drug, drugID).Error; err != nil {
		return ds.PrescriptionDrug{}, err
	}
	dose := CalculatePediatricDose(height, weight, drug.DosePerM2Mg, drug.MaxDailyMg)

	err = r.db.Model(&item).Updates(map[string]interface{}{
		"height_cm":         height,
		"weight_kg":         weight,
		"pediatric_dose_mg": dose,
	}).Error
	if err != nil {
		return ds.PrescriptionDrug{}, err
	}
	r.db.Where("prescription_id = ? AND drug_id = ?", prescriptionID, drugID).
		Preload("Drug").First(&item)
	return item, nil
}

func (r *Repository) EditPrescription(id int, j serializer.PrescriptionEditJSON) (ds.Prescription, error) {
	var rx ds.Prescription
	err := r.db.Where("prescription_id = ? AND status != ?", id, "deleted").First(&rx).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ds.Prescription{}, fmt.Errorf("%w: рецепт с id %d", ErrNotFound, id)
		}
		return ds.Prescription{}, err
	}
	if rx.Status != "draft" {
		return ds.Prescription{}, fmt.Errorf("%w: можно редактировать только черновик", ErrNotAllowed)
	}
	updates := map[string]interface{}{}
	if j.DoctorFullName != nil && *j.DoctorFullName != "" {
		updates["doctor_full_name"] = *j.DoctorFullName
	}
	if j.Notes != nil {
		updates["notes"] = j.Notes
	}
	if len(updates) == 0 {
		r.db.Where("prescription_id = ?", id).First(&rx)
		return rx, nil
	}
	if err := r.db.Model(&rx).Updates(updates).Error; err != nil {
		return ds.Prescription{}, err
	}
	r.db.Where("prescription_id = ?", id).First(&rx)
	return rx, nil
}

func (r *Repository) FormPrescription(id int) (ds.Prescription, error) {
	rx, err := r.GetSinglePrescription(id)
	if err != nil {
		return ds.Prescription{}, err
	}
	if rx.Status != "draft" {
		return ds.Prescription{}, fmt.Errorf("%w: только черновик можно сформировать", ErrNotAllowed)
	}
	if rx.CreatorID != uint(r.GetCreatorID()) {
		return ds.Prescription{}, fmt.Errorf("%w: вы не создатель этого рецепта", ErrNotAllowed)
	}

	items, err := r.GetPrescriptionItems(int(rx.PrescriptionID))
	if err != nil {
		return ds.Prescription{}, err
	}
	if len(items) == 0 {
		return ds.Prescription{}, errors.New("нельзя сформировать пустой рецепт")
	}
	if rx.DoctorFullName == "" {
		return ds.Prescription{}, errors.New("укажите ФИО врача перед формированием")
	}

	for _, item := range items {
		var drug ds.Drug
		if err := r.db.First(&drug, item.DrugID).Error; err != nil {
			return ds.Prescription{}, err
		}
		dose := CalculatePediatricDose(item.HeightCm, item.WeightKg, drug.DosePerM2Mg, drug.MaxDailyMg)
		r.db.Model(&ds.PrescriptionDrug{}).
			Where("prescription_id = ? AND drug_id = ?", rx.PrescriptionID, item.DrugID).
			Update("pediatric_dose_mg", dose)
	}

	formingDate := time.Now()
	err = r.db.Model(&rx).Updates(map[string]interface{}{
		"status":       "formed",
		"forming_date": formingDate,
	}).Error
	if err != nil {
		return ds.Prescription{}, err
	}
	rx.Status = "formed"
	rx.FormingDate = &formingDate
	return rx, nil
}

func (r *Repository) FinishPrescription(id int, status string) (ds.Prescription, error) {
	if status != "completed" && status != "rejected" {
		return ds.Prescription{}, errors.New("неверный статус: допустимы completed или rejected")
	}
	user, err := r.GetUserByID(r.GetUserID())
	if err != nil {
		return ds.Prescription{}, err
	}
	if !user.IsModerator {
		return ds.Prescription{}, fmt.Errorf("%w: вы не модератор", ErrNotAllowed)
	}
	rx, err := r.GetSinglePrescription(id)
	if err != nil {
		return ds.Prescription{}, err
	}
	if rx.Status != "formed" {
		return ds.Prescription{}, fmt.Errorf("%w: завершить или отклонить можно только сформированный рецепт", ErrNotAllowed)
	}
	finishDate := time.Now()
	err = r.db.Model(&rx).Updates(map[string]interface{}{
		"status":       status,
		"finish_date":  finishDate,
		"moderator_id": user.UserID,
	}).Error
	if err != nil {
		return ds.Prescription{}, err
	}
	rx.Status = status
	rx.FinishDate = sql.NullTime{Time: finishDate, Valid: true}
	uid := user.UserID
	rx.ModeratorID = &uid
	return rx, nil
}

func (r *Repository) DeletePrescription(prescriptionID int) (ds.Prescription, error) {
	rx, err := r.GetSinglePrescription(prescriptionID)
	if err != nil {
		return ds.Prescription{}, err
	}
	if rx.Status != "draft" {
		return ds.Prescription{}, fmt.Errorf("%w: удалить можно только черновик", ErrNotAllowed)
	}
	if rx.CreatorID != uint(r.GetCreatorID()) {
		return ds.Prescription{}, fmt.Errorf("%w: вы не создатель этого рецепта", ErrNotAllowed)
	}
	formingDate := time.Now()
	err = r.db.Model(&rx).Updates(map[string]interface{}{
		"status":       "deleted",
		"forming_date": formingDate,
	}).Error
	if err != nil {
		return ds.Prescription{}, err
	}
	rx.Status = "deleted"
	rx.FormingDate = &formingDate
	return rx, nil
}

func (r *Repository) IsDraftPrescription(prescriptionID int, creatorID uint) (bool, error) {
	var rx ds.Prescription
	err := r.db.Select("status").Where("prescription_id = ? AND creator_id = ?",
		prescriptionID, creatorID).First(&rx).Error
	if err != nil {
		return false, err
	}
	return rx.Status == "draft", nil
}
