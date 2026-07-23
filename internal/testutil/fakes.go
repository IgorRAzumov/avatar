package testutil

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"avatar/internal/adapter/objectkey"
	"avatar/internal/domain/model"
)

type MemoryAvatarStore struct {
	mu      sync.Mutex
	Avatars map[string]*model.Avatar
	NextID  int
}

func NewMemoryAvatarStore() *MemoryAvatarStore {
	return &MemoryAvatarStore{Avatars: make(map[string]*model.Avatar), NextID: 1}
}

func (m *MemoryAvatarStore) Create(_ context.Context, avatar *model.Avatar) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if avatar.ID == "" {
		avatar.ID = fmt.Sprintf("avatar-%d", m.NextID)
		m.NextID++
	}
	now := time.Now().UTC()
	avatar.CreatedAt = now
	avatar.UpdatedAt = now
	copyAvatar := *avatar
	m.Avatars[avatar.ID] = &copyAvatar
	return nil
}

func (m *MemoryAvatarStore) GetByID(_ context.Context, id string) (*model.Avatar, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	a, ok := m.Avatars[id]
	if !ok {
		return nil, model.ErrNotFound
	}
	copyAvatar := *a
	return &copyAvatar, nil
}

func (m *MemoryAvatarStore) GetLatestByUserID(ctx context.Context, userID string) (*model.Avatar, error) {
	list, err := m.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, model.ErrNotFound
	}
	return list[0], nil
}

func (m *MemoryAvatarStore) ListByUserID(_ context.Context, userID string) ([]*model.Avatar, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var list []*model.Avatar
	for _, a := range m.Avatars {
		if a.UserID == userID {
			copyAvatar := *a
			list = append(list, &copyAvatar)
		}
	}
	return list, nil
}

func (m *MemoryAvatarStore) SoftDelete(_ context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.Avatars[id]; !ok {
		return model.ErrNotFound
	}
	delete(m.Avatars, id)
	return nil
}

func (m *MemoryAvatarStore) UpdateProcessingStatus(_ context.Context, id, status string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	a, ok := m.Avatars[id]
	if !ok {
		return model.ErrNotFound
	}
	a.ProcessingStatus = status
	a.UpdatedAt = time.Now().UTC()
	return nil
}

func (m *MemoryAvatarStore) CompleteProcessing(_ context.Context, id string, width, height int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	a, ok := m.Avatars[id]
	if !ok {
		return model.ErrNotFound
	}
	a.Width = width
	a.Height = height
	a.ProcessingStatus = model.ProcessingStatusCompleted
	a.UpdatedAt = time.Now().UTC()
	return nil
}

// ProcessingStatus returns the current processing status of an avatar in a
// concurrency-safe way, for use by tests exercising async code paths.
func (m *MemoryAvatarStore) ProcessingStatus(id string) string {
	m.mu.Lock()
	defer m.mu.Unlock()

	if a, ok := m.Avatars[id]; ok {
		return a.ProcessingStatus
	}
	return ""
}

func (m *MemoryAvatarStore) Ping(context.Context) error { return nil }

type MemoryStorage struct {
	mu      sync.Mutex
	Objects map[string][]byte
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{Objects: make(map[string][]byte)}
}

func (m *MemoryStorage) SaveOriginal(_ context.Context, avatarID string, data []byte, _ string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Objects[objectkey.OriginalObjectKey(avatarID)] = append([]byte(nil), data...)
	return nil
}

func (m *MemoryStorage) SaveThumbnail(_ context.Context, avatarID, size string, data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Objects[objectkey.ThumbnailObjectKey(avatarID, size)] = append([]byte(nil), data...)
	return nil
}

func (m *MemoryStorage) OpenOriginal(_ context.Context, avatarID string) ([]byte, error) {
	return m.download(objectkey.OriginalObjectKey(avatarID))
}

func (m *MemoryStorage) OpenThumbnail(_ context.Context, avatarID, size string) ([]byte, error) {
	return m.download(objectkey.ThumbnailObjectKey(avatarID, size))
}

func (m *MemoryStorage) DeleteAll(_ context.Context, avatarID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	prefix := objectkey.AvatarDir(avatarID) + "/"
	for key := range m.Objects {
		if strings.HasPrefix(key, prefix) {
			delete(m.Objects, key)
		}
	}
	return nil
}

func (m *MemoryStorage) download(key string) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, ok := m.Objects[key]
	if !ok {
		return nil, fmt.Errorf("%w", model.ErrNotFound)
	}
	return append([]byte(nil), data...), nil
}

func (m *MemoryStorage) Ping(context.Context) error { return nil }

type NoopPublisher struct{}

func (NoopPublisher) PublishUploadEvent(context.Context, model.AvatarUploadEvent) error { return nil }
func (NoopPublisher) PublishDeleteEvent(context.Context, model.AvatarDeleteEvent) error { return nil }
