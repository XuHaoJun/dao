package dao

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/sessions"
	"github.com/nu7hatch/gouuid"
	"github.com/xuhaojun/oauth2"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	"net/http"
)

type AccountRegisterFrom struct {
	Username string `form:"username" binding:"required"`
	Password string `form:"password" binding:"required"`
	Email    string `form:"email" binding:"required"`
}

func (accRegiForm *AccountRegisterFrom) IsVaild() bool {
	if len(accRegiForm.Username) < 4 ||
		len(accRegiForm.Password) < 4 {
		return false
	}
	return true
}

func (s *Server) DBHandler() martini.Handler {
	return func(c martini.Context) {
		dbSessionClone := s.db.CloneSession()
		c.Map(dbSessionClone)
		defer dbSessionClone.Close()
		c.Next()
	}
}

func handleAccountRegister(form AccountRegisterFrom, db *DaoDB, r render.Render, configs *DaoConfigs) {
	username := form.Username
	password := form.Password
	email := form.Email
	if !form.IsVaild() {
		clientErr := []interface{}{"duplicated account!"}
		clientCall := &ClientCall{
			Receiver: "world",
			Method:   "handleErrorLoginAccount",
			Params:   clientErr,
		}
		r.JSON(200, clientCall)
		return
	}
	queryAcc := bson.M{"username": username}
	err := db.accounts.Find(queryAcc).Select(bson.M{"_id": 1}).One(&struct{}{})
	var clientCall *ClientCall
	if err != nil && err != mgo.ErrNotFound {
		panic(err)
	} else if err != mgo.ErrNotFound {
		clientErr := []interface{}{"duplicated account!"}
		clientCall = &ClientCall{
			Receiver: "world",
			Method:   "handleErrorLoginAccount",
			Params:   clientErr,
		}
	} else if err == mgo.ErrNotFound {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 10)
		if err != nil {
			panic(err)
		}
		acc := NewAccount(username, string(hashedPassword))
		acc.email = email
		acc.maxChars = configs.AccountConfigs.MaxChars
		acc.SaveByOtherDB(db)
		go db.UpdateAccountIndex()
		clientParams := []interface{}{"success register a new account!"}
		clientCall = &ClientCall{
			Receiver: "world",
			Method:   "handleSuccessRegisterAccount",
			Params:   clientParams,
		}
	}
	r.JSON(200, clientCall)
}

func handleAccountInfo(db *DaoDB, session sessions.Session, r render.Render, w *World) {
	var clientCall *ClientCall
	username := session.Get("username")
	if username == nil {
		clientCall = &ClientCall{
			Receiver: "world",
			Method:   "handleWebAccountInfo",
			Params:   []interface{}{map[string]string{"error": "not login!"}},
		}
		r.JSON(200, clientCall)
		return
	} else {
		foundAcc := &AccountDumpDB{}
		queryAcc := bson.M{"username": username.(string)}
		err := db.accounts.Find(queryAcc).One(foundAcc)
		if err != nil && err != mgo.ErrNotFound {
			panic(err)
		}
		if err == mgo.ErrNotFound || err != nil {
			clientCall = &ClientCall{
				Receiver: "world",
				Method:   "handleWebAccountInfo",
				Params: []interface{}{map[string]string{
					"error": "wrong username"}},
			}
			r.JSON(200, clientCall)
			return
		}
		acc := foundAcc.Load(w)
		charClients := make([]interface{}, len(acc.chars))
		for i, char := range acc.chars {
			charClients[i] = char.CharClient()
		}
		param := map[string]interface{}{
			"username":    acc.username,
			"email":       acc.email,
			"maxChars":    acc.maxChars,
			"charConfigs": charClients,
		}
		clientCall = &ClientCall{
			Receiver: "world",
			Method:   "handleWebAccountInfo",
			Params:   []interface{}{param},
		}
	}
	r.JSON(200, clientCall)
}

func handleAccountRegisterByFacebook(db *DaoDB, r render.Render, tokens oauth2.Tokens, configs *DaoConfigs) {
	url := "https://graph.facebook.com/me?access_token=" + tokens.Access()
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalln(err)
	}
	req.Header.Add("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	decoder := json.NewDecoder(resp.Body)
	v := &FacebookProfile{}
	err = decoder.Decode(v)
	if err != nil {
		log.Panicln(err)
	}
	resp.Body.Close()
	hasher := md5.New()
	hasher.Write([]byte("facebook" + v.Id))
	username := hex.EncodeToString(hasher.Sum(nil))
	password := v.Id + "facebook"
	handleAccountRegister(
		AccountRegisterFrom{username, password, ""},
		db, r, configs)
}

func handleAccountLoginByFacebook(params martini.Params, db *DaoDB, r render.Render, tokens oauth2.Tokens, session sessions.Session) {
	ltype := params["ltype"]
	switch ltype {
	case "Web":
		handleAccountLoginWebByFacebook(db, r, tokens, session)
	case "Game":
		handleAccountLoginGameByFacebook(db, r, tokens, session)
	default:
		r.JSON(200, map[string]string{"error": "not match anything!"})
	}
}

func handleAccountLoginWebByFacebook(db *DaoDB, r render.Render, tokens oauth2.Tokens, session sessions.Session) {
	if tokens.Expired() || tokens.ProviderName() != "Facebook" {
		r.Redirect("oauth2login?next=#home", 302)
		return
	}
	url := "https://graph.facebook.com/me?access_token=" + tokens.Access()
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalln(err)
	}
	req.Header.Add("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	decoder := json.NewDecoder(resp.Body)
	v := &struct {
		Id   string `json:"id"`
		Name string `json:"name"`
	}{}
	err = decoder.Decode(v)
	if err != nil {
		log.Panicln(err)
	}
	resp.Body.Close()
	hasher := md5.New()
	hasher.Write([]byte("facebook" + v.Id))
	username := hex.EncodeToString(hasher.Sum(nil))
	session.Set("username", username)
	clientCall := &ClientCall{
		Receiver: "world",
		Method:   "handleSetLastUsername",
		Params:   []interface{}{username},
	}
	r.JSON(200, clientCall)
}

func handleAccountLoginGameByFacebook(db *DaoDB, r render.Render, tokens oauth2.Tokens, session sessions.Session) {
	if tokens.Expired() || tokens.ProviderName() != "Facebook" {
		r.Redirect("oauth2login?next=#loginFacebook", 302)
		return
	}
	url := "https://graph.facebook.com/me?access_token=" + tokens.Access()
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalln(err)
	}
	req.Header.Add("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	decoder := json.NewDecoder(resp.Body)
	v := &FacebookProfile{}
	err = decoder.Decode(v)
	if err != nil {
		log.Panicln(err)
	}
	resp.Body.Close()
	hasher := md5.New()
	hasher.Write([]byte("facebook" + v.Id))
	username := hex.EncodeToString(hasher.Sum(nil))
	password := v.Id + "facebook"
	// TODO
	// find other way to do login!
	session.Set("username", username)
	clientCall := &ClientCall{
		Receiver: "world",
		Method:   "loginAccount",
		Params:   []interface{}{username, password},
	}
	r.JSON(200, clientCall)
}

type AccountLoginForm struct {
	Username string `form:"username" binding:"required"`
	Password string `form:"password" binding:"required"`
}

func handleAccountLogin(params martini.Params, form AccountLoginForm, session sessions.Session, r render.Render, db *DaoDB) {
	ltype := params["ltype"]
	switch ltype {
	case "Web":
		handleAccountLoginWeb(form, session, r, db)
	case "Game":
		handleAccountLoginGame(form, session, r, db)
	default:
		r.JSON(200, map[string]string{"error": "not match anything!"})
	}
}

func handleAccountLoginWeb(form AccountLoginForm, session sessions.Session, r render.Render, db *DaoDB) {
	username := form.Username
	password := form.Password
	foundAcc := struct {
		Password string `bson:"password"`
	}{}
	queryAcc := bson.M{"username": username}
	err := db.accounts.Find(queryAcc).Select(bson.M{"password": 1}).One(&foundAcc)
	if err != nil && err != mgo.ErrNotFound {
		panic(err)
	}
	passwordErr := bcrypt.CompareHashAndPassword([]byte(foundAcc.Password), []byte(password))
	if err == mgo.ErrNotFound || passwordErr != nil {
		clientErr := []interface{}{"wrong username or password"}
		clientCall := &ClientCall{
			Receiver: "world",
			Method:   "handleErrorLoginAccount",
			Params:   clientErr,
		}
		r.JSON(200, clientCall)
		return
	}
	session.Set("username", username)
	clientCall := &ClientCall{
		Receiver: "world",
		Method:   "handleSetLastUsername",
		Params:   []interface{}{username},
	}
	r.JSON(200, clientCall)
}

func handleAccountLoginGameBySession(session sessions.Session, r render.Render, w *World) {
	username := session.Get("username")
	if username == nil {
		clientErr := []interface{}{"wrong username or password"}
		clientCall := &ClientCall{
			Receiver: "world",
			Method:   "handleErrorLoginAccount",
			Params:   clientErr,
		}
		r.JSON(200, clientCall)
		return
	}
	base, _ := uuid.NewV4()
	token := base.String()
	w.addAccountLoginBySession <- AccountLoginBySession{
		username.(string),
		token,
	}
	clientCall := &ClientCall{
		Receiver: "world",
		Method:   "loginAccountBySessionToken",
		Params:   []interface{}{username, token},
	}
	session.Set("username", username.(string))
	r.JSON(200, clientCall)
}

func handleAccountLoginGame(form AccountLoginForm, session sessions.Session, r render.Render, db *DaoDB) {
	username := form.Username
	password := form.Password
	foundAcc := struct {
		Password string `bson:"password"`
	}{}
	queryAcc := bson.M{"username": username}
	err := db.accounts.Find(queryAcc).Select(bson.M{"password": 1}).One(&foundAcc)
	if err != nil && err != mgo.ErrNotFound {
		panic(err)
	}
	err = bcrypt.CompareHashAndPassword([]byte(foundAcc.Password), []byte(password))
	if err == mgo.ErrNotFound || err != nil {
		clientErr := []interface{}{"wrong username or password"}
		clientCall := &ClientCall{
			Receiver: "world",
			Method:   "handleErrorLoginAccount",
			Params:   clientErr,
		}
		r.JSON(200, clientCall)
		return
	}
	clientCall := &ClientCall{
		Receiver: "world",
		Method:   "loginAccount",
		Params:   []interface{}{username, password},
	}
	session.Set("username", username)
	r.JSON(200, clientCall)
}

func haldeAccountIsLogined(session sessions.Session, r render.Render) {
	username := session.Get("username")
	if username == nil {
		r.JSON(200, map[string]string{"error": "not logined!"})
		return
	}
	r.JSON(200, map[string]string{"success": "logined!",
		"username": username.(string)})
}

func hanldeAccountLogout(session sessions.Session, r render.Render) {
	username := session.Get("username")
	if username != nil {
		session.Delete("username")
		log.Println(username)
		r.JSON(200, map[string]string{"success": "delete session"})
		return
	}
	r.JSON(200, map[string]string{"error": "not logined!"})
}

func handleClientVersion(r render.Render, s *Server) {
	clientCall := &ClientCall{
		Receiver: "world",
		Method:   "handleSyncClient",
		Params: []interface{}{
			map[string]string{
				"version": s.configs.ServerConfigs.ClientVersion,
			},
		},
	}
	r.JSON(200, clientCall)
}

func handleWebsocketPort(r render.Render, s *Server) {
	r.JSON(200, s.configs.ServerConfigs.WebsocketPort)
}
