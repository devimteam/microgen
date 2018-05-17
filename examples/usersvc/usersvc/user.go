package usersvc

type User struct {
	Id       string
	Name     string
	Gender   int
	Comments []Comment
}

type Comment struct {
	Id   string
	Text string
}
