package resource

import (
	"database/sql"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

// Nullable String that overrides sql.NullString
type NullString struct {
	sql.NullString
}

func (ns NullString) MarshalJSON() ([]byte, error) {
	if ns.Valid {
		return json.Marshal(ns.String)
	}
	return json.Marshal(nil)
}

func (ns *NullString) UnmarshalJSON(data []byte) error {
	var s *string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s != nil {
		ns.Valid = true
		ns.String = *s
	} else {
		ns.Valid = false
	}
	return nil
}

type Gitmoji struct {
	Emoji       string     `json:"emoji"`
	Entity      string     `json:"entity"`
	Code        string     `json:"code"`
	Description string     `json:"description"`
	DescEsp     string     `json:"descEs"`
	Name        string     `json:"name"`
	Semver      NullString `json:"semver"`
}

//go:embed gitmojis.json
var strmoji string

var gitmojis base

type base struct {
	Schema   string    `json:"$schema"`
	Gitmojis []Gitmoji `json:"gitmojis"`
}

func Gitmojis() []Gitmoji {
	return gitmojis.Gitmojis
}

func init() {
	err := readJSON(strmoji, &gitmojis)
	if err != nil {
		panic(err)
	}
}

// readJSON lee un archivo de la ruta `path` y lo asigna en la variable `dst`
func readJSON(body string, dst any) error { // {{{
	dec := json.NewDecoder(strings.NewReader(string(body)))
	dec.DisallowUnknownFields()
	err := dec.Decode(dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("el texto contiene JSON mal formado (en el carácter %d)", syntaxError.Offset)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return fmt.Errorf("el texto contiene JSON mal formado: %w", err)
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf(
					"el texto contiene un tipo JSON incorrecto para el campo `%q`",
					unmarshalTypeError.Field,
				)

			}
			return fmt.Errorf(
				"el texto contiene un tipo JSON incorrecto (en el carácter `%d`)",
				unmarshalTypeError.Offset,
			)
		case errors.Is(err, io.EOF):
			return fmt.Errorf("el texto no debe estar vacío: %w", err)
		// Si el JSON contiene un campo que no puede ser mapeado al destino entonces Decode() devolverá ahora un mensaje
		// de error con el formato "json: unknown field "<name>". Comprobamos esto, extraemos el nombre del campo del
		// error y lo interpolamos en nuestro mensaje de error personalizado. Tenga en cuenta que hay un tema abierto en
		// <https://github.com/golang/go/issues/29035> con respecto a convertir esto en un tipo de error distinto en el
		// futuro.
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("el texto contiene una clave desconocida `%s`", fieldName)
		case errors.As(err, &invalidUnmarshalError):
			return fmt.Errorf("decodificación inválida: %w", err)
		default:
			return fmt.Errorf("error general: %w", err)
		}
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return fmt.Errorf("el texto sólo debe contener un único valor JSON: %w", err)
	}

	return nil
} // }}}
