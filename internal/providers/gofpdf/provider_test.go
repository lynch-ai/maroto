package gofpdf_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/johnfercher/maroto/v2/internal/fixture"
	"github.com/johnfercher/maroto/v2/internal/merror"
	"github.com/johnfercher/maroto/v2/mocks"
	"github.com/johnfercher/maroto/v2/pkg/consts/extension"
	"github.com/johnfercher/maroto/v2/pkg/consts/protection"
	"github.com/stretchr/testify/mock"

	"github.com/johnfercher/maroto/v2/internal/providers/gofpdf"

	"github.com/johnfercher/maroto/v2/pkg/core/entity"
	"github.com/stretchr/testify/assert"
)

const (
	codeContent = "code"
)

func TestNew(t *testing.T) {
	// Act
	sut := gofpdf.New(&gofpdf.Dependencies{})

	// Assert
	assert.NotNil(t, sut)
	assert.Equal(t, "*gofpdf.provider", fmt.Sprintf("%T", sut))
}

func TestProvider_AddText(t *testing.T) {
	// Arrange
	txtContent := "text"
	cell := &entity.Cell{}
	prop := fixture.TextProp()

	text := &mocks.Text{}
	text.EXPECT().Add(txtContent, cell, &prop)

	dep := &gofpdf.Dependencies{
		Text: text,
	}
	sut := gofpdf.New(dep)

	// Act
	sut.AddText(txtContent, cell, &prop)

	// Assert
	text.AssertNumberOfCalls(t, "Add", 1)
}

func TestProvider_GetTextHeight(t *testing.T) {
	// Arrange
	fontHeightToReturn := 10.0
	prop := fixture.FontProp()

	font := &mocks.Font{}
	font.EXPECT().GetHeight(prop.Family, prop.Style, prop.Size).Return(fontHeightToReturn)

	dep := &gofpdf.Dependencies{
		Font: font,
	}
	sut := gofpdf.New(dep)

	// Act
	fontHeight := sut.GetTextHeight(&prop)

	// Assert
	font.AssertNumberOfCalls(t, "GetHeight", 1)
	assert.Equal(t, fontHeightToReturn, fontHeight)
}

func TestProvider_AddLine(t *testing.T) {
	// Arrange
	cell := &entity.Cell{}
	prop := fixture.LineProp()

	line := &mocks.Line{}
	line.EXPECT().Add(cell, &prop)

	dep := &gofpdf.Dependencies{
		Line: line,
	}
	sut := gofpdf.New(dep)

	// Act
	sut.AddLine(cell, &prop)

	// Assert
	line.AssertNumberOfCalls(t, "Add", 1)
}

// nolint: dupl
func TestProvider_AddMatrixCode(t *testing.T) {
	t.Run("when cannot find image on cache and cannot generate data matrix, should apply error message", func(t *testing.T) {
		// Arrange
		cell := &entity.Cell{}
		prop := fixture.RectProp()

		cache := &mocks.Cache{}
		cache.EXPECT().GetImage(codeContent, extension.Jpg).Return(nil, errors.New("anyError1"))

		code := &mocks.Code{}
		code.EXPECT().GenDataMatrix(codeContent).Return(nil, errors.New("anyError2"))

		text := &mocks.Text{}
		text.EXPECT().Add("could not generate matrixcode", cell, merror.DefaultErrorText)

		dep := &gofpdf.Dependencies{
			Cache: cache,
			Code:  code,
			Text:  text,
		}

		sut := gofpdf.New(dep)

		// Act
		sut.AddMatrixCode(codeContent, cell, &prop)

		// Assert
		cache.AssertNumberOfCalls(t, "GetImage", 1)
		code.AssertNumberOfCalls(t, "GenDataMatrix", 1)
		text.AssertNumberOfCalls(t, "Add", 1)
	})
	t.Run("when can find image on cache but cannot add image, should apply error message", func(t *testing.T) {
		// Arrange
		cell := &entity.Cell{}
		prop := fixture.RectProp()

		img := &entity.Image{Bytes: []byte{1, 2, 3}}

		cache := &mocks.Cache{}
		cache.EXPECT().GetImage(codeContent, extension.Jpg).Return(img, nil)
		cache.EXPECT().AddImage(codeContent, img)

		text := &mocks.Text{}
		text.EXPECT().Add("could not add matrixcode to document", cell, merror.DefaultErrorText)

		cfg := &entity.Config{
			Margins: &entity.Margins{
				Left:   10,
				Top:    10,
				Right:  10,
				Bottom: 10,
			},
		}

		image := &mocks.Image{}
		image.EXPECT().Add(img, cell, cfg.Margins, &prop, extension.Jpg, false).Return(errors.New("anyError"))

		fpdf := &mocks.Fpdf{}
		fpdf.EXPECT().ClearError()

		dep := &gofpdf.Dependencies{
			Cache: cache,
			Text:  text,
			Image: image,
			Fpdf:  fpdf,
			Cfg:   cfg,
		}

		sut := gofpdf.New(dep)

		// Act
		sut.AddMatrixCode(codeContent, cell, &prop)

		// Assert
		cache.AssertNumberOfCalls(t, "GetImage", 1)
		cache.AssertNumberOfCalls(t, "AddImage", 1)
		image.AssertNumberOfCalls(t, "Add", 1)
		text.AssertNumberOfCalls(t, "Add", 1)
	})
	t.Run("when can find image on cache and can add image, should not apply error message", func(t *testing.T) {
		// Arrange
		cell := &entity.Cell{}
		prop := fixture.RectProp()

		img := &entity.Image{Bytes: []byte{1, 2, 3}}

		cache := &mocks.Cache{}
		cache.EXPECT().GetImage(codeContent, extension.Jpg).Return(img, nil)
		cache.EXPECT().AddImage(codeContent, img)

		cfg := &entity.Config{
			Margins: &entity.Margins{
				Left:   10,
				Top:    10,
				Right:  10,
				Bottom: 10,
			},
		}

		image := &mocks.Image{}
		image.EXPECT().Add(img, cell, cfg.Margins, &prop, extension.Jpg, false).Return(nil)

		fpdf := &mocks.Fpdf{}
		fpdf.EXPECT().ClearError()

		dep := &gofpdf.Dependencies{
			Cache: cache,
			Image: image,
			Fpdf:  fpdf,
			Cfg:   cfg,
		}

		sut := gofpdf.New(dep)

		// Act
		sut.AddMatrixCode(codeContent, cell, &prop)

		// Assert
		cache.AssertNumberOfCalls(t, "GetImage", 1)
		cache.AssertNumberOfCalls(t, "AddImage", 1)
		image.AssertNumberOfCalls(t, "Add", 1)
	})
}

// nolint: dupl
func TestProvider_AddQrCode(t *testing.T) {
	t.Run("when cannot find image on cache and cannot generate qr code, should apply error message", func(t *testing.T) {
		// Arrange
		cell := &entity.Cell{}
		prop := fixture.RectProp()

		cache := &mocks.Cache{}
		cache.EXPECT().GetImage(codeContent, extension.Jpg).Return(nil, errors.New("anyError1"))

		code := &mocks.Code{}
		code.EXPECT().GenQr(codeContent).Return(nil, errors.New("anyError2"))

		text := &mocks.Text{}
		text.EXPECT().Add("could not generate qrcode", cell, merror.DefaultErrorText)

		dep := &gofpdf.Dependencies{
			Cache: cache,
			Code:  code,
			Text:  text,
		}

		sut := gofpdf.New(dep)

		// Act
		sut.AddQrCode(codeContent, cell, &prop)

		// Assert
		cache.AssertNumberOfCalls(t, "GetImage", 1)
		code.AssertNumberOfCalls(t, "GenQr", 1)
		text.AssertNumberOfCalls(t, "Add", 1)
	})
	t.Run("when can find image on cache but cannot add image, should apply error message", func(t *testing.T) {
		// Arrange
		cell := &entity.Cell{}
		prop := fixture.RectProp()

		img := &entity.Image{Bytes: []byte{1, 2, 3}}

		cache := &mocks.Cache{}
		cache.EXPECT().GetImage(codeContent, extension.Jpg).Return(img, nil)
		cache.EXPECT().AddImage(codeContent, img)

		text := &mocks.Text{}
		text.EXPECT().Add("could not add qrcode to document", cell, merror.DefaultErrorText)

		cfg := &entity.Config{
			Margins: &entity.Margins{
				Left:   10,
				Top:    10,
				Right:  10,
				Bottom: 10,
			},
		}

		image := &mocks.Image{}
		image.EXPECT().Add(img, cell, cfg.Margins, &prop, extension.Jpg, false).Return(errors.New("anyError"))

		fpdf := &mocks.Fpdf{}
		fpdf.EXPECT().ClearError()

		dep := &gofpdf.Dependencies{
			Cache: cache,
			Text:  text,
			Image: image,
			Fpdf:  fpdf,
			Cfg:   cfg,
		}

		sut := gofpdf.New(dep)

		// Act
		sut.AddQrCode(codeContent, cell, &prop)

		// Assert
		cache.AssertNumberOfCalls(t, "GetImage", 1)
		cache.AssertNumberOfCalls(t, "AddImage", 1)
		image.AssertNumberOfCalls(t, "Add", 1)
		text.AssertNumberOfCalls(t, "Add", 1)
	})
	t.Run("when can find image on cache and can add image, should not apply error message", func(t *testing.T) {
		// Arrange
		cell := &entity.Cell{}
		prop := fixture.RectProp()

		img := &entity.Image{Bytes: []byte{1, 2, 3}}

		cache := &mocks.Cache{}
		cache.EXPECT().GetImage(codeContent, extension.Jpg).Return(img, nil)
		cache.EXPECT().AddImage(codeContent, img)

		cfg := &entity.Config{
			Margins: &entity.Margins{
				Left:   10,
				Top:    10,
				Right:  10,
				Bottom: 10,
			},
		}

		image := &mocks.Image{}
		image.EXPECT().Add(img, cell, cfg.Margins, &prop, extension.Jpg, false).Return(nil)

		fpdf := &mocks.Fpdf{}
		fpdf.EXPECT().ClearError()

		dep := &gofpdf.Dependencies{
			Cache: cache,
			Image: image,
			Fpdf:  fpdf,
			Cfg:   cfg,
		}

		sut := gofpdf.New(dep)

		// Act
		sut.AddQrCode(codeContent, cell, &prop)

		// Assert
		cache.AssertNumberOfCalls(t, "GetImage", 1)
		cache.AssertNumberOfCalls(t, "AddImage", 1)
		image.AssertNumberOfCalls(t, "Add", 1)
	})
}

// nolint: dupl
func TestProvider_AddBarCode(t *testing.T) {
	t.Run("when cannot find image on cache and cannot generate bar code, should apply error message", func(t *testing.T) {
		// Arrange
		cell := &entity.Cell{}
		prop := fixture.BarcodeProp()

		cache := &mocks.Cache{}
		cache.EXPECT().GetImage(codeContent, extension.Jpg).Return(nil, errors.New("anyError1"))

		code := &mocks.Code{}
		code.EXPECT().GenBar(codeContent, cell, &prop).Return(nil, errors.New("anyError2"))

		text := &mocks.Text{}
		text.EXPECT().Add("could not generate barcode", cell, merror.DefaultErrorText)

		dep := &gofpdf.Dependencies{
			Cache: cache,
			Code:  code,
			Text:  text,
		}

		sut := gofpdf.New(dep)

		// Act
		sut.AddBarCode(codeContent, cell, &prop)

		// Assert
		cache.AssertNumberOfCalls(t, "GetImage", 1)
		code.AssertNumberOfCalls(t, "GenBar", 1)
		text.AssertNumberOfCalls(t, "Add", 1)
	})
	t.Run("when can find image on cache but cannot add image, should apply error message", func(t *testing.T) {
		// Arrange
		cell := &entity.Cell{}
		prop := fixture.BarcodeProp()

		img := &entity.Image{Bytes: []byte{1, 2, 3}}

		cache := &mocks.Cache{}
		cache.EXPECT().GetImage(codeContent, extension.Jpg).Return(img, nil)
		cache.EXPECT().AddImage(codeContent, img)

		text := &mocks.Text{}
		text.EXPECT().Add("could not add barcode to document", cell, merror.DefaultErrorText)

		cfg := &entity.Config{
			Margins: &entity.Margins{
				Left:   10,
				Top:    10,
				Right:  10,
				Bottom: 10,
			},
		}

		image := &mocks.Image{}
		image.EXPECT().Add(img, cell, cfg.Margins, prop.ToRectProp(), extension.Jpg, false).Return(errors.New("anyError"))

		fpdf := &mocks.Fpdf{}
		fpdf.EXPECT().ClearError()

		dep := &gofpdf.Dependencies{
			Cache: cache,
			Text:  text,
			Image: image,
			Fpdf:  fpdf,
			Cfg:   cfg,
		}

		sut := gofpdf.New(dep)

		// Act
		sut.AddBarCode(codeContent, cell, &prop)

		// Assert
		cache.AssertNumberOfCalls(t, "GetImage", 1)
		cache.AssertNumberOfCalls(t, "AddImage", 1)
		image.AssertNumberOfCalls(t, "Add", 1)
		text.AssertNumberOfCalls(t, "Add", 1)
	})
	t.Run("when can find image on cache and can add image, should not apply error message", func(t *testing.T) {
		// Arrange
		cell := &entity.Cell{}
		prop := fixture.BarcodeProp()

		img := &entity.Image{Bytes: []byte{1, 2, 3}}

		cache := &mocks.Cache{}
		cache.EXPECT().GetImage(codeContent, extension.Jpg).Return(img, nil)
		cache.EXPECT().AddImage(codeContent, img)

		cfg := &entity.Config{
			Margins: &entity.Margins{
				Left:   10,
				Top:    10,
				Right:  10,
				Bottom: 10,
			},
		}

		image := &mocks.Image{}
		image.EXPECT().Add(img, cell, cfg.Margins, prop.ToRectProp(), extension.Jpg, false).Return(nil)

		fpdf := &mocks.Fpdf{}
		fpdf.EXPECT().ClearError()

		dep := &gofpdf.Dependencies{
			Cache: cache,
			Image: image,
			Fpdf:  fpdf,
			Cfg:   cfg,
		}

		sut := gofpdf.New(dep)

		// Act
		sut.AddBarCode(codeContent, cell, &prop)

		// Assert
		cache.AssertNumberOfCalls(t, "GetImage", 1)
		cache.AssertNumberOfCalls(t, "AddImage", 1)
		image.AssertNumberOfCalls(t, "Add", 1)
	})
}

func TestProvider_CreateRow(t *testing.T) {
	// Arrange
	height := 10.0

	fpdf := &mocks.Fpdf{}
	fpdf.EXPECT().Ln(height)

	dep := &gofpdf.Dependencies{
		Fpdf: fpdf,
	}

	sut := gofpdf.New(dep)

	// Act
	sut.CreateRow(height)

	// Assert
	fpdf.AssertNumberOfCalls(t, "Ln", 1)
}

func TestProvider_CreateCol(t *testing.T) {
	// Arrange
	width := 10.0
	height := 20.0
	cfg := &entity.Config{}
	prop := fixture.CellProp()

	cellWriter := &mocks.CellWriter{}
	cellWriter.EXPECT().Apply(width, height, cfg, &prop)

	dep := &gofpdf.Dependencies{
		CellWriter: cellWriter,
	}

	sut := gofpdf.New(dep)

	// Act
	sut.CreateCol(width, height, cfg, &prop)

	// Assert
	cellWriter.AssertNumberOfCalls(t, "Apply", 1)
}

func TestProvider_SetProtection(t *testing.T) {
	t.Run("when protection is nil, should ignore protection", func(t *testing.T) {
		// Act
		dep := &gofpdf.Dependencies{}
		sut := gofpdf.New(dep)

		// Act
		sut.SetProtection(nil)
	})
	t.Run("when protection is valid, should apply protection", func(t *testing.T) {
		// Arrange
		p := &entity.Protection{
			Type:          protection.Print,
			UserPassword:  "userPassword",
			OwnerPassword: "ownerPassword",
		}

		fpdf := &mocks.Fpdf{}
		fpdf.EXPECT().SetProtection(byte(p.Type), p.UserPassword, p.OwnerPassword)

		dep := &gofpdf.Dependencies{
			Fpdf: fpdf,
		}

		sut := gofpdf.New(dep)

		// Act
		sut.SetProtection(p)

		// Assert
		fpdf.AssertNumberOfCalls(t, "SetProtection", 1)
	})
}

func TestProvider_SetCompression(t *testing.T) {
	// Arrange
	fpdf := &mocks.Fpdf{}
	fpdf.EXPECT().SetCompression(true)

	dep := &gofpdf.Dependencies{
		Fpdf: fpdf,
	}

	sut := gofpdf.New(dep)

	// Act
	sut.SetCompression(true)

	// Assert
	fpdf.AssertNumberOfCalls(t, "SetCompression", 1)
}

func TestProvider_SetMetadata(t *testing.T) {
	t.Run("when metadata is nil, should avoid process", func(t *testing.T) {
		// Arrange
		dep := &gofpdf.Dependencies{}

		sut := gofpdf.New(dep)

		// Act
		sut.SetMetadata(nil)
	})
	t.Run("when metadata is filled, should apply", func(t *testing.T) {
		// Arrange
		timeNow := time.Now()

		fpdf := &mocks.Fpdf{}
		fpdf.EXPECT().SetAuthor("author", true)
		fpdf.EXPECT().SetCreator("creator", true)
		fpdf.EXPECT().SetSubject("subject", true)
		fpdf.EXPECT().SetTitle("title", true)
		fpdf.EXPECT().SetCreationDate(timeNow)

		dep := &gofpdf.Dependencies{
			Fpdf: fpdf,
		}
		sut := gofpdf.New(dep)

		// Act
		sut.SetMetadata(&entity.Metadata{
			Author: &entity.Utf8Text{
				Text: "author",
				UTF8: true,
			},
			Creator: &entity.Utf8Text{
				Text: "creator",
				UTF8: true,
			},
			Subject: &entity.Utf8Text{
				Text: "subject",
				UTF8: true,
			},
			Title: &entity.Utf8Text{
				Text: "title",
				UTF8: true,
			},
			CreationDate: &timeNow,
		})

		// Assert
		fpdf.AssertNumberOfCalls(t, "SetAuthor", 1)
		fpdf.AssertNumberOfCalls(t, "SetCreator", 1)
		fpdf.AssertNumberOfCalls(t, "SetSubject", 1)
		fpdf.AssertNumberOfCalls(t, "SetTitle", 1)
		fpdf.AssertNumberOfCalls(t, "SetCreationDate", 1)
	})
}

func TestProvider_GenerateBytes(t *testing.T) {
	// Arrange
	fpdf := &mocks.Fpdf{}
	fpdf.EXPECT().Output(mock.Anything).Return(errors.New("anyError"))

	dep := &gofpdf.Dependencies{
		Fpdf: fpdf,
	}
	sut := gofpdf.New(dep)

	// Act
	bytes, err := sut.GenerateBytes()

	// Assert
	assert.Nil(t, bytes)
	assert.NotNil(t, err)
	fpdf.AssertNumberOfCalls(t, "Output", 1)
}

func TestProvider_AddImageFromBytes(t *testing.T) {
	t.Run("when image is invalid, should apply message error", func(t *testing.T) {
		// Arrange
		prop := fixture.RectProp()
		cell := &entity.Cell{}

		text := &mocks.Text{}
		text.EXPECT().Add("could not parse image bytes", cell, merror.DefaultErrorText)

		dep := &gofpdf.Dependencies{
			Text: text,
		}

		sut := gofpdf.New(dep)

		// Act
		sut.AddImageFromBytes([]byte{1, 2, 3}, cell, &prop, "invalid")

		// Assert
		text.AssertNumberOfCalls(t, "Add", 1)
	})
	t.Run("when image is valid but cannot add to document, should apply message error", func(t *testing.T) {
		// Arrange
		img := &entity.Image{
			Bytes:     []byte{1, 2, 3},
			Extension: extension.Jpg,
		}
		prop := fixture.RectProp()
		cell := &entity.Cell{}

		cfg := &entity.Config{
			Margins: &entity.Margins{
				Left:   10,
				Top:    10,
				Right:  10,
				Bottom: 10,
			},
		}

		text := &mocks.Text{}
		text.EXPECT().Add("could not add image to document", cell, merror.DefaultErrorText)

		image := &mocks.Image{}
		image.EXPECT().Add(img, cell, cfg.Margins, &prop, img.Extension, false).Return(errors.New("anyError"))

		fpdf := &mocks.Fpdf{}
		fpdf.EXPECT().ClearError()

		dep := &gofpdf.Dependencies{
			Text:  text,
			Image: image,
			Fpdf:  fpdf,
			Cfg:   cfg,
		}

		sut := gofpdf.New(dep)

		// Act
		sut.AddImageFromBytes(img.Bytes, cell, &prop, img.Extension)

		// Assert
		text.AssertNumberOfCalls(t, "Add", 1)
		image.AssertNumberOfCalls(t, "Add", 1)
		fpdf.AssertNumberOfCalls(t, "ClearError", 1)
	})
	t.Run("when image is valid and can add to document, should not apply", func(t *testing.T) {
		// Arrange
		img := &entity.Image{
			Bytes:     []byte{1, 2, 3},
			Extension: extension.Jpg,
		}
		prop := fixture.RectProp()
		cell := &entity.Cell{}

		cfg := &entity.Config{
			Margins: &entity.Margins{
				Left:   10,
				Top:    10,
				Right:  10,
				Bottom: 10,
			},
		}

		image := &mocks.Image{}
		image.EXPECT().Add(img, cell, cfg.Margins, &prop, img.Extension, false).Return(nil)

		dep := &gofpdf.Dependencies{
			Image: image,
			Cfg:   cfg,
		}

		sut := gofpdf.New(dep)

		// Act
		sut.AddImageFromBytes(img.Bytes, cell, &prop, img.Extension)

		// Assert
		image.AssertNumberOfCalls(t, "Add", 1)
	})
}

func TestProvider_AddBackgroundImageFromBytes(t *testing.T) {
	t.Run("when image is invalid, should apply message error", func(t *testing.T) {
		// Arrange
		prop := fixture.RectProp()
		cell := &entity.Cell{}

		text := &mocks.Text{}
		text.EXPECT().Add("could not parse image bytes", cell, merror.DefaultErrorText)

		dep := &gofpdf.Dependencies{
			Text: text,
		}

		sut := gofpdf.New(dep)

		// Act
		sut.AddBackgroundImageFromBytes([]byte{1, 2, 3}, cell, &prop, "invalid")

		// Assert
		text.AssertNumberOfCalls(t, "Add", 1)
	})
	t.Run("when image is valid but cannot add to document, should apply message error", func(t *testing.T) {
		// Arrange
		img := &entity.Image{
			Bytes:     []byte{1, 2, 3},
			Extension: extension.Jpg,
		}
		prop := fixture.RectProp()
		cell := &entity.Cell{}

		cfg := &entity.Config{
			Margins: &entity.Margins{
				Left:   10,
				Top:    10,
				Right:  10,
				Bottom: 10,
			},
		}

		text := &mocks.Text{}
		text.EXPECT().Add("could not add image to document", cell, merror.DefaultErrorText)

		image := &mocks.Image{}
		image.EXPECT().Add(img, cell, cfg.Margins, &prop, img.Extension, true).Return(errors.New("anyError"))

		fpdf := &mocks.Fpdf{}
		fpdf.EXPECT().ClearError()
		fpdf.EXPECT().SetHomeXY()

		dep := &gofpdf.Dependencies{
			Text:  text,
			Image: image,
			Fpdf:  fpdf,
			Cfg:   cfg,
		}

		sut := gofpdf.New(dep)

		// Act
		sut.AddBackgroundImageFromBytes(img.Bytes, cell, &prop, img.Extension)

		// Assert
		text.AssertNumberOfCalls(t, "Add", 1)
		image.AssertNumberOfCalls(t, "Add", 1)
		fpdf.AssertNumberOfCalls(t, "ClearError", 1)
		fpdf.AssertNumberOfCalls(t, "SetHomeXY", 1)
	})
	t.Run("when image is valid and can add to document, should not apply message error", func(t *testing.T) {
		// Arrange
		img := &entity.Image{
			Bytes:     []byte{1, 2, 3},
			Extension: extension.Jpg,
		}
		prop := fixture.RectProp()
		cell := &entity.Cell{}

		cfg := &entity.Config{
			Margins: &entity.Margins{
				Left:   10,
				Top:    10,
				Right:  10,
				Bottom: 10,
			},
		}

		image := &mocks.Image{}
		image.EXPECT().Add(img, cell, cfg.Margins, &prop, img.Extension, true).Return(nil)

		fpdf := &mocks.Fpdf{}
		fpdf.EXPECT().SetHomeXY()

		dep := &gofpdf.Dependencies{
			Image: image,
			Fpdf:  fpdf,
			Cfg:   cfg,
		}

		sut := gofpdf.New(dep)

		// Act
		sut.AddBackgroundImageFromBytes(img.Bytes, cell, &prop, img.Extension)

		// Assert
		image.AssertNumberOfCalls(t, "Add", 1)
		fpdf.AssertNumberOfCalls(t, "SetHomeXY", 1)
	})
}

/*func TestProvider_AddImageFromFile(t *testing.T) {
	t.Run("when cannot find image in cache and cannot load image, should apply error message", func(t *testing.T) {
		// Arrange
		file := "file.jpg"
		cell := &entity.Cell{}
		prop := fixture.RectProp()

		cache := &mocks.Cache{}
		cache.EXPECT().GetImage(file, extension.Jpg)

		dep := &gofpdf.Dependencies{
			Cache: cache,
		}

		sut := gofpdf.New(dep)

		// Act
		sut.AddImageFromFile(file, cell, &prop)
	})
}
*/
