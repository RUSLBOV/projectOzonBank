package api

import "context"

// Shortener описывает бизнес-логику, необходимую HTTP-хендлерам.
// Реализуется app.LinkService — хендлеры зависят от узкого интерфейса,
// а не от конкретного типа, что упрощает тестирование через фейки.
type Shortener interface {
	// Shorten создаёт короткий код для originalURL. Если URL уже был
	// сокращён ранее, возвращает существующий код без ошибки.
	Shorten(ctx context.Context, originalURL string) (code string, err error)

	// Resolve возвращает оригинальный URL по короткому коду.
	Resolve(ctx context.Context, code string) (originalURL string, err error)
}
