package ko

//$ file xfs.ko
//xfs.ko: ELF 64-bit LSB relocatable, x86-64, version 1 (SYSV), BuildID[sha1]=bcb5e287509cedbb0c5ece383e0b97fb99e4781e, not stripped
//
//$ readelf -h xfs.ko
//ELF Header:
//  Magic:   7f 45 4c 46 02 01 01 00 00 00 00 00 00 00 00 00
//  Class:                             ELF64
//  Data:                              2's complement, little endian
//  Version:                           1 (current)
//  OS/ABI:                            UNIX - System V
//  ABI Version:                       0
//  Type:                              REL (Relocatable file)
//  Machine:                           Advanced Micro Devices X86-64
//  Version:                           0x1
//  Entry point address:               0x0
//  Start of program headers:          0 (bytes into file)
//  Start of section headers:          1829088 (bytes into file)
//  Flags:                             0x0
//  Size of this header:               64 (bytes)
//  Size of program headers:           0 (bytes)
//  Number of program headers:         0
//  Size of section headers:           64 (bytes)
//  Number of section headers:         45
//  Section header string table index: 44
//
//For the kernel, an easy way is by compiling one and looking at vmlinux:
//
//$ file vmlinux
//vmlinux: ELF 64-bit LSB executable, x86-64, version 1 (SYSV), statically linked, BuildID[sha1]=eaf006a7ccfedbc40a6feddb04088bdb2ef0112f, with debug_info, not stripped
//
//$ readelf -h vmlinux
//ELF Header:
//  Magic:   7f 45 4c 46 02 01 01 00 00 00 00 00 00 00 00 00
//  Class:                             ELF64
//  Data:                              2's complement, little endian
//  Version:                           1 (current)
//  OS/ABI:                            UNIX - System V
//  ABI Version:                       0
//  Type:                              EXEC (Executable file) ## NOTE:FIRST DIFF
//  Machine:                           Advanced Micro Devices X86-64
//  Version:                           0x1
//  Entry point address:               0x1000000 ## NOTE: SECOND DIFF 0x0
//  Start of program headers:          64 (bytes into file)
//  Start of section headers:          171602920 (bytes into file) ## NOTE:DIF
//  Flags:                             0x0
//  Size of this header:               64 (bytes)
//  Size of program headers:           56 (bytes) ## NOTE: DIFF 0 bytes on ko
//  Number of program headers:         5          ## NOTE: DIFF 0
//  Size of section headers:           64 (bytes)
//  Number of section headers:         43         ## NOTE: DEFF 45 on ko
//  Section header string table index: 42         ## NOTE: DIFF 45 on ko
//

// e_shoff  +------------------+
//     +----+ ELF header       |
//     |    +------------------+ <------+
//     |    |                  |        |
//     |    | section 1        |        |
//     |    |                  |        |
//     |    +------------------+        |
//     |    | section 2        | <---+  |
//     |    +------------------+     |  |
//     |    | section 3        | <+  |  |
//     +--> +------------------+  |  |  |
//         | section header 1 +--------+
//         +------------------+  |  |
//         | section header 2 +-----+
//         +------------------+  | sh_offset
//         | section header 3 +--+
//         +------------------+

//typedef struct
//{
//    unsigned char e_ident[16]; /* ELF identification */
//    Elf64_Half e_type; /* Object file type */
//    Elf64_Half e_machine; /* Machine type */
//    Elf64_Word e_version; /* Object file version */
//    Elf64_Addr e_entry; /* Entry point address */
//    Elf64_Off e_phoff; /* Program header offset */
//    Elf64_Off e_shoff; /* Section header offset */
//    Elf64_Word e_flags; /* Processor-specific flags */
//    Elf64_Half e_ehsize; /* ELF header size */
//    Elf64_Half e_phentsize; /* Size of program header entry */
//    Elf64_Half e_phnum; /* Number of program header entries */
//    Elf64_Half e_shentsize; /* Size of section header entry */
//    Elf64_Half e_shnum; /* Number of section header entries */
//    Elf64_Half e_shstrndx; /* Section name string table index */
//
//} Elf64_Ehdr; The comment describe every filed's meaning.

// This shows the EXPORT_SYMBOL definition. Though seems complicated, we will uses our example to instantiate it. Think our EXPORT_SYMBOL(testexport). After expand this macro, we get this(the __CRC_SYMBOL(sym, sec) is left later):
//
// static const char __kstrtab_testexport[] = "testexport";
// const struct kernel_symbol __ksymtab_testexport =
// {(unsigned long)&testexport, __kstrtab_testexport}
// The second structure represented:
// struct kernel_symbol
// {
// unsigned long value;
// const char *name;
// };
//
// So here we can see, the EXPORT_SYMBOL just define variables, the ‘value’ is the address of this symbol in memory and ‘name’ is the name of this symbol. Not like ordinary defination, the export function’s name is stored in section “__ksymtab_strings”, and the kernel_symbol variable is stored in section “___ksymtab+testexport”. If you look at the ELF file section, you will not find “___ksymtab+testexport” section. It is converted in “__ksymtab” in <scripts/module-common.lds>:
//
// SECTIONS {
// /DISCARD/ : { *(.discard) }
//
// __ksymtab  : { *(SORT(___ksymtab+*)) }
// __ksymtab_gpl  : { *(SORT(___ksymtab_gpl+*)) }
// __ksymtab_unused : { *(SORT(___ksymtab_unused+*)) }
// __ksymtab_unused_gpl : { *(SORT(___ksymtab_unused_gpl+*)) }
// __ksymtab_gpl_future : { *(SORT(___ksymtab_gpl_future+*)) }
// __kcrctab  : { *(SORT(___kcrctab+*)) }
// __kcrctab_gpl  : { *(SORT(___kcrctab_gpl+*)) }
// __kcrctab_unused : { *(SORT(___kcrctab_unused+*)) }
// __kcrctab_unused_gpl : { *(SORT(___kcrctab_unused_gpl+*)) }
// __kcrctab_gpl_future : { *(SORT(___kcrctab_gpl_future+*)) }
// }
//
// As for EXPORT_SYMBOL_GPL and EXPORT_SYMBOL_GPL_FUTURE, the only difference is the section added by “_gpl” and “_gpl_future”. In order to let the kernel uses these sections to find the exported symbol, the linker must export the address of these section. See <include/asm-generic/vmlinux.lds.h>:
//
// /* Kernel symbol table: Normal symbols */   \
// __ksymtab         : AT(ADDR(__ksymtab) - LOAD_OFFSET) {  \
//   VMLINUX_SYMBOL(__start___ksymtab) = .;   \
//   *(SORT(___ksymtab+*))     \
//   VMLINUX_SYMBOL(__stop___ksymtab) = .;   \
// }        \
//         \
// /* Kernel symbol table: GPL-only symbols */   \
// __ksymtab_gpl     : AT(ADDR(__ksymtab_gpl) - LOAD_OFFSET) { \
//   VMLINUX_SYMBOL(__start___ksymtab_gpl) = .;  \
//   *(SORT(___ksymtab_gpl+*))    \
//   VMLINUX_SYMBOL(__stop___ksymtab_gpl) = .;  \
// }        \
//         \
// /* Kernel symbol table: Normal unused symbols */  \
// __ksymtab_unused  : AT(ADDR(__ksymtab_unused) - LOAD_OFFSET) { \
//   VMLINUX_SYMBOL(__start___ksymtab_unused) = .;  \
//   *(SORT(___ksymtab_unused+*))    \
//   VMLINUX_SYMBOL(__stop___ksymtab_unused) = .;  \
// }        \
// ...
