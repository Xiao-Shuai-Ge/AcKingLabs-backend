package model

type Template struct {
	CommonModel
	Body string `gorm:"column:body;type:varchar(255);not null;comment:'内容'"`
}

func (t *Template) TableName() string {
	return "template"
}
