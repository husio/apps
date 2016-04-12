package gallery

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image"
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

func handleImageTags(ctx context.Context, w http.ResponseWriter, r *http.Request) {
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

	fd, err := header.Open()
	if err != nil {
		log.Error("cannot open uploaded file",
			"name", header.Filename,
			"error", err.Error())
		web.StdJSONResp(w, http.StatusInternalServerError)
		return
	}
	defer fd.Close()

	img, image, err := prepareImage(fd)
	if err != nil {
		log.Error("cannot extract image metadata", "error", err.Error())
		web.StdJSONResp(w, http.StatusInternalServerError)
		return
	}

	// encode image as JPEG
	var b bytes.Buffer
	if err := jpeg.Encode(&b, img, &jpeg.Options{100}); err != nil {
		log.Error("cannot store image", "error", err.Error())
		web.StdJSONResp(w, http.StatusInternalServerError)
		return
	}
	imgb := b.Bytes()

	// compute image hash from image content
	oid := sha256.New()
	if _, err := oid.Write(imgb); err != nil {
		log.Error("cannot hash file",
			"name", header.Filename,
			"error", err.Error())
		web.StdJSONResp(w, http.StatusInternalServerError)
		return
	}
	image.ImageID = encode(oid)

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

	// store image locally
	path := fmt.Sprintf("/tmp/%s.jpg", image.ImageID)
	if err := imaging.Save(img, path); err != nil {
		log.Error("cannot store image",
			"path", path,
			"error", err.Error())
		web.StdJSONResp(w, http.StatusInternalServerError)
		return
	}
	log.Debug("image file created",
		"id", image.ImageID,
		"path", path)

	web.JSONResp(w, image, http.StatusCreated)
}

func prepareImage(r io.ReadSeeker) (image.Image, *Image, error) {
	img, err := jpeg.Decode(r)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot decode JPEG: %s", err)
	}
	image := Image{
		Width:  img.Bounds().Max.X,
		Height: img.Bounds().Max.Y,
	}

	if _, err := r.Seek(0, os.SEEK_SET); err != nil {
		return nil, nil, fmt.Errorf("cannot seek: %s", err)
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
				switch o {
				case 1:
					// rotation is ok
				case 3:
					img = imaging.Rotate180(img)
				case 8:
					img = imaging.Rotate90(img)
				case 6:
					img = imaging.Rotate270(img)
				default:
					log.Debug("unknown image orientation",
						"decoder", "EXIF",
						"value", fmt.Sprint(o))
				}
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
				image.Created, err = time.Parse("2006:01:02 15:04:05", raw)
				if err != nil {
					log.Debug("cannot parse datetime original",
						"decoder", "EXIF",
						"value", raw,
						"error", err.Error())
				}
			}
		}
	}

	return img, &image, nil
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

	imageID := web.Args(ctx).ByIndex(0)
	switch err := ImageExists(db, imageID); err {
	case nil:
		// all good
	case sq.ErrNotFound:
		web.JSONErr(w, "parent image does not exist", http.StatusBadRequest)
		return
	default:
		log.Error("database error",
			"image", imageID,
			"error", err.Error())
		web.StdJSONResp(w, http.StatusInternalServerError)
		return
	}

	tag, err := CreateTag(db, Tag{
		ImageID: imageID,
		Name:    input.Name,
		Value:   input.Value,
	})
	switch err {
	case nil:
		web.JSONResp(w, tag, http.StatusCreated)
	case sq.ErrConflict:
		web.JSONResp(w, tag, http.StatusOK)
	default:
		log.Error("cannot create object", "error", err.Error())
		web.StdJSONResp(w, http.StatusInternalServerError)
	}
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

	// TODO: real dir
	path := fmt.Sprintf("/tmp/%s.jpg", img.ImageID)
	fd, err := os.Open(path)
	if err != nil {
		log.Error("cannot open image file",
			"image", img.ImageID,
			"path", path,
			"error", err.Error())
		web.StdJSONResp(w, http.StatusInternalServerError)
		return
	}
	defer fd.Close()

	image, err := jpeg.Decode(fd)
	if err != nil {
		log.Error("cannot read image file",
			"image", img.ImageID,
			"path", path,
			"error", err.Error())
		web.StdJSONResp(w, http.StatusInternalServerError)
		return
	}

	if resize := r.URL.Query().Get("resize"); resize != "" {
		var w, h int
		if _, err := fmt.Sscanf(resize, "%dx%d", &w, &h); err == nil {
			image = imaging.Fill(image, w, h, imaging.Center, imaging.Linear)
		}
	}

	w.Header().Set("X-Image-ID", img.ImageID)
	w.Header().Set("X-Image-Width", fmt.Sprint(img.Width))
	w.Header().Set("X-Image-Height", fmt.Sprint(img.Height))
	w.Header().Set("X-Image-Created", img.Created.Format(time.RFC3339))
	w.Header().Set("Content-Type", "image/jpeg")
	imaging.Encode(w, image, imaging.JPEG)
}
