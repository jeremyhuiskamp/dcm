package log

import (
	"github.com/Sirupsen/logrus"
)

var dcmLog = logrus.New()

const CategoryKey = "category"

func Category(category string) *logrus.Entry {
	return dcmLog.WithField(CategoryKey, category)
}

// TODO: per-category filtering
// - do this by creating a new logrus hook that filters based on fields
// - replace dcmLog output with Discard
// - new hook prints to stdout, but with filtering
