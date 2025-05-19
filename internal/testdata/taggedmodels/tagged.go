//go:build tag
// +build tag

package taggedmodels

type TaggedModel struct {
	ID int `gorm:"primaryKey"`
}
