package controller

import (
	"github.com/bassista/go_spin/internal/cache"
	"github.com/bassista/go_spin/internal/repository"
	"github.com/go-playground/validator/v10"
)

// GroupCrudService implements CrudService for groups.
type GroupCrudService struct {
	Store cache.GroupStore
}

func (s *GroupCrudService) All() ([]repository.Group, error) {
	doc, err := s.Store.Snapshot()
	if err != nil {
		return nil, err
	}
	return doc.Groups, nil
}

func (s *GroupCrudService) Add(item repository.Group) ([]repository.Group, error) {
	doc, err := s.Store.AddGroup(item)
	if err != nil {
		return nil, err
	}
	return doc.Groups, nil
}

func (s *GroupCrudService) Remove(name string) ([]repository.Group, error) {
	doc, err := s.Store.RemoveGroup(name)
	if err != nil {
		return nil, err
	}
	return doc.Groups, nil
}

// GroupCrudValidator implements CrudValidator for groups.
type GroupCrudValidator struct {
	validator *validator.Validate
}

func (v *GroupCrudValidator) Validate(item repository.Group) error {
	return v.validator.Struct(item)
}
