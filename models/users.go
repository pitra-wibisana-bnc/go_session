package models

type Users struct {
	ID        int64  `xorm:"int not null unique 'id' autoincr pk" json:"id"`
	Username  string `xorm:"varchar(50) not null 'username'" json:"username"`
	FirstName string `xorm:"varchar(200) not null 'first_name'" json:"first_name"`
	LastName  string `xorm:"varchar(200) not null 'last_name'" json:"last_name"`
	Password  string `xorm:"varchar(120) not null 'password'" json:"-"`
}
