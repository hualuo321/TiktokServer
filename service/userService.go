package service

type UserService interface {
	/*
		个人使用
	*/
	// GetTableUserList GetUserList 获得全部TableUser对象
	GetTableUserList(users *[]TableUser) bool

	// GetTableUserByUsername GetUserByUsername 根据username获得TableUser对象
	GetTableUserByUsername(name string) bool

	// GetTableUserById GetUserById 根据user_id获得TableUser对象
	GetTableUserById(id int64) bool

	/*
		他人使用
	*/
	// GetUserById 未登录情况下,根据user_id获得User对象
	//(调用方法:user.GetUserById,user会被填充;若填充失败,则返回false,成功,返回true)
	GetUserById(id int64) (User, error)

	// GetUserByIdWithCurId 已登录(curID)情况下,根据user_id获得User对象
	GetUserByIdWithCurId(id int64, curId int64) (User, error)

	// 根据token返回id
	// 接口:auth中间件,解析完token,将userid放入context
	//(调用方法:直接在context内拿参数"user_id"的值)
}

type TableUser struct {
	Id       int64
	Name     string
	Password string
}

/*type User struct {
	Id            int64  `json:"id,omitempty"`
	Name          string `json:"name,omitempty"`
	FollowCount   int64  `json:"follow_count,omitempty"`
	FollowerCount int64  `json:"follower_count,omitempty"`
	IsFollow      bool   `json:"is_follow,omitempty"`
}
*/
