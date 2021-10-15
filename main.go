package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/mgo.v2/bson"
	"log"
	"math/rand"
	"time"

	"net/http"

	"html/template"
)
var client *mongo.Client
var store = sessions.NewCookieStore([]byte("super-secret-password"))
func main() {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, _ = mongo.Connect(ctx, clientOptions)
	fmt.Println("Connected to MongoDB!")
	s := mux.NewRouter()
	s.HandleFunc("/register",register)
	s.HandleFunc("/user", createProfile).Methods("POST")//register page
	s.HandleFunc("/portal", portal)
	s.HandleFunc("/login", checkProfile).Methods("POST")//login page
	s.HandleFunc("/home",studenthomepage) //student dashboard home page
	s.HandleFunc("/grades1",studentgrade) //student dashboard grade page
	s.HandleFunc("/student-attendance",student_attendance) //student dashboard attendance page
	s.HandleFunc("/grade_students", gradestudents).Methods("GET") //student grade page
	s.HandleFunc("/att_students", getstudents).Methods("GET") //student attendance page
	s.HandleFunc("/users", getallusers).Methods("GET") //to get all records from register table
	s.HandleFunc("/student", getrecord).Methods("GET")
	s.HandleFunc("/staff-home",staffhomepage) //staff dashboard home page
	s.HandleFunc("/attendance",attendance) //staff dashboard attendance page
	s.HandleFunc("/att", attendanceProfile).Methods("POST") //to enter student attendance details by staff
	s.HandleFunc("/att_users", getusers).Methods("GET") //to display attendance table in staff dashboard
	s.HandleFunc("/grades",grades) //staff dashboard grade page
	s.HandleFunc("/grade", gradeProfile).Methods("POST") //to enter student grade details by staff
	s.HandleFunc("/grade_users", gradeusers).Methods("GET") //to display grades table in staff dashboard
	s.HandleFunc("/logout",logout) //logout page
	s.PathPrefix("/").Handler(http.FileServer(http.Dir("./UI")))
	log.Fatal(http.ListenAndServe(":8014", s))
}
func register(w http.ResponseWriter, r *http.Request) {
	tmpl1 := template.Must(template.ParseFiles("./UI/html/register.html"))
	tmpl1.Execute(w, nil)
	return
}
func Getrandomstring(n int) string{
	b := make([]byte,n)
	if _,err := rand.Read(b);err!=nil{
		panic(err)
	}
	s := fmt.Sprintf("%X",b)
	return s
}
type User struct {
	ID            string   `json:"ID,omitempty" bson:"ID,omitempty"`
	Firstname      string `json:"firstname,omitempty" bson:"firstname,omitempty"`
	Lastname       string `json:"lastname,omitempty" bson:"lastname,omitempty"`
	Email        string    `json:"email,omitempty" bson:"email,omitempty"`
	Password	  string	`json:"password,omitempty" bson:"password,omitempty"`
	Usertype     string    `json:"usertype,omitempty" bson:"usertype,omitempty"`
}
func createProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	id := Getrandomstring(5)
	var req User
	req=User{ID: id}
	fmt.Println("Connected")
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		fmt.Print(err)
	}
	collection := client.Database("student").Collection("register")
	insertResult, err := collection.InsertOne(context.TODO(), req)
	if err != nil {
		log.Fatal(err)
	}
	json.NewEncoder(w).Encode(insertResult.InsertedID)
	fmt.Println("added Document to DB")
}
func portal(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("./UI/html/login.html"))
	tmpl.Execute(w, nil)
	return
}
type Userreq struct {
	Email string
	Password string
}
type stud struct {
	ID            string   `json:"ID,omitempty" bson:"ID,omitempty"`
	Firstname      string `json:"firstname,omitempty" bson:"firstname,omitempty"`
}
type student struct{
	ID string
	Firstname string
}
type staff struct{
	ID string
	Firstname string
}
type staf struct {
	ID string `json:"ID,omitempty"bson:"ID,omitempty"`
}
func checkProfile(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "ramya")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	session.Options = &sessions.Options{
		Path:   "/",
		MaxAge: 86400 * 7,
		// MaxAge:   5,
		HttpOnly: true,
	}
	var req Userreq
	r.ParseForm()
	req.Email = r.FormValue("email")
	req.Password = r.FormValue("password")
	if req.Email != ""{
		session.Values["email"] = req.Email
	}
	if req.Password != ""{
		session.Values["password"] = req.Password
	}
	fmt.Println("session:", session)
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	collection := client.Database("student").Collection("register")
	count, err := collection.CountDocuments(context.TODO(),bson.M{"email" :req.Email,"password" : req.Password,"usertype":"student"})
	count1, err := collection.CountDocuments(context.TODO(),bson.M{"email" :req.Email,"password" : req.Password,"usertype":"staff"})
	count2, err := collection.CountDocuments(context.TODO(),bson.M{"email" :req.Email,"password" : req.Password,"usertype":"admin"})
	if err != nil{
		fmt.Println(err)
	}
	if count>=1 {
		var person stud
		collections := client.Database("student").Collection("register")
		a := collections.FindOne(context.TODO(),bson.M{"usertype":"student","email":session.Values["email"]}).Decode(&person)
		if a != nil {
			fmt.Println(a)
		}
		fmt.Println(person)
		var req student
		req.ID=person.ID
		req.Firstname=person.Firstname
		if req.ID != ""{
			session.Values["student_id"] = req.ID
		}
		if req.Firstname != ""{
			session.Values["stu_name"] = req.Firstname
		}
		fmt.Println("session:", session)
		err = session.Save(r, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Println("Document exists")
		tmpl := template.Must(template.ParseFiles("./UI/html/student/home.html"))
		tmpl.Execute(w, session.Values["email"])
		return
	} else if count1 >= 1 {
		var person staf
		collections := client.Database("student").Collection("register")
		a := collections.FindOne(context.TODO(),bson.M{"usertype":"staff","email":session.Values["email"]}).Decode(&person)
		if a != nil {
			fmt.Println(a)
		}
		fmt.Println(person)
		var req staff
		req.ID=person.ID
		if req.ID != ""{
			session.Values["staff_id"] = req.ID
		}
		fmt.Println("session:", session)
		err = session.Save(r, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Println("Document exists")
		tmpl := template.Must(template.ParseFiles("./UI/html/staff/dashboard1.html"))
		tmpl.Execute(w, session.Values["email"])
		return
	} else if count2>=1 {
		fmt.Println("Document exists")
		tmpl := template.Must(template.ParseFiles("./UI/html/admin/dashboard2.html"))
		tmpl.Execute(w, session.Values["email"])
		return
	} else {
		fmt.Println("Document Not Exist")
		portal(w,r)
	}
}


func studenthomepage(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "ramya")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	session.Options = &sessions.Options{
		Path:   "/",
		MaxAge: 86400 * 7,
		// MaxAge:   5,
		HttpOnly: true,
	}
	tmpl := template.Must(template.ParseFiles("./UI/html/student/home.html"))
	tmpl.Execute(w, session.Values["email"])
	return
}
func staffhomepage(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "ramya")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	session.Options = &sessions.Options{
		Path:   "/",
		MaxAge: 86400 * 7,
		// MaxAge:   5,
		HttpOnly: true,
	}
	tmpl := template.Must(template.ParseFiles("./UI/html/staff/dashboard1.html"))
	tmpl.Execute(w, session.Values["email"])
	return
}
func studentgrade(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "ramya")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	session.Options = &sessions.Options{
		Path:   "/",
		MaxAge: 86400 * 7,
		// MaxAge:   5,
		HttpOnly: true,
	}
	tmpl := template.Must(template.ParseFiles("./UI/html/student/grades.html"))
	tmpl.Execute(w, session.Values["email"])
	return
}

func student_attendance(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "ramya")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	session.Options = &sessions.Options{
		Path:   "/",
		MaxAge: 86400 * 7,
		// MaxAge:   5,
		HttpOnly: true,
	}
	tmpl := template.Must(template.ParseFiles("./UI/html/student/attendance1.html"))
	tmpl.Execute(w, session.Values["email"])
	return
}

func gradestudents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	session, err := store.Get(r, "ramya")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	session.Options = &sessions.Options{
		Path:   "/",
		MaxAge: 86400 * 7,
		// MaxAge:   5,
		HttpOnly: true,
	}
	var result []primitive.M
	usercollection := client.Database("student").Collection("grades")
	count, s := usercollection.Find(context.TODO(),bson.M{"stu_id":session.Values["student_id"],"student_name":session.Values["stu_name"]})
	if s != nil {
		fmt.Println(err)
	}
	for count.Next(context.TODO()) {
		var elem primitive.M
		s1 := count.Decode(&elem)
		if s1 != nil {
			log.Fatal(s1)
		}
		result = append(result, elem)
	}
	count.Close(context.TODO())
	fmt.Println(result)
	json.NewEncoder(w).Encode(result)
}

func getstudents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	session, err := store.Get(r, "ramya")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	session.Options = &sessions.Options{
		Path:   "/",
		MaxAge: 86400 * 7,
		// MaxAge:   5,
		HttpOnly: true,
	}
	var result []primitive.M
	usercollection := client.Database("student").Collection("attendance")
	count, s := usercollection.Find(context.TODO(),bson.M{"stu_id":session.Values["student_id"],"student_name":session.Values["stu_name"]})
	if s != nil {
		fmt.Println(err)
	}
	for count.Next(context.TODO()) {
		var elem primitive.M
		s1 := count.Decode(&elem)
		if s1 != nil {
			log.Fatal(s1)
		}
		result = append(result, elem)
	}
	count.Close(context.TODO())
	fmt.Println(result)
	json.NewEncoder(w).Encode(result)
}
func attendance(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "ramya")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	session.Options = &sessions.Options{
		Path:   "/",
		MaxAge: 86400 * 7,
		// MaxAge:   5,
		HttpOnly: true,
	}
	tmpl := template.Must(template.ParseFiles("./UI/html/staff/attendance.html"))
	tmpl.Execute(w, session.Values["email"])
	return
}
func getallusers(response http.ResponseWriter, r *http.Request) {
	response.Header().Set("Content-Type", "application/json")
	var results []primitive.M
	collection := client.Database("student").Collection("register")
	cur, err := collection.Find(context.TODO(),bson.M{"usertype":"student"})
	if err != nil {
		fmt.Println(err)
	}
	for cur.Next(context.TODO()) {
		var elem primitive.M
		err := cur.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}
		results = append(results, elem)
	}
	cur.Close(context.TODO())
	json.NewEncoder(response).Encode(results)
}
func getrecord(response http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "ramya")
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}
	session.Options = &sessions.Options{
		Path:   "/",
		MaxAge: 86400 * 7,
		// MaxAge:   5,
		HttpOnly: true,
	}
	response.Header().Set("Content-Type", "application/json")
	var results []primitive.M
	collection := client.Database("student").Collection("register")
	cur, a := collection.Find(context.TODO(),bson.M{"usertype":"student","email":session.Values["email"],"password":session.Values["password"]})
	if a != nil {
		fmt.Println(a)
	}
	for cur.Next(context.TODO()) {
		var elem primitive.M
		err := cur.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}
		results = append(results, elem)
	}
	cur.Close(context.TODO())
	fmt.Println(results)
	json.NewEncoder(response).Encode(results)
}
type Attend struct {
	StuId string  `json:"stu_id,omitempty" bson:"stu_id,omitempty"`
	StudentName string  `json:"student_name,omitempty" bson:"student_name,omitempty"`
	Date string	`json:"date,omitempty" bson:"date,omitempty"`
	Subject string  `json:"subject,omitempty" bson:"subject,omitempty"`
	Attended string  `json:"attended,omitempty" bson:"attended,omitempty"`
	StaffID string `json:"staff_Id,omitempty" bson:"staff_Id,omitempty"`
}
func attendanceProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	session, err := store.Get(r, "ramya")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	session.Options = &sessions.Options{
		Path:   "/",
		MaxAge: 86400 * 7,
		// MaxAge:   5,
		HttpOnly: true,
	}
	var person Attend
	var staff_iD = fmt.Sprintf("%v",session.Values["staff_id"])
	person=Attend{StaffID: staff_iD}
	fmt.Println(person)
	fmt.Println("Connected")
	a := json.NewDecoder(r.Body).Decode(&person)
	if a != nil {
		fmt.Print(a)
	}
	collection := client.Database("student").Collection("attendance")
	insertResult, err := collection.InsertOne(context.TODO(), person)
	if err != nil {
		log.Fatal(err)
	}
	json.NewEncoder(w).Encode(insertResult.InsertedID)
	fmt.Println("added Document to DB")
}
func getusers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	session, err := store.Get(r, "ramya")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	session.Options = &sessions.Options{
		Path:   "/",
		MaxAge: 86400 * 7,
		// MaxAge:   5,
		HttpOnly: true,
	}
	var result []primitive.M
	usercollection := client.Database("student").Collection("attendance")
	cur, err := usercollection.Find(context.TODO(),bson.M{"staff_Id":session.Values["staff_id"]})
	if err != nil {
		fmt.Println(err)
	}
	for cur.Next(context.TODO()) {
		var elem primitive.M
		err := cur.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}
		result = append(result, elem)
	}
	cur.Close(context.TODO())
	json.NewEncoder(w).Encode(result)
}
func grades(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "ramya")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	session.Options = &sessions.Options{
		Path:   "/",
		MaxAge: 86400 * 7,
		// MaxAge:   5,
		HttpOnly: true,
	}
	tmpl := template.Must(template.ParseFiles("./UI/html/staff/grades1.html"))
	tmpl.Execute(w, session.Values["email"])
	return
}
type Usergrade struct {
	ID string  `json:"stu_id,omitempty" bson:"stu_id,omitempty"`
	Firstname string `json:"student_name,omitempty" bson:"student_name,omitempty"`
	Date string `json:"date,omitempty" bson:"date,omitempty"`
	Subject string  `json:"subject,omitempty" bson:"subject,omitempty"`
	Grade string 	`json:"grade,omitempty" bson:"grade,omitempty"`
	StaffId string `json:"staff_id,omitempty" bson:"staff_id,omitempty"`
}
func gradeProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	session, err := store.Get(r, "ramya")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	session.Options = &sessions.Options{
		Path:   "/",
		MaxAge: 86400 * 7,
		// MaxAge:   5,
		HttpOnly: true,
	}
	var person Usergrade
	var staff_iD = fmt.Sprintf("%v",session.Values["staff_id"])
	person=Usergrade{StaffId: staff_iD}
	fmt.Println(person)
	fmt.Println("Connected")
	a := json.NewDecoder(r.Body).Decode(&person)
	if a != nil {
		fmt.Print(a)
	}
	collection := client.Database("student").Collection("grades")
	insertResult, err := collection.InsertOne(context.TODO(), person)
	if err != nil {
		log.Fatal(err)
	}
	json.NewEncoder(w).Encode(insertResult.InsertedID)
	fmt.Println("added Document to DB")
}
func gradeusers(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "ramya")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	session.Options = &sessions.Options{
		Path:   "/",
		MaxAge: 86400 * 7,
		// MaxAge:   5,
		HttpOnly: true,
	}
	w.Header().Set("Content-Type", "application/json")
	var result []primitive.M
	usercollection := client.Database("student").Collection("grades")
	cur, err := usercollection.Find(context.TODO(),bson.M{"staff_id":session.Values["staff_id"]})
	if err != nil {
		fmt.Println(err)
	}
	for cur.Next(context.TODO()) {
		var elem primitive.M
		err := cur.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}
		result = append(result, elem)
	}
	cur.Close(context.TODO())
	json.NewEncoder(w).Encode(result)
}
func logout(w http.ResponseWriter,r *http.Request){
	session, _ := store.Get(r, "session-name")
	session.Options.MaxAge = -1
	fmt.Println("session:", session)
	session.Save(r, w)
	tmpl := template.Must(template.ParseFiles("./UI/html/login.html"))
	tmpl.Execute(w, nil)
	return
}