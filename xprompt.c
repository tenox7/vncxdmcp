// xprompt draws a one line text input box and prints what was typed to
// stdout. It exists because every ready made X dialog is either gone from
// modern distributions (xdialog, gxmessage) or would drag a full GTK stack
// into the image (zenity, yad); this needs nothing but libX11, which is
// already here. With no window manager running the window places itself and
// takes the input focus by hand, since PointerRoot focus would otherwise
// send the keystrokes to whatever sits under the pointer.

#include <X11/Xlib.h>
#include <X11/Xutil.h>
#include <X11/keysym.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#define MAXLEN 255
#define PAD 12
#define COLS 40

static const char *fontnames[] = {"10x20", "9x15", "fixed", NULL};

int main(int argc, char **argv) {
	const char *prompt = argc > 1 ? argv[1] : "Input:";
	char text[MAXLEN + 1] = "";
	int len = 0, focused = 0;

	Display *d = XOpenDisplay(NULL);
	if (!d) {
		fprintf(stderr, "xprompt: cannot open display\n");
		return 2;
	}
	int s = DefaultScreen(d);

	XFontStruct *f = NULL;
	for (const char **n = fontnames; *n && !f; n++)
		f = XLoadQueryFont(d, *n);
	if (!f) {
		fprintf(stderr, "xprompt: no usable font\n");
		return 2;
	}

	int fw = f->max_bounds.width, fh = f->ascent + f->descent;
	size_t cols = strlen(prompt) > COLS ? strlen(prompt) : COLS;
	int w = cols * fw + 4 * PAD, h = 2 * fh + 5 * PAD;
	int fieldy = 2 * PAD + fh;

	XSetWindowAttributes a;
	a.override_redirect = True;
	a.background_pixel = WhitePixel(d, s);
	a.border_pixel = BlackPixel(d, s);
	a.event_mask = ExposureMask | KeyPressMask;
	Window win = XCreateWindow(d, RootWindow(d, s),
		(DisplayWidth(d, s) - w) / 2, (DisplayHeight(d, s) - h) / 2, w, h, 2,
		CopyFromParent, InputOutput, CopyFromParent,
		CWOverrideRedirect | CWBackPixel | CWBorderPixel | CWEventMask, &a);
	GC gc = XCreateGC(d, win, 0, NULL);
	XSetFont(d, gc, f->fid);
	XSetForeground(d, gc, BlackPixel(d, s));
	XMapRaised(d, win);

	for (;;) {
		XEvent e;
		XNextEvent(d, &e);
		if (e.type == Expose) {
			XClearWindow(d, win);
			XDrawString(d, win, gc, 2 * PAD, PAD + f->ascent, prompt, strlen(prompt));
			XDrawRectangle(d, win, gc, PAD, fieldy, w - 2 * PAD, fh + 2 * PAD);
			XDrawString(d, win, gc, 2 * PAD, fieldy + PAD + f->ascent, text, len);
			XFillRectangle(d, win, gc, 2 * PAD + XTextWidth(f, text, len),
				fieldy + PAD, fw, fh);
			if (!focused) { // only legal once the window is viewable
				XSetInputFocus(d, win, RevertToPointerRoot, CurrentTime);
				focused = XGrabKeyboard(d, win, True, GrabModeAsync,
					GrabModeAsync, CurrentTime) == GrabSuccess;
			}
			continue;
		}
		if (e.type != KeyPress)
			continue;
		KeySym ks;
		char buf[8];
		int n = XLookupString(&e.xkey, buf, sizeof buf, &ks, NULL);
		if (ks == XK_Return || ks == XK_KP_Enter)
			break;
		if (ks == XK_Escape)
			return 1;
		if (ks == XK_BackSpace || ks == XK_Delete) {
			if (len)
				text[--len] = '\0';
		} else if (n == 1 && buf[0] >= ' ' && buf[0] < 127 && len < MAXLEN) {
			text[len++] = buf[0];
			text[len] = '\0';
		}
		XClearArea(d, win, 0, 0, 0, 0, True);
	}
	printf("%s\n", text);
	return 0;
}
