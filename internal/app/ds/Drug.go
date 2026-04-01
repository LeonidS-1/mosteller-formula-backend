package ds

type Drug struct {
	DrugID      uint    `gorm:"primaryKey;column:drug_id"`
	Title       string  `gorm:"type:varchar(255);not null"`
	Description string  `gorm:"type:varchar(1000);not null"`
	IsDeleted   bool    `gorm:"type:boolean;not null;default:false"`
	PhotoURL    string  `gorm:"column:photo_url;type:varchar(255)"`
	Video       string  `gorm:"type:varchar(255)"`
	AdultDoseMg float64 `gorm:"column:adult_dose_mg;type:numeric(12,2);not null"`
	DosePerM2Mg float64 `gorm:"column:dose_per_m2_mg;type:numeric(12,2);not null"`
	MaxDailyMg  float64 `gorm:"column:max_daily_mg;type:numeric(12,2);not null"`
}

func (Drug) TableName() string {
	return "drugs"
}
