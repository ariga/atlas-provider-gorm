package models

import "ariga.io/atlas-provider-gorm/gormschema"

type TestModelIndexType struct {
	ID      string `gorm:"column:id"`
	Name    string `gorm:"column:name"`
	Profile string `gorm:"column:profile"`
}

func (model TestModelIndexType) TableName() string {
	return "test_model_index_type"
}

func (model *TestModelIndexType) Indexes() []gormschema.IndexDefinition[TestModelIndexType] {
	return []gormschema.IndexDefinition[TestModelIndexType]{
		{
			Name: "idx_test_model_index_type_name_gin",
			Type: "gin",
			Columns: []gormschema.Col[TestModelIndexType]{
				{Sel: func(m *TestModelIndexType) any { return &m.Name }},
			},
		},
		{
			Name: "idx_test_model_index_type_name_profile_gin",
			Type: "gin",
			Columns: []gormschema.Col[TestModelIndexType]{
				{Sel: func(m *TestModelIndexType) any { return &m.Name }},
				{Sel: func(m *TestModelIndexType) any { return &m.Profile }},
			},
		},
	}
}
