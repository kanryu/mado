# What is Mado's architecture?

The question of what Mado's architecture should be is an essential question.

At the initial stage, the functionality provided by the Mado package is almost the same as that of gioui, with the exception of some APIs that were unavoidably added to support glfw compatibility mode.

Although I think that the API of the Mado package should eventually mature and be used by many external products, for the time being we will prioritize support for the glfw layer as a stub.

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

