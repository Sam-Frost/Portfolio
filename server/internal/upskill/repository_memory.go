package upskill

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Sam-Frost/portfolio/internal/apperr"
	"github.com/Sam-Frost/portfolio/internal/id"
)

type MemoryRepository struct {
	mu        sync.Mutex
	topics    map[string]Topic
	subtopics map[string]Subtopic
	resources map[string]Resource
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		topics:    make(map[string]Topic),
		subtopics: make(map[string]Subtopic),
		resources: make(map[string]Resource),
	}
}

func (r *MemoryRepository) CreateTopic(_ context.Context, topic Topic) (Topic, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	topic.ID = id.New()
	topic.DateAdded = time.Now().UTC()
	r.topics[topic.ID] = topic
	return topic, nil
}

func (r *MemoryRepository) ListTopics(_ context.Context) ([]Topic, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	topics := make([]Topic, 0, len(r.topics))
	for _, t := range r.topics {
		t.SubtopicCount, t.DoneCount = r.progressLocked(t.ID)
		topics = append(topics, t)
	}

	sort.Slice(topics, func(i, j int) bool {
		return topics[i].DateAdded.After(topics[j].DateAdded)
	})

	return topics, nil
}

// progressLocked must be called with r.mu already held.
func (r *MemoryRepository) progressLocked(topicID string) (total, done int) {
	for _, s := range r.subtopics {
		if s.TopicID != topicID {
			continue
		}
		total++
		if s.Done {
			done++
		}
	}
	return total, done
}

func (r *MemoryRepository) GetTopic(_ context.Context, topicID string) (Topic, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	t, ok := r.topics[topicID]
	if !ok {
		return Topic{}, apperr.NotFound("topic not found")
	}
	t.SubtopicCount, t.DoneCount = r.progressLocked(t.ID)
	return t, nil
}

func (r *MemoryRepository) UpdateTopic(_ context.Context, topicID string, input UpdateTopicInput) (Topic, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	t, ok := r.topics[topicID]
	if !ok {
		return Topic{}, apperr.NotFound("topic not found")
	}

	if input.Name != nil {
		t.Name = strings.TrimSpace(*input.Name)
	}
	if input.TargetDate != nil {
		if *input.TargetDate == "" {
			t.TargetDate = nil
		} else {
			t.TargetDate = input.TargetDate
		}
	}

	r.topics[topicID] = t
	t.SubtopicCount, t.DoneCount = r.progressLocked(t.ID)
	return t, nil
}

func (r *MemoryRepository) DeleteTopic(_ context.Context, topicID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.topics[topicID]; !ok {
		return apperr.NotFound("topic not found")
	}
	delete(r.topics, topicID)

	for sid, s := range r.subtopics {
		if s.TopicID != topicID {
			continue
		}
		delete(r.subtopics, sid)
		for rid, res := range r.resources {
			if res.SubtopicID == sid {
				delete(r.resources, rid)
			}
		}
	}

	return nil
}

func (r *MemoryRepository) CreateSubtopic(_ context.Context, subtopic Subtopic) (Subtopic, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.topics[subtopic.TopicID]; !ok {
		return Subtopic{}, apperr.InvalidInput("topic not found")
	}

	subtopic.ID = id.New()
	subtopic.DateAdded = time.Now().UTC()
	subtopic.Done = false

	resources := make([]Resource, 0, len(subtopic.Resources))
	for _, res := range subtopic.Resources {
		res.ID = id.New()
		res.SubtopicID = subtopic.ID
		r.resources[res.ID] = res
		resources = append(resources, res)
	}
	subtopic.Resources = resources

	r.subtopics[subtopic.ID] = subtopic
	return subtopic, nil
}

func (r *MemoryRepository) ListSubtopics(_ context.Context, topicID string) ([]Subtopic, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	subtopics := make([]Subtopic, 0)
	for _, s := range r.subtopics {
		if s.TopicID != topicID {
			continue
		}
		s.Resources = r.resourcesForLocked(s.ID)
		subtopics = append(subtopics, s)
	}

	sort.Slice(subtopics, func(i, j int) bool {
		return subtopics[i].DateAdded.After(subtopics[j].DateAdded)
	})

	return subtopics, nil
}

// resourcesForLocked must be called with r.mu already held.
func (r *MemoryRepository) resourcesForLocked(subtopicID string) []Resource {
	resources := make([]Resource, 0)
	for _, res := range r.resources {
		if res.SubtopicID == subtopicID {
			resources = append(resources, res)
		}
	}
	sort.Slice(resources, func(i, j int) bool { return resources[i].ID < resources[j].ID })
	return resources
}

func (r *MemoryRepository) UpdateSubtopic(_ context.Context, subtopicID string, input UpdateSubtopicInput) (Subtopic, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	s, ok := r.subtopics[subtopicID]
	if !ok {
		return Subtopic{}, apperr.NotFound("subtopic not found")
	}

	if input.Name != nil {
		s.Name = strings.TrimSpace(*input.Name)
	}
	if input.TargetDate != nil {
		if *input.TargetDate == "" {
			s.TargetDate = nil
		} else {
			s.TargetDate = input.TargetDate
		}
	}
	if input.Done != nil {
		s.Done = *input.Done
	}

	r.subtopics[subtopicID] = s
	s.Resources = r.resourcesForLocked(s.ID)
	return s, nil
}

func (r *MemoryRepository) DeleteSubtopic(_ context.Context, subtopicID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.subtopics[subtopicID]; !ok {
		return apperr.NotFound("subtopic not found")
	}
	delete(r.subtopics, subtopicID)

	for rid, res := range r.resources {
		if res.SubtopicID == subtopicID {
			delete(r.resources, rid)
		}
	}

	return nil
}

func (r *MemoryRepository) AddResource(_ context.Context, subtopicID string, resource Resource) (Resource, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.subtopics[subtopicID]; !ok {
		return Resource{}, apperr.InvalidInput("subtopic not found")
	}

	resource.ID = id.New()
	resource.SubtopicID = subtopicID
	r.resources[resource.ID] = resource
	return resource, nil
}

func (r *MemoryRepository) DeleteResource(_ context.Context, resourceID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.resources[resourceID]; !ok {
		return apperr.NotFound("resource not found")
	}
	delete(r.resources, resourceID)
	return nil
}
