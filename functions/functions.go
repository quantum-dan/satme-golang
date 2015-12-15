package functions

// Functions and structs with the exception of routing functions
// For the SATme server

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"errors"
	"golang.org/x/crypto/bcrypt"
)

var dbstr = "localhost:27017"

type User struct { // For logging in.
        Username string `schema:"username" bson:"username"`
        Password string `schema:"password" bson:"password"`
        Role string `schema:"role" bson:"role"`
        Id int `schema:"id" bson:"_id"`  // Not generally used beyond the database but will be useful for deleting users.
}

type DbUser struct { // Uses []byte password
        Username string `bson:"username"`
        Password []byte `bson:"password"`
        Role string `bson:"role"`
        Id int `bson:"_id"`
}

type SuccessLogin struct { // Used to pass information to Create Account
        Success bool
        Username string
        Role string
        Execute bool
}

type Question struct { // Quiz question
        Question string `schema:"question" bson:"question"`
        Answers []string `schema:"answers" bson:"answers"`
        AnswerChosen string `schema:"answer"`
        CorrectIndex int `schema:"correct" bson:"correct"`
        Id string `schema:"id" bson:"_id"`
}

type Quiz struct { // Quiz
        Id string `schema:"id" bson:"_id"`
        Title string `schema:"title" bson:"title"`
        Questions []Question `schema:"questions" bson:"questions"`
}

type DbQuiz struct { // Quiz without ID
	Title string `bson:"title"`
	Questions []Question `bson:"questions"`
}

func NewQuiz(title string) DbQuiz {
	return DbQuiz {title, []Question{}}
}

func NewQuestion(question string, answers []string, correct int) Question {
	return Question{Question: question, Answers: answers, CorrectIndex: correct}
}

func RetrieveQuiz(target string) (Quiz, error) {
        // Retrieves quiz with the given ID
        db, err := mgo.Dial(dbstr)
        defer db.Close()
        if err != nil {
                return *new(Quiz), err
        }
        c := db.DB("server").C("quiz")
        result := new(Quiz)
        err = c.Find(bson.M{"_id": bson.ObjectId(target)}).One(&result)
        if err != nil {
                return *new(Quiz), err
        }
        return *result, nil
}

func UpdateQuiz(quiz Quiz) error {
	db, err := mgo.Dial(dbstr)
	defer db.Close()
	if err != nil {
		return err
	}
	c := db.DB("server").C("quiz")
	err = c.Update(bson.M{"_id": bson.ObjectId(quiz.Id)}, &quiz)
	return err
}

func AddQuestion (id string, question Question) error {
	quiz, err := RetrieveQuiz(id)
	if err != nil {
		return err
	}
	quiz.Questions = append(quiz.Questions, question)
	return UpdateQuiz(quiz)
}

func InsertQuiz(quiz DbQuiz) error {
        db, err := mgo.Dial(dbstr)
        defer db.Close()
        if err != nil {
                return err
        }
        c := db.DB("server").C("quiz")
        err = c.Insert(&quiz)
        return err
}

func RetrieveQuizzes(title string) ([]Quiz, error) {
        // Retrieves all quizzes from the database
        db, err := mgo.Dial(dbstr)
        defer db.Close()
        if err != nil {
                return []Quiz{}, err
        }
        c := db.DB("server").C("quiz")
        var dbresult *mgo.Iter
        if title != "" {
                dbresult = c.Find(bson.M{"title": title}).Limit(10).Iter()
        } else {
                dbresult = c.Find(nil).Limit(10).Iter()
        }
        var result []Quiz
        err = dbresult.All(&result)
        if err != nil {
                return []Quiz{}, err
        }
        return result, nil
}

func (quiz Quiz) Grade() (float32, error) {
        // Grades a quiz
        db, err := mgo.Dial(dbstr)
        defer db.Close()
        if err != nil {
                return 0.0, err
        }
        c := db.DB("server").C("quiz")
        compare := new(Quiz)
        err = c.Find(bson.M{"_id": bson.ObjectId(quiz.Id)}).One(&compare)
        if err != nil {
                return 0.0, err
        }
        var sum float32 = 0.0
        var total float32 = float32(len(quiz.Questions))
        for i := 0; i < len(quiz.Questions); i++ {
                if quiz.Questions[i].AnswerChosen == compare.Questions[i].Answers[compare.Questions[i].CorrectIndex] {
                        sum += 1.0
                }
        }
        return sum * 100 / total, nil
}

func CreateAccount(user User) error {
        db, err := mgo.Dial(dbstr)
        defer db.Close()
        if err != nil {
                return err
        }
        c := db.DB("server").C("users")
        result := new(User)
        err = c.Find(bson.M{"username": user.Username}).One(&result)
        if err != nil && err.Error() != "not found" {
                return err
        } else if err == nil {
                return errors.New("user already exists")
        } else {
                password_bytestr, err := bcrypt.GenerateFromPassword([]byte(user.Password), 15) // Note: excessive costs may cause unreasonable delays on the client-side while the password is hashed server-side.  15 is reasonable (3-5 seconds), 20 is excessive.
                user.Password = string(password_bytestr)
                err = c.Insert(&user)
                return err
        }
}

func CheckLogin(user User) (User, error) {
        db, err := mgo.Dial(dbstr)
        defer db.Close()
        if err != nil {
                return User{}, err
        }
        c := db.DB("server").C("users")
        dbresult := new(DbUser)
        err = c.Find(bson.M{"username": user.Username}).One(dbresult)
        if err != nil && err.Error() != "not found" {
                return User{}, err
        } else if err != nil {
                return User{}, errors.New("login failed")
        }
        if dbresult.Username == user.Username && bcrypt.CompareHashAndPassword(dbresult.Password, []byte(user.Password)) == nil {
                user.Role = dbresult.Role
                return user, nil
        } else {
                return User{}, errors.New("login failed")
        }
}
