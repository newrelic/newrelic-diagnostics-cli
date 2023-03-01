package repository

import (
	"github.com/newrelic/newrelic-diagnostics-cli/domain/entity"
)

type IValidateKeys interface {
	ValidateSchUseStrongCryptoKeys(path string) (*int, error)
	ValidateTLSRegKeys() (*entity.TLSRegKey, error)
}
