package imports

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func importUrssafHandler(c *gin.Context) {
	path := viper.GetString("source.urssafpath")
	reader, err := tarFileReader(path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	err = runHandlers(c, reader)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, "ok")
}

func runHandlers(ctx context.Context, reader *tar.Reader) error {
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		if header.Typeflag == tar.TypeReg {
			handler := selectHandler(header)
			if handler != nil {
				n, err := handler(ctx, reader)
				slog.Info("lignes insérées", slog.Any("type", header.Name), slog.Any("number", n))
				if err != nil {
					return err
				}
			} else {
				slog.Info("pas de handler trouvé, on passe au suivant", slog.Any("type", header.Name))
			}
		}
	}
	return nil
}

func tarFileReader(srcFile string) (*tar.Reader, error) {
	file, err := os.Open(srcFile)
	if err != nil {
		return nil, err
	}

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return nil, err
	}

	return tar.NewReader(gzReader), nil
}

func selectHandler(header *tar.Header) func(ctx context.Context, reader *tar.Reader) (int64, error) {
	switch {
	case strings.HasSuffix(header.Name, "debit.csv"):
		return importDebit
	case strings.HasSuffix(header.Name, "cotisation.csv"):
		return importCotisation
	case strings.HasSuffix(header.Name, "delai.csv"):
		return importDelai
	case strings.HasSuffix(header.Name, "effectif.csv"):
		return importEffectif
	case strings.HasSuffix(header.Name, "procol.csv"):
		return importProcol
	default:
		return nil
	}
}

func parsePeriode(input string) (periode [2]time.Time, err error) {
	if len(input) != 39 {
		return periode, errors.New("moins de 40 caractères")
	}
	periode[0], err = time.Parse("2006-01-02 15:04:05", input[0:19])
	if err != nil {
		return
	}
	periode[1], err = time.Parse("2006-01-02 15:04:05", input[20:39])
	if err != nil {
		return
	}
	return
}
