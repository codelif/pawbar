#include <librsvg/rsvg.h>
#include <cairo/cairo.h>
#include <stdlib.h>
#include <string.h>
#include <math.h>

typedef struct {
    unsigned char* data;
    int width;
    int height;
    int stride;
} render_result_t;


// Renders SVG with applied custom CSS
render_result_t* render_svg_css(const char* svg_data, int svg_len, const char* css, int width, int height);

// Frees result
void free_render_result(render_result_t *result);



int set_svg_stylesheet(RsvgHandle *handle, const char *css);