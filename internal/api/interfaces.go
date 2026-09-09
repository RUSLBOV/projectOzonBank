package api

import "context"

// Shortener описывает бизнес-логику, необходимую HTTP-хендлерам.
// Реализуется app.LinkService - хендлеры зависят от узкого интерфейса,
type Shortener interface {
	// Shorten создаёт короткий код для originalURL. Если URL уже был
	Shorten(ctx context.Context, originalURL string) (code string, err error)

	// Resolve возвращает оригинальный URL по короткому коду.
	Resolve(ctx context.Context, code string) (originalURL string, err error)
}
