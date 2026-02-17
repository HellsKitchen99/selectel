package example

import (
	"go.uber.org/zap"
)

func invalidUpperCase() {
	zap.L().Info("Information")
}
