package main

import (
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
)

const (
	cells   = 100
	xyrange = 30.0
	angle   = math.Pi / 6
)

var width, height float64
var peakColor, lowlandColor string
var formType string
var sin30, cos30 = math.Sin(angle), math.Cos(angle)

func main() {
	http.HandleFunc("/", showFigure)
	http.HandleFunc("/saddle", showFigure)
	http.HandleFunc("/climb", showFigure)
	http.HandleFunc("/egg", showFigure)
	log.Fatal(http.ListenAndServe("127.0.0.1:26221", nil))
}

func showFigure(w http.ResponseWriter, r *http.Request) {
	if strings.Contains(r.URL.RequestURI(), "egg") {
		formType = "egg"
	} else if strings.Contains(r.URL.RequestURI(), "saddle") {
		formType = "saddle"
	} else if strings.Contains(r.URL.RequestURI(), "climb") {
		formType = "climb"
	} else {
		formType = "drop"
	}

	getParams(r)

	w.Header().Set("Content-Type", "image/svg+xml")

	createFigure(w)
}

func createFigure(out io.Writer) {
	_, _ = fmt.Fprintf(out, "<svg xmlns='http://www.w3.org/2000/svg' "+
		"style='stroke: grey; fill: white; stroke-width: 0.7' "+
		"width='%d' height='%d'>", int(width), int(height))

	for i := 0; i < cells; i++ {
		for j := 0; j < cells; j++ {
			ax, ay, ct := corner(i+1, j)
			bx, by, ct1 := corner(i, j)
			cx, cy, ct2 := corner(i, j+1)
			dx, dy, ct3 := corner(i+1, j+1)

			var color string

			switch {
			case ct == 1 || ct1 == 1 || ct2 == 1 || ct3 == 1:
				color = peakColor
			case ct == 2 || ct1 == 2 || ct2 == 2 || ct3 == 2:
				color = lowlandColor
			case ct == 0 || ct1 == 0 || ct2 == 0 || ct3 == 0:
				color = "#00ff00"
			default:
				color = "#00ff00"
			}
			_, _ = fmt.Fprintf(out, "<polygon fill='%s' stroke-width='0.4' points='%g,%g %g,%g %g,%g %g,%g'/>\n",
				color, ax, ay, bx, by, cx, cy, dx, dy)
		}
	}

	_, _ = fmt.Fprintln(out, "</svg>")
}

func getParams(r *http.Request) {
	width = 600
	if canvasWidth, err := strconv.ParseFloat(r.URL.Query().Get("w"), 64); err == nil {
		width = canvasWidth
	}

	height = 320
	if canvasHeight, err := strconv.ParseFloat(r.URL.Query().Get("h"), 64); err == nil {
		height = canvasHeight
	}

	peakColor = "#0000FF"
	if pc := r.URL.Query().Get("pc"); pc != "" {
		peakColor = pc
	}

	lowlandColor = "#FF0000"
	if lc := r.URL.Query().Get("lc"); lc != "" {
		lowlandColor = lc
	}
}

func corner(i, j int) (float64, float64, int) {
	x := xyrange * (float64(i)/cells - 0.5)
	y := xyrange * (float64(j)/cells - 0.5)
	z, ct := f(x, y)

	xyscale, zscale := calcSize(width, height)

	sx := width/2 + (x-y)*cos30*xyscale
	sy := height/2 + (x+y)*sin30*xyscale - z*zscale

	return sx, sy, ct
}

func calcSize(w, h float64) (xyscale, zscale float64) {
	return w / 2 / xyrange,
		h * 0.4
}

func f(x, y float64) (float64, int) {
	var out float64
	switch formType {
	case "drop":
		out = defaultView(x, y)
	case "climb":
		out = climb(x, y)
	case "saddle":
		out = saddle(x, y)
	case "egg":
		out = eggform(x, y)
	}

	ct := 0

	if out < 0. {
		ct = 2
	} else if out > 0. {
		ct = 1
	} else {
		ct = 0
	}

	return out, ct

}

func defaultView(x, y float64) float64 {
	r := math.Hypot(x, y)
	return math.Sin(r) / r
}

func climb(x, y float64) float64 {
	return (math.Sin(x) / x) * (math.Sin(y) / y)
}

func eggform(x, y float64) float64 {
	return math.Pow(2, math.Sin(x)) * math.Pow(2, math.Sin(y)) / 12
}

func saddle(x, y float64) float64 {
	return math.Pow(x, 2)/math.Pow(25, 2) - math.Pow(y, 2)/math.Pow(17, 2)
}
