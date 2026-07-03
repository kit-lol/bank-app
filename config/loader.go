package config

import (
	"bank-app/models"
	"encoding/json"
	"os"
)

func LoadDepositConfig(path string) ([]models.DepositTypeExtended, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config models.DepositConfig
	err = json.Unmarshal(file, &config)
	if err != nil {
		return nil, err
	}

	return config.Deposits, nil
}

// Функция для конвертации Extended в обычный DepositType (для БД)
func ConvertToDepositType(ext models.DepositTypeExtended) models.DepositType {
	return models.DepositType{
		ID:           ext.ID,
		Name:         ext.Name,
		InterestRate: ext.InterestRate,
		MinAmount:    ext.MinAmount,
		CanWithdraw:  ext.CanWithdraw,
		CanDeposit:   ext.CanDeposit,
		Description:  ext.Description,
	}
}
