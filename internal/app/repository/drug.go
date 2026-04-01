package repository

import (
	"fmt"
	"web_backend/internal/app/ds"
)

func (r *Repository) GetDrugs() ([]ds.Drug, error) {
	var drugs []ds.Drug
	err := r.db.Where("is_deleted = ?", false).Find(&drugs).Error
	if err != nil {
		return nil, err
	}
	if len(drugs) == 0 {
		return nil, fmt.Errorf("список препаратов пуст")
	}
	return drugs, nil
}

func (r *Repository) GetDrug(id int) (ds.Drug, error) {
	var drug ds.Drug
	err := r.db.Where("drug_id = ? AND is_deleted = ?", id, false).First(&drug).Error
	if err != nil {
		return ds.Drug{}, err
	}
	return drug, nil
}

func (r *Repository) GetDrugsByTitle(query string) ([]ds.Drug, error) {
	var drugs []ds.Drug
	err := r.db.Where("(title ILIKE ? OR description ILIKE ?) AND is_deleted = ?",
		"%"+query+"%", "%"+query+"%", false).Find(&drugs).Error
	if err != nil {
		return nil, err
	}
	return drugs, nil
}
