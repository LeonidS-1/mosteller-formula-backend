package serializer

import "web_backend/internal/app/ds"

type PrescriptionDrugJSON struct {
	PrescriptionID  uint     `json:"prescription_id"`
	DrugID          uint     `json:"drug_id"`
	HeightCm        int      `json:"height_cm"`
	WeightKg        int      `json:"weight_kg"`
	PediatricDoseMg *float64 `json:"pediatric_dose_mg"`
}

type PrescriptionDrugDetailJSON struct {
	PrescriptionID  uint     `json:"prescription_id"`
	DrugID          uint     `json:"drug_id"`
	HeightCm        int      `json:"height_cm"`
	WeightKg        int      `json:"weight_kg"`
	PediatricDoseMg *float64 `json:"pediatric_dose_mg"`
	Drug            DrugJSON `json:"drug"`
}

func PrescriptionDrugToJSON(item ds.PrescriptionDrug) PrescriptionDrugJSON {
	return PrescriptionDrugJSON{
		PrescriptionID:  item.PrescriptionID,
		DrugID:          item.DrugID,
		HeightCm:        item.HeightCm,
		WeightKg:        item.WeightKg,
		PediatricDoseMg: item.PediatricDoseMg,
	}
}

func PrescriptionDrugDetailToJSON(item ds.PrescriptionDrug) PrescriptionDrugDetailJSON {
	return PrescriptionDrugDetailJSON{
		PrescriptionID:  item.PrescriptionID,
		DrugID:          item.DrugID,
		HeightCm:        item.HeightCm,
		WeightKg:        item.WeightKg,
		PediatricDoseMg: item.PediatricDoseMg,
		Drug:            DrugToJSON(item.Drug),
	}
}

func PrescriptionDrugFromJSON(j PrescriptionDrugJSON) ds.PrescriptionDrug {
	return ds.PrescriptionDrug{
		HeightCm: j.HeightCm,
		WeightKg: j.WeightKg,
	}
}
