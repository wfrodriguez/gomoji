package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"unicode/utf8"

	"github.com/wfrodriguez/console/input"
	"github.com/wfrodriguez/console/output"
	"github.com/wfrodriguez/gomoji/internal/resource"
)

var scopes = [][]string{
	{
		"add - añade uno o varios archivos.",
		"del - elimina uno o varios archivos.",
		"upd - actualiza uno o varios archivos.",
		"feat - cuando se añade una nueva funcionalidad.",
		"fix - cuando se arregla un error.",
		"chore - tareas rutinarias que no sean específicas de una feature o un error.",
		"test - si añadimos o arreglamos tests.",
		"docs - cuando solo se modifica documentación.",
		"build - cuando el cambio afecta al compilado del proyecto.",
		"ci - el cambio afecta a ficheros de configuración y scripts relacionados con la integración continua.",
		"style - cambios de legibilidad o formateo de código que no afecta a funcionalidad.",
		"refac -cambio de código que no corrige errores ni añade funcionalidad, pero mejora el código.",
		"perf - usado para mejoras de rendimiento.",
		"revert - si el commit revierte un commit anterior. Debería indicarse el hash del commit que se revierte.",
	},
	{"add", "del", "upd", "feat", "fix", "chore", "test", "docs", "build", "ci", "style", "refactor", "perf", "revert"},
}

func git(args ...string) (stdout string, stderr string, exitCode int) {
	defaultFailedCode := 1
	var outbuf, errbuf bytes.Buffer
	cmd := exec.Command("git", args...)
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf

	err := cmd.Run()
	stdout = outbuf.String()
	stderr = errbuf.String()

	if err != nil {
		// try to get the exit code
		if exitError, ok := err.(*exec.ExitError); ok {
			ws := exitError.Sys().(syscall.WaitStatus)
			exitCode = ws.ExitStatus()
		} else {
			log.Printf("Could not get exit code for failed program: git, %v", args)
			exitCode = defaultFailedCode
			if stderr == "" {
				stderr = err.Error()
			}
		}
	} else {
		// success, exitCode should be 0 if go is ok
		ws := cmd.ProcessState.Sys().(syscall.WaitStatus)
		exitCode = ws.ExitStatus()
	}

	return
}

func getIntents() [][]string {
	items := resource.Gitmojis()
	labels := make([]string, len(items))
	codes := make([]string, len(items))
	for i, item := range items {
		labels[i] = fmt.Sprintf("%s - %s", item.Emoji, item.DescEsp)
		codes[i] = item.Code
	}
	return [][]string{labels, codes}
}

func perror(err error) {
	if err != nil {
		output.Error([]string{
			"Ha ocurrido un error:",
			"    " + err.Error(),
			"Por favor intentalo de nuevo",
		}...)

		os.Exit(1)
	}
}

func replaceAtIndex(in string, r rune, i int) string {
	out := []rune(in)
	out[i] = r
	return string(out)
}

func drawRuler(s int) string {
	sbn := strings.Builder{}
	sbl := strings.Builder{}
	for i := 0; i < s; i = i + 10 {
		sbn.WriteString(fmt.Sprintf("%10d", i+10))
		sbl.WriteString("┄┄┄┄┄┄┄┄┄┼")
	}

	strn := sbn.String()
	strn = replaceAtIndex(strn, '0', 0)
	strl := sbl.String()
	strl = replaceAtIndex(strl, '├', 0)
	strl = replaceAtIndex(strl, '┤', utf8.RuneCountInString(strl)-1)

	return fmt.Sprintf("╔══╗%s\n╚══╝%s", strn, strl)
}

func main() {
	intents := getIntents()

	idxe, _, erre := input.Select(
		input.WithSelectLabel("Emoji"),
		input.WithItems(intents[0]),
		input.WithSize(10),
	)
	perror(erre)

	idxs, _, errs := input.Select(
		input.WithSelectLabel("Scope"),
		input.WithItems(scopes[0]),
		input.WithSize(8),
	)
	perror(errs)

	fmt.Println()
	fmt.Println(output.Sprintf("<input>Asunto:</>"))
	fmt.Println(drawRuler(50))
	subject, errc := input.Ask(
		input.WithLabel(""),
		input.WithValidate(func(s string) error {
			lc := utf8.RuneCountInString(s)
			if lc < 3 || lc > 50 {
				return fmt.Errorf("el asunto debe tener entre 3 y 50 caracteres")
			}
			return nil
		}),
	)
	perror(errc)

	var msg = make([]string, 0)
	var err error
	if input.Confirm("¿Desea incluir un mensaje más detallado?", 2) {
		msg, err = input.TextArea("Mensaje", 80)
		perror(err)
		tmp := make([]string, len(msg))
		for i, m := range msg {
			tmp[i] = strings.TrimSpace(m)
		}
		msg = tmp
	}

	fmt.Println()
	preview := []string{
		fmt.Sprintf("%s [%s] %s", intents[1][idxe], scopes[1][idxs], subject),
	}
	if len(msg) > 0 {
		preview = append(preview, "")
		preview = append(preview, msg...)
	}
	output.PrintSection("darkseagreen", "Previsualización", preview...)

	if input.Confirm("¿Crear commit?", 2) {
		cmd := []string{"commit", "-m", fmt.Sprintf("%s [%s] %s", intents[1][idxe], scopes[1][idxs], subject)}
		for _, c := range msg {
			cmd = append(cmd, "-m", c)
		}
		out, strerr, sc := git(cmd...)
		if sc != 0 {
			fmt.Println(output.Sprintf("<indianred>%s</>", strerr))
			return
		}
		fmt.Println(output.Sprintf("<darkseagreen>%s</>\nCommit creado", out))
	}

}
