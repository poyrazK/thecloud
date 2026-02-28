# ADR-016: ARM64/UEFI Support for Libvirt Compute

## Status
Accepted

## Context
To support modern compute workloads and provide parity with public cloud providers, the platform needs to support ARM64 architecture for Virtual Machines. ARM64 virtualization in QEMU/Libvirt differs significantly from x86_64, particularly regarding machine types and firmware.

Standard x86_64 VMs often use BIOS or Legacy boot, while ARM64 standardizes on UEFI (`AAVMF`). Additionally, ARM64 requires the `virt` machine type, which does not support certain legacy features like ACPI or APIC in the same way as `pc` or `q35`.

## Decision
We will enable ARM64/UEFI support in the Libvirt adapter.

### Technical Details
1.  **Machine Type**: For ARM64, we use the `virt` machine type.
2.  **Firmware**: We introduce a UEFI loader configuration pointing to the standard `AAVMF_CODE.fd` firmware path.
3.  **Feature Flags**: We conditionally disable ACPI and APIC when UEFI is not used or when running specific ARM64 cloud images that expect a pure device-tree or UEFI environment.
4.  **Template Modernization**: Updated the Libvirt XML templates to support conditional logic based on the host/guest architecture.

## Consequences
- **Pros**: Allows the platform to run on ARM64 hardware (e.g., Apple Silicon via TCG, Graviton, Ampere). Better compatibility with modern Linux distributions.
- **Cons**: Requires UEFI firmware packages to be installed on the host. Slightly more complex Libvirt XML templates.
- **Portability**: Cloud images must now be architecture-aware.
