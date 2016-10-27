package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"runtime"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/HouzuoGuo/tiedot/db"
)

//Db is the SQLITE database object
var Db *sql.DB
var port = ":8080"

//var UseStdOut = true

//Person  is a single person
type Person struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func main() {
	logParam := flag.Bool("log", false, "a bool")
	sqliteParam := flag.Bool("sqlite", false, "a bool")

	if *logParam {
		InitLog()
		log.Println("Writing log file to : logfile.txt")
	} else {
		log.SetOutput(os.Stdout)
		log.Println("Writing log file to stdOut")
	}

	if *sqliteParam {
		log.Println("SQLite database in memory only")
		InitDb()
	} else {
		log.Println("SQLite database on disk")
		OpenDb()
	}
	ConfigRuntime()
	StartGin()
}

func OpenDb() {
	var err error
	//Db, err = sql.Open("sqlite3", "./heroes.sqlite")
	Db,err := db.OpenDB("heroes")

	if err != nil {
		log.Fatal(err)
	}
	//defer Db.Close()
  if err = myDB.Create("Heroes"); err == nil {
		LoadDb()
	}
}

func InitLog() {

	var err error
	var f *os.File
	f, err = os.OpenFile("logfile.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		//fmt.fprintln("error opening file: %v", err)
		fmt.Printf("%v\n", err)
	}
	defer f.Close()
	log.SetOutput(f)
}

//InitDb intialize databases
func InitDb() {
	var err error
	//Db, err = sql.Open("sqlite3", ":memory:")
	Db,err := db.OpenDB("heroes")
	if err != nil {
		log.Fatal(err)
	}

}
func LoadDb() {
	var err error
	var strs = `[{ "id": 11, "name": "Mr. Nice" },
{ "id": 12, "name": "Narco" },
{ "id": 13, "name": "Bombasto" },
{ "id": 14, "name": "Celeritas" },
{ "id": 15, "name": "Magneta" },
{ "id": 16, "name": "RubberMan" },
{ "id": 17, "name": "Dynama" },
{ "id": 18, "name": "Dr IQ" },
{ "id": 19, "name": "Magma" },
{ "id": 20, "name": "Tornado" }]`

	//fmt.Fprintln(os.Stderr, "hello world")
	//os.Stdout.Write([]byte(strs))
	//os.Stdout.WriteString("\n")
	//fmt.Println("hello world")

	//var vals []Person
	vals := make([]Person, 0)
	json.Unmarshal([]byte(strs), &vals)

	//err := os.Remove("./foo.Db")

	//if err != nil {
	//fmt.Println(err)
	//fmt.Println("Creating SQLite Db", "Log")
	//return
	//}

	//log.Println("Creating SQLite Db", "Log")
	//Db, err := sql.Open("sqlite3", "./foo.Db")
	//defer Db.Close()

	sqlStmt := `
	create table heroes (id INTEGER PRIMARY KEY AUTOINCREMENT, name text);
	delete from heroes;
	`
	//log.Println("Executing SQLite Db", "Log")
	_, err = Db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}
	//log.Println("1", "Log")
	tx, err := Db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	stmt, err := tx.Prepare("insert into heroes(id, name) values(?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	for _, k := range vals {
		//os.Stdout.WriteString(strconv.Itoa(k.Id) + ", " + k.Name + "\n")
		//log.Println(strconv.Itoa(k.ID) + ", " + k.Name + "\n")
		_, err = stmt.Exec(k.ID, fmt.Sprintf("%s", k.Name))
		if err != nil {
			log.Fatal(err)
		}
	}

	/*
		for i := 0; i < 100; i++ {
			_, err = stmt.Exec(i, fmt.Sprintf("ROW %d", i))
			if err != nil {
				log.Fatal(err)
			}
		}
	*/
	tx.Commit()
}

//ConfigRuntime print out configuration details
func ConfigRuntime() {
	nuCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(nuCPU)
	log.Printf("Running with %d CPUs\n", nuCPU)
	host, _ := os.Hostname()
	addrs, _ := net.LookupIP(host)
	for _, addr := range addrs {
		if ipv4 := addr.To4(); ipv4 != nil {
			fmt.Println("IPv4: ", ipv4)
		}
	}
	ip, err := externalIP()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Public IP: " + ip)
}

func externalIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "", errors.New("are you connected to the network?")
}

//func postHeroes(c *gin.Context) {
//	fmt.Println("Posting heroes")
//}
func getHeroes(c *gin.Context) {
	rows, err := Db.Query("select id, name from heroes")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	heroes := make([]Person, 0)
	for rows.Next() {
		var id int
		var name string
		err = rows.Scan(&id, &name)
		if err != nil {
			log.Fatal(err)
		}
		//log.Println(id, name)
		p := Person{id, name}
		heroes = append(heroes, p)

	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	//str, err := json.Marshal(heroes)
	//if err != nil {
	//	log.Println("error:", err)
	//}
	//log.Println(string(str[:]))
	//fmt.Fprintln(c, string(str))
	c.JSON(http.StatusOK, heroes)
}

//StartGin starts up gin and sets the routes
func StartGin() {
	/*

		router := gin.New()
		router.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "pong",
			})
		})
	*/
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	//router.POST("/heroes", postHeroes)

	log.Println("Starting up Gin on port: " + port)
	//router.Use(rateLimit, gin.Recovery())
	// Creates a gin router with default middleware:
	// logger and recovery (crash-free) middleware
	//router := gin.Default()
	router.HEAD("/", func(c *gin.Context) {
		fmt.Println("Head")
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Header("Access-Control-Allow-Headers",
			"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		//c.Header("Content-Type", "application/json")
	})

	router.OPTIONS("/heroes/:id", func(c *gin.Context) {
		fmt.Println("Options")
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Header("Access-Control-Allow-Headers",
			"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	})
	router.OPTIONS("/heroes", func(c *gin.Context) {
		fmt.Println("Options")
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Header("Access-Control-Allow-Headers",
			"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	})

	router.GET("/heroes", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Header("Access-Control-Allow-Headers",
			"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Header("Content-Type", "application/json")
		getHeroes(c)

		//return nil
		/*
		   c.JSON(200, gin.H{
		               "message": "pong",
		           })
		*/
	})

	router.POST("/heroes", func(c *gin.Context) {
		//fmt.Println("Post")
		//name := c.PostForm("")
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Header("Access-Control-Allow-Headers",
			"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Header("Content-Type", "application/json")

		var req json.RawMessage
		if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
			log.Panic(err)
			//panic(err)
		}
		cMap := make(map[string]string)

		e := json.Unmarshal(req, &cMap)
		if e != nil {
			log.Panic(e)
		}
		log.Println("Posting id: " + cMap["name"])

		log.Println(cMap)
		stmt, err := Db.Prepare("INSERT INTO heroes(name) values(?)")

		if err != nil {
			log.Panic(err)
		}
		//id := "(select max(id)+1 from heroes)"
		//id := 99
		res, err := stmt.Exec(cMap["name"])
		if err != nil {
			log.Panic(err)
		}

		affect, err := res.RowsAffected()
		if err != nil {
			log.Panic(err)
		}

		log.Println(affect)
		if err != nil {
			log.Panic(err)
		}
		/*
			c.JSON(http.StatusOK, gin.H{
				"status":  "1",
				"message": "OK",
			})
		*/
		getHeroes(c)
	})
	router.PUT("/heroes/:id", func(c *gin.Context) {
		fmt.Println("Put")
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Header("Access-Control-Allow-Headers",
			"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Header("Content-Type", "application/json")
		var req json.RawMessage
		//var error error
		if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
			log.Panic(err)
			//panic(err)
		}
		var hero Person
		//err := json.NewDecoder().Decode(&vals)
		err := json.Unmarshal(req, &hero)

		//heroes := make([]Person, 0)
		//cMap := make(map[string]string)

		//err := json.Unmarshal(req, &cMap)
		if err != nil {
			log.Panic(err)
		}

		log.Println("Updating id: " + strconv.Itoa(hero.ID) + " with name=" + hero.Name)
		//return
		stmt, err := Db.Prepare("update heroes set name=? where id=?")
		if err != nil {
			log.Panic(err)
		}
		//id := "(select max(id)+1 from heroes)"
		//id := 99

		res, err := stmt.Exec(hero.Name, hero.ID)
		if err != nil {
			log.Panic(err)
		}

		affect, err := res.RowsAffected()
		if err != nil {
			log.Panic(err)
		}

		log.Println(affect)
		if err != nil {
			log.Panic(err)
		}
		//log.Println("Updating id " + strconv.Itoa(hero.ID))
		/*
			c.JSON(http.StatusOK, gin.H{
				"status": "1",
				"id":     c.Param("id"),
			})
		*/
		getHeroes(c)
	})

	router.DELETE("/heroes/:id", func(c *gin.Context) {
		//fmt.Println("Delete")
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Header("Access-Control-Allow-Headers",
			"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Header("Content-Type", "application/json")

		log.Println("Deleting id: " + c.Param("id"))
		stmt, err := Db.Prepare("DELETE from heroes where id=?")
		if err != nil {
			log.Panic(err)
		}
		//id := "(select max(id)+1 from heroes)"
		//id := 99

		res, err := stmt.Exec(c.Param("id"))
		if err != nil {
			log.Panic(err)
		}

		affect, err := res.RowsAffected()
		if err != nil {
			log.Panic(err)
		}

		log.Println(affect)
		if err != nil {
			log.Panic(err)
		}
		//log.Println("Deleting id " + c.Param("id"))
		/*
			c.JSON(http.StatusOK, gin.H{
				"status": "1",
				"id":     c.Param("id"),
			})
		*/
		getHeroes(c)
	})

	/*
	   router.GET("/someGet", getting)
	   router.POST("/somePost", posting)
	   router.PUT("/somePut", putting)
	   router.DELETE("/someDelete", deleting)
	   router.PATCH("/somePatch", patching)
	   router.HEAD("/someHead", head)
	   router.OPTIONS("/someOptions", options)
	*/
	router.Run(port)

}

/*
	rows, err := Db.Query("select id, name from foo")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var name string
		err = rows.Scan(&id, &name)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(id, name)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	stmt, err = Db.Prepare("select name from foo where id = ?")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	var name string
	err = stmt.QueryRow("3").Scan(&name)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(name)

	_, err = Db.Exec("delete from foo")
	if err != nil {
		log.Fatal(err)
	}

	_, err = Db.Exec("insert into foo(id, name) values(1, 'foo'), (2, 'bar'), (3, 'baz')")
	if err != nil {
		log.Fatal(err)
	}

	rows, err = Db.Query("select id, name from foo")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var name string
		err = rows.Scan(&id, &name)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(id, name)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Done with SQLite Db", "Log")
*/
/*
	red := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		Db:       1,
	})

	pong, err := red.Ping().Result()
	if pong != "PONG" {
		fmt.Printf("Failed to connect to Redis", err)
		os.Exit(1)
	}
*/
//r := gin.Default()

// Ping test

//r.GET("/ping", func(c *gin.Context) {

//	c.String(200, "pong")

//})
// Listen and Server in 0.0.0.0:8080

//r.Run(":8080")

/*
	router := routing.New()

	router.Use(
		access.Logger(log.Printf),
		slash.Remover(http.StatusMovedPermanently),
		fault.Recovery(log.Printf),
	)

	api := router.Group("/v1")

	api.Use(
		content.TypeNegotiator(content.JSON),
	)
	api.Get("/heroes", func(c *routing.Context) error {
		rows, err := Db.Query("select id, name from heroes")
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()
		heroes := make([]Person, 0)
		for rows.Next() {
			var id int
			var name string
			err = rows.Scan(&id, &name)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(id, name)
			p := Person{id, name}
			heroes = append(heroes, p)

		}
		err = rows.Err()
		if err != nil {
			log.Fatal(err)
		}
		c.Response.Header().Set("Access-Control-Allow-Origin", "*")
		c.Response.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Response.Header().Set("Access-Control-Allow-Headers",
			"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Response.Header().Set("Content-Type", "application/json")
		str, err := json.Marshal(heroes)
		if err != nil {
			fmt.Println("error:", err)
		}
		//fmt.Println(string(str[:]))
		fmt.Fprintln(c.Response, string(str))

		return nil
	})

	api.Post("/heroes", func(c *routing.Context) error {
		var req json.RawMessage
		if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
			return err
		}

		cMap := make(map[string]string)

		e := json.Unmarshal(req, &cMap)
		if e != nil {
			panic(e)
		}

		fmt.Println(cMap)
		stmt, err := Db.Prepare("INSERT INTO heroes(name) values(?)")

		if err != nil {
			panic(err)
		}
		//id := "(select max(id)+1 from heroes)"
		//id := 99
		res, err := stmt.Exec(cMap["name"])
		if err != nil {
			panic(err)
		}

		affect, err := res.RowsAffected()
		if err != nil {
			panic(err)
		}

		fmt.Println(affect)
		if err != nil {
			panic(err)
		}
		c.Response.Header().Set("Access-Control-Allow-Origin", "*")
		c.Response.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Response.Header().Set("Access-Control-Allow-Headers",
			"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Response.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(c.Response, "{1}")
		return nil
	})

	api.Delete(`/heroes/<id:\d+>`, func(c *routing.Context) error {
		fmt.Println("Deleting id: " + c.Param("id"))
		stmt, err := Db.Prepare("DELETE from heroes where id=?")
		if err != nil {
			panic(err)
		}
		//id := "(select max(id)+1 from heroes)"
		//id := 99

		res, err := stmt.Exec(c.Param("id"))
		if err != nil {
			panic(err)
		}

		affect, err := res.RowsAffected()
		if err != nil {
			panic(err)
		}

		fmt.Println(affect)
		if err != nil {
			panic(err)
		}
		c.Response.Header().Set("Access-Control-Allow-Origin", "*")
		c.Response.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Response.Header().Set("Access-Control-Allow-Headers",
			"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Response.Header().Set("Content-Type", "application/json")
		//fmt.Fprintln(c.Response, "{1}")

		//c.Write(json.NewEncoder(c.Response).Encode(dir))
		//fmt.Fprintf(c.Response, "ret %v", json.NewEncoder(c.Response).Encode(dir))
		fmt.Fprintln(c.Response, "{1}")
		return nil
	})
	api.Put(`/heroes/<id:\d+>`, func(c *routing.Context) error {
		return c.Write("update user " + c.Param("id"))
	})
*/
/*
   	api.Get(`/hgetall/<id:\D+>`, func(c *routing.Context) error {

   		redHash, err := red.HGetAll(c.Param("id")).Result()
   		if err == redis.Nil {
   			return c.Write("None")
   		}

   		return c.Write(redHash)
   	})

   	api.Get(`/hmset/<id:\D+>`, func(c *routing.Context) error {

   		var req json.RawMessage
   		if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
   			return err
   		}

   		cMap := make(map[string]string)

   		e := json.Unmarshal(req, &cMap)
   		if e != nil {
   			panic(e)
   		}

   		//for keys, vals := range cMap {

   		//red.HMSet(c.Param("id"), vals) //vals
   		//}

   		return c.Write(cMap)
   	})

   	api.Get(`/set/<id>/<value>`, func(c *routing.Context) error {
   		err := red.Set(c.Param("id"), c.Param("value"), 0).Err()
   		if err != nil {
   			panic(err)
   		}

   		dataGet, err := red.Get(c.Param("id")).Result()
   		if err == redis.Nil {
   			return c.Write("None")
   		}

   		return c.Write(dataGet)
   	})

   	api.Get(`/get/<id:\D+>`, func(c *routing.Context) error {
   		c.Write(c.Param("id"))
   		val, err := red.Get(c.Param("id")).Result()

   		if err == redis.Nil {
   			return c.Write("None")
   		}

   		return c.Write(val)
   	})
   	api.Get("/json", func(c *routing.Context) error {
   		dir := []string{"user", "doc", "bin", "src"}

   		//c.Write(json.NewEncoder(c.Response).Encode(dir))
   		fmt.Fprintf(c.Response, "ret %v", json.NewEncoder(c.Response).Encode(dir))
   		return nil
   	})

   	api.Get("/getall", func(c *routing.Context) error {
   		//redHash, err := red.HGetAll("heroes").Result()
   		strs, err := red.Get("heroes").Result()
   		if err == redis.Nil {
   			return c.Write("None")
   		}
   		c.Response.Header().Set("Access-Control-Allow-Origin", "*")
   		c.Response.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
   		c.Response.Header().Set("Access-Control-Allow-Headers",
   			"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
   		c.Response.Header().Set("Content-Type", "application/json")

   		//vals := make([]Person, 0)
   		var vals []Person
   		//err := json.NewDecoder().Decode(&vals)
   		json.Unmarshal([]byte(strs), &vals)


   		//	b, err := json.Marshal(redHash)
   		//	if err != nil {
   		//		fmt.Println("error:", err)
   		//	}

   		//for _, k := range vals {
   		//	os.Stdout.WriteString(strconv.Itoa(k.Id) + ", " + k.Name + "\n")
   		//}
   		//str := json.NewEncoder(c.Response).Encode(vals)
   		str, err := json.Marshal(vals)
   		if err != nil {
   			fmt.Println("error:", err)
   		}
   		fmt.Println(string(str[:]))
   		//os.Stdout.WriteString(str[:])
   		//str, _ = "var heroes=" + string(str)
   		//return c.Write(redHash)
   		//c.Write(string(str))
   		//return str

   		//return str
   		fmt.Fprintln(c.Response, string(str))
   		//fmt.Fprintln(c.Response, json.NewEncoder(c.Response).Encode(vals))
   		//return c.Write(redHash)
   		return nil
   	})

   	api.Get("/setall", func(c *routing.Context) error {
   		var strs = `[{ "id": 11, "name": "Mr. Nice" },
   { "id": 12, "name": "Narco" },
   { "id": 13, "name": "Bombasto" },
   { "id": 14, "name": "Celeritas" },
   { "id": 15, "name": "Magneta" },
   { "id": 16, "name": "RubberMan" },
   { "id": 17, "name": "Dynama" },
   { "id": 18, "name": "Dr IQ" },
   { "id": 19, "name": "Magma" },
   { "id": 20, "name": "Tornado" }]`

   		//fmt.Fprintln(os.Stderr, "hello world")
   		//os.Stdout.Write([]byte(strs))
   		//os.Stdout.WriteString("\n")
   		//fmt.Println("hello world")

   		//var vals []Person
   		vals := make([]Person, 0)
   		json.Unmarshal([]byte(strs), &vals)
   		//fmt.Printf("%#v\n", vals)
   		os.Stdout.WriteString("Unmarshalled:\n")
   		for _, k := range vals {
   			os.Stdout.WriteString(strconv.Itoa(k.Id) + ", " + k.Name + "\n")
   		}
   		//for i := 0; i < len(vals); i++ {
   		//	os.Stdout.WriteString(string(vals[i].Id) + ", " + vals[i].Name + "\n")
   		//}
   		//vals := make(map[*Person]bool)

   		//c.Param("value")
   		var key = "heroes"
   		b, err := json.Marshal(vals)
   		if err != nil {
   			fmt.Println("error:", err)
   		}
   		//c.Param("id")
   		err1 := red.Set(key, b, 0).Err()
   		if err1 != nil {
   			panic(err1)
   		}
   		os.Stdout.WriteString("JSON values:  " + string(b))

   		//red.Set("heroes", vals) //vals

   		//fmt.Fprintln(c.Response, json.NewEncoder(c.Response).Encode(redHash))
   		return c.Write(vals)
   		//return nil
   	})

   	api.Get("/load", func(c *routing.Context) error {
   		var strs = `[{ "id": 11, "name": "Mr. Nice" },
   { "id": 12, "name": "Narco" },
   { "id": 13, "name": "Bombasto" },
   { "id": 14, "name": "Celeritas" },
   { "id": 15, "name": "Magneta" },
   { "id": 16, "name": "RubberMan" },
   { "id": 17, "name": "Dynama" },
   { "id": 18, "name": "Dr IQ" },
   { "id": 19, "name": "Magma" },
   { "id": 20, "name": "Tornado" }]`

   		//fmt.Fprintln(os.Stderr, "hello world")
   		//os.Stdout.Write([]byte(strs))
   		//os.Stdout.WriteString("\n")
   		//fmt.Println("hello world")

   		//var vals []Person
   		vals := make([]Person, 0)
   		json.Unmarshal([]byte(strs), &vals)
   		//fmt.Printf("%#v\n", vals)
   		os.Stdout.WriteString("Unmarshalled:\n")
   		for _, k := range vals {
   			os.Stdout.WriteString(strconv.Itoa(k.Id) + ", " + k.Name + "\n")
   			//c.Param("id")
   			//err1 := red.Set(key, b, 0).Err()
   			//if err1 != nil {
   			//	panic(err1)
   			//}

   		}
   		//for i := 0; i < len(vals); i++ {
   		//	os.Stdout.WriteString(string(vals[i].Id) + ", " + vals[i].Name + "\n")
   		//}
   		//vals := make(map[*Person]bool)

   		//c.Param("value")
   		//os.Stdout.WriteString("JSON values:  " + string(b))

   		//red.Set("heroes", vals) //vals

   		//fmt.Fprintln(c.Response, json.NewEncoder(c.Response).Encode(redHash))
   		return c.Write(vals)
   		//return nil
   	})

   	api.Get("/heroes", func(c *routing.Context) error {
   		//redHash, err := red.HGetAll("heroes").Result()
   		strs, err := red.Get("heroes").Result()
   		if err == redis.Nil {
   			return c.Write("None")
   		}
   		c.Response.Header().Set("Access-Control-Allow-Origin", "*")
   		c.Response.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
   		c.Response.Header().Set("Access-Control-Allow-Headers",
   			"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
   		c.Response.Header().Set("Content-Type", "application/json")

   		//vals := make([]Person, 0)
   		var vals []Person
   		//err := json.NewDecoder().Decode(&vals)
   		json.Unmarshal([]byte(strs), &vals)


   		//	b, err := json.Marshal(redHash)
   		//	if err != nil {
   		//		fmt.Println("error:", err)
   		//	}

   		//for _, k := range vals {
   		//	os.Stdout.WriteString(strconv.Itoa(k.Id) + ", " + k.Name + "\n")
   		//}
   		//str := json.NewEncoder(c.Response).Encode(vals)
   		str, err := json.Marshal(vals)
   		if err != nil {
   			fmt.Println("error:", err)
   		}
   		fmt.Println(string(str[:]))
   		//os.Stdout.WriteString(str[:])
   		//str, _ = "var heroes=" + string(str)
   		//return c.Write(redHash)
   		//c.Write(string(str))
   		//return str

   		//return str
   		fmt.Fprintln(c.Response, string(str))
   		//fmt.Fprintln(c.Response, json.NewEncoder(c.Response).Encode(vals))
   		//return c.Write(redHash)
   		return nil
   	})
*/

// serve index file

//router.Get("/", file.Content("index.html"))
// serve files under the "ui" subdirectory
//router.Get("/*", file.Server(file.PathMap{
//	"/": "/",
//}))

//http.Handle("/", router)
//http.ListenAndServe("0.0.0.0:8080", nil)

//fmt.Fprintln("Starting server")
//}

/*
func _main() {
	router := routing.New()

	router.Use(
		// all these handlers are shared by every route
		access.Logger(log.Printf),
		slash.Remover(http.StatusMovedPermanently),
		fault.Recovery(log.Printf),
	)

	// serve RESTful APIs
	api := router.Group("/api")
	api.Use(
		// these handlers are shared by the routes in the api group only
		content.TypeNegotiator(content.JSON, content.XML),
	)
	api.Get("/users/<username>", func(c *routing.Context) error {
		fmt.Fprintf(c.Response, "Name: %v", c.Param("username"))
		return nil
		//return c.Write("user list")
	})
	api.Post("/users", func(c *routing.Context) error {
		return c.Write("create a new user")
	})
	api.Put(`/users/<id:\d+>`, func(c *routing.Context) error {
		return c.Write("update user " + c.Param("id"))
	})

	// serve index file
	router.Get("/", file.Content("app/app.component.html"))
	// serve files under the "ui" subdirectory
	router.Get("/*", file.Server(file.PathMap{
		"/": "/app/",
	}))

	http.Handle("/", router)
	http.ListenAndServe(":8080", nil)
}
*/
