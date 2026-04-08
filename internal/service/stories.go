package service

import (
	"context"

	"novelai/internal/novelai"
)

type StoriesService struct {
	Client novelai.StoriesClient
}

func (s *StoriesService) GetKeystore(ctx context.Context, token string) (map[string]any, error) {
	return s.Client.GetKeystore(ctx, token)
}

func (s *StoriesService) PutKeystore(ctx context.Context, token string, payload map[string]any) (map[string]any, error) {
	return s.Client.PutKeystore(ctx, token, payload)
}

func (s *StoriesService) GetSubscription(ctx context.Context, token string) (map[string]any, error) {
	return s.Client.GetSubscription(ctx, token)
}

func (s *StoriesService) ListObjects(ctx context.Context, token string, objectType string) (map[string]any, error) {
	if err := novelai.ValidateObjectType(objectType); err != nil {
		return nil, err
	}
	return s.Client.ListObjects(ctx, token, objectType)
}

func (s *StoriesService) PutObject(ctx context.Context, token string, objectType string, payload map[string]any) (map[string]any, error) {
	if err := novelai.ValidateObjectType(objectType); err != nil {
		return nil, err
	}
	return s.Client.PutObject(ctx, token, objectType, payload)
}

func (s *StoriesService) PatchObject(ctx context.Context, token string, objectType string, id string, payload map[string]any) (map[string]any, error) {
	if err := novelai.ValidateObjectType(objectType); err != nil {
		return nil, err
	}
	return s.Client.PatchObject(ctx, token, objectType, id, payload)
}

func (s *StoriesService) DeleteObject(ctx context.Context, token string, objectType string, id string) error {
	if err := novelai.ValidateObjectType(objectType); err != nil {
		return err
	}
	return s.Client.DeleteObject(ctx, token, objectType, id)
}

