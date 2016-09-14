package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
	"unicode"
)

var (
	f_pkg = flag.String("p", "main", "package name for generated file")
	f_out = flag.String("o", "-", "output file")
)

var (
	reDefineIntConst = regexp.MustCompile(`#define\s+([A-Za-z_]+)\s+(\(?\d+(?:U?LL)?(?:\s*<<\s*\d+)?\)?)`)
)

func constName(s string) string {
	return s
}

func process(w io.Writer, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	r := bufio.NewReader(file)

	var (
		comment            = false
		firstComment       = true
		firstLineInComment = false
	)

	nl := true
	defer fmt.Fprintln(w, ")")
	for {
		line, err := r.ReadBytes('\n')
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			if !nl {
				nl = true
				w.Write([]byte("\n"))
			}
			continue
		}
		nl = false

		if bytes.HasPrefix(line, []byte("/*")) {
			comment = true
			firstLineInComment = true
			line = bytes.TrimPrefix(line, []byte("/*"))
		}
		if comment {
			ends := bytes.HasSuffix(line, []byte("*/"))
			if ends {
				comment = false
				line = bytes.TrimSuffix(line, []byte("*/"))
			}
			line = bytes.TrimLeft(line, " \t*")
			if len(line) > 0 {
				if !firstComment {
					w.Write([]byte("\t"))
				}
				w.Write([]byte("// "))
				if firstLineInComment {
					line[0] = byte(unicode.ToUpper(rune(line[0])))
				}
				line = bytes.Replace(line, []byte("  "), []byte(" "), -1)
				w.Write(line)
				w.Write([]byte("\n"))
				firstLineInComment = false
			}
			if ends && firstComment {
				firstComment = false
				fmt.Fprint(w, "\nconst (\n")
				nl = true
			}
			firstLineInComment = firstLineInComment && !ends
			continue
		}
		if bytes.HasPrefix(line, []byte("#define")) {
			sub := reDefineIntConst.FindStringSubmatch(string(line))
			if len(sub) > 0 {
				name, val := sub[1], sub[2]
				val = strings.Replace(val, "ULL", "", -1)
				fmt.Fprintf(w, "\t%s = %s\n", constName(name), val)
				continue
			}
		}
	}
}

func main() {
	flag.Parse()
	var w io.Writer = os.Stdout
	if path := *f_out; path != "" && path != "-" {
		file, err := os.Create(path)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		w = file
	}

	fmt.Fprintf(w, "package %s\n", *f_pkg)
	for _, path := range flag.Args() {
		if err := process(w, path); err != nil {
			log.Fatal(err)
		}
	}
}
