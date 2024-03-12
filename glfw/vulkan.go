package glfw

import (
	"unsafe"
)

// VulkanSupported reports whether the Vulkan loader has been found. This check is performed by Init.
//
// The availability of a Vulkan loader does not by itself guarantee that window surface creation or
// even device creation is possible. Call GetRequiredInstanceExtensions to check whether the
// extensions necessary for Vulkan surface creation are available and GetPhysicalDevicePresentationSupport
// to check whether a queue family of a physical device supports image presentation.
func VulkanSupported() bool {
	return true
}

// GetVulkanGetInstanceProcAddress returns the function pointer used to find Vulkan core or
// extension functions. The return value of this function can be passed to the Vulkan library.
//
// Note that this function does not work the same way as the glfwGetInstanceProcAddress.
func GetVulkanGetInstanceProcAddress() unsafe.Pointer {
	return nil
}

// GetRequiredInstanceExtensions returns a slice of Vulkan instance extension names required
// by GLFW for creating Vulkan surfaces for GLFW windows. If successful, the list will always
// contain VK_KHR_surface, so if you don't require any additional extensions you can pass this list
// directly to the VkInstanceCreateInfo struct.
//
// If Vulkan is not available on the machine, this function returns nil. Call
// VulkanSupported to check whether Vulkan is available.
//
// If Vulkan is available but no set of extensions allowing window surface creation was found, this
// function returns nil. You may still use Vulkan for off-screen rendering and compute work.
func (window *Window) GetRequiredInstanceExtensions() []string {
	return nil
}

// CreateWindowSurface creates a Vulkan surface for this window.
func (window *Window) CreateWindowSurface(instance interface{}, allocCallbacks unsafe.Pointer) (surface uintptr, err error) {
	return
}
