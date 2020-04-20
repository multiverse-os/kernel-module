# Building the Kernel Module

```C
In <kernel/module.c> we can see the declaration:

/* Provided by the linker */
extern const struct kernel_symbol __start___ksymtab[];
extern const struct kernel_symbol __stop___ksymtab[];
extern const struct kernel_symbol __start___ksymtab_gpl[];
extern const struct kernel_symbol __stop___ksymtab_gpl[];
extern const struct kernel_symbol __start___ksymtab_gpl_future[];
extern const struct kernel_symbol __stop___ksymtab_gpl_future[];
```

So after this, the kernel can use `__start___ksymtab` and other variables without any errorsNow let’s talk more about the ELF file about section `__ksymtab`. Firstly dump this section:

```C
# readelf --hex-dump=_ksymtab hello.ko
readelf: Warning: Section '_ksymtab' was not dumped because it does not exist!
# readelf --hex-dump=__ksymtab hello.ko

Hex dump of section '__ksymtab':
NOTE: This section has relocations against it, but these have NOT been applied to this dump.
  0x00000000 00000000 00000000 00000000 00000000 ................
```

Interesting, they are all zeros! Where is our data. If you look the section headers more carefully, you can see some sections begin with `.rela`. There is a `.rela__ksymtab` section:

# readelf -S hello.ko
There are 33 section headers, starting at offset `0x45ab8`:

```C
Section Headers:
  [Nr] Name              Type             Address           Offset
      Size              EntSize          Flags  Link  Info  Align
  [ 0]                   NULL             0000000000000000  00000000
      0000000000000000  0000000000000000           0     0     0
  [ 1] .note.gnu.build-i NOTE             0000000000000000  00000040
      0000000000000024  0000000000000000   A       0     0     4
  [ 2] .text             PROGBITS         0000000000000000  00000070
      0000000000000051  0000000000000000  AX       0     0     16
  [ 3] .rela.text        RELA             0000000000000000  00025be8
      00000000000000d8  0000000000000018   I      30     2     8
  [ 4] __ksymtab         PROGBITS         0000000000000000  000000d0
      0000000000000010  0000000000000000   A       0     0     16
  [ 5] .rela__ksymtab    RELA             0000000000000000  00025cc0
      0000000000000030  0000000000000018   I      30     4     8
  [ 6] __kcrctab         PROGBITS         0000000000000000  000000e0
      0000000000000008  0000000000000000   A       0     0     8
  [ 7] .rela__kcrctab    RELA             0000000000000000  00025cf0
```

`.rela__ksymtab` section’s type is `RELA`. This means this section contains relocation data which data will be and how to be modified when the final executable is loaded to kernel. section of `.rela__ksymtab` contains the `__ksymtab` relocation data.

```C
# readelf  -r hello.ko | head -20

Relocation section '.rela.text' at offset 0x25be8 contains 9 entries:
  Offset          Info           Type           Sym. Value    Sym. Name + Addend
000000000001  001f00000002 R_X86_64_PC32     0000000000000000 __fentry__ - 4
000000000008  00050000000b R_X86_64_32S      0000000000000000 .rodata.str1.1 + 0
00000000000d  002400000002 R_X86_64_PC32     0000000000000000 printk - 4
000000000021  001f00000002 R_X86_64_PC32     0000000000000000 __fentry__ - 4
000000000028  00050000000b R_X86_64_32S      0000000000000000 .rodata.str1.1 + f
00000000002d  002400000002 R_X86_64_PC32     0000000000000000 printk - 4
000000000041  001f00000002 R_X86_64_PC32     0000000000000000 __fentry__ - 4
000000000048  00050000000b R_X86_64_32S      0000000000000000 .rodata.str1.1 + 1f
00000000004d  002400000002 R_X86_64_PC32     0000000000000000 printk - 4

Relocation section '.rela__ksymtab' at offset 0x25cc0 contains 2 entries:
  Offset          Info           Type           Sym. Value    Sym. Name + Addend
000000000000  002300000001 R_X86_64_64       0000000000000000 testexport + 0
000000000008  000600000001 R_X86_64_64       0000000000000000 __ksymtab_strings + 0

Relocation section '.rela__kcrctab' at offset 0x25cf0 contains 1 entries:
  Offset          Info           Type           Sym. Value    Sym. Name + Addend
```

Here we can see in section `.rela__ksymtab` there is 2 entries. I will not dig into the RELA section format, just notice the `0x23` and `0x06` is used to index the `.symtab` section. So when the `.ko` is loaded into the kernel, the first 8 bytes of section `__ksymtab` will be replaced by the actual address of testexport, and the second 8 bytes of section `__ksymtab` will be replaced by the actual address of the string at `__ksymtab_strings+0` which is `testexport`. So this is what the structure `kernel_symbol`—through `EXPORT_SYMBOL`—does.

### Module load process
`init_module` system call is used to load the kernel module to kernel. User space application loads the `.ko` file into user space and then pass the address and size of `.ko` and the arguments of the kernel module will use to this system call. In `init_module`, it just allocates the memory space and copys the user’s data to kernel, then call the actual work function `load_module`. In general we can split the `load_module` function up to two logical part. The first part completes the load work such as reallocation the memory to hold kernel module, resolve the symbol, apply relocations and so on. The second part later do other work such as call the module’s init function, cleanup the allocated resource and so on. Before we go to the first part, let’s first look at a very important structure `struct module`: `<include/linux/module.h>`

```C
struct module {
enum module_state state;

/* Member of list of modules */
struct list_head list;

/* Unique handle for this module */
char name[MODULE_NAME_LEN];

/* Sysfs stuff. */
struct module_kobject mkobj;
struct module_attribute *modinfo_attrs;
const char *version;
const char *srcversion;
struct kobject *holders_dir;

/* Exported symbols */
const struct kernel_symbol *syms;
const unsigned long *crcs;
unsigned int num_syms;

/* Kernel parameters. */
struct kernel_param *kp;
unsigned int num_kp;
...
}
```

Here I just list some of the fields of `struct module`, it represents a module in kernel, contains the infomation of the kernel module. For example, `state` indicates the status of the module, it will change with the load process, the ‘list’ links all of the modules in kernel and ‘name’ contains the module name. Below lists some important function the load_module calls.

```C
load_module
  -->layout_and_allocate
    -->setup_load_info
      -->rewrite_section_headers
    -->layout_sections
    -->layout_symtab
    -->move_module
  -->find_module_sections
  -->simplify_symbols
  -->apply_relocations
  -->parse_args
  -->do_init_module
```

The rewrite_section_headers function replace the sections header field `sh_addr` with the real address in the memory. Then in function setup_load_info, `mod` is initialized with the `.gnu.linkonce.this_module` section’s real address. Actually, this contains the data compiler setup for us. In the source directory, we can see a `hello.mod.c` file:

```C
__visible struct module __this_module
__attribute__((section(".gnu.linkonce.this_module"))) = {
    .name = KBUILD_MODNAME,
    .init = init_module,
#ifdef CONFIG_MODULE_UNLOAD
    .exit = cleanup_module,
#endif
    .arch = MODULE_ARCH_INIT,
};
```

So here we can see the `mod` will have some field. The interesting here is that we can see the init function is `init_module`, not the same as our `hello_init`. The magic is caused by module_init as follows(`include/linux/init.h`):

```C
/* Each module must use one module_init(). */
#define module_init(initfn)     \
static inline initcall_t __inittest(void)  \
{ return initfn; }     \
int init_module(void) __attribute__((alias(#initfn)));
```

From here we can see the compiler will set the ‘init_module’s alias to our init function name which is `hello_init` in our example. Next in the function `layout_sections`, it will caculate the `core` size and `init` size of the ELF file. Then according where define the `CONFIG_KALLSYMS`, `layout_symtab` will be called and the symbol info will be added to the core section. After caculate the core and init section, it will allocate space for core and init section in function `move_module` and then copy the origin section data to the new space. So the sections’s `sh_addr` should also be updated. Then the `mod`s address should be updated.

```C
mod = (void *)info->sechdrs[info->index.mod].sh_addr;

                    core section
                    +------------+ <-----mod->module_core
                +-> |            |
                |   +------------+
+------------+ +---> |            |
| ELF header | | |   +------------+
+------------+ | |   |            |
| section 0  +---+   +------------+
+------------+ |
| section 1  +----+
+------------+ |  |  init section
| section 2  +----+  +------------+ <-----mod->module_init
+------------+ | +-> |            |
| section 3  +-+ |   +------------+
+------------+   +-> |           ||
|sec head table      +------------+
+------------+       |            |
                    |            |
                    +------------+
```

So for now , we have this section.

Later `load_module` call `find_module_sections` to get the export symbol. Next, it calls `simplify_symbols` to fix up the symbols. The function call chain is `simplify_symbols–>resolve_symbol_wait–> –>resolve_symbol–>find_symbol–>each_symbol_section` In the last function, it will first iterate the kernel’s export symbol and then iterate the loaded modules symbol. If `resolve_symbol` successful, it will call `ref_module` to establish the dependency between current load module and the module of the symbol it uses. This is done in `add_module_usage`


```C
static int add_module_usage(struct module *a, struct module *b)
{
struct module_use *use;

pr_debug("Allocating new usage for %s.\n", a->name);
use = kmalloc(sizeof(*use), GFP_ATOMIC);
if (!use) {
  pr_warn("%s: out of memory loading\n", a->name);
  return -ENOMEM;
}

use->source = a;
use->target = b;
list_add(&use->source_list, &b->source_list);
list_add(&use->target_list, &a->target_list);
return 0;
}
```

Here a is current loading module, and b is the module a uses its symbol. module->source_list links the modules depend on module, and module->target_list links the modules it depends on.

After fix up the symbols, the `load_module` function will do relocation by calling function `apply_relocations`. If the section’s type is `SHT_REL` or `SHT_RELA`, function `apply_relocations` will call the arch-spec function. As the symbol table has been solved, this relocation is much simple. So now the module’s export symbol address has been corrected the right value.

Next the `load_module` function will call `parse_args` to parse module parameters. Let’s first look at how to define parameter in kernel module.

```C
static bool __read_mostly fasteoi = 1;
module_param(fasteoi, bool, S_IRUGO);

#define module_param(name, type, perm)    \
module_param_named(name, name, type, perm)

#define module_param_named(name, value, type, perm)      \
param_check_##type(name, &(value));       \
module_param_cb(name, &param_ops_##type, &value, perm);     \
__MODULE_PARM_TYPE(name, #type)

#define module_param_cb(name, ops, arg, perm)          \
__module_param_call(MODULE_PARAM_PREFIX, name, ops, arg, perm, -1, 0)

#define __module_param_call(prefix, name, ops, arg, perm, level, flags) \
/* Default value instead of permissions? */   \
static const char __param_str_##name[] = prefix #name; \
static struct kernel_param __moduleparam_const __param_##name \
__used        \
    __attribute__ ((unused,__section__ ("__param"),aligned(sizeof(void *)))) \
= { __param_str_##name, ops, VERIFY_OCTAL_PERMISSIONS(perm), \
    level, flags, { arg } }

Let’s try an example using the ‘fasteoi’.

param_check_bool(fasteoi, &(fasteoi));
static const char __param_str_bool[] = "fasteoi";
static struct kernel_param __moduleparam_const __param_fasteoi \
__used
    __attribute__ ((unused,__section__ ("__param"),aligned(sizeof(void *)))) \
  = { __param_str_fasteoi, param_ops_bool, VERIFY_OCTAL_PERMISSIONS(perm), \
    -1, 0, { &fasteoi} }
```

So here we can see `module_param(fasteoi, bool, S_IRUGO);` define a variable which is `struct kernel_param` and store it in section `__param`.

```C
struct kernel_param {
const char *name;
const struct kernel_param_ops *ops;
u16 perm;
s8 level;
u8 flags;
union {
  void *arg;
  const struct kparam_string *str;
  const struct kparam_array *arr;
};
};
```

the union `arg` will contain the kernel parameter’s address.

The user space will pass the specific arguments to load_module in the `uargs` argument. In `parse_args`, it will pass one by one parameter, and compare it will the data in section `__param`, and then write it will the user specific value.

```C
int param_set_bool(const char *val, const struct kernel_param *kp)
{
/* No equals means "set"... */
if (!val) val = "1";

/* One of =[yYnN01] */
return strtobool(val, kp->arg);
}
int strtobool(const char *s, bool *res)
{
switch (s[0]) {
case 'y':
case 'Y':
case '1':
  *res = true;
  break;
case 'n':
case 'N':
case '0':
  *res = false;
  break;
default:
  return -EINVAL;
}
return 0;
}
```

### Version control
One thing we have lost is version control. Version control is used to keep consistency between kernel and module. We can’t load modules compiled for 2.6 kernel into 3.2 kernel. That’s why version control needed. Kernel and module uses CRC checksum to do this. The idea behind this is so easy, the build tools will generate CRC checksum for every exported function and for every function module reference. Then in `load_module` function, these two CRC will be checked if there are the same. In order to support this mechism, the kernel config must contain `CONFIG_MODVERSIONS`. In `EXPORT_SYMBOL` macro, there is a `__CRC_SYMBOL` definition.

```C
#ifdef CONFIG_MODVERSIONS
/* Mark the CRC weak since genksyms apparently decides not to
* generate a checksums for some symbols */
#define __CRC_SYMBOL(sym, sec)     \
extern __visible void *__crc_##sym __attribute__((weak));  \
static const unsigned long __kcrctab_##sym  \
__used       \
__attribute__((section("___kcrctab" sec "+" #sym), unused)) \
= (unsigned long) &__crc_##sym;
#else
#define __CRC_SYMBOL(sym, sec)
#endif
```

Expand it.

```C
extern __visible void *__crc_textexport;
static const unsigned long __kcrctab_testexport = (unsigned long) &__crc_textexport;

So for every export symbol, build tools will generate a CRC checksum and store it in section ‘_kcrctab’.

The time for module load process. In hello.mod.c we can see the below:

static const struct modversion_info ____versions[]
__used
__attribute__((section("__versions"))) = {
    { 0x21fac097, __VMLINUX_SYMBOL_STR(module_layout) },
    { 0x27e1a049, __VMLINUX_SYMBOL_STR(printk) },
    { 0xbdfb6dbb, __VMLINUX_SYMBOL_STR(__fentry__) },
};

struct modversion_info {
unsigned long crc;
char name[MODULE_NAME_LEN];
};
```

The `ELF` will have an array of struct modversion stored in section `__versions`, and every element in this array have a crc and name to indicate the module references symbol.

In `check_version`, when it finds the symbole it will call `check_version`. Function `check_version` iterates the `__versions` and compare the finded symble’s CRC checksum. If it is the same, it passes the check.

### Modinfo
`.ko` file will also contain a `.modinfo` section which stores some of the module information. modinfo program can show these info. In the source code, one can use `MODULE_INFO` to add this information.

```C
#define MODULE_INFO(tag, info) __MODULE_INFO(tag, tag, info)

#ifdef MODULE
#define __MODULE_INFO(tag, name, info)       \
static const char __UNIQUE_ID(name)[]       \
  __used __attribute__((section(".modinfo"), unused, aligned(1)))   \
  = __stringify(tag) "=" info
#else  /* !MODULE */
/* This struct is here for syntactic coherency, it is not used */
#define __MODULE_INFO(tag, name, info)       \
  struct __UNIQUE_ID(name) {}
#endif
```

`MODULE_INFO` just define a key-value data in `.modinfo` section once the `MODULE` is defined. `MODULE_INFO` is used several places, such as license, vermagic:

```C
#define MODULE_LICENSE(_license) MODULE_INFO(license, _license)

/*
* Author(s), use "Name <email>" or just "Name", for multiple
* authors use multiple MODULE_AUTHOR() statements/lines.
*/
#define MODULE_AUTHOR(_author) MODULE_INFO(author, _author)

/* What your module does. */
#define MODULE_DESCRIPTION(_description) MODULE_INFO(description, _description)

MODULE_INFO(vermagic, VERMAGIC_STRING);
```

### Vermagic
vermagic is a string generated by kernel configuration information. ‘load_module’ will check this in ‘layout_and_allocate’->’check_modinfo’->’same_magic’. ‘VERMAGIC_STRING’ is generated by the kernel configuration.

```
#define VERMAGIC_STRING        \
UTS_RELEASE " "       \
MODULE_VERMAGIC_SMP MODULE_VERMAGIC_PREEMPT     \
MODULE_VERMAGIC_MODULE_UNLOAD MODULE_VERMAGIC_MODVERSIONS \
MODULE_ARCH_VERMAGI
```

After doing the tough work, `load_module` goes to the final work to call `do_init_module`. If the module has an init function, `do_init_module` will call it in function `do_one_initcall`. Then change the module’s state to `MODULE_STATE_LIVE`, and call the function registered in `module_notify_list` list and finally free the INIT section of module.

### Unload module
Unload module is quite easy, it is done by syscall `delete_module`, which takes only the module name argument. First find the module in modules list and then check whether it is depended by other modules then call module exit function and finally notify the modules who are interested module unload by iterates `module_notify_list`
