package main

import (
	"log"

	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"
	"github.com/johnfercher/maroto/v2/pkg/repository"

	"github.com/johnfercher/maroto/v2"

	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
)

func main() {
	m := GetMaroto("docs/assets/fonts/Cairo-Regular.ttf")
	document, err := m.Generate()
	if err != nil {
		log.Fatal(err.Error())
	}

	err = document.Save("docs/assets/pdf/rtlmodev2.pdf")
	if err != nil {
		log.Fatal(err.Error())
	}

	err = document.GetReport().Save("docs/assets/text/rtlmodev2.txt")
	if err != nil {
		log.Fatal(err.Error())
	}
}

func GetMaroto(customFontFile string) core.Maroto {
	customFont := "cairo"

	customFonts, err := repository.New().
		AddUTF8Font(customFont, fontstyle.Normal, customFontFile).
		AddUTF8Font(customFont, fontstyle.Italic, customFontFile).
		AddUTF8Font(customFont, fontstyle.Bold, customFontFile).
		AddUTF8Font(customFont, fontstyle.BoldItalic, customFontFile).
		Load()
	if err != nil {
		log.Fatal(err.Error())
	}

	cfg := config.NewBuilder().
		WithCustomFonts(customFonts).
		WithRTLMode(true).
		Build()

	mrt := maroto.New(cfg)
	m := maroto.NewMetricsDecorator(mrt)

	err = m.RegisterHeader(text.NewRow(20, "مرحبا بالعالم", props.Text{
		Size:   12,
		Align:  align.Right,
		Style:  fontstyle.Bold,
		Family: "cairo",
	}))
	if err != nil {
		log.Fatal(err)
	}

	return m
}
