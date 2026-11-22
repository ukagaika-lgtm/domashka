package handlers

import (
	"html/template"
	"net/http"
	"path/filepath"
	"strings"

	"zadanie2/internal/models"
	"zadanie2/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	userRepo     *repository.UserRepository
	registerTmpl *template.Template
	loginTmpl    *template.Template
}

func NewAuthHandler(userRepo *repository.UserRepository, tmplDir string) (*AuthHandler, error) {
	registerTmpl, err := template.ParseFiles(
		filepath.Join(tmplDir, "layout.html"),
		filepath.Join(tmplDir, "register.html"),
	)
	if err != nil {
		return nil, err
	}

	loginTmpl, err := template.ParseFiles(
		filepath.Join(tmplDir, "layout.html"),
		filepath.Join(tmplDir, "login.html"),
	)
	if err != nil {
		return nil, err
	}

	return &AuthHandler{
		userRepo:     userRepo,
		registerTmpl: registerTmpl,
		loginTmpl:    loginTmpl,
	}, nil
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		if err := h.registerTmpl.ExecuteTemplate(w, "layout", map[string]string{
			"Title": "Регистрация",
		}); err != nil {
			http.Error(w, "Ошибка отображения страницы", http.StatusInternalServerError)
		}
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Некорректные данные", http.StatusBadRequest)
		return
	}

	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")

	if email == "" || password == "" {
		http.Error(w, "Заполни все поля", http.StatusBadRequest)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Ошибка обработки пароля", http.StatusInternalServerError)
		return
	}

	user := &models.User{
		Email:        email,
		PasswordHash: string(hash),
	}

	if err := h.userRepo.Create(user); err != nil {
		http.Error(w, "Не удалось создать пользователя. Возможно, такой email уже зарегистрирован.", http.StatusConflict)
		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		if err := h.loginTmpl.ExecuteTemplate(w, "layout", map[string]string{
			"Title": "Вход",
		}); err != nil {
			http.Error(w, "Ошибка отображения страницы", http.StatusInternalServerError)
		}
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Некорректные данные", http.StatusBadRequest)
		return
	}

	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")

	user, err := h.userRepo.GetByEmail(email)
	if err != nil {
		http.Error(w, "Неверный email или пароль", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		http.Error(w, "Неверный email или пароль", http.StatusUnauthorized)
		return
	}

	sessionID, err := createSession(user.ID)
	if err != nil {
		http.Error(w, "Ошибка создания сессии", http.StatusInternalServerError)
		return
	}

	setSessionCookie(w, sessionID)

	http.Redirect(w, r, "/expenses", http.StatusSeeOther)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err == nil {
		deleteSession(cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
