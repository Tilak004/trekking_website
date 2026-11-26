package main

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// Photo uploads from the admin console (drag and drop or a file chooser), so
// photos no longer have to be copied into the server's photos folder by hand.

const maxAdminPhotoBytes = 15 << 20

// photoFolder is where an upload may go: the shared "uploads" folder (trek cover
// photos), "leads", or one trek's own gallery, "treks/<slug>".
func photoFolder(v string) (string, bool) {
	v = strings.Trim(strings.TrimSpace(v), "/")
	switch {
	case v == "":
		return "uploads", true
	case v == "uploads", v == "leads":
		return v, true
	case strings.HasPrefix(v, "treks/"):
		slug := strings.TrimPrefix(v, "treks/")
		return v, slug != "" && slug == slugify(slug)
	}
	return "", false
}

func (a *App) adminUploadPhoto(w http.ResponseWriter, r *http.Request) {
	fail := func(status int, msg string) { apiJSON(w, status, map[string]string{"error": msg}) }
	r.Body = http.MaxBytesReader(w, r.Body, maxAdminPhotoBytes+(1<<20))
	if err := r.ParseMultipartForm(8 << 20); err != nil {
		fail(http.StatusBadRequest, "That file is too large (15 MB at most) or didn't upload properly.")
		return
	}
	folder, ok := photoFolder(r.FormValue("folder"))
	if !ok {
		fail(http.StatusBadRequest, "Unknown photo folder.")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		fail(http.StatusBadRequest, "Choose a photo to upload.")
		return
	}
	defer file.Close()
	name := strings.TrimSuffix(header.Filename, path.Ext(header.Filename))
	rel, err := a.savePhoto(file, folder, name)
	if err != nil {
		fail(http.StatusBadRequest, "Upload the photo as a JPG, PNG or WebP image.")
		return
	}
	apiJSON(w, http.StatusOK, map[string]string{"path": rel})
}

// adminDeletePhoto removes a photo from a trek's own gallery. Cover and lead
// photos are never deleted here, since a trek or lead may still be using them.
func (a *App) adminDeletePhoto(w http.ResponseWriter, r *http.Request) {
	rel := path.Clean(strings.TrimSpace(r.FormValue("path")))
	if !strings.HasPrefix(rel, "treks/") || strings.Count(rel, "/") != 2 || !isImage(rel) {
		apiJSON(w, http.StatusBadRequest, map[string]string{"error": "That photo can't be removed here."})
		return
	}
	if err := os.Remove(filepath.Join(a.photosDir, filepath.FromSlash(rel))); err != nil && !os.IsNotExist(err) {
		a.serverError(w, err)
		return
	}
	apiJSON(w, http.StatusOK, map[string]string{"removed": rel})
}
