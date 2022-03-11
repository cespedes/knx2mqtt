package main

import (
	"archive/zip"
	"errors"
	"io"
	"path"
	"regexp"
)

type ETS struct {
	Project io.ReadCloser
}

func Uncompress(filename string) (*ETS, error) {
	var e ETS

	archive, err := zip.OpenReader(filename)
	if err != nil {
		return nil, err
	}
	// defer archive.Close()

	projectMetaFileRe := regexp.MustCompile("^(p|P)-([0-9a-zA-Z]+)/(p|P)roject.xml$")
	projectFileBaseRe := regexp.MustCompile("^(\\d).xml$")

	for _, file := range archive.File {
		// fmt.Printf("Name: %s size=%d\n", file.Name, file.UncompressedSize64)
		if projectMetaFileRe.MatchString(file.Name) {
			// fmt.Printf("** Project file (meta): %q\n", file.Name)
			projectDir := path.Dir(file.Name)
			for _, file2 := range archive.File {
				if path.Dir(file2.Name) != projectDir {
					continue
				}
				if projectFileBaseRe.MatchString(path.Base(file2.Name)) {
					// fmt.Printf("** Project file (real): %q\n", file2.Name)
					e.Project, err = file2.Open()
					if err != nil {
						return nil, err
					}
					return &e, nil
				}
			}

		}
	}
	return nil, errors.New("no project found")
}

/*
var (
        projectMetaFileRe  = regexp.MustCompile("^(p|P)-([0-9a-zA-Z]+)/(p|P)roject.xml$")
        manufacturerFileRe = regexp.MustCompile("^(m|M)-([0-9a-zA-Z]+)/(m|M)-([^.]+).xml$")

        // TODO: Figure out if '/' is a universal path seperator in ZIP files.
)

func (ex *ExportArchive) findFiles() error {
        for _, file := range ex.archive.File {
                if projectMetaFileRe.MatchString(file.Name) {
                        ex.ProjectFiles = append(ex.ProjectFiles, newProjectFile(ex.archive, file))
                } else if matches := manufacturerFileRe.FindStringSubmatch(file.Name); matches != nil {
                        ex.ManufacturerFiles = append(ex.ManufacturerFiles, ManufacturerFile{
                                File:           file,
                                ManufacturerID: "M-" + matches[2],
                                ContentID:      "M-" + matches[4],
                        })
                }
        }

        return nil
}
*/
