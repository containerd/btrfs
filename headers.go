/*
   Copyright The containerd Authors.

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

package btrfs

//go:generate go run ./internal/cmd/hgen.go -u -g -t BTRFS_ -p btrfs -cs=treeKeyType:uint32=_KEY,objectID:uint64=_OBJECTID -cp=fileType=FT_,fileExtentType=FILE_EXTENT_,devReplaceItemState=DEV_REPLACE_ITEM_STATE_,blockGroup:uint64=BLOCK_GROUP_ -o btrfs_tree_hc.go btrfs_tree.h
//go:generate ltag -t ./license-templates btrfs_tree_hc.go
//go:generate sed -i -e "s/SPDX-License-Identifier: GPL-2.0 WITH Linux-syscall-note/SPDX-License-Identifier: GPL-2.0 WITH Linux-syscall-note (see headers.go)/" btrfs_tree_hc.go
//go:generate gofmt -l -w btrfs_tree_hc.go

/*
btrfs_tree.h can be found at https://github.com/torvalds/linux/blob/v5.13/include/uapi/linux/btrfs_tree.h
btrfs_tree.h is licensed under the terms of "GPL-2.0 WITH Linux-syscall-note": https://github.com/torvalds/linux/blob/v5.13/LICENSES/exceptions/Linux-syscall-note

containerd/btrfs shall be considered as "user programs that use kernel services by normal system calls" mentioned in the note above,
and "does *not* fall under the heading of \"derived work\"" of the GPL-2.0 code.
*/
