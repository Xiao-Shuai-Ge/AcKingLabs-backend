package model

type User struct {
	ID int64 `json:"id" gorm:"column:id;primaryKey;type:bigint;type:bigint"`
	TimeModel
	Username string `json:"username" gorm:"column:username;type:varchar(255);comment:用户名"`
	Password string `json:"password" gorm:"column:password;type:varchar(255);comment:密码"`
	Email    string `json:"email" gorm:"column:email;type:varchar(255);comment:邮箱"`
	Avatar   string `json:"avatar" gorm:"column:avatar;type:varchar(255);comment:头像URL"`

	Xp int `json:"xp" gorm:"column:xp;type:int;comment:经验值"`

	Grade     int    `json:"grade" gorm:"column:grade;type:int;comment:年级"`
	StudentNo string `json:"student_no" gorm:"column:student_no;type:varchar(255);comment:学号"`
	RealName  string `json:"real_name" gorm:"column:real_name;type:varchar(255);comment:真实姓名"`

	CodeforcesID     string `json:"codeforces_id" gorm:"column:codeforces_id;type:varchar(255);comment:codeforces ID"`
	CodeforcesRating int    `json:"codeforces_rating" gorm:"column:codeforces_rating;type:int;comment:codeforces 分数"`
}
