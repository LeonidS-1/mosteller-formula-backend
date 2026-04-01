package serializer

import "web_backend/internal/app/ds"

type DrugJSON struct {
	DrugID      uint    `json:"drug_id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	IsDeleted   bool    `json:"is_deleted"`
	PhotoURL    string  `json:"photo_url"`
	Video       string  `json:"video"`
	AdultDoseMg float64 `json:"adult_dose_mg"`
	DosePerM2Mg float64 `json:"dose_per_m2_mg"`
	MaxDailyMg  float64 `json:"max_daily_mg"`
}

func DrugToJSON(d ds.Drug) DrugJSON {
	return DrugJSON{
		DrugID:      d.DrugID,
		Title:       d.Title,
		Description: d.Description,
		IsDeleted:   d.IsDeleted,
		PhotoURL:    d.PhotoURL,
		Video:       d.Video,
		AdultDoseMg: d.AdultDoseMg,
		DosePerM2Mg: d.DosePerM2Mg,
		MaxDailyMg:  d.MaxDailyMg,
	}
}

func DrugFromJSON(j DrugJSON) ds.Drug {
	return ds.Drug{
		Title:       j.Title,
		Description: j.Description,
		AdultDoseMg: j.AdultDoseMg,
		DosePerM2Mg: j.DosePerM2Mg,
		MaxDailyMg:  j.MaxDailyMg,
	}
}
