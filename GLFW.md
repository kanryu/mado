#　What does GLFW mean to Mado?

For Mado, GLFW is the original purpose of development, a reference, and ultimately a competitor that must be overcome.

GLFW has had a critical problem for many years with the lack of input support for CJK (Chinese, Japanese Korean) languages, and has not shown any improvement from 2015 to March 2024.

They claim that this issue will be resolved by the release of GLFW-2.5, but I know this is not possible.
This is because the contributions of the [clear-node/glfw](https://github.com/clear-code/glfw) team, which has been trying to add IME support to glfw for many years, are also incomplete.

## GLFW compatibility layer in mado

glfw in Mado actually refers to (go-gl/glfw)[https://github.com/go-gl/glfw], not (glfw/glfw)[https://github.com/glfw/glfw].
This is because Mado is a library for Go language users and is not intended to be ported to other languages.

go-gl/glfw is essentially a Go wrapper for glfw and should expose the C API, but it doesn't.

In particular, Window is objectified as glfw.Window, and the glfw API and many callbacks related to Window are provided through this object.

mado/glfw has attempted to be a complete port of glfw.Window from the beginning, and currently has a fairly high level of compatibility.

On the other hand, most of the glfw APIs that are not related to Windows are still unimplemented, and the glfw applications that many users have written so far will not work.
But don't worry. The tasks will be cleared one by one.

## OpenGL environment provided by GLFW

When you write an application that makes any use of OpenGL without GLFW or glut support, you are faced with a frightening gap.

Historically, GLFW's OpenGL initialization routines are among the best maintained and most used programs in the world.

Initially, when I forked gioui wasn't accounted for that OpenGL insanity, so the OpenGL triangles weren't drawn at all.

I searched for other promising implementations in this field and discovered [Ebitengine](https://github.com/hajimehoshi/ebiten). I was impressed with how well it was organized, but when it came to the important thing about his OpenGL initialization, he incorporated GLFW. However, it had been heavily ported to Go, so he decided to port this implementation to Mado.

I installed Windows and Mac by the first release of Mado.
I've verified that the triangle drawing sample works, but Linux support is still in its infancy, and there's still a lot of work to be done to provide the kind of support that GLFW provides for OpenGL. This is a topic that you who actually draw using Mado know more about than me who created Mado.

Please post issues or pull requests on the Mado project. We welcome your cooperation.


## Problems with IME products of clear-node/glfw

I have previously tried implementing [Fyne](https://github.com/fyne-io/fyne)'s IME support using clear-node/glfw.
At that time, I faced many problems and realized that I could not solve them this way.

### Does not support MS-Windows windowed rendering.

Direct3D/OpenGL has Full screen Mode and Windowed Mode, and in Full screen Mode, the application must perform all drawing, including drawing the IME, so their implementation is correct.

However, for most other uses, including original toolkits, applications and the OS work together to support IME. In other words, the IME application provided by OS is drawn with the coordinates superimposed on the application window. clear-node/glfw cannot make any assumptions about this style.

### Incomplete support for MacOS

Their implementation cannot receive IME ON-Off events.
More precisely, their events are redundant, and turning the IME ON once will cause the IME ON event to be clicked 8 times.

### Incomplete support for Linux X11

An event is not issued when the IME is turned on. This event is an important moment for applications to specify what coordinates the IME text cursor should be at, and it is difficult to implement IME support without this event. Strictly speaking, the preedit text timing is altanative, but since the preedit text event is issued every time the user inputs text through the IME, it would be excessive timing to consider the coordinate calculation of the IME text cursor. IME Window Redrawing is a fairly expensive calculation.

### Not at all for Linux Wayland

Wayland's text input is not considered at all.

So, I have decided that it is impossible to continue to maintain the glfw/glfw implementation with the aim of improving it, including the many problems with the X11 event handler implementation.

## GIOUI

I considered alternatives to solving this problem and ultimately decided that [gioui](https://github.com/gioui/gio) was the most appropriate starting point.

- Window support for Windows, Mac, Linux X11, Linux Wayland, Android, iOS
- Support for OpenGL, OpenGL ES, Vulcan, Metal, Direct3D11
- Have a high level of IME support in each environment
- Without GLFW and has a large proportion of Go code
