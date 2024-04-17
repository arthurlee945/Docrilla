package store

import (
	"context"
	"database/sql"
	"sync"

	"github.com/arthurlee945/Docrilla/internal/errors"
	"github.com/arthurlee945/Docrilla/internal/model"
	"github.com/jmoiron/sqlx"
)

const (
	ErrProjectFailedGet    = errors.Error("project_failed_get: couldn't find the project.")
	ErrProjectFailedCreate = errors.Error("project_failed_create: project couldn't be created.")
	ErrProjectFailedUpdate = errors.Error("project_failed_update: project couldn't update.")
	ErrProjectFailedDelete = errors.Error("project_failed_delete: project couldn't delete.")
)

type Store struct {
	db *sqlx.DB
}

func NewStore(db *sqlx.DB) *Store {
	return &Store{
		db,
	}
}

func (ps *Store) GetProjectOverview(ctx context.Context, user *model.User, uuid string) (*model.Project, error) {
	proj := new(model.Project)
	if err := ps.db.GetContext(ctx, proj, `
	SELECT uuid, title, description, archived, created_at, visited_at 
	FROM project WHERE uuid = $1 AND user_id = $2
	`, uuid, user.ID); err != nil {
		return nil, err
	}
	return proj, nil
}

func (ps *Store) GetProjectDetail(ctx context.Context, user *model.User, uuid string) (*model.Project, error) {
	proj, fields := new(model.Project), []model.Field{}
	if err := ps.db.GetContext(ctx, proj, `SELECT * FROM project WHERE uuid = $1 AND user_id = $2`, uuid, user.ID); err != nil {
		return nil, err
	}
	if err := ps.db.SelectContext(ctx, &fields, `SELECT * FROM field WHERE project_id = $1`, proj.ID); err != nil {
		return nil, err
	}
	proj.Fields = fields
	return proj, nil
}

func (ps *Store) CreateProject(ctx context.Context, user *model.User, proj *model.Project) (*model.Project, error) {
	rows, err := ps.db.NamedQueryContext(ctx, `
		INSERT INTO project ( user_id, title, description, document_url) 
		VALUES (:user_id, :title, :description, :document_url) RETURNING *
		`, proj)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, ErrProjectFailedCreate.Wrap(errors.ErrNotFound)
	}
	newProj := new(model.Project)
	if err := rows.StructScan(newProj); err != nil {
		return nil, errors.ErrUnknown.Wrap(err)
	}
	return newProj, nil
}

func (ps *Store) UpdateProject(ctx context.Context, user *model.User, proj *model.Project) error {
	txCtx, txCancel := context.WithCancel(ctx)
	defer txCancel()
	tx, err := ps.db.BeginTxx(txCtx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback()
	proj.UserID = user.ID
	errChan, waitChan := make(chan error), make(chan struct{})

	wg := sync.WaitGroup{}
	wg.Add(len(proj.Fields) + 1)
	go func() {
		go func() {
			defer func() {
				txCancel()
				wg.Done()
			}()
			if _, err := tx.NamedExecContext(txCtx,
				`UPDATE project
			SET
			title = COALESCE(:title, title),
			description = COALESCE(:description, description),
			document_url = COALESCE(:document_url, document_url),
			archived = COALESCE(:archived, archived),
			visitedAt = COALESCE(:visited_at, visited_at)
			WHERE id=:id AND user_id=:user_id
			`, proj); err != nil {
				errChan <- errors.Error("Project Update Error").Wrap(err)
			}
		}()

		for _, f := range proj.Fields {
			field := f
			go func() {
				defer func() {
					txCancel()
					wg.Done()
				}()
				if _, err := tx.NamedExecContext(txCtx,
					`UPDATE field
				SET
				x1 = COALESCE(:x1, x1),
				y1 = COALESCE(:y1, y1),
				x2 = COALESCE(:x2, x2),
				y2 = COALESCE(:y2, y2),
				page = COALESCE(:page, page),
				type = COALESCE(:type, type)
				WHERE id=:id AND project_id=:project_id
				`,
					field); err != nil {
					errChan <- errors.Error("Field Update Error").Wrap(err)
				}
			}()
		}

		wg.Wait()
		close(waitChan)
	}()

	select {
	case err := <-errChan:
		return ErrProjectFailedUpdate.Wrap(err)
	case <-waitChan:
		if err := tx.Commit(); err != nil {
			return ErrProjectFailedUpdate.Wrap(err)
		}
		return nil
	}
}

func (ps *Store) DeleteProject(ctx context.Context, user *model.User, proj *model.Project) error {
	proj.UserID = user.ID
	if _, err := ps.db.NamedExecContext(ctx, `
	DELETE FROM project
	WHERE uuid = :uuid AND user_id = :user_id
	`, proj); err != nil {
		return ErrProjectFailedDelete.Wrap(err)
	}
	return nil
}
