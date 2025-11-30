package main

import "fyne.io/fyne/v2"

// Star icon (outline) - uses path to draw outline shape that Fyne can theme
var resourceStarSvg = &fyne.StaticResource{
	StaticName:    "star.svg",
	StaticContent: []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path fill="#000000" d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2zm0 2.82L9.69 9.63l-5.28.77 3.82 3.72-.9 5.26L12 16.89l4.67 2.49-.9-5.26 3.82-3.72-5.28-.77L12 4.82z"/></svg>`),
}

// Star icon (filled)
var resourceStarFilledSvg = &fyne.StaticResource{
	StaticName:    "star-filled.svg",
	StaticContent: []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path fill="#000000" d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"/></svg>`),
}
