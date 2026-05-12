package models

import "time"

type Role struct {
	ID    uint   `gorm:"primaryKey" json:"id"`
	Name  string `gorm:"size:50;uniqueIndex;not null" json:"name"`
	Users []User `json:"-"`
}

type User struct {
	ID             uint   `gorm:"primaryKey" json:"id"`
	Username       string `gorm:"size:50;uniqueIndex;not null" json:"username"`
	Email          string `gorm:"size:120;uniqueIndex;not null" json:"email"`
	Password       string `gorm:"size:255;not null" json:"-"`
	RoleID         uint   `gorm:"not null" json:"role_id"`
	ProfilePicture string `gorm:"size:255" json:"profile_picture"`

	Role            Role         `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"role"`
	Profile         *Profile     `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"profile"`
	Courses         []Course     `gorm:"foreignKey:TeacherID" json:"courses"`
	EnrolledCourses []Course     `gorm:"many2many:enrollments;" json:"enrolled_courses"`
	Attendances     []Attendance `json:"attendances"`
}

type Profile struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	UserID    uint   `gorm:"uniqueIndex;not null" json:"user_id"`
	FullName  string `gorm:"size:120;not null" json:"full_name"`
	AvatarURL string `gorm:"size:255" json:"avatar_url"`

	User *User `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
}

type Course struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Title       string `gorm:"size:150;not null" json:"title"`
	Description string `gorm:"type:text" json:"description"`
	TeacherID   uint   `gorm:"not null" json:"teacher_id"`

	Teacher     User         `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"teacher"`
	Lessons     []Lesson     `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"lessons"`
	Assignments []Assignment `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"assignments"`
	Students    []User       `gorm:"many2many:enrollments;" json:"students"`
}

type Enrollment struct {
	UserID   uint `gorm:"primaryKey" json:"user_id"`
	CourseID uint `gorm:"primaryKey" json:"course_id"`

	User   User   `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	Course Course `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
}

type Lesson struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	CourseID uint   `gorm:"not null" json:"course_id"`
	Title    string `gorm:"size:150;not null" json:"title"`
	Content  string `gorm:"type:text" json:"content"`

	Course Course `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
}

type Assignment struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	CourseID    uint      `gorm:"not null" json:"course_id"`
	Title       string    `gorm:"size:150;not null" json:"title"`
	Description string    `gorm:"type:text" json:"description"`
	DueDate     time.Time `gorm:"type:date;not null" json:"due_date"`

	Course      Course       `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	Submissions []Submission `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"submissions"`
}

type Submission struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	AssignmentID uint      `gorm:"not null" json:"assignment_id"`
	UserID       uint      `gorm:"not null" json:"user_id"`
	FileURL      string    `gorm:"size:255;not null" json:"file_url"`
	SubmittedAt  time.Time `gorm:"autoCreateTime" json:"submitted_at"`

	Assignment Assignment `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	User       User       `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
}

type Attendance struct {
	ID       uint      `gorm:"primaryKey" json:"id"`
	UserID   uint      `gorm:"not null" json:"user_id"`
	CourseID uint      `gorm:"not null" json:"course_id"`
	Status   string    `gorm:"size:10;not null" json:"status"`
	Date     time.Time `gorm:"type:date;not null" json:"date"`

	User   User   `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	Course Course `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
}
