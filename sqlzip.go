package mback

import (
	"archive/zip"
	"fmt"
	"os"
)

type ZipFileWriter struct {
	zipWriter *zip.Writer
	zipFile   *os.File
}

func NewZipFileWriter(filename string) (*ZipFileWriter, error) {
	zipFile, err := os.Create(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create ZIP file: %w", err)
	}
	zipWriter := zip.NewWriter(zipFile)
	return &ZipFileWriter{zipWriter: zipWriter, zipFile: zipFile}, nil
}

func (w *ZipFileWriter) Write(tableName, content string) error {
	fileWriter, err := w.zipWriter.Create(tableName + ".sql")
	if err != nil {
		return fmt.Errorf("failed to create file in ZIP: %w", err)
	}
	_, err = fileWriter.Write([]byte(content + "\n"))
	return err
}

func (w *ZipFileWriter) Close() error {
	err := w.zipWriter.Close()
	if err != nil {
		return err
	}
	return w.zipFile.Close()
}
