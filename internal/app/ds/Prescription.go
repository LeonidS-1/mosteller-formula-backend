package ds

import (
	"database/sql"
	"time"
)

type Prescription struct {
	PrescriptionID uint         `gorm:"primaryKey;column:prescription_id"`
	Status         string       `gorm:"type:varchar(20);not null"`
	CreatedAt      time.Time    `gorm:"not null"`
	CreatorID      uint         `gorm:"not null"`
	FormingDate    *time.Time   `gorm:"column:forming_date"`
	FinishDate     sql.NullTime `gorm:"column:finish_date"`
	ModeratorID    *uint        `gorm:"column:moderator_id"`
	DoctorFullName string       `gorm:"column:doctor_full_name;type:varchar(200);not null"`
	Notes          *string      `gorm:"type:varchar(2000)"`

	Creator   Users  `gorm:"foreignKey:CreatorID"`
	Moderator *Users `gorm:"foreignKey:ModeratorID"`
}

func (Prescription) TableName() string {
	return "prescriptions"
}
