/*
   Copyright The containerd Authors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

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
	f_pkg      = flag.String("p", "main", "package name for generated file")
	f_out      = flag.String("o", "-", "output file")
	f_unexport = flag.Bool("u", true, "make all definitions unexported")
	f_goname   = flag.Bool("g", true, "rename symbols to follow Go conventions")
	f_trim     = flag.String("t", "", "prefix to trim from names")

	f_constSuf  = flag.String("cs", "", "comma-separated list of constant suffixes to create typed constants")
	f_constPref = flag.String("cp", "", "comma-separated list of constant prefixes to create typed constants")
)

var (
	reDefineIntConst = regexp.MustCompile(`#define\s+([A-Za-z_][A-Za-z\d_]*)\s+(\(?-?\d+(?:U?LL)?(?:\s*<<\s*\d+)?\)?)`)
	reNegULL         = regexp.MustCompile(`-(\d+)ULL`)
)

var (
	constTypes []constType
)

type constType struct {
	Name   string
	Type   string
	Suffix string
	Prefix string
}

func constName(s string) string {
	s = strings.TrimPrefix(s, *f_trim)
	typ := ""
	for _, t := range constTypes {
		if t.Suffix != "" && strings.HasSuffix(s, t.Suffix) {
			//s = strings.TrimSuffix(s, t.Suffix)
			typ = t.Name
			break
		} else if t.Prefix != "" && strings.HasPrefix(s, t.Prefix) {
			typ = t.Name
			break
		}
	}
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
	if typ != "" {
		s += " " + typ
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
						val = strings.Replace(val, s[0], fmt.Sprintf("(1<<64 - %s)", s[1]), -1)
					}
				}
				val = strings.Replace(val, "ULL", "", -1)
				fmt.Fprintf(w, "\t%s = %s\n", constName(name), val)
				continue
			}
		}
	}
}

func regConstTypes(str string, fnc func(*constType, string)) {
	for _, s := range strings.Split(str, ",") {
		kv := strings.Split(s, "=")
		if len(kv) != 2 {
			continue
		}
		st := strings.Split(kv[0], ":")
		typ := "int"
		if len(st) > 1 {
			typ = st[1]
		}
		t := constType{Name: st[0], Type: typ}
		fnc(&t, kv[1])
		constTypes = append(constTypes, t)
	}
}

func main() {
	flag.Parse()
	if suf := *f_constSuf; suf != "" {
		regConstTypes(suf, func(t *constType, v string) { t.Suffix = v })
	}
	if pref := *f_constPref; pref != "" {
		regConstTypes(pref, func(t *constType, v string) { t.Prefix = v })
	}
	var w io.Writer = os.Stdout
	if path := *f_out; path != "" && path != "-" {
		file, err := os.Create(path)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		w = file
	}

	fmt.Fprintf(w, "package %s\n\n", *f_pkg)
	fmt.Fprint(w, "// This code was auto-generated; DO NOT EDIT!\n\n")
	for _, t := range constTypes {
		fmt.Fprintf(w, "type %s %s\n\n", t.Name, t.Type)
	}
	for _, path := range flag.Args() {
		if err := process(w, path); err != nil {
			log.Fatal(err)
		}
	}
}
