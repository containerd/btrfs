/*
  Copyright The containerd Authors

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

#include <linux/version.h>
#if LINUX_VERSION_CODE < KERNEL_VERSION(4,12,0)
#warning "Headers from kernel >= 4.12 are required on compilation time (not on run time)"
#endif
#include <linux/btrfs.h>
#include <linux/btrfs_tree.h>

// unfortunately, we need to define "alignment safe" C structs to populate for
// packed structs that aren't handled by cgo. Fields will be added here, as
// needed.

struct gosafe_btrfs_root_item {
	__u8 uuid[BTRFS_UUID_SIZE];
	__u8 parent_uuid[BTRFS_UUID_SIZE];
	__u8 received_uuid[BTRFS_UUID_SIZE];

	__le64 generation;
	__le64 otransid;
	__le64 flags;
};

void unpack_root_item(struct gosafe_btrfs_root_item* dst, struct btrfs_root_item* src);
/* void unpack_root_ref(struct gosafe_btrfs_root_ref* dst, struct btrfs_root_ref* src); */
