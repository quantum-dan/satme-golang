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
        Username string `schema:"username"`
        Password string `schema:"password"`
        Role string `schema:"role"`
        Id int `schema:"id"`  // Not generally used beyond the database but will be useful for deleting users.
}

type DbUser struct { // Uses []byte password
        Username string
        Password []byte
        Role string
        Id int
}

type SuccessLogin struct { // Used to pass information to Create Account
        Success bool
        Username string
        Role string
        Execute bool
}

type Question struct { // Quiz question
        Question string `schema:"question"`
        Answers []string `schema:"answers"`
        AnswerChosen string `schema:"answer"`
        CorrectIndex int `schema:"correct"`
        Id string `schema:"id"`
}

type Quiz struct { // Quiz
        Title string `schema:"title"`
        Id string `schema:"id"`
        Questions []Question `schema:"-"`
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

func InsertQuiz(quiz Quiz) error {
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
        // Retrieves all quizes from the database
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
