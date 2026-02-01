package controller

import (
	"github.com/bassista/go_spin/internal/cache"
	"github.com/bassista/go_spin/internal/repository"
	"github.com/go-playground/validator/v10"
)

// ContainerCrudService implements CrudService for containers.
type ContainerCrudService struct {
	Store cache.ContainerStore
}

func (s *ContainerCrudService) All() ([]repository.Container, error) {
	doc, err := s.Store.Snapshot()
	if err != nil {
		return nil, err
	}
	return doc.Containers, nil
}

func (s *ContainerCrudService) Add(item repository.Container) ([]repository.Container, error) {
	doc, err := s.Store.AddContainer(item)
	if err != nil {
		return nil, err
	}
	return doc.Containers, nil
}

func (s *ContainerCrudService) Remove(name string) ([]repository.Container, error) {
	doc, err := s.Store.RemoveContainer(name)
	if err != nil {
		return nil, err
	}
	return doc.Containers, nil
}

// ContainerCrudValidator implements CrudValidator for containers.
type ContainerCrudValidator struct {
	validator *validator.Validate
}

func (v *ContainerCrudValidator) Validate(item repository.Container) error {
	return v.validator.Struct(item)
}
