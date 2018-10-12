// Bundle of functions managing the CRUD
package controllers

import (
	"database/sql"
	"net/http"
	"net/smtp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/quorumsco/application"
	. "github.com/quorumsco/jsonapi"
	"github.com/quorumsco/logs"
	"github.com/quorumsco/oauth2/models"
	"github.com/quorumsco/oauth2/views"
	"github.com/quorumsco/router"
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
func NewRegister(w http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {

		req.ParseForm()

		code := GenerateCode()
		fromadmin := false
		fromuser := false

		//by default password = code
		passwordHash, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
		if err != nil {
			panic(err)
			return
		}

		//by default group ID = 0
		groupID, err := strconv.ParseUint("0", 10, 32)
		if err != nil {
			logs.Error(err)
			Error(w, req, err.Error(), http.StatusBadRequest)
			return
		}

		// si register user pas fait par un admin (password existe)
		logs.Debug(req.FormValue("password"))
		logs.Debug(req.FormValue("password") != "")
		if req.FormValue("password") != "" {
			passwordHash, err = bcrypt.GenerateFromPassword([]byte(req.FormValue("password")), bcrypt.DefaultCost)
			if err != nil {
				panic(err)
				return
			} else {
				fromuser = true
			}
		}

		// si register user fait par un admin (group_id existe)
		logs.Debug(req.FormValue("group_id"))
		logs.Debug(req.FormValue("group_id") != "")
		if req.FormValue("group_id") != "" {
			groupID, err = strconv.ParseUint(req.FormValue("group_id"), 10, 32)
			if err != nil {
				logs.Error(err)
				Error(w, req, err.Error(), http.StatusBadRequest)
				return
			} else {
				fromadmin = true
			}
		}

		curTime := time.Now()
		u := &models.User{
			Firstname: sPtr(req.FormValue("firstname")),
			Surname:   sPtr(req.FormValue("surname")),
			Mail:      sPtr(req.FormValue("mail")),
			Phone:     sPtr(req.FormValue("phone")),
			Role:      sPtr(req.FormValue("role")),
			Address:   sPtr(req.FormValue("address")),
			Password:  sPtr(string(passwordHash)),
			// INFO : dans ce cas Cause viens de la requête (gateway) qui est allé chercher le nom de la campagne (différent du code cause)
			//Cause:          sPtr(req.FormValue("cause")),
			Created:        &curTime,
			Validationcode: sPtr(code),
			GroupID:        uint(groupID),
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

			//StatusConflict
			//StatusConflict
			logs.Error(err)
			Error(w, req, err.Error(), http.StatusConflict)
			return
		} else {
			if fromadmin && !fromuser {
				host1 := req.FormValue("serveurWebappForChangingPassword")
				logs.Debug(req.FormValue("cause"))
				// INFO : dans ce cas Cause viens de la requête (gateway) qui est allé chercher le nom de la campagne (différent du code cause)
				urlValidation1 := host1 + "?mail=" + *u.Mail + "&code=" + code + "&new=true"
				SendEmail(req, "NewRegisterFromAdmin", sPtr(req.FormValue("mail")), urlValidation1, req.FormValue("firstname"), req.FormValue("cause"))
			} else if !fromadmin && fromuser {
				host2 := req.FormValue("serveurWebappForValidationMail")
				//génération de l'url de validation
				urlValidation2 := host2 + "?mail=" + *u.Mail + "&code=" + code
				SendEmail(req, "NewRegister", sPtr(req.FormValue("mail")), urlValidation2, req.FormValue("firstname"), "")
			} else {
				Error(w, req, "problème de distinction de type d'enregistrement : contactez le support", http.StatusNotImplemented)
			}
		}
	}
	SuccessOKOr404(w, req, "mes couilles sur la commode")
	// templates := getTemplates(req)
	// if err := templates["users/register"].ExecuteTemplate(w, "base", nil); err != nil {
	// 	logs.Error(err)
	// 	//Error(w, req, err.Error(), http.StatusInternalServerError)
	// }
}

// OLD - Creates a new user
func Register(w http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {
		req.ParseForm()

		passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.FormValue("password")), bcrypt.DefaultCost)
		if err != nil {
			panic(err)
		}

		curTime := time.Now()
		u := &models.User{
			Firstname: sPtr(req.FormValue("firstname")),
			Surname:   sPtr(req.FormValue("surname")),
			Mail:      sPtr(req.FormValue("mail")),
			Phone:     sPtr(req.FormValue("phone")),
			Address:   sPtr(req.FormValue("address")),
			Password:  sPtr(string(passwordHash)),
			Cause:     sPtr(req.FormValue("cause")),
			Created:   &curTime,
		}

		errs := u.Validate()
		if len(errs) > 0 {
			logs.Error(errs)
			Error(w, req, "Vous avez une ou des erreur(s) dans le formulaire d'inscription. vérifiez votre saisie (formatage du mail par exemple)", http.StatusBadRequest)
			//Fail(w, req, "", http.StatusInternalServerError)

		}

		var store = models.UserStore(getDB(req))
		err = store.Save(u)
		if err != nil {
			logs.Error(err)
			Error(w, req, err.Error(), http.StatusBadRequest)
			return
		} else {
			SendEmail(req, "Register", sPtr(req.FormValue("mail")), "", req.FormValue("firstname"), "")
		}
	}

	templates := getTemplates(req)
	if err := templates["users/register"].ExecuteTemplate(w, "base", nil); err != nil {
		logs.Error(err)
	}
}

func RegisterFromAdmin(w http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {
		req.ParseForm()

		passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.FormValue("password")), bcrypt.DefaultCost)
		if err != nil {
			panic(err)
		}

		ID, err := strconv.ParseUint(req.FormValue("group_id"), 10, 64)
		if err != nil {
			logs.Error(err)
			Error(w, req, err.Error(), http.StatusBadRequest)
			return
		}
		//ID = uint(ID)

		curTime := time.Now()
		u := &models.User{
			Firstname: sPtr(req.FormValue("firstname")),
			Surname:   sPtr(req.FormValue("surname")),
			Mail:      sPtr(req.FormValue("mail")),
			Role:      sPtr(req.FormValue("role")),
			Phone:     sPtr(req.FormValue("phone")),
			Address:   sPtr(req.FormValue("address")),
			Password:  sPtr(string(passwordHash)),
			GroupID:   uint(ID),
			Created:   &curTime,
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
		} else {
			SendEmail(req, "RegisterFromAdmin", sPtr(req.FormValue("mail")), "", req.FormValue("firstname"), req.FormValue("cause"))
		}
	}

	templates := getTemplates(req)
	if err := templates["users/register"].ExecuteTemplate(w, "base", nil); err != nil {
		logs.Error(err)
	}
}

func ValidPassword(w http.ResponseWriter, req *http.Request) {
	u := &models.User{
		Mail: sPtr(req.FormValue("mail")),
		// only to pass the "update" control
		Password: sPtr(req.FormValue("code")),
	}
	var store = models.UserStore(getDB(req))
	err := store.First(u)
	if err != nil {
		logs.Error(err)
		Error(w, req, err.Error(), http.StatusBadRequest)
		return
	} else {
		if *sPtr(req.FormValue("code")) == *u.Validationcode {
			//the validation code is correct
			//mise à jour en base du user
			//u.GroupID = u.OldgroupID
			//u.OldgroupID = 99999
			temp := ""
			u.Validationcode = &temp
			err = store.Update(u)
			if err != nil {
				logs.Error(err)
				Error(w, req, err.Error(), http.StatusBadRequest)
				return
			} else {
				SendEmail(req, "Confirmation", sPtr(req.FormValue("mail")), "", *u.Firstname, "")
				//Error(w, req, err.Error(), http.StatusAccepted)
				data := "Validation du changement de mot de passe"
				SuccessOKOr404(w, req, data)
			}
		} else {
			logs.Error("Non correspondance de code de validation")
			id := strconv.FormatInt(u.ID, 10)
			SendEmail(req, "NotMatching", sPtr(req.FormValue("mail")), "", id, "")
			Error(w, req, "URL invalide", http.StatusUnauthorized)
			return
		}
	}
}

func ValidUser(w http.ResponseWriter, req *http.Request) {

	u := &models.User{
		Mail: sPtr(req.FormValue("mail")),
		// only to pass the "update" control
		Password: sPtr(req.FormValue("code")),
	}
	var store = models.UserStore(getDB(req))
	err := store.First(u)
	if err != nil {
		logs.Error(err)
		Error(w, req, err.Error(), http.StatusBadRequest)
		return
	} else {
		if u.Validationcode != nil {
			codeTMP := *u.Validationcode
			if codeTMP == req.FormValue("code") {
				temp := ""
				u.Validationcode = &temp
				err = store.Update(u)
				if err != nil {
					logs.Error(err)
					Error(w, req, err.Error(), http.StatusBadRequest)
					return
				} else {
					SendEmail(req, "ConfirmationUser", sPtr(req.FormValue("mail")), "", *u.Firstname, "")
					//SendEmail(req, "ConfirmationReferent", sPtr(req.FormValue("email_referent")), "", *u.Firstname, "")
					//Error(w, req, err.Error(), http.StatusAccepted)
					data := "Validation du changement de mot de passe"
					SuccessOKOr404(w, req, data)
				}
			} else {
				logs.Error("Non correspondance de code de validation")
				id := strconv.FormatInt(u.ID, 10)
				SendEmail(req, "NotMatching", sPtr(req.FormValue("mail")), "", id, "")
				Error(w, req, "URL invalide", http.StatusUnauthorized)
				return
			}
		} else {
			logs.Error("Non correspondance de code de validation")
			id := strconv.FormatInt(u.ID, 10)
			SendEmail(req, "NotMatching", sPtr(req.FormValue("mail")), "", id, "")
			Error(w, req, "URL invalide", http.StatusUnauthorized)
			return
		}

	}
}

func ExistMail(w http.ResponseWriter, req *http.Request) {

	logs.Debug(req.FormValue("mail"))
	u := &models.User{
		Mail: sPtr(req.FormValue("mail")),

		// only to pass the "Store" control
		Password: sPtr(req.FormValue("mail")),
	}
	var store = models.UserStore(getDB(req))
	err := store.First(u)
	if err != nil {
		data := err
		SuccessOKOr404(w, req, data)
		// logs.Error(err)
		// Error(w, req, err.Error(), http.StatusUnauthorized)
		return
	} else {
		if *sPtr(req.FormValue("mail")) == *u.Mail {

			logs.Error("Correspondance de mail")
			Error(w, req, "Correspondance de mail", http.StatusUnauthorized)
			return

		} else {
			data := "Non correspondance de mail"
			SuccessOKOr404(w, req, data)

		}
	}
}

func GenerateCode() string {
	//generate Code ---------------------------
	hashCode := time.Now().UnixNano()
	code := strconv.FormatInt(hashCode, 10)
	code2, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
	if err != nil {
		logs.Error(err)
		panic(err)
	}
	//n := bytes.IndexByte(code2, 0)
	code = string(code2[:])
	code = strings.Replace(code, ".", "Z", -1)
	return code
}

func SendRequestToReferent(w http.ResponseWriter, req *http.Request) {
	conf := router.Context(req).Env["Application"].(*application.Application).Components["Smtp"].(settings.Smtp)
	req.ParseForm()

	//generate Code ---------------------------
	hashCode := time.Now().UnixNano()
	code := strconv.FormatInt(hashCode, 10)
	code2, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
	if err != nil {
		logs.Error(err)
		panic(err)
	}
	//n := bytes.IndexByte(code2, 0)
	code = string(code2[:])
	code = strings.Replace(code, ".", "Z", -1)

	//génération de l'url de validation
	urlValidation := conf.Host + "/user/validation?mail=" + req.FormValue("mail") + "&code=" + code + "&email_referent=" + req.FormValue("email_referent")

	prenom_nom_mail := strings.Title(req.FormValue("firstname")) + " " + strings.Title(req.FormValue("surname")) + " (" + req.FormValue("mail") + ")"
	req.Form.Add("validationcode", code)

	Update_code(w, req)
	SendEmail(req, "ValidationUser", sPtr(req.FormValue("email_referent")), urlValidation, prenom_nom_mail, "")
	data := "ValidationUser"
	SuccessOKOr404(w, req, data)
}

func Update_code(w http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {
		req.ParseForm()

		//string to uint-----------------
		// temp, err := strconv.ParseUint(req.FormValue("group_id"), 10, 0)
		// if err != nil {
		// 	logs.Debug(err)
		// 	Fail(w, req, map[string]interface{}{"group_id": "not integer"}, http.StatusBadRequest)
		// 	return
		// }
		//groupid := uint(temp)

		u := &models.User{
			Mail: sPtr(req.FormValue("mail")),
			//OldgroupID:     groupid,
			//GroupID:        0000,
			Validationcode: sPtr(req.FormValue("validationcode")),
		}
		//mise à jour en base du user
		var store = models.UserStore(getDB(req))
		err := store.Update(u)
		if err != nil {
			logs.Error(err)
			Error(w, req, err.Error(), http.StatusBadRequest)
			return
		}
	}
}

func SendMailWithUrlForPasswordChange(w http.ResponseWriter, req *http.Request) {
	logs.Debug("SendMailWithUrlForPasswordChange")
	//conf := router.Context(req).Env["Application"].(*application.Application).Components["Smtp"].(settings.Smtp)

	if req.Method == "POST" {
		req.ParseForm()
		// encrypt the new password
		// passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.FormValue("password")), bcrypt.DefaultCost)
		// if err != nil {
		// 	panic(err)
		// }
		host := req.FormValue("Referer")

		//génération d'un code de validation
		hashCode := time.Now().UnixNano()
		code := strconv.FormatInt(hashCode, 10)
		code2, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
		//n := bytes.IndexByte(code2, 0)
		code = string(code2[:])
		code = strings.Replace(code, ".", "Z", -1)

		passwordHash, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)

		//affecte le mail et le nouveau password hashé au user
		u := &models.User{
			Mail:     sPtr(req.FormValue("mail")),
			Password: sPtr(string(passwordHash)),
		}

		//valide la formation du mail
		errs := u.Validate()
		if len(errs) > 0 {
			logs.Error(errs)
			Error(w, req, "Vous avez une ou des erreur(s) dans le mail saisi. vérifiez votre saisie.", http.StatusBadRequest)
			//Fail(w, req, "", http.StatusInternalServerError)
			return
		}

		//génération de l'url de validation
		urlValidation := host + "?mail=" + *u.Mail + "&code=" + code

		//recupération du user par le mail et affectation des différents champs
		var store = models.UserStore(getDB(req))
		err = store.First(u)
		if err != nil {
			logs.Error(err)
			Error(w, req, err.Error(), http.StatusBadRequest)
			return
		} else {
			//u.OldgroupID = u.GroupID
			//u.GroupID = 0000
			u.Validationcode = &code
			//u.Password = sPtr(string(passwordHash))
		}

		//mise à jour en base du user
		err = store.Update(u)
		if err != nil {
			logs.Error(err)
			Error(w, req, err.Error(), http.StatusBadRequest)
			return
		} else {
			//err = store.UpdateGroupIDtoZero(u)
			SendEmail(req, "ValidationPassword", sPtr(req.FormValue("mail")), urlValidation, *u.Firstname, "")

		}
	}

	templates := getTemplates(req)
	if err := templates["users/register"].ExecuteTemplate(w, "base", nil); err != nil {
		logs.Error(err)
	}
}

func NewSavePassword(w http.ResponseWriter, req *http.Request) {

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

	var store = models.UserStore(getDB(req))
	err = store.First(u)
	if err != nil {
		logs.Error(err)
		Error(w, req, err.Error(), http.StatusBadRequest)
		return
	} else {
		if *sPtr(req.FormValue("code")) == *u.Validationcode {
			//the validation code is correct
			//mise à jour en base du user
			//u.GroupID = u.OldgroupID
			//u.OldgroupID = 99999
			temp := ""
			u.Validationcode = &temp
			u.Password = sPtr(string(passwordHash))
			err = store.Update(u)
			if err != nil {
				logs.Error(err)
				Error(w, req, err.Error(), http.StatusBadRequest)
				return
			} else {
				SendEmail(req, "Confirmation", sPtr(req.FormValue("mail")), "", *u.Firstname, "")
				//Error(w, req, err.Error(), http.StatusAccepted)
				data := "Validation du changement de mot de passe"
				SuccessOKOr404(w, req, data)
			}
		} else {
			logs.Error("Non correspondance de code de validation")
			id := strconv.FormatInt(u.ID, 10)
			SendEmail(req, "NotMatching", sPtr(req.FormValue("mail")), "", id, "")
			Error(w, req, "URL invalide", http.StatusUnauthorized)
			return
		}
	}
}

// NEW - Update a user password and set the group_id to "0"
func UpdatePassword(w http.ResponseWriter, req *http.Request) {
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
		urlValidation := conf.Host + "/password/validation?mail=" + *u.Mail + "&code=" + code

		//recupération du user par le mail et affectation des différents champs
		var store = models.UserStore(getDB(req))
		err = store.First(u)
		if err != nil {
			logs.Error(err)
			Error(w, req, err.Error(), http.StatusBadRequest)
			return
		} else {
			//u.OldgroupID = u.GroupID
			//u.GroupID = 0000
			u.Validationcode = &code
			u.Password = sPtr(string(passwordHash))
		}

		//mise à jour en base du user
		err = store.Update(u)
		if err != nil {
			logs.Error(err)
			Error(w, req, err.Error(), http.StatusBadRequest)
			return
		} else {
			SendEmail(req, "ValidationPasswordOld", sPtr(req.FormValue("mail")), urlValidation, *u.Firstname, "")
		}
	}

	templates := getTemplates(req)
	if err := templates["users/register"].ExecuteTemplate(w, "base", nil); err != nil {
		logs.Error(err)
	}
}

func Update(w http.ResponseWriter, req *http.Request) {

	if req.Method == "PATCH" {
		req.ParseForm()

		id, err := strconv.Atoi(req.FormValue("id"))
		if err != nil {
			logs.Debug(err)
			Fail(w, req, map[string]interface{}{"id": "not integer"}, http.StatusBadRequest)
			return
		}
		//affecte le mail et le nouveau password hashé au user
		u := &models.User{
			Mail:      sPtr(req.FormValue("mail")),
			Surname:   sPtr(req.FormValue("surname")),
			Firstname: sPtr(req.FormValue("firstname")),
			Role:      sPtr(req.FormValue("role")),
			Phone:     sPtr(req.FormValue("phone")),
			Address:   sPtr(req.FormValue("address")),
			Password:  sPtr(""),
			ID:        int64(id),
		}

		//valide la formation du mail
		errs := u.ValidateEmail()
		if len(errs) > 0 {
			logs.Error(errs)
			Error(w, req, "vérifiez votre saisie (formatage du mail)", http.StatusBadRequest)
			//Fail(w, req, "", http.StatusInternalServerError)
			return
		}

		//recupération du user par le mail et affectation des différents champs
		var store = models.UserStore(getDB(req))

		//mise à jour en base du user
		err = store.Update(u)
		if err != nil {
			logs.Error(err)
			Error(w, req, err.Error(), http.StatusBadRequest)
			return
		}
		//Success(w, req, views.User{User: u}, http.StatusOK)
	}
}

func Delete(w http.ResponseWriter, req *http.Request) {

	if req.Method == "DELETE" {
		req.ParseForm()

		//id, err := strconv.Atoi(req.FormValue("id"))
		id, err := strconv.Atoi(router.Context(req).Param("id"))
		if err != nil {
			logs.Debug(err)
			Fail(w, req, map[string]interface{}{"id": "not integer"}, http.StatusBadRequest)
			return
		}

		u := &models.User{
			ID: int64(id),
		}

		//recupération du user par le mail et affectation des différents champs
		var store = models.UserStore(getDB(req))

		//mise à jour en base du user
		err = store.Delete(u)
		if err != nil {
			logs.Error(err)
			Error(w, req, err.Error(), http.StatusBadRequest)
			return
		}
		//Success(w, req, views.User{User: u}, http.StatusOK)
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
	if err := userStore.First(&u); err != nil {
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

func RetrieveAllUsersByGroup(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.Atoi(router.Context(r).Param("id"))
	if err != nil {
		logs.Debug(err)
		Fail(w, r, map[string]interface{}{"id": "not integer"}, http.StatusBadRequest)
		return
	}
	r.ParseForm()
	limit, err := strconv.Atoi(r.FormValue("limit"))
	if err != nil {
		logs.Debug(err)
		limit = -1
	}

	offset, err := strconv.Atoi(r.FormValue("offset"))
	if err != nil {
		logs.Debug(err)
		offset = -1
	}

	sort := r.FormValue("sort")
	if sort == "false" {
		sort = "desc"
	} else {
		sort = "asc"
	}

	var (
		db        = getDB(r)
		userStore = models.UserStore(db)
		users2    = models.UserReply{}
		user      = models.User{GroupID: uint(id)}
		//users2.User = models.User{GroupID:id}
	)

	users2.User = &user

	if err := userStore.Find(&users2, limit, offset, sort); err != nil {
		if err == sql.ErrNoRows {
			Fail(w, r, nil, http.StatusNotFound)
			return
		}
		logs.Error(err)
		Error(w, r, err.Error(), http.StatusInternalServerError)
		return
	}
	Success(w, r, views.Users{Users: users2.Users, Count: users2.Count}, http.StatusOK)
}

/*
func RetrieveAllUsersByTeam(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.Atoi(router.Context(r).Param("id"))
	if err != nil {
		logs.Debug(err)
		Fail(w, r, map[string]interface{}{"id": "not integer"}, http.StatusBadRequest)
		return
	}
	r.ParseForm()
	limit, err := strconv.Atoi(r.FormValue("limit"))
	if err != nil {
		logs.Debug(err)
		limit = -1
	}

	offset, err := strconv.Atoi(r.FormValue("offset"))
	if err != nil {
		logs.Debug(err)
		offset = -1
	}

	sort := r.FormValue("sort")
	if sort == "false" {
		sort = "desc"
	} else {
		sort = "asc"
	}

	var (
		db        = getDB(r)
		userStore = models.UserStore(db)
		users2    = models.UserReply{}
		team      = models.Team{ID_team: uint(id)}
		//users2.User = models.User{GroupID:id}
	)
	users2.Team = &team
	//users2.Teams[0] = team

	if err := userStore.FindByTeam(&users2, limit, offset, sort); err != nil {
		if err == sql.ErrNoRows {
			Fail(w, r, nil, http.StatusNotFound)
			return
		}
		logs.Error(err)
		Error(w, r, err.Error(), http.StatusInternalServerError)
		return
	}
	Success(w, r, views.Users{Users: users2.Users, Count: users2.Count}, http.StatusOK)
}
*/
func SendEmail(r *http.Request, type_mail string, to *string, url string, prenom string, campagne string) {
	//var settings string
	conf := router.Context(r).Env["Application"].(*application.Application).Components["Smtp"].(settings.Smtp)
	// Set up authentication information.
	auth := smtp.PlainAuth("", conf.User, conf.Password, conf.Smtpserver)

	// Connect to the server, authenticate, set the sender and recipient,
	// and send the email all in one step.
	//to := []string{"jbd@quorumapp.co"}

	var msg []byte
	if type_mail == "Confirmation" {
		msg = []byte("To: " + *to + "\r\n" +
			"Subject: QUORUM | Confirmation de changement de votre mot de passe !\r\n" +
			"\r\n" +
			"Bonjour " + prenom + ",\r" +
			"Votre mot de passe a été modifié avec succès. Vous pouvez dès à present vous connecter aux applications QUORUM et reprendre la mobilisation !\r" +
			"Attention ! Si vous n’êtes pas l’auteur de la demande de changement de mot de passe,  merci de nous contacter au plus vite par mail support@quorum.co ou directement au 01 79 73 40 04.\r" +
			"\r\rL'équipe Quorum\rMobilisons, sans limites.\rteam@quorum.co\n")
	} else if type_mail == "ConfirmationReferent" {
		msg = []byte("To: " + *to + "\r\n" +
			"Subject: QUORUM | Confirmation de création de compte !\r\n" +
			"\r\n" +
			"Bonjour,\r" +
			"le compte de " + prenom + " a été créé avec succès.\r\r" +
			"Attention ! Si vous n’êtes pas l’auteur de la validation du compte,  merci de nous contacter au plus vite par mail support@quorum.co ou directement au 01 79 73 40 04.\r" +
			"\r\rL'équipe Quorum\rMobilisons, sans limites.\rteam@quorum.co\n")
	} else if type_mail == "ConfirmationUser" {
		msg = []byte("To: " + *to + "\r\n" +
			"Subject: QUORUM | Confirmation de création de votre compte sur Quorum !\r\n" +
			"\r\n" +
			"Bonjour " + prenom + ",\r" +
			"Votre compte a été créé avec succès! Vous pouvez dès à present vous connecter aux applications QUORUM et reprendre la mobilisation !\r" +
			"Attention ! Si vous n’êtes pas l’auteur de la demande de compte,  merci de nous contacter au plus vite par mail team@quorum.co ou directement au 01 79 73 40 04.\r" +
			"\r\rL'équipe Quorum\rMobilisons, sans limites.\team@quorum.co\n")
	} else if type_mail == "NotMatching" {
		*to = conf.User
		msg = []byte("To: " + *to + "\r\n" +
			"Subject: Not matching CODE\r\n" +
			"\r\n" +
			"Tentative de matching d'un mauvais code. User ID:" +
			prenom +
			"\r\n")
	} else if type_mail == "Register" {
		msg = []byte("To: " + *to + "\r\n" +
			"Subject: demande de compte en cours\r\n" +
			"\r\n" +
			"Bonjour " + prenom + ",\r" +
			"Votre compte sera activé dès que votre référent l'aura fait.\r" +
			"l'équipe QUORUM" +
			"\r\n")
	} else if type_mail == "NewRegister" {
		msg = []byte("To: " + *to + "\r\n" +
			"Subject: demande de compte\r\n" +
			"\r\n" +
			"Bonjour " + prenom + ",\r" +
			"Merci de cliquer sur le lien ci dessous afin que votre compte soit activé.\r" +
			url + "\r" +
			"l'équipe QUORUM" +
			"\r\n")
	} else if type_mail == "RegisterFromAdmin" {
		msg = []byte("To: " + *to + "\r\n" +
			"Subject: demande de compte\r\n" +
			"\r\n" +
			"Bravo " + prenom + "!\r" +
			"Vous faites maintenant partie de la campagne de mobilisation '" + campagne + "'.\r" +
			"Afin de pouvoir accéder à votre application, merci d'initialiser votre mot de passe via 'mot de passe oublié' sur votre écran d'authentification.\r" +
			"l'équipe QUORUM" +
			"\r\n")
	} else if type_mail == "NewRegisterFromAdmin" {
		msg = []byte("To: " + *to + "\r\n" +
			"Subject: demande de compte\r\n" +
			"\r\n" +
			"Bravo " + prenom + "!\r" +
			"Vous faites maintenant partie de la campagne de mobilisation '" + campagne + "'.\r" +
			"Afin de pouvoir accéder à votre application, merci de cliquer sur le lien ci dessous afin d'initialiser votre mot de passe.\r" +
			url + "\r" +
			"l'équipe QUORUM" +
			"\r\n")
	} else if type_mail == "ValidationUser" {
		msg = []byte("To: " + *to + "\r\n" +
			"Subject: demande de validation de compte\r\n" +
			"\r\n" +
			"Bonjour,\r" +
			prenom + " vient de faire une demande de compte pour la campagne.\r" +
			"Pour valider cette demande, veuillez cliquer sur le lien ci dessous:\r" +
			url +
			"\rA très vite," +
			"\r\rL'équipe Quorum\rteam@quorum.co\n")
	} else if type_mail == "ValidationPasswordOld" {
		msg = []byte("To: " + *to + "\r\n" +
			"Subject: QUORUM | Validez le changement de votre mot de passe !\r\n" +
			"\r\n" +
			"Bonjour " + prenom + ",\r" +
			"Vous venez de faire une demande de changement de mot de passe.\rPour valider cette demande, veuillez cliquer sur le lien ci dessous:\r" +
			url +
			"\rAttention ! Votre mot de passe changera seulement si vous cliquez sur le lien." +
			"\rSi vous n’avez pas fait de demande de changement de mot de passe, merci de ne pas cliquer sur le lien." +
			"\rA très vite," +
			"\r\rL'équipe Quorum\rteam@quorum.co\n")
	} else if type_mail == "ValidationPassword" {
		msg = []byte("To: " + *to + "\r\n" +
			"Subject: QUORUM | Demande de changement de mot de passe !\r\n" +
			"\r\n" +
			"Bonjour " + prenom + ",\r" +
			"Vous venez de faire une demande de changement de mot de passe.\rPour valider cette demande, veuillez cliquer sur le lien ci dessous:\r" +
			url +
			"\rSi vous n’avez pas fait de demande de changement de mot de passe, merci de ne pas cliquer sur le lien." +
			"\rA très vite," +
			"\r\rL'équipe Quorum\rteam@quorum.co\n")
	}
	toBis := []string{*to}
	err := smtp.SendMail(conf.Smtpserver+":"+conf.Port, auth, conf.User, toBis, msg)
	if err != nil {
		logs.Error(err)
	}
}
