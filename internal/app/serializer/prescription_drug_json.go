package serializer

import "web_backend/internal/app/ds"

type PrescriptionDrugJSON struct {
	PrescriptionID uint     `json:"prescription_id"`
	DrugID         uint     `json:"drug_id"`
	HeightCm       int      `json:"height_cm"`
	WeightKg       int      `json:"weight_kg"`
	Dose           *float64 `json:"dose"`
}

type PrescriptionDrugDetailJSON struct {
	PrescriptionID uint     `json:"prescription_id"`
	DrugID         uint     `json:"drug_id"`
	HeightCm       int      `json:"height_cm"`
	WeightKg       int      `json:"weight_kg"`
	Dose           *float64 `json:"dose"`
	Drug           DrugJSON `json:"drug"`
}

func PrescriptionDrugToJSON(item ds.PrescriptionDrug) PrescriptionDrugJSON {
	return PrescriptionDrugJSON{
		PrescriptionID: item.PrescriptionID,
		DrugID:         item.DrugID,
		HeightCm:       item.HeightCm,
		WeightKg:       item.WeightKg,
		Dose:           item.PediatricDoseMg,
	}
}

func PrescriptionDrugDetailToJSON(item ds.PrescriptionDrug) PrescriptionDrugDetailJSON {
	return PrescriptionDrugDetailJSON{
		PrescriptionID: item.PrescriptionID,
		DrugID:         item.DrugID,
		HeightCm:       item.HeightCm,
		WeightKg:       item.WeightKg,
		Dose:           item.PediatricDoseMg,
		Drug:           DrugToJSON(item.Drug),
	}
}

func PrescriptionDrugFromJSON(j PrescriptionDrugJSON) ds.PrescriptionDrug {
	return ds.PrescriptionDrug{
		HeightCm: j.HeightCm,
		WeightKg: j.WeightKg,
	}
}
