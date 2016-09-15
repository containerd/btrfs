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
	"strconv"
	"strings"
	"unicode"
)

var (
	f_pkg      = flag.String("p", "main", "package name for generated file")
	f_out      = flag.String("o", "-", "output file")
	f_unexport = flag.Bool("u", true, "make all definitions unexported")
	f_goname   = flag.Bool("g", true, "rename symbols to follow Go conventions")
	f_trim     = flag.String("t", "", "prefix to trim from names")
)

var (
	reDefineIntConst = regexp.MustCompile(`#define\s+([A-Za-z_]+)\s+(\(?-?\d+(?:U?LL)?(?:\s*<<\s*\d+)?\)?)`)
	reNegULL         = regexp.MustCompile(`-(\d+)ULL`)
)

func constName(s string) string {
	s = strings.TrimPrefix(s, *f_trim)
	if *f_goname {
		buf := bytes.NewBuffer(nil)
		buf.Grow(len(s))
		up := !*f_unexport
		for _, r := range s {
			if r == '_' {
				up = true
				continue
			}
			if up {
				up = false
				r = unicode.ToUpper(r)
			} else {
				r = unicode.ToLower(r)
			}
			buf.WriteRune(r)
		}
		s = buf.String()
	} else if *f_unexport {
		s = "_" + s
	}
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
	fmt.Fprint(w, "// This code was auto-generated; DO NOT EDIT!\n\n")
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
				if sub := reNegULL.FindAllStringSubmatch(val, -1); len(sub) > 0 {
					for _, s := range sub {
						v, err := strconv.ParseInt(s[1], 10, 64)
						if err != nil {
							panic(err)
						}
						val = strings.Replace(val, s[0], fmt.Sprintf("0x%x /* -%s */", uint64(-v), s[1]), -1)
					}
				}
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
