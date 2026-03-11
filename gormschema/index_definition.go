package gormschema

import (
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// ---------- Public API ----------

// Column selector + per-column options.
type Col[T any] struct {
	Sel   func(*T) any // MUST return a *pointer* to the struct field (e.g., `&m.TenantID`)
	Sort  string       // "", "asc", "desc"
	Nulls string       // "", "first", "last" (used as `sort:desc nulls last`)
}

func Field[T any](sel func(*T) any) Col[T] { return Col[T]{Sel: sel} }
func Asc[T any](c Col[T]) Col[T]           { c.Sort = "asc"; return c }
func Desc[T any](c Col[T]) Col[T]          { c.Sort = "desc"; return c }
func NullsFirst[T any](c Col[T]) Col[T]    { c.Nulls = "first"; return c }
func NullsLast[T any](c Col[T]) Col[T]     { c.Nulls = "last"; return c }

// IndexDefinition declares a composite (or single-column) index.
type IndexDefinition[T any] struct {
	Name    string
	Columns []Col[T] // order => priority:1..N
	Type    string   // e.g. "gin", "btree", "brin"
	Unique  bool
	Where   string // e.g. "deleted_at IS NULL"
}

// AutoMigrateModel inspects 'model' for an Indexes() method.
// If present, it uses those definitions to synthesize index tags on a
// cloned runtime type, then runs AutoMigrate on that clone.
// If not, it falls back to db.AutoMigrate(model).
func AutoMigrateModel(db *gorm.DB, model any) error {
	if model == nil {
		return fmt.Errorf("nil model")
	}

	base := indirectType(reflect.TypeOf(model))
	if base.Kind() != reflect.Struct {
		return fmt.Errorf("model must be a struct or *struct, got %v", base.Kind())
	}

	// Find Indexes method on a *pointer* receiver if needed.
	mv := reflect.ValueOf(model)
	var recv reflect.Value
	if mv.Kind() == reflect.Ptr {
		recv = mv
	} else {
		// create addressable copy to access pointer-receiver methods
		p := reflect.New(mv.Type())
		p.Elem().Set(mv)
		recv = p
	}

	method := recv.MethodByName("Indexes")
	if !method.IsValid() {
		// No Indexes() -> regular migration
		return db.AutoMigrate(model)
	}
	if method.Type().NumIn() != 0 || method.Type().NumOut() != 1 {
		// Unexpected signature; ignore gracefully.
		return db.AutoMigrate(model)
	}

	// Call Indexes() reflectively; result is a slice of IndexDefinition[T] (unknown T).
	out := method.Call(nil)[0]
	if out.Kind() != reflect.Slice {
		return db.AutoMigrate(model)
	}
	if out.Len() == 0 {
		return db.AutoMigrate(model)
	}

	// Build field -> index-tag fragments from the returned definitions.
	fieldToIndexTags, err := collectIndexTagsFromIndexesValue(base, out)
	if err != nil {
		return err
	}

	// Build cloned struct type with merged tags.
	fields := make([]reflect.StructField, 0, base.NumField())
	for i := 0; i < base.NumField(); i++ {
		sf := base.Field(i)
		// Keep only exported fields; GORM ignores unexported columns anyway.
		if sf.PkgPath != "" {
			continue
		}
		newTag := mergeIndexIntoGormTag(sf.Tag, fieldToIndexTags[sf.Name])
		fields = append(fields, reflect.StructField{
			Name:      sf.Name,
			Type:      sf.Type,
			Tag:       newTag,
			Anonymous: sf.Anonymous,
		})
	}

	dyn := reflect.StructOf(fields)
	ptr := reflect.New(dyn).Interface()

	// Respect custom table name if model implements Tabler.
	if tabler, ok := any(model).(schema.Tabler); ok {
		return db.Table(tabler.TableName()).AutoMigrate(ptr)
	}
	// Also handle pointer-receiver TableName() methods by asserting on *T when model is T.
	mt := reflect.TypeOf(model)
	var ptrModel any
	if mt.Kind() == reflect.Ptr {
		ptrModel = model
	} else {
		ptrModel = reflect.New(mt).Interface()
	}
	if tabler, ok := ptrModel.(schema.Tabler); ok {
		return db.Table(tabler.TableName()).AutoMigrate(ptr)
	}
	return db.AutoMigrate(ptr)
}

// -------- internals --------

func collectIndexTagsFromIndexesValue(baseStruct reflect.Type, defsSlice reflect.Value) (map[string][]string, error) {
	fieldToIndexTags := map[string][]string{}

	for i := 0; i < defsSlice.Len(); i++ {
		def := defsSlice.Index(i)
		if def.Kind() == reflect.Pointer {
			def = def.Elem()
		}
		if def.Kind() != reflect.Struct {
			return nil, fmt.Errorf("Indexes()[%d] is not a struct", i)
		}

		// Expect fields: Name string, Columns []Col[?], Unique bool, Where string.
		// Type is optional for backward compatibility with older reflected shapes.
		nameF := def.FieldByName("Name")
		colsF := def.FieldByName("Columns")
		typeF := def.FieldByName("Type")
		uniqueF := def.FieldByName("Unique")
		whereF := def.FieldByName("Where")

		if !nameF.IsValid() || !colsF.IsValid() || !uniqueF.IsValid() || !whereF.IsValid() {
			return nil, fmt.Errorf("Indexes()[%d] doesn't look like IndexDefinition", i)
		}
		name := nameF.String()
		indexType := ""
		if typeF.IsValid() {
			indexType = strings.TrimSpace(typeF.String())
		}
		unique := uniqueF.Bool()
		where := strings.TrimSpace(whereF.String())

		if colsF.Kind() != reflect.Slice {
			return nil, fmt.Errorf("Index %q: Columns is not a slice", name)
		}
		for j := 0; j < colsF.Len(); j++ {
			col := colsF.Index(j)
			if col.Kind() == reflect.Pointer {
				col = col.Elem()
			}
			if col.Kind() != reflect.Struct {
				return nil, fmt.Errorf("Index %q column %d: not a struct", name, j+1)
			}

			selF := col.FieldByName("Sel")   // func(*T) any
			sortF := col.FieldByName("Sort") // string
			nullF := col.FieldByName("Nulls")

			if !selF.IsValid() {
				return nil, fmt.Errorf("Index %q column %d: missing Sel", name, j+1)
			}
			fname, err := fieldNameFromSelectorValue(selF)
			if err != nil {
				return nil, fmt.Errorf("index %q column %d: %w", name, j+1, err)
			}

			parts := []string{
				"index:" + name,
				fmt.Sprintf("priority:%d", j+1),
			}
			if s := strings.TrimSpace(sortF.String()); s != "" {
				val := s
				if n := strings.TrimSpace(nullF.String()); n != "" {
					val = val + " nulls " + n
				}
				parts = append(parts, "sort:"+val)
			}
			if j == 0 && unique {
				parts = append(parts, "unique")
			}
			if j == 0 && indexType != "" {
				parts = append(parts, "type:"+indexType)
			}
			if j == 0 && where != "" {
				parts = append(parts, "where:"+where)
			}

			fieldToIndexTags[fname] = append(fieldToIndexTags[fname], strings.Join(parts, ","))
		}
	}
	return fieldToIndexTags, nil
}

func fieldNameFromSelectorValue(sel reflect.Value) (string, error) {
	if sel.Kind() != reflect.Func {
		return "", fmt.Errorf("Sel is not a func")
	}
	ft := sel.Type()
	if ft.NumIn() != 1 || ft.In(0).Kind() != reflect.Ptr || ft.NumOut() != 1 {
		return "", fmt.Errorf("Sel must be func(*T) any")
	}

	// Make zero *T and call the selector.
	ptrToT := reflect.New(ft.In(0).Elem()) // *T
	out := sel.Call([]reflect.Value{ptrToT})
	if len(out) != 1 {
		return "", fmt.Errorf("Sel returned unexpected values")
	}

	// IMPORTANT: unwrap interface{} -> underlying pointer
	res := out[0]
	if res.Kind() == reflect.Interface {
		if res.IsNil() {
			return "", fmt.Errorf("Sel returned a nil interface")
		}
		res = res.Elem()
	}

	if res.Kind() != reflect.Ptr || res.IsNil() {
		return "", fmt.Errorf("Sel must return a *field (pointer)")
	}
	retPtr := res.Pointer()

	// Compare against addresses of top-level exported fields on T
	v := ptrToT.Elem()
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		sf := t.Field(i)
		if sf.PkgPath != "" { // unexported
			continue
		}
		fv := v.Field(i)
		if fv.CanAddr() && fv.Addr().Pointer() == retPtr {
			return sf.Name, nil
		}
	}
	return "", fmt.Errorf("Sel didn't point to a top-level exported field on %s", t.Name())
}

var tagKV = regexp.MustCompile(`(\w+):"([^"]*)"`)

func parseStructTag(tag reflect.StructTag) map[string]string {
	out := map[string]string{}
	s := string(tag)
	for _, m := range tagKV.FindAllStringSubmatch(s, -1) {
		out[m[1]] = m[2]
	}
	return out
}

func buildStructTag(kv map[string]string) reflect.StructTag {
	if len(kv) == 0 {
		return ""
	}
	parts := make([]string, 0, len(kv))
	for k, v := range kv {
		parts = append(parts, fmt.Sprintf(`%s:"%s"`, k, v))
	}
	sort.Strings(parts) // deterministic
	return reflect.StructTag(strings.Join(parts, " "))
}

// Remove any existing index/uniqueIndex fragments so we don't duplicate them.
func stripIndexPiecesFromGormTag(gormTag string) string {
	if gormTag == "" {
		return ""
	}
	parts := strings.Split(gormTag, ";")
	keep := parts[:0]
	for _, p := range parts {
		pp := strings.TrimSpace(p)
		if strings.HasPrefix(pp, "index:") || strings.HasPrefix(pp, "uniqueIndex:") {
			continue
		}
		keep = append(keep, pp)
	}
	return strings.Join(keep, ";")
}

func mergeIndexIntoGormTag(orig reflect.StructTag, toAdd []string) reflect.StructTag {
	kv := parseStructTag(orig)
	gormVal := stripIndexPiecesFromGormTag(kv["gorm"])
	if len(toAdd) > 0 {
		add := strings.Join(toAdd, ";")
		if gormVal == "" {
			gormVal = add
		} else {
			gormVal = gormVal + ";" + add
		}
	}
	if gormVal == "" {
		delete(kv, "gorm")
	} else {
		kv["gorm"] = gormVal
	}
	return buildStructTag(kv)
}

func indirectType(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t
}
