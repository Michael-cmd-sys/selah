package bible

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const helloaoBase = "https://api.helloao.org/api"

type helloaoClient struct {
	http *http.Client
}

type helloaoBook struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	CommonName      string `json:"commonName"`
	Order           int    `json:"order"`
	NumberOfChapters int   `json:"numberOfChapters"`
}

type helloaoTranslation struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	ShortName string `json:"shortName"`
	Language  string `json:"language"`
}

type helloaoVerse struct {
	Number int    `json:"number"`
	Text   string `json:"text"`
}

type helloaoChapter struct {
	TranslationID string         `json:"translationId"`
	BookID        string         `json:"bookId"`
	BookName      string         `json:"bookName"`
	Chapter       int            `json:"chapter"`
	Verses        []helloaoVerse `json:"verses"`
	NextChapter   *struct {
		BookID  string `json:"bookId"`
		Chapter int    `json:"chapter"`
	} `json:"nextChapter"`
	PreviousChapter *struct {
		BookID  string `json:"bookId"`
		Chapter int    `json:"chapter"`
	} `json:"previousChapter"`
}

func (c *helloaoClient) get(path string, out any) error {
	resp, err := c.http.Get(helloaoBase + path)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("helloao returned %d for %s", resp.StatusCode, path)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, out)
}

func (c *helloaoClient) translations() ([]helloaoTranslation, error) {
	var result struct {
		Translations []helloaoTranslation `json:"translations"`
	}
	if err := c.get("/available_translations", &result); err != nil {
		return nil, err
	}
	return result.Translations, nil
}

func (c *helloaoClient) books(translationID string) ([]helloaoBook, error) {
	var result struct {
		Books []helloaoBook `json:"books"`
	}
	if err := c.get(fmt.Sprintf("/%s/books", translationID), &result); err != nil {
		return nil, err
	}
	return result.Books, nil
}

func (c *helloaoClient) chapter(translationID, bookID string, chapter int) (*helloaoChapter, error) {
	var result helloaoChapter
	path := fmt.Sprintf("/%s/%s/%d", translationID, bookID, chapter)
	if err := c.get(path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *helloaoClient) search(translationID, query string) ([]helloaoVerse, error) {
	var result struct {
		Verses []helloaoVerse `json:"verses"`
	}
	path := fmt.Sprintf("/%s/search?q=%s", translationID, url.QueryEscape(query))
	if err := c.get(path, &result); err != nil {
		return nil, err
	}
	return result.Verses, nil
}
