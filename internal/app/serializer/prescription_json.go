package serializer

import (
	"time"

	"web_backend/internal/app/ds"
)

type PrescriptionJSON struct {
	PrescriptionID         uint       `json:"prescription_id"`
	Status                 string     `json:"status"`
	CreatedAt              time.Time  `json:"created_at"`
	CreatorLogin           string     `json:"creator_login"`
	ModeratorLogin         *string    `json:"moderator_login"`
	FormingDate            *time.Time `json:"forming_date"`
	FinishDate             *time.Time `json:"finish_date"`
	DoctorFullName         string     `json:"doctor_full_name"`
	CompletedDoseLineCount int        `json:"completed_dose_line_count"`
}

func PrescriptionToJSON(p ds.Prescription, creatorLogin, moderatorLogin string, completedDoseLineCount int) PrescriptionJSON {
	var mLogin *string
	if moderatorLogin != "" {
		mLogin = &moderatorLogin
	}
	var finishDate *time.Time
	if p.FinishDate.Valid {
		finishDate = &p.FinishDate.Time
	}
	return PrescriptionJSON{
		PrescriptionID:         p.PrescriptionID,
		Status:                 p.Status,
		CreatedAt:              p.CreatedAt,
		CreatorLogin:           creatorLogin,
		ModeratorLogin:         mLogin,
		FormingDate:            p.FormingDate,
		FinishDate:             finishDate,
		DoctorFullName:         p.DoctorFullName,
		CompletedDoseLineCount: completedDoseLineCount,
	}
}

type PrescriptionEditJSON struct {
	DoctorFullName *string `json:"doctor_full_name"`
}

type StatusJSON struct {
	Status string `json:"status"`
}
