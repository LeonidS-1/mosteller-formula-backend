package ds

type PrescriptionDrug struct {
	PrescriptionID  uint     `gorm:"primaryKey;column:prescription_id"`
	DrugID          uint     `gorm:"primaryKey;column:drug_id"`
	HeightCm        int      `gorm:"not null"`
	WeightKg        int      `gorm:"not null"`
	PediatricDoseMg *float64 `gorm:"column:pediatric_dose_mg;type:numeric(12,2)"`

	Drug         Drug         `gorm:"foreignKey:DrugID"`
	Prescription Prescription `gorm:"foreignKey:PrescriptionID"`
}

func (PrescriptionDrug) TableName() string {
	return "prescription_drugs"
}
