package shorturl

type ShortURL interface {
	Shorten(URL string) (string, error)
}
