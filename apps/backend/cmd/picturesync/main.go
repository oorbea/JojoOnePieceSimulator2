// Command picturesync fills in missing Devil Fruit pictures from the One
// Piece wiki, with a human review step in the middle. It talks to a running
// backend over its public API only (no DB access), so it works against prod
// with an ADMIN access token.
//
//	picturesync list       -> out/missing.json        (read-only)
//	picturesync candidates -> out/candidates/*, out/review.html
//	picturesync upload     -> PATCH /devil-fruits/{id}/picture per selection
//
// Candidate images are only ever stored under --out (gitignored); nothing is
// committed. The token comes from PICTURE_SYNC_TOKEN or --token-file, never a
// flag value, so it stays out of shell history. See
// ObsidianVault/devil-fruit-picture-sync.md.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	defaultBaseURL = "https://jojo-one-piece-simulator.duckdns.org"
	wikiAPI        = "https://onepiece.fandom.com/api.php"
	userAgent      = "JOPS-picturesync/1.0 (hobby project; contact via repo owner)"
	// maxImageBytes stays under the backend's default 5 MiB
	// PICTURE_MAX_BYTES so an accepted candidate never gets a 413.
	maxImageBytes  = 4*1024*1024 + 512*1024
	maxCandidates  = 3
	uploadPause    = 1500 * time.Millisecond
	maxRateRetries = 5
	httpTimeout    = 60 * time.Second
)

type fruit struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Rarity string `json:"rarity"`
}

type candidate struct {
	File   string `json:"file"`   // relative to --out
	Source string `json:"source"` // original wiki URL
	Page   string `json:"page"`   // wiki page title the image came from
}

type candidateEntry struct {
	fruit
	Candidates []candidate `json:"candidates"`
}

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	cmd := os.Args[1]
	fs := flag.NewFlagSet(cmd, flag.ExitOnError)
	base := fs.String("base-url", defaultBaseURL, "backend base URL (no trailing /api)")
	out := fs.String("out", "picturesync-out", "working directory for missing.json, candidates and review.html")
	tokenFile := fs.String("token-file", "", "file holding the ADMIN access token (default: $PICTURE_SYNC_TOKEN)")
	excludeRarity := fs.String("exclude-rarity", "EPIC,LEGENDARY,MYTHICAL", "list: rarities left for admins to upload by hand")
	dryRun := fs.Bool("dry-run", false, "upload: log what would be sent without sending")
	_ = fs.Parse(os.Args[2:])

	ctx := context.Background()
	c := &client{base: strings.TrimRight(*base, "/"), http: &http.Client{Timeout: httpTimeout}}
	var err error
	switch cmd {
	case "list":
		c.token, err = loadToken(*tokenFile)
		if err == nil {
			err = runList(ctx, c, *out, splitSet(*excludeRarity))
		}
	case "candidates":
		err = runCandidates(ctx, c, *out)
	case "upload":
		c.token, err = loadToken(*tokenFile)
		if err == nil {
			err = runUpload(ctx, c, *out, *dryRun)
		}
	default:
		usage()
	}
	if err != nil {
		log.Fatalf("picturesync %s: %v", cmd, err)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: picturesync list|candidates|upload [flags]")
	os.Exit(2)
}

func loadToken(file string) (string, error) {
	if file != "" {
		b, err := os.ReadFile(file)
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(b)), nil
	}
	if t := strings.TrimSpace(os.Getenv("PICTURE_SYNC_TOKEN")); t != "" {
		return t, nil
	}
	return "", errors.New("no token: set PICTURE_SYNC_TOKEN or pass --token-file")
}

func splitSet(csv string) map[string]bool {
	set := map[string]bool{}
	for _, s := range strings.Split(csv, ",") {
		if s = strings.TrimSpace(s); s != "" {
			set[strings.ToUpper(s)] = true
		}
	}
	return set
}

// ---------------------------------------------------------------- backend

type client struct {
	base  string
	token string
	http  *http.Client
}

type fruitRow struct {
	fruit
	Picture string `json:"picture"`
}

func (c *client) listFruits(ctx context.Context) ([]fruitRow, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.base+"/api/v1/devil-fruits", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("User-Agent", userAgent)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET /devil-fruits: %s", resp.Status)
	}
	// The unpaginated form returns either a bare array or {items: [...]}.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var rows []fruitRow
	if err := json.Unmarshal(body, &rows); err == nil {
		return rows, nil
	}
	var page struct {
		Items []fruitRow `json:"items"`
	}
	if err := json.Unmarshal(body, &page); err != nil {
		return nil, fmt.Errorf("decoding /devil-fruits: %w", err)
	}
	return page.Items, nil
}

// missingFruits keeps rows without a picture whose rarity is not reserved
// for manual admin upload.
func missingFruits(rows []fruitRow, excludeRarity map[string]bool) []fruit {
	var out []fruit
	for _, r := range rows {
		if r.Picture != "" || excludeRarity[strings.ToUpper(r.Rarity)] {
			continue
		}
		out = append(out, r.fruit)
	}
	return out
}

func runList(ctx context.Context, c *client, out string, excludeRarity map[string]bool) error {
	rows, err := c.listFruits(ctx)
	if err != nil {
		return err
	}
	missing := missingFruits(rows, excludeRarity)
	withoutAny := 0
	for _, r := range rows {
		if r.Picture == "" {
			withoutAny++
		}
	}
	if err := writeJSON(filepath.Join(out, "missing.json"), missing); err != nil {
		return err
	}
	log.Printf("%d fruits total, %d without picture, %d go to the wiki flow (rest left for admins)",
		len(rows), withoutAny, len(missing))
	return nil
}

func (c *client) upload(ctx context.Context, id, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	for attempt := 0; ; attempt++ {
		var buf bytes.Buffer
		mw := multipart.NewWriter(&buf)
		h := textproto.MIMEHeader{}
		h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="picture"; filename="%s"`, filepath.Base(path)))
		h.Set("Content-Type", http.DetectContentType(data))
		part, err := mw.CreatePart(h)
		if err != nil {
			return err
		}
		if _, err := part.Write(data); err != nil {
			return err
		}
		if err := mw.Close(); err != nil {
			return err
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodPatch, c.base+"/api/v1/devil-fruits/"+id+"/picture", &buf)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.token)
		req.Header.Set("Content-Type", mw.FormDataContentType())
		req.Header.Set("User-Agent", userAgent)
		resp, err := c.http.Do(req)
		if err != nil {
			return err
		}
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		resp.Body.Close()
		switch {
		case resp.StatusCode == http.StatusAccepted || resp.StatusCode == http.StatusOK:
			return nil
		case resp.StatusCode == http.StatusTooManyRequests && attempt < maxRateRetries:
			time.Sleep(time.Duration(attempt+1) * 5 * time.Second)
		default:
			return fmt.Errorf("%s: %s", resp.Status, strings.TrimSpace(string(msg)))
		}
	}
}

func runUpload(ctx context.Context, c *client, out string, dryRun bool) error {
	var selections map[string]string // fruit id -> candidate file (relative to out), "" = skip
	if err := readJSON(filepath.Join(out, "selections.json"), &selections); err != nil {
		return fmt.Errorf("reading selections.json (export it from review.html): %w", err)
	}
	var entries []candidateEntry
	if err := readJSON(filepath.Join(out, "candidates.json"), &entries); err != nil {
		return err
	}
	names := map[string]string{}
	for _, e := range entries {
		names[e.ID] = e.Name
	}
	// Re-check live state so a retry never re-uploads over a picture an
	// admin (or a previous partial run) already set.
	rows, err := c.listFruits(ctx)
	if err != nil {
		return err
	}
	has := map[string]bool{}
	for _, r := range rows {
		has[r.ID] = r.Picture != ""
	}
	var done, skipped, failed int
	for id, rel := range selections {
		name := names[id]
		switch {
		case rel == "":
			skipped++
		case has[id]:
			log.Printf("skip %s: already has a picture", name)
			skipped++
		case dryRun:
			log.Printf("dry-run: would upload %s -> %s", rel, name)
		default:
			if err := c.upload(ctx, id, filepath.Join(out, rel)); err != nil {
				log.Printf("FAIL %s: %v", name, err)
				failed++
				if strings.HasPrefix(err.Error(), "401") {
					return errors.New("token expired: get a fresh one and re-run (finished entries are skipped)")
				}
				continue
			}
			log.Printf("ok %s", name)
			done++
			time.Sleep(uploadPause)
		}
	}
	log.Printf("uploaded %d, skipped %d, failed %d", done, skipped, failed)
	if failed > 0 {
		return fmt.Errorf("%d uploads failed", failed)
	}
	return nil
}

// ------------------------------------------------------------------- wiki

// wikiTitle turns "Hira Hira no mi" / "Inu Inu no mi: Model Hound" into the
// capitalisation the wiki uses for its page titles.
func wikiTitle(name string) string {
	base := name
	model := ""
	if i := strings.Index(name, ":"); i >= 0 {
		base, model = strings.TrimSpace(name[:i]), strings.TrimSpace(name[i+1:])
	}
	base = strings.Replace(base, " no mi", " no Mi", 1)
	if model != "" {
		return base + ", Model: " + strings.TrimPrefix(model, "Model ")
	}
	return base
}

// imageStem is the file-name prefix the wiki uses for a fruit's images,
// e.g. "Hira_Hira_no_Mi".
func imageStem(name string) string {
	base := name
	if i := strings.Index(name, ":"); i >= 0 {
		base = name[:i]
	}
	return strings.ReplaceAll(strings.Replace(strings.TrimSpace(base), " no mi", " no Mi", 1), " ", "_")
}

type wikiClient struct {
	http *http.Client
}

func (w *wikiClient) get(ctx context.Context, q url.Values, dst any) error {
	q.Set("format", "json")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, wikiAPI+"?"+q.Encode(), nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", userAgent)
	resp, err := w.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("wiki: %s", resp.Status)
	}
	return json.NewDecoder(resp.Body).Decode(dst)
}

// resolvePage finds the wiki page for a fruit: exact title first, search as
// a fallback (model variants have irregular titles).
func (w *wikiClient) resolvePage(ctx context.Context, f fruit) (string, error) {
	var res struct {
		Query struct {
			Pages map[string]struct {
				Title   string    `json:"title"`
				Missing json.RawMessage `json:"missing"`
			} `json:"pages"`
		} `json:"query"`
	}
	if err := w.get(ctx, url.Values{"action": {"query"}, "titles": {wikiTitle(f.Name)}, "redirects": {"1"}}, &res); err != nil {
		return "", err
	}
	for _, p := range res.Query.Pages {
		if len(p.Missing) == 0 {
			return p.Title, nil
		}
	}
	var s struct {
		Query struct {
			Search []struct {
				Title string `json:"title"`
			} `json:"search"`
		} `json:"query"`
	}
	if err := w.get(ctx, url.Values{"action": {"query"}, "list": {"search"}, "srsearch": {wikiTitle(f.Name)}, "srlimit": {"1"}}, &s); err != nil {
		return "", err
	}
	if len(s.Query.Search) == 0 {
		return "", nil
	}
	return s.Query.Search[0].Title, nil
}

// pageImages returns original image URLs on a page: the infobox image first,
// then other files whose name starts with the fruit's stem.
func (w *wikiClient) pageImages(ctx context.Context, title, stem string) ([]string, error) {
	var res struct {
		Query struct {
			Pages map[string]struct {
				Original *struct {
					Source string `json:"source"`
				} `json:"original"`
				Images []struct {
					Title string `json:"title"`
				} `json:"images"`
			} `json:"pages"`
		} `json:"query"`
	}
	q := url.Values{"action": {"query"}, "titles": {title}, "prop": {"pageimages|images"}, "piprop": {"original"}, "imlimit": {"50"}}
	if err := w.get(ctx, q, &res); err != nil {
		return nil, err
	}
	var urls []string
	var files []string
	for _, p := range res.Query.Pages {
		if p.Original != nil && !isPlaceholder(p.Original.Source) {
			urls = append(urls, p.Original.Source)
		}
		for _, im := range p.Images {
			name := strings.ReplaceAll(strings.TrimPrefix(im.Title, "File:"), " ", "_")
			if strings.HasPrefix(strings.ToLower(name), strings.ToLower(stem)) && isImageFile(name) {
				files = append(files, im.Title)
			}
		}
	}
	for _, f := range files {
		if len(urls) >= maxCandidates*2 {
			break
		}
		var info struct {
			Query struct {
				Pages map[string]struct {
					ImageInfo []struct {
						URL string `json:"url"`
					} `json:"imageinfo"`
				} `json:"pages"`
			} `json:"query"`
		}
		if err := w.get(ctx, url.Values{"action": {"query"}, "titles": {f}, "prop": {"imageinfo"}, "iiprop": {"url"}}, &info); err != nil {
			continue
		}
		for _, p := range info.Query.Pages {
			for _, ii := range p.ImageInfo {
				urls = append(urls, ii.URL)
			}
		}
	}
	return dedupe(urls), nil
}

func isPlaceholder(u string) bool { return strings.Contains(u, "NoPicAvailable") }

func isImageFile(name string) bool {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".png", ".jpg", ".jpeg", ".webp", ".gif":
		return true
	}
	return false
}

func dedupe(in []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range in {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}

func (w *wikiClient) download(ctx context.Context, src string) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, src, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("User-Agent", userAgent)
	// The wiki's image CDN 403s hotlinks without a wiki Referer.
	req.Header.Set("Referer", "https://onepiece.fandom.com/")
	resp, err := w.http.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("%s", resp.Status)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxImageBytes+1))
	if err != nil {
		return nil, "", err
	}
	if len(data) > maxImageBytes {
		return nil, "", errors.New("too large")
	}
	ext := map[string]string{"image/png": ".png", "image/jpeg": ".jpg", "image/webp": ".webp", "image/gif": ".gif"}[http.DetectContentType(data)]
	if ext == "" {
		return nil, "", errors.New("not an accepted image type")
	}
	return data, ext, nil
}

func runCandidates(ctx context.Context, c *client, out string) error {
	var missing []fruit
	if err := readJSON(filepath.Join(out, "missing.json"), &missing); err != nil {
		return fmt.Errorf("reading missing.json (run `list` first): %w", err)
	}
	w := &wikiClient{http: c.http}
	var entries []candidateEntry
	for _, f := range missing {
		e := candidateEntry{fruit: f}
		title, err := w.resolvePage(ctx, f)
		if err != nil || title == "" {
			log.Printf("no wiki page: %s (%v)", f.Name, err)
			entries = append(entries, e)
			continue
		}
		urls, err := w.pageImages(ctx, title, imageStem(f.Name))
		if err != nil {
			log.Printf("images failed: %s: %v", f.Name, err)
		}
		for _, u := range urls {
			if len(e.Candidates) >= maxCandidates {
				break
			}
			data, ext, err := w.download(ctx, u)
			if err != nil {
				log.Printf("skip image for %s: %v", f.Name, err)
				continue
			}
			rel := filepath.ToSlash(filepath.Join("candidates", f.ID, fmt.Sprintf("%d%s", len(e.Candidates)+1, ext)))
			if err := os.MkdirAll(filepath.Dir(filepath.Join(out, rel)), 0o755); err != nil {
				return err
			}
			if err := os.WriteFile(filepath.Join(out, rel), data, 0o644); err != nil {
				return err
			}
			e.Candidates = append(e.Candidates, candidate{File: rel, Source: u, Page: title})
		}
		log.Printf("%s: %d candidate(s) from %q", f.Name, len(e.Candidates), title)
		entries = append(entries, e)
		time.Sleep(300 * time.Millisecond)
	}
	if err := writeJSON(filepath.Join(out, "candidates.json"), entries); err != nil {
		return err
	}
	if err := writeReview(filepath.Join(out, "review.html"), entries); err != nil {
		return err
	}
	log.Printf("open %s, pick one per fruit, export selections.json into %s, then run `upload`", filepath.Join(out, "review.html"), out)
	return nil
}

// ---------------------------------------------------------------- helpers

func writeJSON(path string, v any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

func readJSON(path string, dst any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, dst)
}
