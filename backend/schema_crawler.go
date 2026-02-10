package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Field represents a database field
type Field struct {
	Name         string            `json:"name"`
	Type         string            `json:"type"`
	GoType       string            `json:"go_type"`
	Tags         map[string]string `json:"tags,omitempty"`
	IsPrimaryKey bool              `json:"is_primary_key"`
	IsNullable   bool              `json:"is_nullable"`
	IsUnique     bool              `json:"is_unique"`
	IsForeignKey bool              `json:"is_foreign_key"`
	References   string            `json:"references,omitempty"`
}

// Table represents a database table
type Table struct {
	Name        string            `json:"name"`
	GoStructName string           `json:"go_struct_name"`
	Fields      []Field           `json:"fields"`
	Indexes     []string          `json:"indexes,omitempty"`
	Relations   map[string]string `json:"relations,omitempty"`
}

// Schema represents the complete database schema
type Schema struct {
	Tables   []Table           `json:"tables"`
	Metadata map[string]string `json:"metadata"`
}

var (
	dirPath    = flag.String("dir", ".", "Directory to scan for Go files")
	outputFile = flag.String("output", "schema.json", "Output file for schema (JSON)")
	sqlOutput  = flag.String("sql", "", "Output SQL file (optional)")
	verbose    = flag.Bool("v", false, "Verbose output")
)

func main() {
	flag.Parse()

	schema := &Schema{
		Tables:   make([]Table, 0),
		Metadata: make(map[string]string),
	}

	fmt.Printf("🔍 Scanning directory: %s\n", *dirPath)

	err := filepath.Walk(*dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip vendor, node_modules, and hidden directories
		if info.IsDir() && (info.Name() == "vendor" || info.Name() == "node_modules" || strings.HasPrefix(info.Name(), ".")) {
			return filepath.SkipDir
		}

		// Only process .go files (excluding test files)
		if !info.IsDir() && strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			if *verbose {
				fmt.Printf("  📄 Parsing: %s\n", path)
			}
			tables, err := parseGoFile(path)
			if err != nil {
				fmt.Printf("  ⚠️  Error parsing %s: %v\n", path, err)
				return nil
			}
			schema.Tables = append(schema.Tables, tables...)
		}

		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error walking directory: %v\n", err)
		os.Exit(1)
	}

	if len(schema.Tables) == 0 {
		fmt.Println("⚠️  No database models found")
		return
	}

	fmt.Printf("✅ Found %d table(s)\n\n", len(schema.Tables))

	// Output JSON schema
	jsonData, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
		os.Exit(1)
	}

	err = os.WriteFile(*outputFile, jsonData, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing JSON file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("📝 Schema written to: %s\n", *outputFile)

	// Output SQL if requested
	if *sqlOutput != "" {
		sql := generateSQL(schema)
		err = os.WriteFile(*sqlOutput, []byte(sql), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing SQL file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("📝 SQL schema written to: %s\n", *sqlOutput)
	}

	// Print summary
	printSummary(schema)
}

func parseGoFile(filename string) ([]Table, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	tables := make([]Table, 0)

	ast.Inspect(node, func(n ast.Node) bool {
		typeSpec, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}

		structType, ok := typeSpec.Type.(*ast.StructType)
		if !ok {
			return true
		}

		// Check if this struct looks like a database model
		if !isDBModel(structType) {
			return true
		}

		table := parseStruct(typeSpec.Name.Name, structType)
		if len(table.Fields) > 0 {
			tables = append(tables, table)
		}

		return true
	})

	return tables, nil
}

func isDBModel(structType *ast.StructType) bool {
	// Look for common ORM tags or field patterns
	for _, field := range structType.Fields.List {
		if field.Tag != nil {
			tag := field.Tag.Value
			// Check for common ORM tags
			if strings.Contains(tag, "gorm:") || strings.Contains(tag, "db:") ||
				strings.Contains(tag, "sql:") || strings.Contains(tag, "bun:") {
				return true
			}
		}

		// Check for common DB field names
		for _, name := range field.Names {
			fieldName := strings.ToLower(name.Name)
			if fieldName == "id" || fieldName == "createdat" || fieldName == "updatedat" ||
				fieldName == "deletedat" {
				return true
			}
		}
	}
	return false
}

func parseStruct(structName string, structType *ast.StructType) Table {
	table := Table{
		GoStructName: structName,
		Name:         toSnakeCase(structName),
		Fields:       make([]Field, 0),
		Relations:    make(map[string]string),
	}

	for _, field := range structType.Fields.List {
		if len(field.Names) == 0 {
			// Embedded field - could be gorm.Model or similar
			if ident, ok := field.Type.(*ast.Ident); ok {
				if ident.Name == "Model" {
					// Add common gorm.Model fields
					table.Fields = append(table.Fields, getGormModelFields()...)
				}
			}
			continue
		}

		for _, name := range field.Names {
			f := Field{
				Name:   toSnakeCase(name.Name),
				GoType: getTypeName(field.Type),
				Tags:   make(map[string]string),
			}

			// Parse tags
			if field.Tag != nil {
				parseTags(field.Tag.Value, &f, &table)
			}

			// Determine SQL type
			f.Type = goTypeToSQLType(f.GoType)

			table.Fields = append(table.Fields, f)
		}
	}

	return table
}

func parseTags(tagString string, field *Field, table *Table) {
	// Remove backticks
	tagString = strings.Trim(tagString, "`")

	// Parse different tag formats
	tagRe := regexp.MustCompile(`(\w+):"([^"]*)"`)
	matches := tagRe.FindAllStringSubmatch(tagString, -1)

	for _, match := range matches {
		if len(match) < 3 {
			continue
		}

		tagName := match[1]
		tagValue := match[2]
		field.Tags[tagName] = tagValue

		switch tagName {
		case "gorm":
			parseGormTag(tagValue, field, table)
		case "db", "sql":
			parseDBTag(tagValue, field)
		case "bun":
			parseBunTag(tagValue, field, table)
		}
	}
}

func parseGormTag(tagValue string, field *Field, table *Table) {
	parts := strings.Split(tagValue, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		
		if part == "primaryKey" || part == "primary_key" {
			field.IsPrimaryKey = true
		} else if part == "unique" {
			field.IsUnique = true
		} else if strings.HasPrefix(part, "column:") {
			field.Name = strings.TrimPrefix(part, "column:")
		} else if strings.HasPrefix(part, "type:") {
			field.Type = strings.TrimPrefix(part, "type:")
		} else if strings.HasPrefix(part, "foreignKey:") {
			field.IsForeignKey = true
		} else if strings.HasPrefix(part, "references:") {
			field.References = strings.TrimPrefix(part, "references:")
		} else if strings.HasPrefix(part, "index") {
			indexName := part
			if strings.Contains(part, ":") {
				indexName = strings.Split(part, ":")[1]
			}
			table.Indexes = append(table.Indexes, indexName)
		}
	}
}

func parseDBTag(tagValue string, field *Field) {
	parts := strings.Split(tagValue, ",")
	if len(parts) > 0 && parts[0] != "" && parts[0] != "-" {
		field.Name = parts[0]
	}
}

func parseBunTag(tagValue string, field *Field, table *Table) {
	parts := strings.Split(tagValue, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		
		if part == "pk" {
			field.IsPrimaryKey = true
		} else if part == "unique" {
			field.IsUnique = true
		} else if strings.HasPrefix(part, "type:") {
			field.Type = strings.TrimPrefix(part, "type:")
		}
	}
}

func getTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + getTypeName(t.X)
	case *ast.ArrayType:
		return "[]" + getTypeName(t.Elt)
	case *ast.SelectorExpr:
		return getTypeName(t.X) + "." + t.Sel.Name
	default:
		return "unknown"
	}
}

func goTypeToSQLType(goType string) string {
	// Remove pointer indicator
	goType = strings.TrimPrefix(goType, "*")

	switch goType {
	case "string":
		return "VARCHAR(255)"
	case "int", "int32":
		return "INTEGER"
	case "int64":
		return "BIGINT"
	case "uint", "uint32":
		return "INTEGER UNSIGNED"
	case "uint64":
		return "BIGINT UNSIGNED"
	case "bool":
		return "BOOLEAN"
	case "float32":
		return "FLOAT"
	case "float64":
		return "DOUBLE"
	case "time.Time":
		return "TIMESTAMP"
	case "[]byte":
		return "BLOB"
	default:
		if strings.HasPrefix(goType, "[]") {
			return "JSON"
		}
		return "VARCHAR(255)"
	}
}

func getGormModelFields() []Field {
	return []Field{
		{Name: "id", Type: "BIGINT UNSIGNED", GoType: "uint", IsPrimaryKey: true},
		{Name: "created_at", Type: "TIMESTAMP", GoType: "time.Time"},
		{Name: "updated_at", Type: "TIMESTAMP", GoType: "time.Time"},
		{Name: "deleted_at", Type: "TIMESTAMP", GoType: "*time.Time", IsNullable: true},
	}
}

func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

func generateSQL(schema *Schema) string {
	var sql strings.Builder

	sql.WriteString("-- Auto-generated SQL Schema\n")
	sql.WriteString("-- Generated from Go backend models\n\n")

	for _, table := range schema.Tables {
		sql.WriteString(fmt.Sprintf("CREATE TABLE %s (\n", table.Name))

		for i, field := range table.Fields {
			sql.WriteString(fmt.Sprintf("  %s %s", field.Name, field.Type))

			if field.IsPrimaryKey {
				sql.WriteString(" PRIMARY KEY")
			}
			if !field.IsNullable && !field.IsPrimaryKey {
				sql.WriteString(" NOT NULL")
			}
			if field.IsUnique {
				sql.WriteString(" UNIQUE")
			}

			if i < len(table.Fields)-1 {
				sql.WriteString(",")
			}
			sql.WriteString("\n")
		}

		sql.WriteString(");\n\n")

		// Add indexes
		for _, index := range table.Indexes {
			sql.WriteString(fmt.Sprintf("CREATE INDEX %s ON %s;\n", index, table.Name))
		}
		if len(table.Indexes) > 0 {
			sql.WriteString("\n")
		}
	}

	return sql.String()
}

func printSummary(schema *Schema) {
	fmt.Println("\n📊 Schema Summary:")
	fmt.Println(strings.Repeat("=", 60))

	for _, table := range schema.Tables {
		fmt.Printf("\n📋 Table: %s (struct: %s)\n", table.Name, table.GoStructName)
		fmt.Printf("   Fields: %d\n", len(table.Fields))

		for _, field := range table.Fields {
			fmt.Printf("   - %s: %s", field.Name, field.Type)
			if field.IsPrimaryKey {
				fmt.Print(" [PK]")
			}
			if field.IsUnique {
				fmt.Print(" [UNIQUE]")
			}
			if field.IsForeignKey {
				fmt.Printf(" [FK -> %s]", field.References)
			}
			fmt.Println()
		}

		if len(table.Indexes) > 0 {
			fmt.Printf("   Indexes: %v\n", table.Indexes)
		}
	}

	fmt.Println(strings.Repeat("=", 60))
}
