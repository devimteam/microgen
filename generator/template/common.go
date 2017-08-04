package template

import (
	. "github.com/dave/jennifer/jen"
	"github.com/devimteam/microgen/parser"
	"github.com/devimteam/microgen/util"
)

// Renders struct field
//
// Visit *entity.Visit `json:"visit"`
//
func structField(field *parser.FuncField) Code {
	s := Id(util.ToUpperFirst(field.Name))

	s.Add(fieldType(field))

	s.Tag(map[string]string{"json": util.ToSnakeCase(field.Name)})

	return s
}

// Renders func params
//
// visit *entity.Visit
//
func funcParams(fields []*parser.FuncField) Code {
	c := Make()

	for _, field := range fields {
		c.Id(field.Name).Add(fieldType(field))
	}

	return c
}

// Renders field type for given func field
//
// *repository.Visit
//
func fieldType(field *parser.FuncField) Code {
	c := Make()

	if field.IsArray {
		c.Index()
	}

	if field.IsPointer {
		c.Op("*")
	}

	if field.Package != nil {
		c.Qual(field.Package.Path, field.Type)
	} else {
		c.Id(field.Type)
	}

	return c
}
