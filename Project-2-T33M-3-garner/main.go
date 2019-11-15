package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"os/exec"

	dbconnection "github.com/NGKlaure/Project-2-T33M-3-garner/dbConnection"
	//dbconnection "dbConnection"
)

//localstruct
type localstruct struct {
	FILES []string
}

//Users a
type Users struct {
	Username      string
	Password      string
	Nametooshort  bool
	Namenotunique bool
	Pwtooshort    bool
	//
}

//LoginInfo a
type LoginInfo struct {
	CurrentUser string
	Loggedin    bool
	Invalidname bool
	Invalidpw   bool
	//
}

//ViewInfo a
type ViewInfo struct {
	Usr        []Users
	Singleuser Users
	Login      LoginInfo
}

var remoteUsername, password string

const remoteHostname = "3.86.179.34"
const localuser = "ubuntu"
const tcpLine = "3.86.179.34:8081"                              //This is the tcp connection to the user-server.go program on the server
const amazon = "ubuntu@ec2-3-86-179-34.compute-1.amazonaws.com" //This is the amazon aws hostname@IP address.
const key = "t33mkey.pem"                                       //This is the amazon key to ssh into the server.
//Signin variable
var Signin = LoginInfo{}

func main() {

	db := dbconnection.DbConnection()
	ping(db)
	getAll(db)

	http.HandleFunc("/", index)
	http.HandleFunc("/remotefiles.html", remotefiles)
	http.HandleFunc("/localfiles.html", localfiles)
	http.HandleFunc("/downloader", downloader)
	http.HandleFunc("/uploader", uploader)
	//http.HandleFunc("/createfilelocal", createfilelocal)

	http.HandleFunc("/registrationform", registrationForm)
	http.HandleFunc("/register", register)
	http.HandleFunc("/login", login)
	http.HandleFunc("/logout", logout)
	Signin.Loggedin = false

	fmt.Println(" Open browser to localhost:7005")
	http.ListenAndServe(":7005", nil)
}

//index runs the index page
func index(response http.ResponseWriter, request *http.Request) {
	temp, _ := template.ParseFiles("html/index.html")
	response.Header().Set("Content-Type", "text/html; charset=utf-8")
	temp.Execute(response, Signin)
}

func registrationForm(response http.ResponseWriter, request *http.Request) {
	temp, _ := template.ParseFiles("html/registrationform.html")
	temp.Execute(response, nil)
}

func register(response http.ResponseWriter, request *http.Request) {
	db := dbconnection.DbConnection()
	defer db.Close()

	temp, _ := template.ParseFiles("html/register.html")
	user := Users{}
	user.Username = request.FormValue("uname")
	user.Password = request.FormValue("pwd")

	// check if username is too short or is already taken or if password is too short
	if len(user.Username) < 3 {
		user.Nametooshort = true

	} else if uniqueName(user.Username) == false {
		user.Namenotunique = true
	} else if len(user.Password) < 3 {
		user.Pwtooshort = true
	} else { // insert username and password into database if acceptable
		statement := "INSERT INTO users (username, password)"
		statement += " VALUES ($1, $2);"
		_, err := db.Exec(statement, user.Username, user.Password)
		if err != nil {
			panic(err)
		}
		remoteUsername = user.Username
		remote3 := exec.Command("ssh", "-i", key, amazon, "mkdir", "-p", "home/"+remoteUsername+"/servercatchbox")
		remote3.Run()
	}
	temp.Execute(response, user)
}

//handler for logging in
func login(response http.ResponseWriter, request *http.Request) {
	db := dbconnection.DbConnection()
	defer db.Close()
	temp, _ := template.ParseFiles("html/login.html")

	user := Users{}
	view := ViewInfo{}
	login := LoginInfo{}

	if !Signin.Loggedin { // if not logged in then check if username and password are in database
		user.Username = request.FormValue("uname")
		user.Password = request.FormValue("pwd")
		if uniqueName(user.Username) == true {
			login.Invalidname = true
		} else if passwordMatches(user.Username, user.Password) == false {
			login.Invalidpw = true
		} else {
			Signin.CurrentUser = user.Username
			Signin.Loggedin = true
			remoteUsername = user.Username
		}
	} else {
		user.Username = Signin.CurrentUser
	}

	view.Singleuser = user
	view.Login = login

	//connect to this socket
	conn, _ := net.Dial("tcp", tcpLine)

	// send to socket
	fmt.Fprintf(conn, user.Username+"\n")
	temp.Execute(response, view)
}

//logout handles logout listener
func logout(response http.ResponseWriter, request *http.Request) {
	temp, _ := template.ParseFiles("html/index.html")
	Signin.Loggedin = false
	temp.Execute(response, Signin)
}

// uniqueName checks if the username does not already exist in database
func uniqueName(name string) bool {
	db := dbconnection.DbConnection()
	defer db.Close()

	rows, _ := db.Query("SELECT username FROM users")
	for rows.Next() {
		var username string
		rows.Scan(&username)
		if name == username {
			return false
		}
	}
	return true // name is not already in the db
}

//passwordMatches check if passwords matches password stored with username
func passwordMatches(name string, password string) bool {
	db := dbconnection.DbConnection()
	defer db.Close()
	var pw string
	row := db.QueryRow("SELECT password FROM users WHERE username = $1", name)
	row.Scan(&pw)
	if password == pw {
		return true
	}
	return false
}

//ping tests if connection with database
func ping(db *sql.DB) {
	err := db.Ping()

	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully connected!")
}

func getAll(db *sql.DB) {
	rows, _ := db.Query("SELECT * FROM users")
	for rows.Next() {
		var remoteUsername string
		var password string
		rows.Scan(&remoteUsername, &password)
		fmt.Println(remoteUsername, password)
	}
}

//remotefiles displays remote computer files to html page.
func remotefiles(response http.ResponseWriter, request *http.Request) {
	temp, _ := template.ParseFiles("html/remotefiles.html")
	response.Header().Set("Content-Type", "text/html; charset=utf-8")

	remote3 := exec.Command("ssh", "-i", key, amazon, "mkdir", "-p", "home/"+remoteUsername+"/servercatchbox")
	remote3.Run()

	remote, err := exec.Command("ssh", "-i", key, amazon, "ls", "home/"+remoteUsername+"/servercatchbox", ">", "file1", ";", "cat", "file1").Output()
	g := localstruct{FILES: make([]string, 1)}
	length := 0
	if err != nil {
		fmt.Println(err)
	}
	for l := 0; l < len(remote); l = l + 1 {
		if remote[l] != 10 {
			g.FILES[length] = g.FILES[length] + string(remote[l])
		} else {
			g.FILES = append(g.FILES, "\n")
			length = length + 1
		}
	}
	temp.Execute(response, g)
}

//localfiles displays host computer files to html page.
func localfiles(response http.ResponseWriter, request *http.Request) {
	temp, _ := template.ParseFiles("html/localfiles.html")
	response.Header().Set("Content-Type", "text/html; charset=utf-8")

	local4 := exec.Command("mkdir", "-p", "/home/"+localuser+"/servercatchbox")
	local4.Run()

	remote, err := exec.Command("ls", "/home/"+localuser+"/servercatchbox", ">", "file1", ";", "cat", "file1").Output()
	g := localstruct{FILES: make([]string, 1)}
	length := 0

	if err != nil {
		fmt.Println(err)
		fmt.Print("some error")
	}

	for l := 0; l < len(remote); l = l + 1 {
		if remote[l] != 10 {
			g.FILES[length] = g.FILES[length] + string(remote[l])
		} else {
			g.FILES = append(g.FILES, "\n")
			length = length + 1
		}
	}
	temp.Execute(response, g)
}

//uploader Uploads file from user's servercatchbox to the server's user's catchbox.
func uploader(response http.ResponseWriter, request *http.Request) {
	var upload1 = request.FormValue("upload1")
	fmt.Println("The file to upload is: " + upload1)

	remote2 := exec.Command("scp", "-i", key, "/home/"+localuser+"/servercatchbox/"+upload1, amazon+":home/"+remoteUsername+"/servercatchbox")
	remote2.Run()

	temp, _ := template.ParseFiles("html/remotefiles.html")
	response.Header().Set("Content-Type", "text/html; charset=utf-8")
	remote, err := exec.Command("ssh", "-i", key, amazon, "ls", "home/"+remoteUsername+"/servercatchbox", ">", "file1", ";", "cat", "file1").Output()

	g := localstruct{FILES: make([]string, 1)}
	length := 0
	if err != nil {
		fmt.Println(err)
	}
	for l := 0; l < len(remote); l = l + 1 {
		if remote[l] != 10 {
			g.FILES[length] = g.FILES[length] + string(remote[l])
		} else {
			g.FILES = append(g.FILES, "\n")
			length = length + 1
		}
	}
	remote2.Run()
	temp.Execute(response, g)
}

//downloader Downloads file from server user catchbox to local user catchbox.
func downloader(response http.ResponseWriter, request *http.Request) {
	var download1 = request.FormValue("download1")
	fmt.Println("The file to download is: " + download1)

	remote2 := exec.Command("scp", "-i", key, amazon+":home/"+remoteUsername+"/servercatchbox/"+download1, "/home/"+localuser+"/servercatchbox")
	remote2.Run()

	temp, _ := template.ParseFiles("html/localfiles.html")
	response.Header().Set("Content-Type", "text/html; charset=utf-8")
	remote, err := exec.Command("ls", "/home/"+localuser+"/servercatchbox", ">", "file1", ";", "cat", "file1").Output()

	g := localstruct{FILES: make([]string, 1)}
	length := 0
	if err != nil {
		fmt.Println(err)
	}
	for l := 0; l < len(remote); l = l + 1 {
		if remote[l] != 10 {
			g.FILES[length] = g.FILES[length] + string(remote[l])
		} else {
			g.FILES = append(g.FILES, "\n")
			length = length + 1
		}
	}
	temp.Execute(response, g)
}

func loginpage(response http.ResponseWriter, request *http.Request) {
	login, _ := template.ParseFiles("webpage/loginpage.html", "webpage/loginpage.css")
	login.Execute(response, nil)
}

func mainpage(response http.ResponseWriter, request *http.Request) {
	mainpage, _ := template.ParseFiles("webpage/mainpage.html", "webpage/mainpage.css")
	mainpage.Execute(response, nil)
}

//CreateFolder creates a new folder inside the server.
func CreateFolder(n string) {
	exec.Command("mkdir " + n)
}

//DeleteFolder will delete a folder inside the server.
func DeleteFolder(n string) {
	exec.Command("rm -rf " + n)
}

//NewFile will create a new file inside the current folder.
func NewFile(n string) {
	exec.Command("touch " + n)
}

//DeleteFile will delete a file inside the current folder.
func DeleteFile(n string) {
	exec.Command("rm " + n)

}
