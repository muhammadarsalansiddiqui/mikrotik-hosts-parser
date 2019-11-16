//+build ignore

// Source: <https://github.com/wso2/product-apim-tooling/tree/master/import-export-cli/box>, <https://nuancesprog.ru/p/4894/amp/>

package main

import (
	"bytes"
	"fmt"
	"go/format"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

const in_dir = "data"
const resources_out = "data.go"

var packageTemplate = template.Must(template.New("").Funcs(map[string]interface{}{"conv": FormatByteSlice}).Parse(`// Code generated by go generate; DO NOT EDIT.
// generated using files from resources directory
//
// !!! DO NOT COMMIT this file !!!

package resources

func init() {
	{{- range $name, $file := . }}
    	Resources.Add("{{ $name }}", []byte{ {{ conv $file }} })
	{{- end }}
}
`))

func FormatByteSlice(sl []byte) string {
	builder := strings.Builder{}
	for _, v := range sl {
		builder.WriteString(fmt.Sprintf("%d,", int(v)))
	}
	return builder.String()
}

func main() {
	log.SetFlags(log.Lshortfile | log.Lmicroseconds)
	log.Printf("Packing resources into [%s] file", resources_out)

	if _, err := os.Stat(in_dir); os.IsNotExist(err) {
		log.Fatal("Resources directory does not exists")
	}

	resources := make(map[string][]byte)
	var totalSize int64 = 0
	err := filepath.Walk(in_dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Println("Error :", err)
			return err
		}
		relativePath := filepath.ToSlash(strings.TrimPrefix(path, "resources"))
		if info.IsDir() {
			log.Println(path, "is a directory, skip")
			return nil
		} else {
			log.Println(path, "is a file, packing..")
			b, err := ioutil.ReadFile(path)
			if err != nil {
				log.Printf("Error reading %s: %s", path, err)
				return err
			}
			resources["/"+strings.TrimLeft(relativePath, "/")] = b
			totalSize += int64(len(b))
		}
		return nil
	})

	if err != nil {
		log.Fatal("Error walking through resources directory:", err)
	}

	f, err := os.Create(resources_out)
	if err != nil {
		log.Fatal("Error creating blob file:", err)
	}
	defer f.Close()

	builder := &bytes.Buffer{}

	err = packageTemplate.Execute(builder, resources)
	if err != nil {
		log.Fatal("Error executing template", err)
	}

	data, err := format.Source(builder.Bytes())
	if err != nil {
		log.Fatal("Error formatting generated code", err)
	}
	err = ioutil.WriteFile(resources_out, data, os.ModePerm)
	if err != nil {
		log.Fatal("Error writing blob file", err)
	}

	log.Printf("Resources packing done! Total size: %d KiB", totalSize/1024)
}