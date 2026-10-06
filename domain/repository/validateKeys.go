package repository

import (
	"github.com/newrelic/newrelic-diagnostics-cli/domain/entity"
)

type IValidateKeys interface {
	// ValidateSchUseStrongCryptoKeys returns nil with no error when SchUseStrongCrypto isn't set at path.
	ValidateSchUseStrongCryptoKeys(path string) (*int, error)
	ValidateTLSRegKeys() (*entity.TLSRegKey, error)
}
