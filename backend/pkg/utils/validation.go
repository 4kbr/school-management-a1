package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/go-playground/validator/v10"
)

// validate adalah instance validator global (singleton).
// Best practice: jangan buat validator.New() baru tiap request — reuse satu instance.
var validate = validator.New()

// ValidationError menggambarkan satu kegagalan validasi pada satu field.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidateStruct memvalidasi sebuah struct (berdasarkan tag `validate`).
// Mengembalikan daftar error per field, atau nil kalau valid.
func ValidateStruct(s interface{}) []ValidationError {
	// 1. Jalankan validasi struct
	err := validate.Struct(s)
	// 2. Kalau tidak ada error → valid
	if err == nil {
		return nil
	}
	// 3. Terjemahkan error validator ke daftar pesan ramah
	return translateErrors(err)
}

// ValidateSlice memvalidasi setiap elemen sebuah slice of structs (untuk batch).
func ValidateSlice[T any](items []T) []ValidationError {
	// 1. Loop tiap item
	for i, item := range items {
		// 2. Validasi item; kalau ada error, tandai nomor indeksnya
		if errs := validate.Struct(item); errs != nil {
			translated := translateErrors(errs)
			// 3. Prefix field dengan indeks biar client tahu item yang mana
			prefixed := make([]ValidationError, len(translated))
			for j, e := range translated {
				prefixed[j] = ValidationError{
					Field:   fmt.Sprintf("[%d].%s", i, e.Field),
					Message: e.Message,
				}
			}
			return prefixed
		}
	}
	// 4. Semua valid
	return nil
}

// DecodeJSON mendecode body JSON ke dst.
// Kalau tipe field tidak cocok (mis. class dikasih angka padahal string),
// mengembalikan ValidationError terstruktur alih-alih pesan teks polos.
func DecodeJSON(body io.Reader, dst any) []ValidationError {
	// 1. Decode body
	err := json.NewDecoder(body).Decode(dst)
	// 2. Decode sukses → tidak ada error validasi
	if err == nil {
		return nil
	}

	// 3. Deteksi error tipe mismatch (field ada, tapi tipenya salah)
	var typeErr *json.UnmarshalTypeError
	if errors.As(err, &typeErr) {
		return []ValidationError{{
			Field:   toSnake(typeErr.Field),
			Message: fmt.Sprintf("must be %s, got %s", typeErr.Type, typeErr.Value),
		}}
	}

	// 4. Error JSON lainnya (malformed, dll.)
	return []ValidationError{{
		Field:   "request",
		Message: "invalid request body",
	}}
}

// translateErrors mengubah validator.ValidationErrors menjadi []ValidationError.
func translateErrors(err error) []ValidationError {
	// 1. Konversi error ke tipe ValidationErrors
	verrs, ok := err.(validator.ValidationErrors)
	if !ok {
		return []ValidationError{{Field: "request", Message: "invalid input"}}
	}

	// 2. Siapkan slice hasil
	results := make([]ValidationError, 0, len(verrs))
	for _, ve := range verrs {
		results = append(results, ValidationError{
			Field:   toSnake(ve.Field()),
			Message: messageFor(ve),
		})
	}
	return results
}

// messageFor memetakan tag validasi ke pesan yang ramah.
func messageFor(ve validator.FieldError) string {
	switch ve.Tag() {
	case "required":
		return "field is required"
	case "email":
		return "must be a valid email address"
	case "min":
		return fmt.Sprintf("must be at least %s", ve.Param())
	case "max":
		return fmt.Sprintf("must be at most %s", ve.Param())
	default:
		return "is invalid"
	}
}

// toSnake mengubah nama field Go ke snake_case (mis. FirstName -> first_name,
// ID -> id). Tidak menambah underscore sebelum huruf besar yang didahului
// huruf besar (akronim), jadi "ID" jadi "id", bukan "i_d".
func toSnake(s string) string {
	var b strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' && s[i-1] >= 'a' && s[i-1] <= 'z' {
			b.WriteByte('_')
		}
		b.WriteRune(r)
	}
	return strings.ToLower(b.String())
}
