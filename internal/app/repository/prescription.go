package repository

import (
	"errors"
	"fmt"
	"math"
	"time"
	"web_backend/internal/app/ds"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// CalculatePediatricDose вычисляет детскую дозу по формуле Mosteller.
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
	draftDemoHeightCm = 120
	draftDemoWeightKg = 22
	draftDoctorName   = "Смирнова Анна Владимировна"
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

func (r *Repository) GetPrescription(id int, creatorID uint) ([]ds.PrescriptionDrug, *ds.Prescription, error) {
	var p ds.Prescription
	err := r.db.Where("prescription_id = ? AND creator_id = ? AND status != ?",
		id, creatorID, "deleted").First(&p).Error
	if err != nil {
		return nil, nil, err
	}

	var items []ds.PrescriptionDrug
	err = r.db.Where("prescription_id = ?", id).
		Preload("Drug").
		Find(&items).Error
	if err != nil {
		return nil, nil, err
	}

	return items, &p, nil
}

func (r *Repository) AddDrug(drugID uint, creatorID uint) error {
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

	var count int64
	r.db.Model(&ds.PrescriptionDrug{}).
		Where("prescription_id = ? AND drug_id = ?", rx.PrescriptionID, drugID).
		Count(&count)

	if count > 0 {
		return nil
	}

	var drug ds.Drug
	if err := r.db.First(&drug, drugID).Error; err != nil {
		return err
	}

	dose := CalculatePediatricDose(draftDemoHeightCm, draftDemoWeightKg, drug.DosePerM2Mg, drug.MaxDailyMg)
	row := ds.PrescriptionDrug{
		PrescriptionID:  rx.PrescriptionID,
		DrugID:          drugID,
		HeightCm:        draftDemoHeightCm,
		WeightKg:        draftDemoWeightKg,
		PediatricDoseMg: &dose,
	}
	return r.db.Create(&row).Error
}

// DeletePrescription выполняет логическое удаление рецепта сырым SQL UPDATE (без ORM).
func (r *Repository) DeletePrescription(prescriptionID uint) error {
	query := `
		UPDATE prescriptions
		SET status = 'deleted'
		WHERE prescription_id = $1;
	`
	result := r.db.Exec(query, prescriptionID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("prescription %d not found", prescriptionID)
	}
	return nil
}

func (r *Repository) IsDraftPrescription(prescriptionID int, creatorID uint) (bool, error) {
	var p ds.Prescription
	err := r.db.Select("status").Where("prescription_id = ? AND creator_id = ?",
		prescriptionID, creatorID).First(&p).Error
	if err != nil {
		return false, err
	}
	return p.Status == "draft", nil
}
