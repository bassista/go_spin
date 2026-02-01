package repository

import "context"

// Saver persists a DataDocument.
// Small interface used by background jobs like the persistence scheduler.
type Saver interface {
	Save(ctx context.Context, doc *DataDocument) error
}

// Repository abstracts persistence and watching of the data file.
// JSONRepository implements this interface.
type Repository interface {
	Saver
	Load(ctx context.Context) (*DataDocument, error)
	StartWatcher(ctx context.Context, cacheStore CacheStore) error
}
