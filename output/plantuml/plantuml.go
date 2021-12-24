package plantuml

import (
	"embed"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"

	"github.com/HMasataka/tbls/config"
	"github.com/HMasataka/tbls/output"
	"github.com/HMasataka/tbls/schema"
	"github.com/pkg/errors"
)

//go:embed templates/*
var tmpl embed.FS

// PlantUML struct
type PlantUML struct {
	config *config.Config
	tmpl   embed.FS
}

// New return PlantUML
func New(c *config.Config) *PlantUML {
	return &PlantUML{
		config: c,
		tmpl:   tmpl,
	}
}

func (p *PlantUML) schemaTemplate() (string, error) {
	if len(p.config.Templates.PUML.Schema) > 0 {
		tb, err := os.ReadFile(p.config.Templates.PUML.Schema)
		if err != nil {
			return "", errors.WithStack(err)
		}
		return string(tb), nil
	} else {
		tb, err := p.tmpl.ReadFile("templates/schema.puml.tmpl")
		if err != nil {
			return "", errors.WithStack(err)
		}
		return string(tb), nil
	}
}

func (p *PlantUML) tableTemplate() (string, error) {
	if len(p.config.Templates.PUML.Table) > 0 {
		tb, err := os.ReadFile(p.config.Templates.PUML.Table)
		if err != nil {
			return "", errors.WithStack(err)
		}
		return string(tb), nil
	} else {
		tb, err := p.tmpl.ReadFile("templates/table.puml.tmpl")
		if err != nil {
			return "", errors.WithStack(err)
		}
		return string(tb), nil
	}
}

// OutputSchema output dot format for full relation.
func (p *PlantUML) OutputSchema(wr io.Writer, s *schema.Schema) error {
	for _, t := range s.Tables {
		err := addPrefix(t)
		if err != nil {
			return err
		}
	}

	ts, err := p.schemaTemplate()
	if err != nil {
		return errors.WithStack(err)
	}
	tmpl := template.Must(template.New(s.Name).Funcs(output.Funcs(&p.config.MergedDict)).Parse(ts))
	err = tmpl.Execute(wr, map[string]interface{}{
		"Schema":      s,
		"showComment": p.config.ER.Comment,
	})
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// OutputTable output dot format for table.
func (p *PlantUML) OutputTable(wr io.Writer, t *schema.Table) error {
	tables, relations, err := t.CollectTablesAndRelations(*p.config.ER.Distance, true)

	if err != nil {
		return errors.WithStack(err)
	}
	for _, t := range tables {
		if err := addPrefix(t); err != nil {
			return errors.WithStack(err)
		}
	}
	ts, err := p.tableTemplate()
	if err != nil {
		return errors.WithStack(err)
	}
	tmpl := template.Must(template.New(t.Name).Funcs(output.Funcs(&p.config.MergedDict)).Parse(ts))
	err = tmpl.Execute(wr, map[string]interface{}{
		"Table":       tables[0],
		"Tables":      tables[1:],
		"Relations":   relations,
		"showComment": p.config.ER.Comment,
	})
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func addPrefix(t *schema.Table) error {
	// PRIMARY KEY
	for _, i := range t.Indexes {
		if strings.Index(i.Def, "PRIMARY") < 0 {
			continue
		}
		for _, c := range i.Columns {
			column, err := t.FindColumnByName(c)
			if err != nil {
				return err
			}
			column.Name = fmt.Sprintf("+ %s", column.Name)
		}
	}
	// Foreign Key (Relations)
	for _, c := range t.Columns {
		if len(c.ParentRelations) > 0 && strings.Index(c.Name, "+") < 0 {
			c.Name = fmt.Sprintf("# %s", c.Name)
		}
	}
	return nil
}

func contains(rs []*schema.Relation, e *schema.Relation) bool {
	for _, r := range rs {
		if e == r {
			return true
		}
	}
	return false
}
