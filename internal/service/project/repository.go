package project

import (
	"context"
	"database/sql"
	"sync"

	"github.com/arthurlee945/Docrilla/internal/errors"
	"github.com/arthurlee945/Docrilla/internal/model"
	"github.com/jmoiron/sqlx"
)

type Repository interface {
	GetAll(ctx context.Context) ([]model.Project, error)
	GetOverviewById(ctx context.Context, uuid string) (*model.Project, error)
	GetDetailById(ctx context.Context, uuid string) (*model.Project, error)
	Create(ctx context.Context, proj *model.Project) (*model.Project, error)
	Update(ctx context.Context, proj *model.Project) error
	Delete(ctx context.Context, uuid string) error
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{
		db,
	}
}

func (r *repository) GetAll(ctx context.Context) ([]model.Project, error) {
	projects := []model.Project{}
	if err := r.db.SelectContext(ctx, &projects, "SELECT uuid, title, description, archived, token, route, created_at, visited_at FROM project LIMIT 10"); err != nil {
		return nil, ErrRepoGet.Wrap(err)
	}
	return projects, nil
}

func (r *repository) GetOverviewById(ctx context.Context, uuid string) (*model.Project, error) {
	proj := new(model.Project)
	if err := r.db.GetContext(ctx, proj, `
	SELECT uuid, title, description, archived, token, route, created_at, visited_at 
	FROM project WHERE uuid = $1
	`, uuid); err != nil {
		return nil, ErrRepoGet.Wrap(err)
	}
	return proj, nil
}

func (r *repository) GetDetailById(ctx context.Context, uuid string) (*model.Project, error) {
	proj, fields := new(model.Project), []model.Field{}
	if err := r.db.GetContext(ctx, proj, `SELECT * FROM project WHERE uuid = $1`, uuid); err != nil {
		return nil, ErrRepoGet.Wrap(err)
	}
	if err := r.db.SelectContext(ctx, &fields, `SELECT * FROM field WHERE project_id = $1`, proj.ID); err != nil {
		return nil, ErrRepoGet.Wrap(err)
	}
	proj.Fields = fields
	return proj, nil
}

func (r *repository) Create(ctx context.Context, proj *model.Project) (*model.Project, error) {
	rows, err := r.db.NamedQueryContext(ctx, `
		INSERT INTO project ( user_id, title, description, document_url, token, route) 
		VALUES (:user_id, :title, :description, :document_url, :token, :route) RETURNING *
		`, proj)
	if err != nil {
		return nil, ErrRepoCreate.Wrap(err)
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, ErrRepoCreate.Wrap(errors.ErrNotFound)
	}
	newProj := new(model.Project)
	if err := rows.StructScan(newProj); err != nil {
		return nil, errors.ErrUnknown.Wrap(err)
	}
	return newProj, nil
}

func (r *repository) Update(ctx context.Context, proj *model.Project) error {
	txCtx, txCancel := context.WithCancel(ctx)
	defer txCancel()
	tx, err := r.db.BeginTxx(txCtx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback()
	errChan, waitChan := make(chan error), make(chan struct{})

	go func() {
		wg := sync.WaitGroup{}
		wg.Add(len(proj.Fields) + 1)
		go func() {
			defer wg.Done()
			if _, err := tx.NamedExecContext(txCtx,
				`UPDATE project
			SET
			title = COALESCE(:title, title),
			description = COALESCE(:description, description),
			document_url = COALESCE(:document_url, document_url),
			route = COALESCE(:route, route),
			token = COALESCE(:token, token),
			archived = COALESCE(:archived, archived),
			visited_at = COALESCE(:visited_at, visited_at)
			WHERE uuid=:uuid
			`, proj); err != nil {
				errChan <- errors.Error("Project Update Error").Wrap(err)
				txCancel()
			}
		}()

		for _, field := range proj.Fields {
			go func(field *model.Field) {
				defer wg.Done()
				if _, err := tx.NamedExecContext(txCtx,
					`UPDATE field
				SET
				x1 = COALESCE(:x1, x1),
				y1 = COALESCE(:y1, y1),
				x2 = COALESCE(:x2, x2),
				y2 = COALESCE(:y2, y2),
				page = COALESCE(:page, page),
				type = COALESCE(:type, type)
				WHERE uuid=:uuid
				`,
					field); err != nil {
					errChan <- errors.Error("Field Update Error").Wrap(err)
					txCancel()
				}
			}(&field)
		}

		wg.Wait()
		close(waitChan)
	}()

	select {
	case err := <-errChan:
		return ErrRepoUpdate.Wrap(err)
	case <-waitChan:
		if err := tx.Commit(); err != nil {
			return ErrRepoUpdate.Wrap(err)
		}
		return nil
	}
}

func (r *repository) Delete(ctx context.Context, uuid string) error {
	if _, err := r.db.ExecContext(ctx, `
	DELETE FROM project
	WHERE uuid = $1 
	`, uuid); err != nil {
		return ErrRepoDelete.Wrap(err)
	}
	return nil
}
