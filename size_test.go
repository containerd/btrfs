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

import (
	"reflect"
	"testing"
	"unsafe"
)

var caseSizes = []struct {
	obj  interface{}
	size int
}{
	{obj: btrfs_ioctl_vol_args{}, size: 4096},
	{obj: btrfs_qgroup_limit{}, size: 40},
	{obj: btrfs_qgroup_inherit{}, size: 72},
	{obj: btrfs_ioctl_qgroup_limit_args{}, size: 48},
	{obj: btrfs_ioctl_vol_args_v2{}, size: 4096},
	{obj: btrfs_scrub_progress{}, size: 120},
	{obj: btrfs_ioctl_scrub_args{}, size: 1024},
	{obj: btrfs_ioctl_dev_replace_start_params{}, size: 2072},
	{obj: btrfs_ioctl_dev_replace_status_params{}, size: 48},
	{obj: btrfs_ioctl_dev_replace_args_u1{}, size: 2600},
	{obj: btrfs_ioctl_dev_replace_args_u2{}, size: 2600},
	{obj: btrfs_ioctl_dev_info_args{}, size: 4096},
	{obj: btrfs_ioctl_fs_info_args{}, size: 1024},
	{obj: btrfs_ioctl_feature_flags{}, size: 24},
	{obj: btrfs_balance_args{}, size: 136},
	{obj: BalanceProgress{}, size: 24},
	{obj: btrfs_ioctl_balance_args{}, size: 1024},
	{obj: btrfs_ioctl_ino_lookup_args{}, size: 4096},
	{obj: btrfs_ioctl_search_key{}, size: 104},
	{obj: btrfs_ioctl_search_header{}, size: 32},
	{obj: btrfs_ioctl_search_args{}, size: 4096},
	{obj: btrfs_ioctl_search_args_v2{}, size: 112},
	{obj: btrfs_ioctl_clone_range_args{}, size: 32},
	{obj: btrfs_ioctl_same_extent_info{}, size: 32},
	{obj: btrfs_ioctl_same_args{}, size: 24},
	{obj: btrfs_ioctl_defrag_range_args{}, size: 48},
	{obj: btrfs_ioctl_space_info{}, size: 24},
	{obj: btrfs_ioctl_space_args{}, size: 16},
	{obj: btrfs_data_container{}, size: 16},
	{obj: btrfs_ioctl_ino_path_args{}, size: 56},
	{obj: btrfs_ioctl_logical_ino_args{}, size: 56},
	{obj: btrfs_ioctl_get_dev_stats{}, size: 1032},
	{obj: btrfs_ioctl_quota_ctl_args{}, size: 16},
	{obj: btrfs_ioctl_qgroup_assign_args{}, size: 24},
	{obj: btrfs_ioctl_qgroup_create_args{}, size: 16},
	{obj: btrfs_ioctl_timespec{}, size: 16},
	{obj: btrfs_ioctl_received_subvol_args{}, size: 200},
	{obj: btrfs_ioctl_send_args{}, size: 72},

	//{obj:btrfs_timespec{},size:12},
	//{obj:btrfs_root_ref{},size:18},
	//{obj:btrfs_root_item{},size:439},
	{obj: btrfs_root_item_raw{}, size: 439},
	{obj: btrfs_root_item_raw_p1{}, size: 439 - 23 - int(unsafe.Sizeof(btrfs_root_item_raw_p3{}))},
	{obj: btrfs_root_item_raw_p3{}, size: 439 - 23 - int(unsafe.Sizeof(btrfs_root_item_raw_p1{}))},
	//{obj:btrfs_inode_item{},size:160},
	{obj: btrfs_inode_item_raw{}, size: 160},
	{obj: timeBlock{}, size: 4 * 12},
}

func TestSizes(t *testing.T) {
	for _, c := range caseSizes {
		if sz := int(reflect.ValueOf(c.obj).Type().Size()); sz != c.size {
			t.Errorf("unexpected size of %T: %d (exp: %d)", c.obj, sz, c.size)
		}
	}
}
