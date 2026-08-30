package ast

import (
	"fmt"
	"strings"

	"github.com/ettle/strcase"
)

// ColumnCustomType represents custom PostgreSQL types such as enums and
// domains. It is used by Column.CustomType to indicate the presence of a custom
// type.
type ColumnCustomType int

// Column represents intermediate column metadata.
type Column struct {
	Name       string
	Type       string
	CustomType ColumnCustomType
	IsNullable bool
	IsArray    bool
	Default    *string
}

// ColumnTypeDefinition defines the model type and the query column formatter
// for a PostgreSQL data type.
type ColumnTypeDefinition struct {
	ModelType       string
	QueryColumnFunc func(
		modelName string,
		col *Column,
		modelType string,
	) string
}

// CustomTypeDefinition defines model type generation and query column
// formatting for user-defined PostgreSQL custom types.
type CustomTypeDefinition struct {
	ModelTypeFunc   func(col *Column) string
	QueryColumnFunc func(
		modelName string,
		col *Column,
		modelType string,
	) string
}

// CustomNone indicates a standard built-in column type.
const CustomNone ColumnCustomType = 0

// CustomEnum indicates a user-defined PostgreSQL enum type.
const CustomEnum ColumnCustomType = 1

// CustomDomain indicates a user-defined PostgreSQL domain type.
const CustomDomain ColumnCustomType = 2

// pgTypeToModelTypeFallback is the fallback Go type when no mapping is found.
var pgTypeToModelTypeFallback = "string"

// mapCustomTypeToCustomTypeDefinition maps ColumnCustomType identifiers to
// their CustomTypeDefinition.
var mapCustomTypeToCustomTypeDefinition =
	map[ColumnCustomType]CustomTypeDefinition{
	CustomEnum: {

		ModelTypeFunc: func(col *Column) string {
			return strcase.ToGoPascal(col.Type)
		},
		QueryColumnFunc: formatTypedColumn("StringColumn"),
	},
	CustomDomain: {
		ModelTypeFunc: func(col *Column) string {
			return strcase.ToGoPascal(col.Type)
		},
		QueryColumnFunc: formatTypedColumn("Column"),
	},
}

// mapPgTypeToColumnTypeDefinition maps PostgreSQL type names to their
// ColumnTypeDefinition.
var mapPgTypeToColumnTypeDefinition = map[string]ColumnTypeDefinition{
	// UUID
	"uuid": {
		ModelType:       "uuid.UUID",
		QueryColumnFunc: formatUUIDColumn,
	},

	// Numeric
	"bigint": {
		ModelType:       "int64",
		QueryColumnFunc: formatTypedColumn("NumericColumn"),
	},
	"bigserial": {
		ModelType:       "int64",
		QueryColumnFunc: formatTypedColumn("NumericColumn"),
	},
	"double precision": {
		ModelType:       "float64",
		QueryColumnFunc: formatTypedColumn("NumericColumn"),
	},
	"float4": {
		ModelType:       "float32",
		QueryColumnFunc: formatTypedColumn("NumericColumn"),
	},
	"float8": {
		ModelType:       "float64",
		QueryColumnFunc: formatTypedColumn("NumericColumn"),
	},
	"int2": {
		ModelType:       "int16",
		QueryColumnFunc: formatTypedColumn("NumericColumn"),
	},
	"int4": {
		ModelType:       "int32",
		QueryColumnFunc: formatTypedColumn("NumericColumn"),
	},
	"int8": {
		ModelType:       "int64",
		QueryColumnFunc: formatTypedColumn("NumericColumn"),
	},
	"integer": {
		ModelType:       "int32",
		QueryColumnFunc: formatTypedColumn("NumericColumn"),
	},
	"real": {
		ModelType:       "float32",
		QueryColumnFunc: formatTypedColumn("NumericColumn"),
	},
	"serial": {
		ModelType:       "int32",
		QueryColumnFunc: formatTypedColumn("NumericColumn"),
	},
	"smallint": {
		ModelType:       "int16",
		QueryColumnFunc: formatTypedColumn("NumericColumn"),
	},
	"smallserial": {
		ModelType:       "int16",
		QueryColumnFunc: formatTypedColumn("NumericColumn"),
	},

	// String
	"bpchar": {
		ModelType:       "string",
		QueryColumnFunc: formatTypedColumn("StringColumn"),
	},
	"char": {
		ModelType:       "string",
		QueryColumnFunc: formatTypedColumn("StringColumn"),
	},
	"name": {
		ModelType:       "string",
		QueryColumnFunc: formatTypedColumn("StringColumn"),
	},
	"text": {
		ModelType:       "string",
		QueryColumnFunc: formatTypedColumn("StringColumn"),
	},
	"varchar": {
		ModelType:       "string",
		QueryColumnFunc: formatTypedColumn("StringColumn"),
	},

	// Date / Time
	"date": {
		ModelType:       "time.Time",
		QueryColumnFunc: formatTypedColumn("TimestampColumn"),
	},
	"time": {
		ModelType:       "time.Time",
		QueryColumnFunc: formatTypedColumn("TimestampColumn"),
	},
	"time with time zone": {
		ModelType:       "time.Time",
		QueryColumnFunc: formatTypedColumn("TimestampColumn"),
	},
	"timestamp": {
		ModelType:       "time.Time",
		QueryColumnFunc: formatTypedColumn("TimestampColumn"),
	},
	"timestamp with time zone": {
		ModelType:       "time.Time",
		QueryColumnFunc: formatTypedColumn("TimestampColumn"),
	},
	"timestamp without time zone": {
		ModelType:       "time.Time",
		QueryColumnFunc: formatTypedColumn("TimestampColumn"),
	},
	"timestamptz": {
		ModelType:       "time.Time",
		QueryColumnFunc: formatTypedColumn("TimestampColumn"),
	},
	"timetz": {
		ModelType:       "time.Time",
		QueryColumnFunc: formatTypedColumn("TimestampColumn"),
	},

	// Decimal & Duration
	"decimal": {
		ModelType:       "pgtype.Numeric",
		QueryColumnFunc: formatTypedColumn("DecimalColumn"),
	},
	"numeric": {
		ModelType:       "pgtype.Numeric",
		QueryColumnFunc: formatTypedColumn("DecimalColumn"),
	},
	"interval": {
		ModelType:       "time.Duration",
		QueryColumnFunc: formatTypedColumn("DurationColumn"),
	},

	// Geometry
	"box": {
		ModelType:       "pgtype.Box",
		QueryColumnFunc: formatTypedColumn("GeoColumn"),
	},
	"circle": {
		ModelType:       "pgtype.Circle",
		QueryColumnFunc: formatTypedColumn("GeoColumn"),
	},
	"line": {
		ModelType:       "pgtype.Line",
		QueryColumnFunc: formatTypedColumn("GeoColumn"),
	},
	"lseg": {
		ModelType:       "pgtype.Lseg",
		QueryColumnFunc: formatTypedColumn("GeoColumn"),
	},
	"path": {
		ModelType:       "pgtype.Path",
		QueryColumnFunc: formatTypedColumn("GeoColumn"),
	},
	"point": {
		ModelType:       "pgtype.Point",
		QueryColumnFunc: formatTypedColumn("GeoColumn"),
	},
	"polygon": {
		ModelType:       "pgtype.Polygon",
		QueryColumnFunc: formatTypedColumn("GeoColumn"),
	},

	// JSON
	"json": {
		ModelType:       "[]byte",
		QueryColumnFunc: formatTypedColumn("JSONColumn"),
	},
	"jsonb": {
		ModelType:       "[]byte",
		QueryColumnFunc: formatTypedColumn("JSONColumn"),
	},

	// Boolean & Binary
	"bool": {
		ModelType:       "bool",
		QueryColumnFunc: formatTypedColumn("Column"),
	},
	"boolean": {
		ModelType:       "bool",
		QueryColumnFunc: formatTypedColumn("Column"),
	},
	"bytea": {
		ModelType:       "[]byte",
		QueryColumnFunc: formatTypedColumn("Column"),
	},

	// Network & System IDs
	"cidr": {
		ModelType:       "pgtype.CIDR",
		QueryColumnFunc: formatTypedColumn("Column"),
	},
	"inet": {
		ModelType:       "pgtype.Inet",
		QueryColumnFunc: formatTypedColumn("Column"),
	},
	"macaddr": {
		ModelType:       "pgtype.Macaddr",
		QueryColumnFunc: formatTypedColumn("Column"),
	},
	"macaddr8": {
		ModelType:       "pgtype.Macaddr",
		QueryColumnFunc: formatTypedColumn("Column"),
	},
	"cid": {
		ModelType:       "uint32",
		QueryColumnFunc: formatTypedColumn("Column"),
	},
	"oid": {
		ModelType:       "uint32",
		QueryColumnFunc: formatTypedColumn("Column"),
	},
	"xid": {
		ModelType:       "uint32",
		QueryColumnFunc: formatTypedColumn("Column"),
	},
	"xid8": {
		ModelType:       "uint64",
		QueryColumnFunc: formatTypedColumn("Column"),
	},
	"pg_lsn": {
		ModelType:       "pgtype.Uint64",
		QueryColumnFunc: formatTypedColumn("Column"),
	},
	"tid": {
		ModelType:       "string",
		QueryColumnFunc: formatTypedColumn("Column"),
	},
	"bit": {
		ModelType:       "pgtype.Varbit",
		QueryColumnFunc: formatTypedColumn("Column"),
	},
	"bit varying": {
		ModelType:       "pgtype.Varbit",
		QueryColumnFunc: formatTypedColumn("Column"),
	},
	"varbit": {
		ModelType:       "pgtype.Varbit",
		QueryColumnFunc: formatTypedColumn("Column"),
	},
	"money": {
		ModelType:       "pgtype.Money",
		QueryColumnFunc: formatTypedColumn("Column"),
	},
}

// formatUUIDColumn formats a single-type-parameter UUIDColumn definition.
func formatUUIDColumn(
	modelName string,
	_ *Column,
	_ string,
) string {
	return fmt.Sprintf("column.UUIDColumn[%s]", modelName)
}

// formatTypedColumn returns a formatter closure for strongly typed query
// column representations.
func formatTypedColumn(
	baseColType string,
) func(modelName string, col *Column, modelType string) string {
	return func(modelName string, col *Column, modelType string) string {
		colType := baseColType
		if col.IsNullable {
			switch baseColType {
			case "NumericColumn", "StringColumn", "JSONColumn", "GeoColumn":
				colType = "Nullable" + baseColType
			default:
				colType = "NullableColumn"
			}
		}
		return fmt.Sprintf(
			"column.%s[%s, %s]",
			colType,
			modelName,
			modelType,
		)
	}
}

// formatArrayColumn formats a 2-type-parameter ArrayColumn definition.
func formatArrayColumn(modelName string, col *Column) string {
	colType := "ArrayColumn"
	if col.IsNullable {
		colType = "Nullable" + colType
	}
	elemCol := &Column{
		Name:       col.Name,
		Type:       col.Type,
		CustomType: col.CustomType,
	}
	return fmt.Sprintf(
		"column.%s[%s, %s]",
		colType,
		modelName,
		elemCol.ToModelType(),
	)
}

// typeDefinition resolves the base model type and query column formatter.
func (c *Column) typeDefinition() (
	string,
	func(string, *Column, string) string,
) {
	customDef, okCustomDef :=
		mapCustomTypeToCustomTypeDefinition[c.CustomType]
	if okCustomDef {
		return customDef.ModelTypeFunc(c), customDef.QueryColumnFunc
	}


	lowerType := strings.ToLower(c.Type)
	pgDef, okPgDef := mapPgTypeToColumnTypeDefinition[lowerType]
	if okPgDef {
		return pgDef.ModelType, pgDef.QueryColumnFunc
	}

	return pgTypeToModelTypeFallback, formatTypedColumn("Column")
}

// ToModelType returns the Go type representation for model and mutator fields.
func (c *Column) ToModelType() string {
	modelType, _ := c.typeDefinition()
	if c.IsArray {
		modelType = "[]" + modelType
	}
	if c.IsNullable {
		return "*" + modelType
	}
	return modelType
}

// ToQueryColumnType returns the type-safe schema query column definition.
func (c *Column) ToQueryColumnType(modelName string) string {
	if c.IsArray {
		return formatArrayColumn(modelName, c)
	}

	modelType, queryFunc := c.typeDefinition()
	return queryFunc(modelName, c, modelType)
}

// ToGoPascalName returns the PascalCase name of the column.
func (c *Column) ToGoPascalName() string {
	return strcase.ToGoPascal(c.Name)
}

// ColumnNames extracts the column names from a slice of *Column.
func ColumnNames(cols []*Column) []string {
	names := make([]string, len(cols))
	for i, c := range cols {
		names[i] = c.Name
	}
	return names
}
