package memstore

import (
	"bytecourses/internal/domain"
	"context"
	"errors"
	"sort"
	"sync"
	"time"
)

type ContentStore struct {
	mu            sync.RWMutex
	itemsByID     map[int64]domain.ContentItem
	itemsByModule map[int64][]int64
	lecturesByID  map[int64]domain.Lecture
	nextContentID int64
}

func NewContentStore() *ContentStore {
	return &ContentStore{
		itemsByID:     make(map[int64]domain.ContentItem),
		itemsByModule: make(map[int64][]int64),
		lecturesByID:  make(map[int64]domain.Lecture),
		nextContentID: 1,
	}
}

func (s *ContentStore) CreateContentItem(ctx context.Context, item *domain.ContentItem) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	item.ID = s.nextContentID
	now := time.Now()
	item.CreatedAt = now
	item.UpdatedAt = now

	moduleItems := s.itemsByModule[item.ModuleID]
	item.Position = len(moduleItems) + 1

	s.itemsByID[item.ID] = *item
	s.itemsByModule[item.ModuleID] = append(moduleItems, item.ID)
	s.nextContentID++

	return nil
}

func (s *ContentStore) GetContentItemByID(ctx context.Context, id int64) (*domain.ContentItem, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if item, ok := s.itemsByID[id]; ok {
		copy := item
		return &copy, true
	}
	return nil, false
}

func (s *ContentStore) UpdateContentItem(ctx context.Context, item *domain.ContentItem) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.itemsByID[item.ID]; !exists {
		return errors.New("content item does not exist")
	}

	item.UpdatedAt = time.Now()
	s.itemsByID[item.ID] = *item
	return nil
}

func (s *ContentStore) DeleteContentItemByID(ctx context.Context, id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, exists := s.itemsByID[id]
	if !exists {
		return errors.New("content item does not exist")
	}

	moduleItems := s.itemsByModule[item.ModuleID]
	for i, iid := range moduleItems {
		if iid == id {
			s.itemsByModule[item.ModuleID] = append(moduleItems[:i], moduleItems[i+1:]...)
			break
		}
	}

	delete(s.itemsByID, id)
	delete(s.lecturesByID, id)
	return nil
}

func (s *ContentStore) ListContentItemsByModuleID(ctx context.Context, moduleID int64) ([]domain.ContentItem, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	itemIDs := s.itemsByModule[moduleID]
	out := make([]domain.ContentItem, 0, len(itemIDs))

	for _, id := range itemIDs {
		if item, ok := s.itemsByID[id]; ok {
			out = append(out, item)
		}
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].Position < out[j].Position
	})

	return out, nil
}

func (s *ContentStore) ReorderContentItems(ctx context.Context, moduleID int64, itemIDs []int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()

	for i, id := range itemIDs {
		item, exists := s.itemsByID[id]
		if !exists {
			return errors.New("content item does not exist")
		}
		if item.ModuleID != moduleID {
			return errors.New("content item does not belong to module")
		}
		item.Position = i + 1
		item.UpdatedAt = now
		s.itemsByID[id] = item
	}

	s.itemsByModule[moduleID] = itemIDs

	return nil
}

func (s *ContentStore) GetLecture(ctx context.Context, contentItemID int64) (*domain.Lecture, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if lecture, ok := s.lecturesByID[contentItemID]; ok {
		copy := lecture
		return &copy, true
	}
	return nil, false
}

func (s *ContentStore) UpsertLecture(ctx context.Context, lecture *domain.Lecture) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.itemsByID[lecture.ContentItemID]; !exists {
		return errors.New("content item does not exist")
	}

	s.lecturesByID[lecture.ContentItemID] = *lecture
	return nil
}

func (s *ContentStore) GetContentItemWithLecture(ctx context.Context, id int64) (*domain.ContentItem, *domain.Lecture, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, ok := s.itemsByID[id]
	if !ok {
		return nil, nil, false
	}

	itemCopy := item
	var lectureCopy *domain.Lecture
	if lecture, ok := s.lecturesByID[id]; ok {
		lec := lecture
		lectureCopy = &lec
	}

	return &itemCopy, lectureCopy, true
}

func (s *ContentStore) ListContentItemsWithLecturesByModuleID(ctx context.Context, moduleID int64) ([]domain.ContentItem, map[int64]*domain.Lecture, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	itemIDs := s.itemsByModule[moduleID]
	items := make([]domain.ContentItem, 0, len(itemIDs))
	lectures := make(map[int64]*domain.Lecture)

	for _, id := range itemIDs {
		if item, ok := s.itemsByID[id]; ok {
			items = append(items, item)
			if lecture, ok := s.lecturesByID[id]; ok {
				lec := lecture
				lectures[id] = &lec
			}
		}
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Position < items[j].Position
	})

	return items, lectures, nil
}
