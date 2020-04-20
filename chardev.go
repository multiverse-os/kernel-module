package ko

//}struct file_operations {
//	struct module *owner;
//	 loff_t(*llseek) (struct file *, loff_t, int);
//	 ssize_t(*read) (struct file *, char __user *, size_t, loff_t *);
//	 ssize_t(*aio_read) (struct kiocb *, char __user *, size_t, loff_t);
//	 ssize_t(*write) (struct file *, const char __user *, size_t, loff_t *);
//	 ssize_t(*aio_write) (struct kiocb *, const char __user *, size_t,
//			      loff_t);
//	int (*readdir) (struct file *, void *, filldir_t);
//	unsigned int (*poll) (struct file *, struct poll_table_struct *);
//	int (*ioctl) (struct inode *, struct file *, unsigned int,
//		      unsigned long);
//	int (*mmap) (struct file *, struct vm_area_struct *);
//	int (*open) (struct inode *, struct file *);
//	int (*flush) (struct file *);
//	int (*release) (struct inode *, struct file *);
//	int (*fsync) (struct file *, struct dentry *, int datasync);
//	int (*aio_fsync) (struct kiocb *, int datasync);
//	int (*fasync) (int, struct file *, int);
//	int (*lock) (struct file *, int, struct file_lock *);
//	 ssize_t(*readv) (struct file *, const struct iovec *, unsigned long,
//			  loff_t *);
//	 ssize_t(*writev) (struct file *, const struct iovec *, unsigned long,
//			   loff_t *);
//	 ssize_t(*sendfile) (struct file *, loff_t *, size_t, read_actor_t,
//			     void __user *);
//	 ssize_t(*sendpage) (struct file *, struct page *, int, size_t,
//			     loff_t *, int);
//	unsigned long (*get_unmapped_area) (struct file *, unsigned long,
//					    unsigned long, unsigned long,
//					    unsigned long);
//};
//
//
//Some operations are not implemented by a driver. For example, a driver that handles a video card won't need to read from a directory structure. The corresponding entries in the file_operations structure should be set to NULL.
//
//There is a gcc extension that makes assigning to this structure more convenient. You'll see it in modern drivers, and may catch you by surprise. This is what the new way of assigning to the structure looks like:
//
//struct file_operations fops = {
//	read: device_read,
//	write: device_write,
//	open: device_open,
//	release: device_release
//};
//
//
//However, there's also a C99 way of assigning to elements of a structure, and this is definitely preferred over using the GNU extension. The version of gcc I'm currently using, 2.95, supports the new C99 syntax. You should use this syntax in case someone wants to port your driver. It will help with compatibility:
//
//struct file_operations fops = {
//	.read = device_read,
//	.write = device_write,
//	.open = device_open,
//	.release = device_release
//};

//int register_chrdev(unsigned int major, const char *name, struct
//file_operations *fops);
