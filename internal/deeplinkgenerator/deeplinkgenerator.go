package deeplinkgenerator

type DeepLinkGenerator interface {
	Create(category string, subcategory string, amount float64) string
}
