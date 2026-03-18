package repository

import (
	"fmt"
	"math"
	"strings"
)

// Repository — хранилище данных (Lab 1: данные в массивах, без БД).
type Repository struct {
}

// NewRepository создаёт новый экземпляр репозитория.
func NewRepository() (*Repository, error) {
	return &Repository{}, nil
}

// Drug — услуга: препарат с рекомендованными дозировками для взрослых.
type Drug struct {
	ID          int
	Title       string
	Description string
	Image       string  // ключ изображения в Minio
	Video       string  // ключ видео в Minio
	AdultDoseMg float64 // рекомендуемая разовая доза для взрослых (мг)
	DosePerM2Mg float64 // доза на м² ППТ (мг/м²)
	MaxDailyMg  float64 // максимальная суточная доза (мг)
}

// Prescription — заявка: расчёт индивидуальной детской дозы для набора препаратов и пациентов.
type Prescription struct {
	ID          int
	Title       string
	Description string // описание заявки текстом
	Drugs       []PrescriptionDrug
	DrugCount   int
}

// PrescriptionDrug — связь м-м: препарат в заявке, рост/вес пациента, рассчитанная детская доза.
type PrescriptionDrug struct {
	Drug            Drug
	HeightCm        int     // рост пациента (см)
	WeightKg        int     // вес пациента (кг)
	PediatricDoseMg float64 // результат: рассчитанная детская доза (мг)
}

// GetDrugs возвращает все препараты.
func (r *Repository) GetDrugs() ([]Drug, error) {
	drugs := []Drug{
		{
			ID:          1,
			Title:       "Парацетамол",
			Description: "Жаропонижающее и обезболивающее. Рекомендован при лихорадке и боли лёгкой и средней интенсивности.",
			Image:       "paracetamol.jpg",
			Video:       "paracetamol.mp4",
			AdultDoseMg: 1000,
			DosePerM2Mg: 250,
			MaxDailyMg:  4000,
		},
		{
			ID:          2,
			Title:       "Ибупрофен",
			Description: "НПВП, жаропонижающее и противовоспалительное. Применяется при боли и воспалении.",
			Image:       "ibuprofen.jpg",
			Video:       "ibuprofen.mp4",
			AdultDoseMg: 400,
			DosePerM2Mg: 100,
			MaxDailyMg:  2400,
		},
		{
			ID:          3,
			Title:       "Амоксициллин",
			Description: "Антибиотик группы пенициллинов. Назначается при бактериальных инфекциях дыхательных путей и ЛОР-органов.",
			Image:       "amoxicillin.jpg",
			Video:       "amoxicillin.mp4",
			AdultDoseMg: 500,
			DosePerM2Mg: 25,
			MaxDailyMg:  1500,
		},
		{
			ID:          4,
			Title:       "Цетиризин",
			Description: "Антигистаминный препарат. Показан при аллергическом рините и крапивнице.",
			Image:       "cetirizine.jpg",
			Video:       "cetirizine.mp4",
			AdultDoseMg: 10,
			DosePerM2Mg: 5,
			MaxDailyMg:  10,
		},
		{
			ID:          5,
			Title:       "Омепразол",
			Description: "Ингибитор протонной помпы. Используется при ГЭРБ и язвенной болезни.",
			Image:       "omeprazole.jpg",
			Video:       "omeprazole.mp4",
			AdultDoseMg: 20,
			DosePerM2Mg: 10,
			MaxDailyMg:  40,
		},
		{
			ID:          6,
			Title:       "Домперидон",
			Description: "Противорвотное, прокинетик. При тошноте и функциональных нарушениях ЖКТ.",
			Image:       "domperidone.jpg",
			Video:       "domperidone.mp4",
			AdultDoseMg: 10,
			DosePerM2Mg: 2.5,
			MaxDailyMg:  30,
		},
	}

	if len(drugs) == 0 {
		return nil, fmt.Errorf("массив препаратов пуст")
	}

	return drugs, nil
}

// GetDrug возвращает препарат по ID.
func (r *Repository) GetDrug(id int) (Drug, error) {
	drugs, err := r.GetDrugs()
	if err != nil {
		return Drug{}, err
	}

	for _, d := range drugs {
		if d.ID == id {
			return d, nil
		}
	}
	return Drug{}, fmt.Errorf("препарат не найден")
}

// GetDrugsByTitle возвращает препараты, содержащие подстроку в названии или описании.
func (r *Repository) GetDrugsByTitle(query string) ([]Drug, error) {
	drugs, err := r.GetDrugs()
	if err != nil {
		return nil, err
	}

	q := strings.ToLower(query)
	var result []Drug
	for _, d := range drugs {
		if strings.Contains(strings.ToLower(d.Title), q) || strings.Contains(strings.ToLower(d.Description), q) {
			result = append(result, d)
		}
	}
	return result, nil
}

// CalculatePediatricDose вычисляет детскую дозу по формуле Mosteller: ППТ = sqrt(рост*вес/3600), доза = ППТ * dosePerM2Mg, ограничена maxDailyMg.
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

// buildPrescription собирает заявку из записей м-м.
// descriptionFormat — строка формата с одним %d для подстановки количества пациентов (например "для %d пациентов").
func (r *Repository) buildPrescription(id int, title, descriptionFormat string, entries []struct {
	DrugID   int
	HeightCm int
	WeightKg int
}) (Prescription, error) {
	drugs, err := r.GetDrugs()
	if err != nil {
		return Prescription{}, err
	}

	drugMap := make(map[int]Drug)
	for _, d := range drugs {
		drugMap[d.ID] = d
	}

	var prescriptionDrugs []PrescriptionDrug

	for _, e := range entries {
		drug, ok := drugMap[e.DrugID]
		if !ok {
			continue
		}
		dose := CalculatePediatricDose(e.HeightCm, e.WeightKg, drug.DosePerM2Mg, drug.MaxDailyMg)
		prescriptionDrugs = append(prescriptionDrugs, PrescriptionDrug{
			Drug:            drug,
			HeightCm:        e.HeightCm,
			WeightKg:        e.WeightKg,
			PediatricDoseMg: dose,
		})
	}

	description := fmt.Sprintf(descriptionFormat, len(prescriptionDrugs))

	return Prescription{
		ID:          id,
		Title:       title,
		Description: description,
		Drugs:       prescriptionDrugs,
		DrugCount:   len(prescriptionDrugs),
	}, nil
}

// GetPrescriptions возвращает все заявки на расчёт доз.
func (r *Repository) GetPrescriptions() ([]Prescription, error) {
	entries := []struct {
		DrugID   int
		HeightCm int
		WeightKg int
	}{
		{1, 120, 22},
		{2, 95, 14},
		{3, 140, 32},
		{4, 110, 20},
		{5, 130, 28},
		{6, 100, 16},
	}

	prescription, err := r.buildPrescription(
		1,
		"Расчёт детских доз для стационарных пациентов",
		"Заявка на расчёт индивидуальных педиатрических доз по площади поверхности тела (формула Mosteller) для %d пациентов с разными антропометрическими данными. Рост и вес указаны в см и кг соответственно.",
		entries,
	)
	if err != nil {
		return nil, err
	}

	return []Prescription{prescription}, nil
}

// GetPrescription возвращает заявку по ID.
func (r *Repository) GetPrescription(id int) (Prescription, error) {
	prescriptions, err := r.GetPrescriptions()
	if err != nil {
		return Prescription{}, err
	}

	for _, p := range prescriptions {
		if p.ID == id {
			return p, nil
		}
	}
	return Prescription{}, fmt.Errorf("заявка не найдена")
}

// GetPrescriptionForDrug ищет заявку, содержащую данный препарат, и возвращает запись м-м.
func (r *Repository) GetPrescriptionForDrug(drugID int) (*PrescriptionDrug, error) {
	prescriptions, err := r.GetPrescriptions()
	if err != nil {
		return nil, err
	}

	for _, p := range prescriptions {
		for _, pd := range p.Drugs {
			if pd.Drug.ID == drugID {
				return &pd, nil
			}
		}
	}
	return nil, fmt.Errorf("препарат не найден в заявках")
}
