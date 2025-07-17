#include "librsvg_wrapper.h"


// Render function with CSS support
render_result_t* render_svg_css(const char* svg_data, int svg_len, const char* css, int width, int height) {
    GError *error = NULL;
    RsvgHandle *handle = rsvg_handle_new_from_data((const guint8*)svg_data, svg_len, &error);
    if (error || !handle) {
        if (error) g_error_free(error);
        return NULL;
    }
    // Apply CSS if provided
    if (css && strlen(css) > 0) {
        if (!set_svg_stylesheet(handle, css)) {
            g_object_unref(handle);
            return NULL;
        }
    }


    gdouble svg_width, svg_height;
    RsvgRectangle viewport;

    if (!rsvg_handle_get_intrinsic_size_in_pixels(handle, &svg_width, &svg_height)) {
        // fallback
        viewport.x = 0;
        viewport.y = 0;
        viewport.width = width > 0 ? width : 100;
        viewport.height = height > 0 ? height : 100;

        RsvgRectangle ink_rect, logical_rect;
        if (rsvg_handle_get_geometry_for_layer(handle, NULL, &viewport, &ink_rect, &logical_rect, &error)) {
            svg_width = logical_rect.width;
            svg_height = logical_rect.height;
        } else {
            // last resort: assume its the provided dim
            svg_width = width > 0 ? width : 100;
            svg_height = height > 0 ? height : 100;
            if (error) {
                g_error_free(error);
                error = NULL;
            }
        }

    }


    int target_width = width > 0 ? width : (int)ceil(svg_width);
    int target_height = height > 0 ? height : (int)ceil(svg_height);

    // Create cairo surface
    cairo_surface_t *surface = cairo_image_surface_create(CAIRO_FORMAT_ARGB32, target_width, target_height);
    cairo_t *cr = cairo_create(surface);

    // Clear to transparent
    cairo_set_operator(cr, CAIRO_OPERATOR_CLEAR);
    cairo_paint(cr);
    cairo_set_operator(cr, CAIRO_OPERATOR_OVER);


    viewport.x = 0;
    viewport.y = 0;
    viewport.width = target_width;
    viewport.height = target_height;

    // Scale if needed
    if (width > 0 && height > 0) {
        double scale_x = (double)target_width / svg_width;
        double scale_y = (double)target_height / svg_height;
        cairo_scale(cr, scale_x, scale_y);

        viewport.width = svg_width;
        viewport.height = svg_height;
    }

    // Render SVG
    // gboolean success = rsvg_handle_render_cairo(handle, cr);
    gboolean success = rsvg_handle_render_document(handle, cr, &viewport, &error);
    if (error) {
        g_error_free(error);
        success = FALSE;
    }
    cairo_destroy(cr);

    if (!success) {
        cairo_surface_destroy(surface);
        g_object_unref(handle);
        return NULL;
    }

    // Get surface data
    cairo_surface_flush(surface);
    unsigned char *surface_data = cairo_image_surface_get_data(surface);


    int surface_width = cairo_image_surface_get_width(surface);
    int surface_height = cairo_image_surface_get_height(surface);
    int stride = cairo_image_surface_get_stride(surface);

    // Allocate result
    render_result_t *result = malloc(sizeof(render_result_t));
    result->width = surface_width;
    result->height = surface_height;
    result->stride = stride;

    int data_size = stride * surface_height;
    result->data = malloc(data_size);


// Convert BGRA to RGBA in-place
for (int i = 0; i < data_size; i += 4) {
    unsigned char temp = surface_data[i];     // Save B
    surface_data[i] = surface_data[i + 2];    // B = R
    surface_data[i + 2] = temp;               // R = B
    // G and A stay in the same positions
}

    memcpy(result->data, surface_data, data_size);

    // Cleanup
    cairo_surface_destroy(surface);
    g_object_unref(handle);

    return result;
}


void free_render_result(render_result_t *result) {
    if (result) {
        if (result->data) {
            free(result->data);
        }
        free(result);
    }
}

// Helper to set CSS style data for symbolic icon recoloring
int set_svg_stylesheet(RsvgHandle *handle, const char *css) {
    GError *error = NULL;
    gboolean success = rsvg_handle_set_stylesheet(handle, (const guint8*)css, strlen(css), &error);
    if (error) {
        g_error_free(error);
        return 0;
    }
    return success ? 1 : 0;
}
