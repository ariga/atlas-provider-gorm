//go:build buildtag
// +build buildtag

package buildtags

type TaggedModel struct {
	ID int `gorm:"primaryKey"`
}
