package model

import (
	"fmt"

	"github.com/arthurlee945/Docrilla/model/enum/field"
)

var FieldSchema = fmt.Sprintf(`
	CREATE TYPE IF NOT EXITS FieldType AS ENUM ('%v', '%v', '%v')

	CREATE TABLE IF NOT EXISTS Field (
		id SERIAL PRIMARY KEY,
		project_id INT NOT NULL,
		x1 NUMERIC NOT NULL,
		y1 NUMERIC NOT NULL,
		x2 NUMERIC NOT NULL,
		y2 NUMERIC NOT NULL,
		page INT NOT NULL,
		type FieldType NOT NULL,
		field_id TEXT NOT NULL,
		value TEXT NOT NULL,
		CONTRAINT fk_Project FOREIGN KEY(project_id) REFERENCES Project(id) ON DELETE CASCADE
	)
`, field.TEXT, field.NUMBER, field.IMAGE)

type Field struct {
	ContentID uint64 `db:"id"`
	X1        float64
	Y1        float64
	X2        float64
	Y2        float64
	Page      uint32
	Type      field.FieldType
	FieldId   string `db:"field_id"`
	Value     string
	Project   Project
}
