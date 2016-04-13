package gallery

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image/jpeg"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/husio/x/log"
	"github.com/husio/x/storage/sq"
	"github.com/husio/x/web"
	"github.com/rwcarlsen/goexif/exif"

	"golang.org/x/net/context"
)

func handleImageDetails(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	db := sq.DB(ctx)
	img, err := ImageByID(db, web.Args(ctx).ByIndex(0))
	switch err {
	case nil:
		// all good
	case sq.ErrNotFound:
		web.StdHTMLResp(w, http.StatusNotFound)
		return
	default:
		log.Error("cannot get image",
			"image", web.Args(ctx).ByIndex(0),
			"error", err.Error())
		web.StdJSONResp(w, http.StatusInternalServerError)
		return
	}

	img.Tags, err = ImageTags(db, img.ImageID)
	if err != nil {
		log.Error("cannot get image",
			"image", web.Args(ctx).ByIndex(0),
			"error", err.Error())
		web.StdJSONResp(w, http.StatusInternalServerError)
		return
	}

	web.JSONResp(w, img, http.StatusOK)
}

func handleListImages(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	offset, _ := strconv.ParseInt(r.URL.Query().Get("offset"), 10, 64)
	opts := ImagesOpts{
		Offset: offset,
		Limit:  200,
	}

	// narrow to images tagged as specified
	for name, values := range r.URL.Query() {
		if !strings.HasPrefix(name, "tag_") {
			continue
		}
		for _, value := range values {
			opts.Tags = append(opts.Tags, KeyValue{
				Key:   name[4:],
				Value: value,
			})
		}
	}

	imgs, err := Images(sq.DB(ctx), opts)
	if err != nil {
		log.Error("cannot list images", "error", err.Error())
		web.StdJSONResp(w, http.StatusInternalServerError)
		return
	}

	resp := struct {
		Images []*Image `json:"images"`
	}{
		Images: imgs,
	}
	web.JSONResp(w, resp, http.StatusOK)
}

func handleUploadImage(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 * megabyte); err != nil {
		web.JSONResp(w, err.Error(), http.StatusBadRequest)
		return
	}

	var header *multipart.FileHeader
	for _, headers := range r.MultipartForm.File {
		for _, h := range headers {
			log.Debug("uploading file", "name", h.Filename)
			if header != nil {
				web.JSONErr(w, "cannot upload more than one time at once", http.StatusBadRequest)
				return
			}
			header = h
		}
	}
	if header == nil {
		web.JSONErr(w, "image file missing", http.StatusBadRequest)
		return
	}
	if !strings.HasSuffix(header.Filename, ".jpg") {
		// XXX this is not the best validation
		web.JSONErr(w, "only JPEG format is allowed", http.StatusBadRequest)
		return
	}

	fd, err := header.Open()
	if err != nil {
		log.Error("cannot open uploaded file",
			"name", header.Filename,
			"error", err.Error())
		web.StdJSONResp(w, http.StatusInternalServerError)
		return
	}
	defer fd.Close()

	image, err := imageMeta(fd)
	if err != nil {
		log.Error("cannot extract image metadata", "error", err.Error())
		web.StdJSONResp(w, http.StatusInternalServerError)
		return
	}

	// store image in database
	db := sq.DB(ctx)
	image, err = CreateImage(db, *image)
	switch err {
	case nil:
		// all good
	case sq.ErrConflict:
		// image already exists, nothing more to do here
		web.JSONResp(w, image, http.StatusOK)
		return
	default:
		log.Error("cannot create object", "error", err.Error())
		web.StdJSONResp(w, http.StatusInternalServerError)
		return
	}

	if _, err := fd.Seek(0, os.SEEK_SET); err != nil {
		log.Error("cannot seek image", "error", err.Error())
		web.StdJSONResp(w, http.StatusInternalServerError)
		return
	}

	fs := FileStore(ctx)
	if err := fs.Put(image, fd); err != nil {
		log.Error("cannot store image", "error", err.Error())
		web.StdJSONResp(w, http.StatusInternalServerError)
		return
	}
	log.Debug("image file created", "id", image.ImageID)

	web.JSONResp(w, image, http.StatusCreated)
}

func imageMeta(r io.ReadSeeker) (*Image, error) {
	conf, err := jpeg.DecodeConfig(r)
	if err != nil {
		return nil, fmt.Errorf("cannot decode JPEG: %s", err)
	}

	// compute image hash from image content
	oid := sha256.New()
	if _, err := io.Copy(oid, r); err != nil {
		return nil, fmt.Errorf("cannot compute SHA: %s", err)
	}
	img := Image{
		ImageID: encode(oid),
		Width:   conf.Width,
		Height:  conf.Height,
	}

	if _, err := r.Seek(0, os.SEEK_SET); err != nil {
		return nil, fmt.Errorf("cannot seek: %s", err)
	}
	if meta, err := exif.Decode(r); err != nil {
		log.Error("cannot extract EXIF metadata", "error", err.Error())
	} else {
		if orientation, err := meta.Get(exif.Orientation); err != nil {
			log.Debug("cannot extract image orientation",
				"decoder", "EXIF",
				"error", err.Error())
		} else {
			if o, err := orientation.Int(0); err != nil {
				log.Debug("cannot format orientation",
					"decoder", "EXIF",
					"error", err.Error())
			} else {
				img.Orientation = o
			}
		}
		if dt, err := meta.Get(exif.DateTimeOriginal); err != nil {
			log.Debug("cannot extract image datetime original",
				"decoder", "EXIF",
				"error", err.Error())
		} else {
			if raw, err := dt.StringVal(); err != nil {
				log.Debug("cannot format datetime original",
					"decoder", "EXIF",
					"error", err.Error())
			} else {
				img.Created, err = time.Parse("2006:01:02 15:04:05", raw)
				if err != nil {
					log.Debug("cannot parse datetime original",
						"decoder", "EXIF",
						"value", raw,
						"error", err.Error())
				}
			}
		}
	}

	return &img, nil
}

const (
	_        = iota
	kilobyte = 1 << (10 * iota)
	megabyte
)

func handleTagImage(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name  string
		Value string
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		web.JSONErr(w, err.Error(), http.StatusBadRequest)
		return
	}

	var errs []string
	if input.Name == "" {
		errs = append(errs, `"name" is required`)
	}
	if input.Value == "" {
		errs = append(errs, `"value" is required`)
	}
	if len(errs) != 0 {
		web.JSONErrs(w, errs, http.StatusBadRequest)
		return
	}

	db := sq.DB(ctx)

	img, err := ImageByID(db, web.Args(ctx).ByIndex(0))
	switch err {
	case nil:
		// all good
	case sq.ErrNotFound:
		web.JSONErr(w, "parent image does not exist", http.StatusBadRequest)
		return
	default:
		log.Error("database error",
			"image", web.Args(ctx).ByIndex(0),
			"error", err.Error())
		web.StdJSONResp(w, http.StatusInternalServerError)
		return
	}

	tag, err := CreateTag(db, Tag{
		ImageID: img.ImageID,
		Name:    input.Name,
		Value:   input.Value,
	})
	switch err {
	case nil:
		// all good, update storage meta
	case sq.ErrConflict:
		web.JSONResp(w, tag, http.StatusOK)
		return
	default:
		log.Error("cannot create object", "error", err.Error())
		web.StdJSONResp(w, http.StatusInternalServerError)
		return
	}

	if img.Tags, err = ImageTags(db, img.ImageID); err != nil {
		log.Error("cannot get image tags",
			"image", img.ImageID,
			"error", err.Error())
		web.StdJSONResp(w, http.StatusInternalServerError)
		return
	}

	fs := FileStore(ctx)
	if err := fs.PutMeta(img); err != nil {
		log.Error("cannot store image metadata",
			"image", img.ImageID,
			"error", err.Error())
		web.StdJSONResp(w, http.StatusInternalServerError)
		return
	}

	web.JSONResp(w, tag, http.StatusCreated)
}

func handleServeImage(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	img, err := ImageByID(sq.DB(ctx), web.Args(ctx).ByIndex(0))
	switch err {
	case nil:
		// all good
	case sq.ErrNotFound:
		web.StdJSONResp(w, http.StatusNotFound)
		return
	default:
		log.Error("cannot get object",
			"object", web.Args(ctx).ByIndex(0),
			"error", err.Error())
		web.StdJSONResp(w, http.StatusInternalServerError)
		return
	}

	if web.CheckLastModified(w, r, img.Created) {
		return
	}

	fs := FileStore(ctx)
	fd, err := fs.Read(img.Created.Year(), img.ImageID)
	if err != nil {
		log.Error("cannot read image file",
			"image", img.ImageID,
			"error", err.Error())
		web.StdJSONResp(w, http.StatusInternalServerError)
		return
	}
	defer fd.Close()

	w.Header().Set("X-Image-ID", img.ImageID)
	w.Header().Set("X-Image-Width", fmt.Sprint(img.Width))
	w.Header().Set("X-Image-Height", fmt.Sprint(img.Height))
	w.Header().Set("X-Image-Created", img.Created.Format(time.RFC3339))
	w.Header().Set("Content-Type", "image/jpeg")

	if r.URL.Query().Get("resize") == "" {
		io.Copy(w, fd)
		return
	}

	image, err := jpeg.Decode(fd)
	if err != nil {
		log.Error("cannot read image file",
			"image", img.ImageID,
			"error", err.Error())
		web.StdJSONResp(w, http.StatusInternalServerError)
		return
	}
	var width, height int
	if _, err := fmt.Sscanf(r.URL.Query().Get("resize"), "%dx%d", &width, &height); err != nil {
		log.Error("cannot resize image",
			"image", img.ImageID,
			"error", err.Error())
	} else {
		image = imaging.Fill(image, width, height, imaging.Center, imaging.Linear)
		switch img.Orientation {
		case 1:
			// all good
		case 3:
			image = imaging.Rotate180(image)
		case 8:
			image = imaging.Rotate90(image)
		case 6:
			image = imaging.Rotate270(image)
		default:
			log.Debug("unknown image orientation",
				"decoder", "EXIF",
				"image", img.ImageID,
				"value", fmt.Sprint(img.Orientation))
		}
	}
	imaging.Encode(w, image, imaging.JPEG)
}
