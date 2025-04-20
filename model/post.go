package model

type Post struct {
	ID int64 `json:"id" gorm:"column:id;primaryKey;type:bigint;type:bigint"`
	TimeModel

	UserID  int64  `json:"user_id" gorm:"column:user_id;type:bigint;comment:用户ID"`
	Title   string `json:"title" gorm:"column:title;type:varchar(255);comment:标题"`
	Content string `json:"content" gorm:"column:content;type:text;comment:内容"`

	Type   string `json:"type" gorm:"column:type;type:varchar(63);comment:类型"`
	Source string `json:"source" gorm:"column:source;type:varchar(255);comment:来源"`

	Likes    int `json:"likes" gorm:"not null;column:likes;type:int;comment:点赞数"`
	Comments int `json:"comments" gorm:"not null;column:comments;type:int;comment:评论数"`

	IsAdminLike bool `json:"is_admin_like" gorm:"not null;column:is_admin_like;type:bool;comment:是否有管理员点赞"`
	IsFeatured  bool `json:"is_featured" gorm:"not null;column:is_featured;type:bool;comment:是否精选"`
	IsPrivate   bool `json:"is_private" gorm:"not null;column:is_private;type:bool;comment:是否私密"`

	// 时间戳+点赞数(半小时)+评论数(半小时)+管理员点赞(7天)
	Weight int64 `json:"weight" gorm:"not null;column:weight;type:bigint;comment:权重;index"`
}

func (Post) TableName() string {
	return "posts"
}

type PostLike struct {
	ID int64 `json:"id" gorm:"column:id;primaryKey;type:bigint;type:bigint"`
	TimeModel

	PostID int64 `json:"post_id" gorm:"column:post_id;type:bigint;comment:帖子ID;uniqueIndex:idx_post_user_unique"`
	UserID int64 `json:"user_id" gorm:"column:user_id;type:bigint;comment:用户ID;uniqueIndex:idx_post_user_unique"`
}

func (PostLike) TableName() string {
	return "post_like"
}

type Comment struct {
	ID int64 `json:"id" gorm:"column:id;primaryKey;type:bigint;type:bigint"`
	TimeModel

	Content     string `json:"content" gorm:"column:content;type:text;comment:内容"`
	PostID      int64  `json:"post_id" gorm:"column:post_id;type:bigint;comment:帖子ID;index"`
	FatherID    int64  `json:"father_id" gorm:"column:father_id;type:bigint;comment:父评论ID"`
	UserID      int64  `json:"user_id" gorm:"column:user_id;type:bigint;comment:用户ID"`
	Likes       int    `json:"likes" gorm:"not null;column:likes;type:int;comment:点赞数"`
	IsAdminLike bool   `json:"is_admin_like" gorm:"not null;column:is_admin_like;type:bool;comment:是否有管理员点赞"`
}

func (Comment) TableName() string {
	return "comments"
}

type CommentLike struct {
	ID int64 `json:"id" gorm:"column:id;primaryKey;type:bigint;type:bigint"`
	TimeModel

	CommentID int64 `json:"comment_id" gorm:"column:comment_id;type:bigint;comment:评论ID;uniqueIndex:idx_comment_user_unique"`
	UserID    int64 `json:"user_id" gorm:"column:user_id;type:bigint;comment:用户ID;uniqueIndex:idx_comment_user_unique"`
}

func (CommentLike) TableName() string {
	return "comment_like"
}
