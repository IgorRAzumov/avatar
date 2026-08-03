package repository

import (
	"testing"
	"time"

	"avatar/internal/domain/model"
	"avatar/internal/imageformat"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRow struct {
	values []any
	err    error
}

func (m mockRow) Scan(dest ...any) error {
	if m.err != nil {
		return m.err
	}
	for i, d := range dest {
		switch ptr := d.(type) {
		case *string:
			*ptr = m.values[i].(string)
		case *int64:
			*ptr = m.values[i].(int64)
		case *int:
			*ptr = m.values[i].(int)
		case *time.Time:
			*ptr = m.values[i].(time.Time)
		case **time.Time:
			*ptr = m.values[i].(*time.Time)
		}
	}
	return nil
}

func TestScanAvatarFromRow(t *testing.T) {
	now := time.Now().UTC()
	row := mockRow{values: []any{
		"id-1", "user@test.com", "photo.jpg", imageformat.MIMEJPEG, int64(100),
		"completed", "processing", 800, 600, now, now, (*time.Time)(nil),
	}}

	avatar, err := scanAvatarFromRow(row)
	require.NoError(t, err)
	assert.Equal(t, "id-1", avatar.ID)
	assert.Equal(t, model.ProcessingStatusProcessing, avatar.ProcessingStatus)
	assert.Equal(t, 800, avatar.Width)
}
