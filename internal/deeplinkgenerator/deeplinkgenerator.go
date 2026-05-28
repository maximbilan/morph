package deeplinkgenerator

import "time"

type DeepLinkGenerator interface {
	Create(category string, subcategory string, account string, amount float64, date time.Time) string
}
