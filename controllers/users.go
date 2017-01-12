// Bundle of functions managing the CRUD
package controllers

import (
	"database/sql"
	"net/http"
	"strconv"
	"net/smtp"
	"time"
	"strings"

	"golang.org/x/crypto/bcrypt"

	. "github.com/quorumsco/jsonapi"
	"github.com/quorumsco/logs"
	"github.com/quorumsco/oauth2/models"
	"github.com/quorumsco/oauth2/views"
	"github.com/quorumsco/router"
	"github.com/quorumsco/application"
	"github.com/quorumsco/settings"
)



// Return a string's pointer
func sPtr(s string) *string {
	if s == "" {
		return nil
	} else {
		return &s
	}
}

// Creates a new user
func Register(w http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {
		req.ParseForm()

		passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.FormValue("password")), bcrypt.DefaultCost)
		if err != nil {
			panic(err)
		}

		u := &models.User{
			Firstname: sPtr(req.FormValue("firstname")),
			Surname:   sPtr(req.FormValue("surname")),
			Mail:      sPtr(req.FormValue("mail")),
			Password:  sPtr(string(passwordHash)),
			Cause:		 sPtr(req.FormValue("cause")),
		}

		errs := u.Validate()
		if len(errs) > 0 {
			logs.Error(errs)
			Error(w, req, "Vous avez une ou des erreur(s) dans le formulaire d'inscription. vérifiez votre saisie (formatage du mail par exemple)", http.StatusBadRequest)
			//Fail(w, req, "", http.StatusInternalServerError)
			return
		}

		var store = models.UserStore(getDB(req))
		err = store.Save(u)
		if err != nil {
			logs.Error(err)
			Error(w, req, err.Error(), http.StatusBadRequest)
			return
		}
	}

	templates := getTemplates(req)
	if err := templates["users/register"].ExecuteTemplate(w, "base", nil); err != nil {
		logs.Error(err)
	}
}

func SendEmail(r *http.Request,to *string,url string,prenom string) {
	//var settings string
	conf := router.Context(r).Env["Application"].(*application.Application).Components["Smtp"].(settings.Smtp)
	// Set up authentication information.
	auth := smtp.PlainAuth("", conf.User, conf.Password, conf.Smtpserver)

	// Connect to the server, authenticate, set the sender and recipient,
	// and send the email all in one step.
	//to := []string{"jbd@quorumapp.co"}
	var msg []byte
	if (url=="Confirmation"){
		msg = []byte("To: "+*to+"\r\n" +
			"Subject: QUORUM - Confirmation de changement de mot de passe\r\n" +
			"\r\n" +
			"Bonjour " + prenom +",\r" +
			"Votre nouveau mot de passe est effectif, vous pouvez vous connecter aux applications QUORUM." +
			"\r\rL'équipe Quorum\r\n")

	}else if (url=="NotMatching"){
		*to=conf.User
		msg = []byte("To: "+*to+"\r\n" +
			"Subject: Not matching CODE\r\n" +
			"\r\n" +
			"Tentative de matching d'un mauvais code. User ID:" +
			prenom +
			"\r\n")
	}else {
		msg = []byte("To: "+*to+"\r\n" +
			"Subject: QUORUM - Demande de changement de mot de passe\r\n" +
			"\r\n" +
			"Bonjour " + prenom +",\r" +
			"Vous venez de faire une demande de changement de mot de passe.\rPour la valider, veuillez cliquer sur le lien ci dessous:\r" +
			url +
			"\rVous recevrez une confirmation par mail, une fois le lien cliqué, de la réussite du changement de mot de passe." +
			"\r\rL'équipe Quorum\r\n")
	}
	toBis:=[]string {*to}
	err := smtp.SendMail(conf.Smtpserver+":"+conf.Port, auth, conf.User, toBis, msg)
	if err != nil {
		logs.Error(err)
	}
}

func ValidPassword(w http.ResponseWriter, req *http.Request) {
	u := &models.User{
		Mail:     sPtr(req.FormValue("mail")),
		// only to pass the "update" control
		Password: sPtr(req.FormValue("code")),
	}
	var store = models.UserStore(getDB(req))
	err := store.First(u)
	if err != nil {
		logs.Error(err)
		Error(w, req, err.Error(), http.StatusBadRequest)
		return
	}else
	{
		if (*sPtr(req.FormValue("code"))==*u.Validationcode) {
			//the validation code is correct
			//mise à jour en base du user
			u.GroupID=u.OldgroupID
			u.OldgroupID = 99999
			temp:="&"
			u.Validationcode = &temp
			err = store.Update(u)
			if err != nil {
				logs.Error(err)
				Error(w, req, err.Error(), http.StatusBadRequest)
				return
			}else{
					SendEmail(req,sPtr(req.FormValue("mail")),"Confirmation",*u.Firstname)
					//Error(w, req, err.Error(), http.StatusAccepted)
					data:="Validation du changement de mot de passe"
					SuccessOKOr404(w, req, data)
			}
		}else
		{
			logs.Error("Non correspondance de code de validation")
			id := strconv.FormatInt(u.ID, 10)
			SendEmail(req,sPtr(req.FormValue("mail")),"NotMatching",id)
			Error(w, req, "URL invalide", http.StatusUnauthorized)
			return
		}
	}
}

// Update a user password and set the group_id to "0"
func Update(w http.ResponseWriter, req *http.Request) {
	conf := router.Context(req).Env["Application"].(*application.Application).Components["Smtp"].(settings.Smtp)
	if req.Method == "POST" {
		req.ParseForm()
		// encrypt the new password
		passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.FormValue("password")), bcrypt.DefaultCost)
		if err != nil {
			panic(err)
		}

		//affecte le mail et le nouveau password hashé au user
		u := &models.User{
			Mail:     sPtr(req.FormValue("mail")),
			Password: sPtr(string(passwordHash)),
		}

		//valide la formation du mail
		errs := u.Validate()
		if len(errs) > 0 {
			logs.Error(errs)
			Error(w, req, "Vous avez une ou des erreur(s) dans le formulaire d'inscription. vérifiez votre saisie (formatage du mail par exemple)", http.StatusBadRequest)
			//Fail(w, req, "", http.StatusInternalServerError)
			return
		}

		//génération d'un code de validation
		hashCode := time.Now().UnixNano()
		code := strconv.FormatInt(hashCode, 10)
		code2, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
		//n := bytes.IndexByte(code2, 0)
		code = string(code2[:])
		code = strings.Replace(code, ".", "Z", -1)

		//génération de l'url de validation
		urlValidation := conf.Host + "/password/validation?mail="+*u.Mail+"&code="+code

		//recupération du user par le mail et affectation des différents champs
		var store = models.UserStore(getDB(req))
		err = store.First(u)
		if err != nil {
			logs.Error(err)
			Error(w, req, err.Error(), http.StatusBadRequest)
			return
		}else
		{
			u.OldgroupID = u.GroupID
			u.GroupID = 0000
			u.Validationcode = &code
			u.Password= sPtr(string(passwordHash))
		}

		//mise à jour en base du user
		err = store.Update(u)
		if err != nil {
			logs.Error(err)
			Error(w, req, err.Error(), http.StatusBadRequest)
			return
		}else{
			err = store.UpdateGroupIDtoZero(u)
			if err != nil {
				logs.Error(err)
				Error(w, req, err.Error(), http.StatusBadRequest)
				return
			}else{
				SendEmail(req,sPtr(req.FormValue("mail")),urlValidation,*u.Firstname)
			}
		}
	}

	templates := getTemplates(req)
	if err := templates["users/register"].ExecuteTemplate(w, "base", nil); err != nil {
		logs.Error(err)
	}
}

// Returns a user
func RetrieveUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(router.Context(r).Param("id"))
	if err != nil {
		logs.Debug(err)
		Fail(w, r, map[string]interface{}{"id": "not integer"}, http.StatusBadRequest)
		return
	}

	var (
		u         = models.User{ID: int64(id)}
		db        = getDB(r)
		userStore = models.UserStore(db)
	)
	if err = userStore.First(&u); err != nil {
		if err == sql.ErrNoRows {
			Fail(w, r, nil, http.StatusNotFound)
			return
		}
		logs.Error(err)
		Error(w, r, err.Error(), http.StatusInternalServerError)
		return
	}
	Success(w, r, views.User{User: &u}, http.StatusOK)
}
